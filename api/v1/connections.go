package v1

import (
	"context"

	"github.com/pgvillage-tools/pgroute66/internal/logging"
	"github.com/pgvillage-tools/pgroute66/pkg/pg"
)

// Connections is a map of connections per route
type Connections map[string]*pg.Conn

// FilteredConnections return a list of connections that conform to a filter
func (rcs Connections) FilteredConnections(ctx context.Context, filter []string) Connections {
	ctx, logger := logging.GetLogComponent(ctx, logging.ServerComponent)
	logger.Debug().Any("filter", filter).Msg("filtering")

	fcs := Connections{}

	for _, host := range filter {
		if conn, ok := rcs[host]; ok {
			fcs[host] = conn
		}
	}

	return fcs
}
