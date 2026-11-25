package identity

import (
	"log/slog"
	"net/http"

	"github.com/georgifotev1/remy/internal/platform/httpx"
)

type IdentityHandlers struct {
	logger *slog.Logger
}

func NewIdentityHandlers(l *slog.Logger) *IdentityHandlers {
	return &IdentityHandlers{
		logger: l,
	}
}

func (h *IdentityHandlers) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /sign-up", h.signUp)
}

func (h *IdentityHandlers) signUp(w http.ResponseWriter, r *http.Request) {
	var payload SignupRequest
	if err := httpx.Read(w, r, &payload); err != nil {
		httpx.HandleError(w, r, h.logger, err)
		return
	}

	if err := payload.Validate(); err != nil {
		httpx.HandleError(w, r, h.logger, err)
		return
	}

	httpx.Write(w, http.StatusCreated, User{Name: payload.User.Name})
}
