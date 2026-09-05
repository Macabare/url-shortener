package stat

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	resp "github.com/Macabare/url-shortener/internal/lib/api/response"
	"github.com/Macabare/url-shortener/internal/lib/logger/sl"
	"github.com/Macabare/url-shortener/internal/model"
	"github.com/Macabare/url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UrlService interface {
	GetURLStat(ctx context.Context, shortCode string) (model.URL, error)
}

type Response struct {
	resp.Response
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	AccessCount int64     `json:"access_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func New(logger *slog.Logger, s UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.stat.New"

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

		urlObj, err := s.GetURLStat(ctx, shortCode)
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

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response:    resp.Ok(),
			OriginalURL: urlObj.OriginalURL,
			ShortCode:   urlObj.ShortCode,
			AccessCount: urlObj.AccessCount,
			CreatedAt:   urlObj.CreatedAt,
			UpdatedAt:   urlObj.UpdatedAt,
		})
	}
}
