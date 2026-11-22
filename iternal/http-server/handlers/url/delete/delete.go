package delete

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Kudzeri/go-url-shortener/iternal/lib/api/response"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-gonic/gin"
)

type URLDeleter interface {
	DeleteUrl(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) gin.HandlerFunc {
	const op = "handlers.url.delete.New"
	log = log.With(slog.String("op", op))

	return func(ctx *gin.Context) {
		alias := ctx.Param("alias")
		if alias == "" {
			log.Info("no alias provided")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, resp.Error("no alias provided"))
			return
		}

		if err := urlDeleter.DeleteUrl(alias); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Info("alias not found")
				ctx.AbortWithStatusJSON(http.StatusNotFound, resp.Error("alias not found"))
				return
			}

			log.Error(err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusNoContent, resp.OK())
	}
}
