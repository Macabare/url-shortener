package health

import (
	"log/slog"
	"net/http"

	resp "github.com/Macabare/url-shortener/internal/lib/api/response"
	"github.com/go-chi/render"
)

func New(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.Ok())
	}
}
