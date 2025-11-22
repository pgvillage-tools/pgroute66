package pg

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Conn objects can connect to PostgreSQL and verify state
type Conn struct {
	connParams Dsn
	endpoint   string
	conn       *pgxpool.Pool
	logger     *zap.SugaredLogger
}

// NewConn can create a Conn object
func NewConn(connParams Dsn, logger *zap.SugaredLogger) (c *Conn) {
	c = &Conn{
		connParams: connParams,
		logger:     logger,
	}
	c.endpoint = fmt.Sprintf("%s:%s", c.Host(), c.Port())

	return c
}

// DSN returns a string value of the COnnection Parameters
func (c *Conn) DSN() (dsn string) {
	pairs := make([]string, 0, len(c.connParams))
	for key, value := range c.connParams {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, connectStringValue(value)))
	}

	return strings.Join(pairs[:], " ")
}

// Host returns the host parameter from the Connection Parameters
func (c *Conn) Host() string {
	value, ok := c.connParams["host"]
	if ok {
		return value
	}

	value = os.Getenv("PGHOST")
	if value != "" {
		return value
	}

	return "localhost"
}

// Port returns the port parameter from the Connection Parameters
func (c *Conn) Port() string {
	value, ok := c.connParams["port"]
	if ok {
		return value
	}

	value = os.Getenv("PGPORT")
	if value != "" {
		return value
	}

	return "5432"
}

// Connect can be used to actually connect the connection
func (c *Conn) Connect(ctx context.Context) (err error) {
	if c.conn != nil {
		return nil
	}

	c.logger.Debugf("Connecting to %s (%v)", c.endpoint, c.DSN())

	poolConfig, err := pgxpool.ParseConfig(c.DSN())
	if err != nil {
		log.Panicf("Unable to parse DSN (%s): %e", c.DSN(), err)
	}

	c.conn, err = pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		c.conn = nil

		return err
	}

	return nil
}

func (c *Conn) runQueryExec(ctx context.Context, query string, args ...any) (affected int64, err error) {
	c.logger.Debugf("Running query `%s` on %s", query, c.endpoint)

	var ct pgconn.CommandTag

	if err = c.Connect(ctx); err != nil {
		return 0, err
	} else if ct, err = c.conn.Exec(ctx, query, args...); err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (c *Conn) runQueryExists(ctx context.Context, query string, args ...any) (exists bool, err error) {
	c.logger.Debugf("Running query `%s` on %s", query, c.endpoint)

	err = c.Connect(ctx)
	if err != nil {
		return false, err
	}

	var answer string
	err = c.conn.QueryRow(ctx, query, args...).Scan(&answer)

	if err == nil {
		c.logger.Debugf("Query `%s` returns rows for %s", query, c.endpoint)
		return true, nil
	} else if err.Error() == pgx.ErrNoRows.Error() {
		c.logger.Debugf("Query `%s` returns no rows for %s", query, c.endpoint)
		return false, nil
	}
	return false, err
}

// GetRows runs a query and returns the results
func (c *Conn) GetRows(
	ctx context.Context,
	query string,
	args ...any,
) ([]map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	c.logger.Debugf("Running SQL: %s with args %v", query, args)
	result, err := c.conn.Query(ctx, query, args...)

	if err != nil {
		result.Close()

		return nil, err
	}
	defer result.Close()

	hdr := make([]string, len(result.FieldDescriptions()))

	for i, col := range result.FieldDescriptions() {
		hdr[i] = string(col.Name)
	}

	var answer []map[string]any

	for result.Next() {
		row := map[string]any{}

		var values []any

		if values, err = result.Values(); err != nil {
			return nil, err
		}
		for i, value := range values {
			row[hdr[i]] = value
		}

		answer = append(answer, row)
	}

	return answer, nil
}

// IsPrimary returns true when this is a primary
func (c *Conn) IsPrimary(ctx context.Context) (bool, error) {
	return c.runQueryExists(ctx, "select 'primary' where not pg_is_in_recovery()")
}

// IsStandby returns true when this is a standby
func (c *Conn) IsStandby(ctx context.Context) (bool, error) {
	return c.runQueryExists(ctx, "select 'standby' where pg_is_in_recovery()")
}
