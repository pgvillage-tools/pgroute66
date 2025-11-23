// Package pg is the library with all PostgreSQL specific code
package pg

import (
	"context"
	"fmt"
)

const (
	// AvcSchema is the schema to store the Availability Checker table
	AvcSchema = "public"
	// AvcTable is the table to store the Availability Checker record
	AvcTable = "pgr66_avc"
	// AvcColumn is the column to store the Availability Checker data
	AvcColumn = "pgr66_avc"
)

// AvcDurationExceededError is raised when AV check duration is expired
type AvcDurationExceededError struct {
	max      float64
	actually float64
}

func fullTableName() string {
	return fmt.Sprintf("%s.%s", identifierNameSQL(AvcSchema), identifierNameSQL(AvcTable))
}

func (der AvcDurationExceededError) Error() string {
	return fmt.Sprintf("should have taken %f msec, but actually took %f msec", der.max, der.actually)
}

func (der AvcDurationExceededError) String() string {
	return fmt.Sprintf("exceeded %f by %f msec", der.max, der.actually)
}

func (c *Conn) avcTableExists(ctx context.Context) (bool, error) {
	if exists, err := c.runQueryExists(ctx, "select relname from pg_class where relname = $1 and relnamespace in "+
		"(select oid from pg_namespace where nspname=$2)",
		AvcTable, AvcSchema); err != nil {
		return false, fmt.Errorf("failed to check for table %s", fullTableName())
	} else if exists {
		return true, nil
	}
	return false, nil
}

// AvcCreateTable is a query builder for the create statement of the AVC table
func (c *Conn) AvcCreateTable(ctx context.Context) error {
	c.logger.Infof("Creating table")

	if exists, err := c.avcTableExists(ctx); err != nil {
		c.logger.Errorf("failed to check if table %s exists: %e", fullTableName(), err)

		return err
	} else if exists {
		return nil
	}

	if _, err := c.runQueryExec(ctx, fmt.Sprintf("create table %s (%s timestamp)",
		fullTableName(), identifierNameSQL(AvcColumn))); err != nil {
		return fmt.Errorf("failed to create table %s", fullTableName())
	}

	if affected, err := c.runQueryExec(ctx, fmt.Sprintf("insert into %s values(now())", fullTableName())); err != nil {
		return fmt.Errorf("failed to create table %s", fullTableName())
	} else if affected != 1 {
		return fmt.Errorf("unexpected result while inserting into table %s", fullTableName())
	}

	return nil
}

func (c *Conn) avCheckerGetDuration(ctx context.Context) (float64, error) {
	fullColName := identifierNameSQL(AvcColumn)

	if exists, err := c.avcTableExists(ctx); err != nil {
		c.logger.Errorf("failed to check if table %s exists: %e", fullTableName(), err)

		return 0, err
	} else if !exists {
		return -1, nil
	}

	qry := fmt.Sprintf("select extract('epoch' from (now()-%s)) duration from %s", fullColName, fullTableName())
	var mSec float64
	var mSecOk bool
	if result, err := c.GetRows(ctx, qry); err != nil {
		c.logger.Errorf("failed to retrieve duration from postgres: %e", err)
		return 0, err
	} else if len(result) != 1 {
		return 0, fmt.Errorf("unexpected result while checking for duration (%d != 1)", len(result))
	} else if value, valueOk := result[0]["duration"]; !valueOk {
		return 0, fmt.Errorf("unexpected result while checking for duration (%d != 1)", len(result))
	} else if mSec, mSecOk = value.(float64); !mSecOk {
		return 0, fmt.Errorf("unexpected result type checking for duration (%T != float64)", value)
	}
	return mSec, nil
}

// AvUpdateDuration can update the AVC column
func (c *Conn) AvUpdateDuration(ctx context.Context) error {
	var affected int64

	if isPrimary, err := c.IsPrimary(ctx); err != nil {
		return err
	} else if !isPrimary {
		c.logger.Infof("skipping update of %s on a standby database server", fullTableName())

		return nil
	} else if err = c.AvcCreateTable(ctx); err != nil {
		return err
	} else if affected, err = c.runQueryExec(ctx, fmt.Sprintf("update %s set %s = now()",
		fullTableName(), identifierNameSQL(AvcColumn))); err != nil {
		return err
	} else if affected != 1 {
		return fmt.Errorf("unexpecetedly updated %d rows instead of 1 for %s", affected, fullTableName())
	}

	return nil
}

// AvCheckDuration will query PostgreSQL to check availability
func (c *Conn) AvCheckDuration(ctx context.Context, maxDuration float64) error {
	var (
		err   error
		since float64
	)

	if since, err = c.avCheckerGetDuration(ctx); err != nil {
		return err
	} else if since < 0 {
		return fmt.Errorf("table %s does not exist", fullTableName())
	} else if since > maxDuration {
		return AvcDurationExceededError{
			max:      maxDuration,
			actually: since,
		}
	}

	return nil
}
