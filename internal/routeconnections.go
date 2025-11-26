package internal

import (
	"github.com/mannemsolutions/pgroute66/pkg/pg"
)

// RouteConnections is a map of connections per route
type RouteConnections map[string]*pg.Conn

// FilteredConnections return a list of connections that conform to a filter
func (rcs RouteConnections) FilteredConnections(filter []string) RouteConnections {
	logger.Debug().Msgf("filtering on name: %s", filter)

	fcs := RouteConnections{}

	for _, host := range filter {
		if conn, ok := rcs[host]; ok {
			fcs[host] = conn
		}
	}

	return fcs
}
