// Package internal holds all unexported code
package internal

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RunAPI will run the gin webserver
func (h PgRouteHandler) RunAPI() {
	var err error

	var cert tls.Certificate

	if !h.config.Debug() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET("/v1/primary", func(ctx *gin.Context) { getPrimary(ctx, h) })
	router.GET("/v1/primaries", func(ctx *gin.Context) { getPrimaries(ctx, h) })
	router.GET("/v1/standbys", func(ctx *gin.Context) { getStandbys(ctx, h) })
	router.GET("/v1/:id/status", func(ctx *gin.Context) { getStatus(ctx, h) })
	router.GET("/v1/:id/availability", func(ctx *gin.Context) { getAvailability(ctx, h) })

	logger.Debug().Msgf("Running on %s", h.config.BindTo())

	if h.config.Ssl.Enabled() {
		logger.Debug().Msg("Running with SSL")

		cert, err = tls.X509KeyPair(h.config.Ssl.MustCertBytes(), h.config.Ssl.MustKeyBytes())
		if err != nil {
			logger.Fatal().Msgf("Error parsing cert and key: %v", err)
		}

		tlsConfig := tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		}
		server := http.Server{Addr: h.config.BindTo(), Handler: router, TLSConfig: &tlsConfig}
		err = server.ListenAndServeTLS("", "")
	} else {
		logger.Debug().Msg("Running without SSL")
		err = router.Run(h.config.BindTo())
	}

	if err != nil {
		logger.Panic().Msgf("Error running API: %s", err.Error())
	}
}

func getPrimary(ctx *gin.Context, h PgRouteHandler) {
	primary := h.GetPrimaries(ctx, ctx.DefaultQuery("group", "all"))
	switch len(primary) {
	case 0:
		ctx.IndentedJSON(http.StatusNotFound, "")
	case 1:
		ctx.IndentedJSON(http.StatusOK, primary[0])
	default:
		ctx.IndentedJSON(http.StatusConflict, "")
	}
}

// getPrimaries responds with the list of all albums as JSON.
func getPrimaries(ctx *gin.Context, h PgRouteHandler) {
	primaries := h.GetPrimaries(ctx, ctx.DefaultQuery("group", "all"))
	ctx.IndentedJSON(http.StatusOK, primaries)
}

// getStandbys responds with the list of all albums as JSON.
func getStandbys(ctx *gin.Context, h PgRouteHandler) {
	ctx.IndentedJSON(http.StatusOK, h.GetStandbys(ctx, ctx.DefaultQuery("group", "all")))
}

func getStatus(ctx *gin.Context, h PgRouteHandler) {
	id := ctx.Param("id")

	status := h.GetNodeStatus(ctx, id)
	switch status {
	case ghStatusPrimary, ghStatusStandby:
		ctx.IndentedJSON(http.StatusOK, status)
	case ghStatusInvalid:
		ctx.IndentedJSON(http.StatusNotFound, status)
	case ghStatusUnavailable:
		ctx.IndentedJSON(http.StatusUnprocessableEntity, status)
	}
}

func getAvailability(ctx *gin.Context, h PgRouteHandler) {
	var (
		limit  float64
		err    error
		id     = ctx.Param("id")
		logger = log.With().Logger()
	)

	if value := ctx.DefaultQuery("limit", "10"); value == "" {
		limit = -1
	} else if limit, err = strconv.ParseFloat(value, bitSize32); err != nil {
		logger.Error().Msgf("invalid value for limit (%s is not an int32)", value)
	}

	status := h.GetNodeAvailability(ctx, id, limit)
	if status == ghStatusOk {
		ctx.IndentedJSON(http.StatusOK, status)
	} else if strings.HasPrefix(status, "exceeded") {
		ctx.IndentedJSON(http.StatusRequestTimeout, status)
	} else {
		ctx.IndentedJSON(http.StatusExpectationFailed, status)
	}
}
