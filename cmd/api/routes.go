package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", app.healthHandler)

	mux.HandleFunc("POST /api/v1/signup", app.signupHandler)
	mux.HandleFunc("POST /api/v1/login", app.loginHandler)
	mux.HandleFunc("POST /api/v1/logout", app.logoutHandler)

	mux.Handle("GET /api/v1/me", app.requireAuth(http.HandlerFunc(app.meHandler)))

	return chain(mux,
		app.requestID,
		app.logRequest,
		app.recoverPanic,
		app.enableCORS,
	)
}

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}
