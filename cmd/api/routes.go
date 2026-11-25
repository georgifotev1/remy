package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", app.healthHandler)

	mux.HandleFunc("POST /api/v1/signup", app.signupHandler)
	mux.HandleFunc("POST /api/v1/login", app.loginHandler)
	mux.HandleFunc("POST /api/v1/logout", app.logoutHandler)

	return chain(mux,
		app.requestID,
		app.logRequest,
		app.recoverPanic,
		app.enableCORS,
	)
}

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}
