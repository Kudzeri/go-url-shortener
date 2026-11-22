package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Kudzeri/go-url-shortener/iternal/lib/api/response"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-gonic/gin"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const op = "handlers.url.redirect.New"
		log = log.With(slog.String("op", op))

		alias := ctx.Param("alias")
		if alias == "" {
			log.Info("no alias provided")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, resp.Error("no alias provided"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrNotFound) {
			log.Info("url not found", "alias", alias)
			ctx.AbortWithStatusJSON(http.StatusNotFound, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", "alias", alias, "err", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, resp.Error("internal server error"))
			return
		}

		log.Info("got url", "alias", alias, "url", resURL)

		http.Redirect(ctx.Writer, ctx.Request, resURL, http.StatusFound)
	}
}
