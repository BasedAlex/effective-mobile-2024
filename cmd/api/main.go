package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/basedalex/effective-mobile-test/internal/api"
	"github.com/basedalex/effective-mobile-test/internal/config"
	"github.com/basedalex/effective-mobile-test/internal/db"
	"github.com/basedalex/effective-mobile-test/internal/router"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := config.New()

	database, err := db.NewPostgres(ctx, cfg.Env.PGDSN)
	if err != nil {
		log.Panic(err)
	}

	apiClient := api.New(cfg.Env.APIHost, cfg.Env.APIPort)

	log.Info("connected to db")
	log.Info("connecting to ", cfg.Env.Port)
	err = router.NewServer(ctx, cfg, database, apiClient)
	if err != nil {
		log.Panic(err)
	}
	
	
}