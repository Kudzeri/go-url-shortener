package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Kudzeri/go-url-shortener/iternal/lib/api/response"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/sl"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/random"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" binding:"required,http_url"`
	Alias string `json:"alias" binding:"omitempty,min=3,max=32,alphanum"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

// TODO: move to config
const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v3@v3.9.1

func New(log *slog.Logger, urlSaver URLSaver) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.save.New"

		log.With(slog.String("op", op),
			slog.String("request_id", requestid.Get(c)))
		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			var errs validator.ValidationErrors

			if errors.As(err, &errs) {
				c.JSON(http.StatusBadRequest, resp.ValidationError(errs))
				return
			}

			c.JSON(http.StatusBadRequest, resp.Error("Ожидается корректный JSON с полями нужных типов"))
			return
		}

		log.Info("Request body is bound", slog.Any("req", req))

		alias := req.Alias
		if alias == "" {
			var err error
			alias, err = random.NewRandomString(aliasLength)
			if err != nil {
				log.Error("failed to generate alias", sl.Error(err))
				c.JSON(http.StatusInternalServerError, resp.Error("Внутренняя ошибка сервера"))
				return
			}
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Info("URL already exists", slog.Any("url", req.URL))

			c.JSON(http.StatusConflict, resp.Error("URL already exists"))
			return
		}
		if err != nil {
			log.Error("failed to save URL", sl.Error(err))
			c.JSON(http.StatusInternalServerError, resp.Error("failed to save URL"))
			return
		}

		log.Info("URL saved", slog.Int64("id", id))
		c.JSON(http.StatusOK, Response{
			Response: *resp.OK(),
			Alias:    alias,
		})
	}
}
