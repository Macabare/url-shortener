package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Macabare/url-shortener/internal/lib/api/response"
	"github.com/Macabare/url-shortener/internal/lib/logger/sl"
	"github.com/Macabare/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UrlService interface {
	GetURL(ctx context.Context, shortCode string) (string, error)
	IncrementAccessCount(ctx context.Context, shortCode string) error
}

type Response struct {
	resp.Response
	URL string `json:"url,omitempty"`
}

func New(logger *slog.Logger, s UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.get.New"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		ctx := r.Context()

		shortCode := chi.URLParam(r, "shortCode")
		if shortCode == "" {
			logger.Error("invalid request", slog.String("shortCode", shortCode))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to process request"))
			return
		}

		url, err := s.GetURL(ctx, shortCode)
		if errors.Is(err, storage.ErrUrlNotFound) {
			logger.Info("cannot find url", slog.String("shortCode", shortCode))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("cannot find URL"))
			return
		}
		if err != nil {
			logger.Error("failed to get URL", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get URL"))
			return
		}

		if err := s.IncrementAccessCount(ctx, shortCode); err != nil {
			logger.Error(
				"failed to increment access count",
				sl.Err(err),
				slog.String("shortCode", shortCode),
			)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Ok(),
			URL:      url,
		})
	}
}
