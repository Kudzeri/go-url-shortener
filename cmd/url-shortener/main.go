package main

import (
	"log/slog"
	"os"

	"github.com/Kudzeri/go-url-shortener/iternal/config"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/sl"
	"github.com/Kudzeri/go-url-shortener/iternal/middleware/requestlog"
	"github.com/Kudzeri/go-url-shortener/iternal/storage/sqlite"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := sl.New(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("error to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := gin.New()

	//middleware
	router.Use(requestid.New(), requestlog.New(log))

	//TODO: run server:
}
