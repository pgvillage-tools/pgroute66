package pg

import "github.com/rs/zerolog/log"

// Dsn is a string map and can hold connection parameters
type Dsn map[string]string

var logger = log.With().Logger()
