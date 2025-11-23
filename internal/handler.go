package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/mannemsolutions/pgroute66/pkg/pg"
)

const (
	ghStatusInvalid     = "invalid"
	ghStatusOk          = "ok"
	ghStatusPrimary     = "primary"
	ghStatusStandby     = "standby"
	ghStatusUnavailable = "unavailable"
)

// PgRouteHandler handles all PostgreSQL connections for a route
type PgRouteHandler struct {
	connections RouteConnections
	config      RouteConfig
}

/*
With Gin, there is no winning with gochecknoglobals.
This seems like a proper way to go.
Also see https://github.com/gothinkster/golang-gin-realworld-example-app/issues/15 for background
*/

// NewPgRouteHandler returns a PgRouteHandler
func NewPgRouteHandler(config RouteConfig) *PgRouteHandler {
	prh := PgRouteHandler{
		connections: map[string]*pg.Conn{},
	}

	prh.config = config

	for name, dsn := range prh.config.Hosts {
		if b64password, exists := dsn["b64password"]; exists {
			sDec, err := base64.StdEncoding.DecodeString(b64password)
			if err != nil {
				logger.Panic().Msgf("Could not decode b64password %s, %s", b64password, err.Error())
			}
			dsn["password"] = string(sDec)
			delete(dsn, "b64password")
		}
		prh.connections[name] = pg.NewConn(dsn)
	}

	return &prh
}

// GetStandbys connects all PostgreSQL servers and returns a list of all that are standby
func (prh PgRouteHandler) GetStandbys(ctx context.Context, group string) (standbys []string) {
	for name, conn := range prh.connections.FilteredConnections(prh.config.GroupHosts(group)) {
		isStandby, err := conn.IsStandby(ctx)
		if err != nil {
			logger.Debug().Msgf("Could not get state of standby %s, %s", name, err.Error())
		}

		if isStandby {
			standbys = append(standbys, name)
		}
	}

	sort.Strings(standbys)

	return standbys
}

// GetPrimaries connects all PostgreSQL servers and returns a list of all that are primary
func (prh PgRouteHandler) GetPrimaries(ctx context.Context, group string) (primaries []string) {
	for name, conn := range prh.connections.FilteredConnections(prh.config.GroupHosts(group)) {
		isPrimary, err := conn.IsPrimary(ctx)
		if err != nil {
			logger.Debug().Msgf("Could not get state of primary %s, %s", name, err.Error())
		}

		if isPrimary {
			primaries = append(primaries, name)
		}
	}

	sort.Strings(primaries)

	return primaries
}

// GetNodeStatus returns a status for a node
func (prh PgRouteHandler) GetNodeStatus(ctx context.Context, name string) string {
	if node, exists := prh.connections[name]; exists {
		isPrimary, err := node.IsPrimary(ctx)
		if err != nil {
			logger.Debug().Msgf("Could not get state of node %s, %s", name, err.Error())

			return ghStatusUnavailable
		} else if isPrimary {
			return ghStatusPrimary
		}
		return ghStatusStandby
	}

	return ghStatusInvalid
}

// UpdateNodeAvailability on the primary
func (prh PgRouteHandler) UpdateNodeAvailability(ctx context.Context) {
	for nodeName, conn := range prh.connections {
		if isPrimary, err := conn.IsPrimary(ctx); err != nil {
			logger.Error().Msgf("failed to check if node %s is primary: %e", nodeName, err)
		} else if !isPrimary {
			continue
		} else if err = conn.AvUpdateDuration(ctx); err != nil {
			logger.Error().Msgf("failed to update availability info on node %s: %e", nodeName, err)
			return
		} else {
			logger.Info().Msgf("updating availability info on node %s", nodeName)

			return
		}
	}
}

// CreateAvailabilityTable creates the AVC table
func (prh PgRouteHandler) CreateAvailabilityTable(ctx context.Context) {
	for nodeName, conn := range prh.connections {
		if isPrimary, err := conn.IsPrimary(ctx); err != nil {
			logger.Error().Msgf("failed to check if node %s is primary: %e", nodeName, err)
		} else if !isPrimary {
			continue
		} else if err = conn.AvcCreateTable(ctx); err != nil {
			logger.Error().Msgf("failed to create availability table on node %s: %e", nodeName, err)
			return
		} else {
			logger.Info().Msgf("creating availability table on node %s", nodeName)
			return
		}
	}
}

// GetNodeAvailability returns the state of one node
func (prh PgRouteHandler) GetNodeAvailability(ctx context.Context, name string, limit float64) string {
	prh.CreateAvailabilityTable(ctx)
	defer prh.UpdateNodeAvailability(ctx)

	if node, exists := prh.connections[name]; exists {
		err := node.AvCheckDuration(ctx, limit)
		if err == nil {
			logger.Error().Msgf("availability of node %s is within limits", name)

			return ghStatusOk
		} else if aErr, ok := err.(pg.AvcDurationExceededError); ok {
			logger.Error().Msgf("Availability limit exceeded for %s: %e", name, aErr)
			return fmt.Sprintf("exceeded (%s)", aErr.String())
		}
		logger.Info().Msgf("unexpeced error occurred while retrieving availability of %s: %e", name, err)
		return err.Error()
	}
	return ghStatusInvalid
}
