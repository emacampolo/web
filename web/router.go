package web

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler func(ctx context.Context, request interface{}) (response interface{}, err error)

type Middleware func(http.Handler) http.Handler

type HandlerOption func(*handler)

func WithDecodeRequestFunc(dec DecodeRequestFunc) HandlerOption {
	return func(h *handler) { h.decoder = dec }
}

func WithEncodeResponseFunc(enc EncodeResponseFunc) HandlerOption {
	return func(h *handler) { h.encoder = enc }
}

func WithErrorEncoder(ee ErrorEncoder) HandlerOption {
	return func(h *handler) { h.errorEncoder = ee }
}

func WithErrorHandler(errorHandler ErrorHandler) HandlerOption {
	return func(h *handler) { h.errorHandler = errorHandler }
}

func WithMiddleware(mw []Middleware) HandlerOption {
	return func(h *handler) { h.middleware = append(h.middleware, mw...) }
}

type handler struct {
	h            Handler
	decoder      DecodeRequestFunc
	encoder      EncodeResponseFunc
	errorEncoder ErrorEncoder
	errorHandler ErrorHandler
	middleware   []Middleware
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := h.decoder(ctx, r)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
		return
	}

	response, err := h.h(ctx, request)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
		return
	}

	if err := h.encoder(ctx, w, response); err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
		return
	}
}

type Router struct {
	mux *chi.Mux
	mw  []Middleware
}

type Config struct {
	Mw []Middleware
}

func NewRouter(cfg Config) *Router {
	return &Router{
		mux: chi.NewRouter(),
		mw:  cfg.Mw,
	}
}

func (a *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *Router) Handle(method, pattern string, handlerFunc Handler, opts ...HandlerOption) {
	handler := handler{
		h:            handlerFunc,
		decoder:      basicDecoder,
		encoder:      EncodeJSONResponse,
		errorEncoder: DefaultErrorEncoder,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(&handler)
	}

	h := chain(chain(handler, a.mw), handler.middleware)
	a.mux.Method(method, pattern, h)
}

func chain(handler http.Handler, mw []Middleware) http.Handler {
	// Loop backwards through the middleware invoking each one. Replace the
	// handler with the new wrapped handler. Looping backwards ensures that the
	// first middleware of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
