package server

import (
	"github.com/pgvillage-tools/pgroute66/pkg/pg"
)

// GroupConnections is a map of Connections grouped per cluster
type GroupConnections map[string]Connections

// Connections is a map of connections per route
type Connections map[string]*pg.Conn
