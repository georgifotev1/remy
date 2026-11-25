package main

import (
	"net/http"
)

func (app *application) meHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.models.Users.GetByID(r.Context(), contextGetUserID(r))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
