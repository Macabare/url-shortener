package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Macabare/url-shortener/internal/lib/api/response"
	"github.com/Macabare/url-shortener/internal/lib/logger/sl"
	"github.com/Macabare/url-shortener/internal/lib/random"
	"github.com/Macabare/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type UrlSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// TODO: maybe config?
const aliasLength = 7

func New(logger *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			logger.Error("Failed to parse request", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to process request"))
			return
		}

		logger.Debug("parsed request", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validatorErr := err.(validator.ValidationErrors)
			logger.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validatorErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrUrlExists) {
			logger.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url alreadyexists"))
			return
		}
		if err != nil {
			logger.Error("failed to save URL", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to save url"))
			return
		}

		logger.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.Ok(),
			Alias:    alias,
		})
	}
}
