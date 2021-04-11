package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"web/web"
)

var ErrEmpty = errors.New("empty string")

type StringService interface {
	Uppercase(string) (string, error)
}

type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V string `json:"v"`
}

func main() {
	r := web.NewRouter(web.Config{})
	r.Handle("GET", "/", makeUppercaseHandler(stringService{}), web.WithDecodeRequestFunc(decodeUppercaseRequest))
	log.Fatal(http.ListenAndServe(":8080", r))
}

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func makeUppercaseHandler(svc StringService) web.Handler {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return nil, err
		}

		return uppercaseResponse{v}, nil
	}
}
