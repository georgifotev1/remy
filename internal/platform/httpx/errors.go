package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/georgifotev1/remy/internal/platform/apperr"
)

const InternalErrorMsg = "the server encountered a problem and could not process your request"

func HandleError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	var (
		verr *apperr.ValidationError
		berr *apperr.BadRequestError
	)

	switch {
	case errors.As(err, &verr):
		Write(w, http.StatusUnprocessableEntity, Envelope{"errors": verr.Fields}, nil)

	case errors.As(err, &berr):
		WriteError(w, http.StatusBadRequest, berr.Msg)

	case errors.Is(err, apperr.ErrNotFound):
		WriteError(w, http.StatusNotFound, "the requested resource could not be found")

	case errors.Is(err, apperr.ErrConflict):
		WriteError(w, http.StatusConflict, "a resource with these details already exists")

	case errors.Is(err, apperr.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, "the request contains invalid data")

	case errors.Is(err, apperr.ErrForbidden):
		WriteError(w, http.StatusForbidden, "access denied")

	case errors.Is(err, context.Canceled):
		log.Info("request cancelled", "method", r.Method, "uri", r.URL.RequestURI())

	case errors.Is(err, context.DeadlineExceeded):
		log.Warn("request timed out",
			"err", err, "method", r.Method, "uri", r.URL.RequestURI())
		WriteError(w, http.StatusServiceUnavailable, "the request timed out, please try again")

	default:
		log.Error("internal error",
			"err", err,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"request_id", r.Header.Get("X-Request-ID"),
		)
		WriteError(w, http.StatusInternalServerError, InternalErrorMsg)
	}
}
