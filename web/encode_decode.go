package web

import (
	"context"
	"encoding/json"
	"net/http"
)

type ErrorEncoder func(ctx context.Context, err error, w http.ResponseWriter)

type DecodeRequestFunc func(context.Context, *http.Request) (request interface{}, err error)

type EncodeResponseFunc func(context.Context, http.ResponseWriter, interface{}) error

type StatusCoder interface {
	StatusCode() int
}

type Headerer interface {
	Headers() http.Header
}

func DefaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	contentType, body := "text/plain; charset=utf-8", []byte(err.Error())
	if m, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := m.MarshalJSON(); marshalErr == nil {
			contentType, body = "application/json; charset=utf-8", jsonBody
		}
	}

	w.Header().Set("Content-Type", contentType)
	if h, ok := err.(Headerer); ok {
		for k, values := range h.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	code := http.StatusInternalServerError
	if sc, ok := err.(StatusCoder); ok {
		code = sc.StatusCode()
	}

	w.WriteHeader(code)
	w.Write(body)
}

func EncodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if h, ok := response.(Headerer); ok {
		for k, values := range h.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	code := http.StatusOK
	if sc, ok := response.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	if code == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

func basicDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	return r, nil
}
