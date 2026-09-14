// Package internal holds all unexported code
package server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pgvillage-tools/pgroute66/internal/logging"
)

// RunAPI will run the gin webserver
func RunAPI() {
	var err error

	var cert tls.Certificate

	_, logger := logging.GetLogComponent(context.Background(), logging.ServerComponent)
	Initialize()

	if globalHandler.config.LogFile != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET("/v1/primary", getPrimary)
	router.GET("/v1/primaries", getPrimaries)
	router.GET("/v1/standbys", getStandbys)
	router.GET("/v1/:id/status", getStatus)
	router.GET("/v1/:id/availability", getAvailability)

	logger.Debug().Str("binding to", globalHandler.config.BindTo()).Msg("")

	if globalHandler.config.Ssl.Enabled() {
		logger.Debug().Msg("Running with SSL")

		var keyBytes []byte
		keyBytes, err = globalHandler.config.Ssl.KeyBytes()
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("Error parsing key bytes")
		}
		cert, err = tls.X509KeyPair(globalHandler.config.Ssl.MustCertBytes(), keyBytes)
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("Error parsing cert and key")
		}

		tlsConfig := tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		}
		server := http.Server{Addr: globalHandler.config.BindTo(), Handler: router, TLSConfig: &tlsConfig}
		err = server.ListenAndServeTLS("", "")
	} else {
		logger.Debug().Msg("Running without SSL")
		err = router.Run(globalHandler.config.BindTo())
	}

	if err != nil {
		log.Panicf("Error running API: %s", err.Error())
	}
}

func getPrimary(c *gin.Context) {
	primary := globalHandler.GetPrimaries(c.Request.Context(), c.DefaultQuery("group", "all"))
	switch len(primary) {
	case 0:
		c.IndentedJSON(http.StatusNotFound, "")
	case 1:
		c.IndentedJSON(http.StatusOK, primary[0])
	default:
		c.IndentedJSON(http.StatusConflict, "")
	}
}

// getPrimaries responds with the list of all albums as JSON.
func getPrimaries(c *gin.Context) {
	primaries := globalHandler.GetPrimaries(c.Request.Context(), c.DefaultQuery("group", "all"))
	c.IndentedJSON(http.StatusOK, primaries)
}

// getStandbys responds with the list of all albums as JSON.
func getStandbys(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, globalHandler.GetStandbys(c.Request.Context(), c.DefaultQuery("group", "all")))
}

func getStatus(c *gin.Context) {
	id := c.Param("id")

	status := globalHandler.GetNodeStatus(c.Request.Context(), c.DefaultQuery("group", "all"), id)
	switch status {
	case ghStatusPrimary, ghStatusStandby:
		c.IndentedJSON(http.StatusOK, status)
	case ghStatusInvalid:
		c.IndentedJSON(http.StatusNotFound, status)
	case ghStatusUnavailable:
		c.IndentedJSON(http.StatusUnprocessableEntity, status)
	}
}

func getAvailability(c *gin.Context) {
	_, logger := logging.GetLogComponent(context.Background(), logging.ServerComponent)
	id := c.Param("id")

	var limit float64

	var err error

	if value := c.DefaultQuery("limit", "10"); value == "" {
		limit = -1
	} else if limit, err = strconv.ParseFloat(value, bitSize32); err != nil {
		logger.Error().Str("value", value).Msg("invalid value for limit (%s is not an int32)")
	}

	status := globalHandler.GetNodeAvailability(c.Request.Context(), c.DefaultQuery("group", "all"), id, limit)
	if status == ghStatusOk {
		c.IndentedJSON(http.StatusOK, status)
	} else if strings.HasPrefix(status, "exceeded") {
		c.IndentedJSON(http.StatusRequestTimeout, status)
	} else {
		c.IndentedJSON(http.StatusExpectationFailed, status)
	}
}
