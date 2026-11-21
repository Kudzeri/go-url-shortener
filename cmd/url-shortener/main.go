package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Kudzeri/go-url-shortener/iternal/config"
	"github.com/Kudzeri/go-url-shortener/iternal/http-server/handlers/url/redirect"
	"github.com/Kudzeri/go-url-shortener/iternal/http-server/handlers/url/save"
	"github.com/Kudzeri/go-url-shortener/iternal/http-server/middleware/requestlog"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/sl"
	"github.com/Kudzeri/go-url-shortener/iternal/storage/sqlite"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	log := sl.New(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("error to init storage", sl.Error(err))
		os.Exit(1)
	}

	router := gin.New()

	//middleware
	router.Use(requestid.New(), requestlog.New(log))
	router.Use(gin.Recovery())

	router.Handle(http.MethodGet, "/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.Handle(http.MethodPost, "/url", save.New(log, storage))
	router.Handle(http.MethodGet, "/{alias}", redirect.New(log, storage))

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info(
		"starting server",
		slog.String("addr", srv.Addr),
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server failed", sl.Error(err))
		os.Exit(1)
	}

	log.Error("server stopped")
}
