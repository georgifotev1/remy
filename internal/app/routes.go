package app

import (
	"net/http"

	"github.com/georgifotev1/remy/internal/identity"
	mw "github.com/georgifotev1/remy/internal/platform/middleware"
)

func (app *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)

	api := http.NewServeMux()
	identity.NewIdentityHandlers(app.logger).Routes(api)

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	return mw.Chain(mux,
		mw.RequestID,
		mw.RequestLogger(app.logger),
		mw.Recoverer(app.logger),
		mw.CORS(app.cfg.CORS.AllowedOrigins),
	)
}

func (app *App) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}
