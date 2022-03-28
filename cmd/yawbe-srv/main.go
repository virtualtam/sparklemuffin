package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	defaultListenAddr = "0.0.0.0:8080"
)

func main() {
	debugMode := flag.Bool("debug", defaultDebugMode, "Enable debugging")
	listenAddr := flag.String("listenAddr", defaultListenAddr, "Listen on this address")
	flag.Parse()

	if *debugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	userRepository := &memory.Repository{}

	userService := user.NewService(userRepository)

	// FIXME: add fixtures
	_ = userRepository.AddUser(user.User{
		Email:        "test@dev.local",
		PasswordHash: "$2y$10$JWIOiD1v77jKe2Q46tCWBelJCr/mwlQJS3kLJMPLKmwwVQssALR76",
		// Password: "test",
	})

	router := mux.NewRouter()
	www.AddRoutes(router, userService)

	// Structured loging
	chain := alice.New(hlog.NewHandler(log.Logger), hlog.AccessHandler(accessLogger))

	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      chain.Then(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Info().Msgf("Listening to http://%s", *listenAddr)
	if err := server.ListenAndServe(); err != nil {
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
