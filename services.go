// Copyright 2017 Axel Etcheverry. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DigitalAnswer/crypto-ticker-server/config"
	"github.com/DigitalAnswer/crypto-ticker-server/controller"
	"github.com/DigitalAnswer/crypto-ticker-server/database"
	"github.com/DigitalAnswer/crypto-ticker-server/server"
	"github.com/DigitalAnswer/crypto-ticker-server/server/authentication"
	"github.com/DigitalAnswer/crypto-ticker-server/services"
	"github.com/asdine/storm"
	"github.com/caarlos0/env"
	"github.com/euskadi31/go-service"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// Service Container
var Service = service.New()

func init() {
	Service.Set("logger", func(c *service.Container) interface{} {
		cfg := c.Get("config").(*config.Configuration)

		logger := zerolog.New(os.Stdout).With().
			Timestamp().
			Str("role", "crypto-ticker-server").
			//Str("host", host).
			Logger()

		stdlog.SetFlags(0)
		stdlog.SetOutput(logger)

		log.Logger = logger

		if cfg.Debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		return logger
	})

	Service.Set("config", func(c *service.Container) interface{} {
		cfg := &config.Configuration{}

		if err := env.Parse(cfg); err != nil {
			log.Fatal().Err(err)
		}

		log.Debug().Msgf("Config: %v", cfg)

		return cfg
	})

	Service.Set("db", func(c *service.Container) interface{} {
		cfg := c.Get("config").(*config.Configuration)
		path := strings.TrimRight(cfg.Path, "/")

		db, err := storm.Open(fmt.Sprintf("%s/hyperpaas.db", path))
		if err != nil {
			log.Fatal().Err(err)
		}

		return db
	})

	Service.Set("user", func(c *service.Container) interface{} {
		db := c.Get("db").(*storm.DB)

		return services.NewUserService(db)
	})

	Service.Set("authentication", func(c *service.Container) interface{} {
		cfg := c.Get("config").(*config.Configuration)
		user := c.Get("user").(*services.UserService)

		auth := services.NewAuthenticationService(cfg.Path, user)

		if !auth.HasKey() {
			log.Info().Msg("Generate RSA key...")

			if err := auth.GenerateKey(); err != nil {
				log.Fatal().Err(err)
			}
		}

		if err := auth.LoadKeys(); err != nil {
			log.Fatal().Err(err)
		}

		return auth
	})

	Service.Set("router", func(c *service.Container) interface{} {
		logger := c.Get("logger").(zerolog.Logger)
		db := c.Get("db").(*storm.DB)
		auth := c.Get("authentication").(*services.AuthenticationService)
		user := c.Get("user").(*services.UserService)

		corsHandler := cors.New(cors.Options{
			AllowCredentials: false,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodOptions, http.MethodPost, http.MethodPut, http.MethodDelete},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			Debug:            true,
		})

		router := server.NewRouter()

		router.Use(hlog.NewHandler(logger))
		router.Use(database.NewHandler(db))
		router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}))
		router.Use(hlog.RemoteAddrHandler("ip"))
		router.Use(hlog.UserAgentHandler("user_agent"))
		router.Use(hlog.RefererHandler("referer"))
		router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
		router.Use(corsHandler.Handler)
		router.Use(authentication.NewAuthHandler(auth))

		router.EnableHealthCheck()

		apiController, err := controller.NewAPIController(auth, user)
		if err != nil {
			log.Fatal().Err(err)
		}

		router.AddController(apiController)

		return router
	})
}

// Run Application
func Run() {
	_ = Service.Get("logger").(zerolog.Logger)

	addr := fmt.Sprintf(":%d", 8080)

	router := Service.Get("router").(*server.Router)

	log.Info().Msgf("Server running on %s", addr)

	log.Fatal().Err(http.ListenAndServe(addr, router))
}
