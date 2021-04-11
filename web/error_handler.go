package web

import (
	"context"
)

// ErrorHandler receives a transport error to be processed for diagnostic purposes.
// Usually this means logging the error.
type ErrorHandler interface {
	Handle(ctx context.Context, err error)
}

var DefaultErrorHandler = &errorHandler{}

type errorHandler struct{}

func (e errorHandler) Handle(ctx context.Context, err error) {
	e.notifyErr(ctx, err)
	e.logErr(ctx, err)
}

func (e errorHandler) notifyErr(ctx context.Context, err error) {}

func (e errorHandler) logErr(ctx context.Context, err error) {}
