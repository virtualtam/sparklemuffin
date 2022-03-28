package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/yawbe/pkg/http/www"
	"github.com/virtualtam/yawbe/pkg/storage/memory"
	"github.com/virtualtam/yawbe/pkg/user"
)

const (
	defaultDebugMode  = false
	defaultHMACKey    = "hmac-secret-key"
	defaultListenAddr = "0.0.0.0:8080"
)

func main() {
	debugMode := flag.Bool("debug", defaultDebugMode, "Enable debugging")
	hmacKey := flag.String("hmacKey", defaultHMACKey, "Secret HMAC key")
	listenAddr := flag.String("listenAddr", defaultListenAddr, "Listen on this address")
	flag.Parse()

	if *debugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	userRepository := &memory.Repository{}

	userService := user.NewService(userRepository, *hmacKey)

	// FIXME: add fixtures
	_ = userRepository.AddUser(user.User{
		Email:        "test@dev.local",
		PasswordHash: "$2y$10$JWIOiD1v77jKe2Q46tCWBelJCr/mwlQJS3kLJMPLKmwwVQssALR76",
		// Password: "test",
	})

	server := www.NewServer(userService)

	// Structured loging
	chain := alice.New(hlog.NewHandler(log.Logger), hlog.AccessHandler(accessLogger))

	httpServer := &http.Server{
		Addr:         *listenAddr,
		Handler:      chain.Then(server),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info().Msgf("Listening to http://%s", *listenAddr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("ListenAndServe")
	}
}

func accessLogger(r *http.Request, status, size int, dur time.Duration) {
	hlog.FromRequest(r).Info().
		Dur("duration_ms", dur).
		Str("host", r.Host).
		Str("path", r.URL.Path).
		Int("size", size).
		Int("status", status).
		Msg("Request")
}
