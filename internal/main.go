package internal

import "github.com/rs/zerolog/log"

const (
	defaultSSLPort   = 8443
	defaultNoSSLPort = 8080
	bitSize32        = 32
	bitSize64        = 64
)

var logger = log.With().Logger()
