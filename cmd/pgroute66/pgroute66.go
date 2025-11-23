// Package main is the main entrypoint for pgroute66
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mannemsolutions/pgroute66/internal"
	"github.com/rs/zerolog/log"
)

const (
	envConfName     = "PGROUTE66CONFIG"
	defaultConfFile = "/etc/pgroute66/config.yaml"
)

func main() {
	var (
		debug      bool
		version    bool
		configFile string
	)

	flag.BoolVar(&debug, "d", false, "Add debugging output")
	flag.BoolVar(&version, "v", false, "Show version information")

	flag.StringVar(&configFile, "c", os.Getenv(envConfName), "Path to configfile")

	flag.Parse()

	if version {
		fmt.Println(internal.AppVersion)
		os.Exit(0)
	}

	if configFile == "" {
		configFile = defaultConfFile
	}
	config, err := internal.NewConfigFromFile(configFile, debug)
	if err != nil {
		logger := log.With().Logger()
		logger.Fatal().Msgf("failed to read config")
	}

	handler := internal.NewPgRouteHandler(config)
	handler.RunAPI()
}
