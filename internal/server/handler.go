package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/pgvillage-tools/pgroute66/internal/config"
	"github.com/pgvillage-tools/pgroute66/internal/logging"
	"github.com/pgvillage-tools/pgroute66/pkg/pg"
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
	groupConns GroupConnections
	config     config.Config
}

/*
With Gin, there is no winning with gochecknoglobals.
This seems like a proper way to go.
Also see https://github.com/gothinkster/golang-gin-realworld-example-app/issues/15 for background
*/
//nolint
var globalHandler *PgRouteHandler

// Initialize this module
func Initialize() {
	globalHandler = NewPgRouteHandler()
}

// NewPgRouteHandler returns a PgRouteHandler
func NewPgRouteHandler() *PgRouteHandler {
	var err error

	_, logger := logging.GetLogComponent(context.Background(), logging.ServerComponent)
	prh := PgRouteHandler{
		groupConns: GroupConnections{},
	}

	prh.config, err = config.NewConfig()
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("Cannot parse config")
	}

	for groupName, hosts := range prh.config.GetHostGroups() {
		conns := Connections{}
		for hostName, dsn := range hosts {
			if b64password, exists := dsn["b64password"]; exists {
				sDec, err := base64.StdEncoding.DecodeString(b64password)
				if err != nil {
					logger.Panic().AnErr("error", err).Str("b64password",
						b64password).Msg("Failed to decode b64password")
				}
				dsn["password"] = string(sDec)
				delete(dsn, "b64password")
			}
			conns[hostName] = pg.NewConn(dsn)
		}
		prh.groupConns[groupName] = conns
	}
	return &prh
}

// GetStandbys connects all PostgreSQL servers and returns a list of all that are standby
func (prh PgRouteHandler) GetStandbys(ctx context.Context, group string) (standbys []string) {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	groupConnections, ok := prh.groupConns[group]
	if !ok {
		return nil
	}
	for name, conn := range groupConnections {
		isStandby, err := conn.IsStandby(ctx)
		if err != nil {
			logger.Debug().
				Str("group", group).
				Str("standby", name).
				AnErr("error", err).
				Msg("Could not get state of standby")
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
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	groupConnections, ok := prh.groupConns[group]
	if !ok {
		logger.Fatal().Str("group", group).Msg("not defined in config")
	}
	for name, conn := range groupConnections {
		isPrimary, err := conn.IsPrimary(ctx)
		if err != nil {
			logger.Debug().
				Str("group", group).
				Str("primary", name).
				AnErr("error", err).
				Msg("Could not get state of primary")
		}

		if isPrimary {
			primaries = append(primaries, name)
		}
	}

	sort.Strings(primaries)

	return primaries
}

// GetNodeStatus returns a status for a node
func (prh PgRouteHandler) GetNodeStatus(ctx context.Context, group string, name string) string {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	nodes, exists := prh.groupConns[group]
	if !exists {
		logger.Fatal().Str("group", group).Msg("not defined in config")
	}
	if node, exists := nodes[name]; exists {
		isPrimary, err := node.IsPrimary(ctx)
		if err != nil {
			logger.Debug().Str("node", name).Msg("Could not get state of node")
			return ghStatusUnavailable
		} else if isPrimary {
			return ghStatusPrimary
		}
		return ghStatusStandby
	}
	return ghStatusInvalid
}

// UpdateNodeAvailability on the primary
func (prh PgRouteHandler) UpdateNodeAvailability(ctx context.Context, group string) {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	nodes, exists := prh.groupConns[group]
	if !exists {
		logger.Fatal().Str("group", group).Msg("not defined in config")
	}
	for nodeName, conn := range nodes {
		if isPrimary, err := conn.IsPrimary(ctx); err != nil {
			logger.Debug().Str("node", nodeName).AnErr("error", err).Msg("failed to check if node %s is primary")
		} else if !isPrimary {
			continue
		} else if err = conn.AvUpdateDuration(ctx); err != nil {
			logger.Debug().Str("node", nodeName).AnErr("error", err).Msg("failed to update availability info on node")
			return
		} else {
			logger.Info().Str("node", nodeName).Msg("updating availability info on node")
			return
		}
	}
}

// CreateAvailabilityTable creates the AVC table
func (prh PgRouteHandler) CreateAvailabilityTable(
	ctx context.Context,
	group string,
) {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	nodes, exists := prh.groupConns[group]
	if !exists {
		logger.Fatal().Str("group", group).Msg("not defined in config")
	}
	for nodeName, conn := range nodes {
		if isPrimary, err := conn.IsPrimary(ctx); err != nil {
			logger.Debug().Str("node", nodeName).AnErr("error", err).Msg("failed to check if node %s is primary")
		} else if !isPrimary {
			continue
		} else if err = conn.AvcCreateTable(ctx); err != nil {
			logger.Debug().Str("node", nodeName).AnErr("error", err).Msg("failed to create availability table")
			return
		} else {
			logger.Info().Str("node", nodeName).Msg("creating availability table")
			return
		}
	}
}

// GetNodeAvailability returns the state of one node
func (prh PgRouteHandler) GetNodeAvailability(ctx context.Context, group string, name string, limit float64) string {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	nodes, exists := prh.groupConns[group]
	if !exists {
		logger.Fatal().Str("group", group).Msg("not defined in config")
	}
	prh.CreateAvailabilityTable(ctx, group)
	defer prh.UpdateNodeAvailability(ctx, group)

	if node, exists := nodes[name]; exists {
		err := node.AvCheckDuration(ctx, limit)
		if err == nil {
			logger.Info().Str("node", name).Msg("availability of node %s is within limits")
			return ghStatusOk
		} else if aErr, ok := err.(pg.AvcDurationExceededError); ok {
			logger.Info().Str("node", name).AnErr("error", err).Msg("Availability limit exceeded")
			return fmt.Sprintf("exceeded (%s)", aErr.String())
		}
		logger.Info().Str("node", name).AnErr("error", err).
			Msg("unexpected error occurred while retrieving availability")
		return err.Error()
	}

	return ghStatusInvalid
}
