package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/georgifotev1/remy/internal/platform/apperr"
)

type Envelope map[string]any

func Write(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

func WriteError(w http.ResponseWriter, status int, msg string) error {
	return Write(w, status, Envelope{"error": msg}, nil)
}

func Read(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return badRequest(err)
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &apperr.BadRequestError{Msg: "body must only contain a single JSON value"}
	}
	return nil
}

func badRequest(err error) error {
	var (
		syntaxErr    *json.SyntaxError
		unmarshalErr *json.UnmarshalTypeError
		invalidErr   *json.InvalidUnmarshalError
		maxBytesErr  *http.MaxBytesError
		msg          string
	)

	switch {
	case errors.As(err, &syntaxErr):
		msg = fmt.Sprintf("body contains badly-formed JSON (at character %d)", syntaxErr.Offset)
	case errors.Is(err, io.ErrUnexpectedEOF):
		msg = "body contains badly-formed JSON"
	case errors.As(err, &unmarshalErr):
		if unmarshalErr.Field != "" {
			msg = fmt.Sprintf("body contains incorrect JSON type for field %q", unmarshalErr.Field)
		} else {
			msg = fmt.Sprintf("body contains incorrect JSON type (at character %d)", unmarshalErr.Offset)
		}
	case errors.Is(err, io.EOF):
		msg = "body must not be empty"
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		msg = "body contains unknown key " + strings.TrimPrefix(err.Error(), "json: unknown field ")
	case errors.As(err, &maxBytesErr):
		msg = fmt.Sprintf("body must not be larger than %d bytes", maxBytesErr.Limit)
	case errors.As(err, &invalidErr):
		panic(err) // programmer error: dst must be a non-nil pointer
	default:
		return err
	}
	return &apperr.BadRequestError{Msg: msg}
}
