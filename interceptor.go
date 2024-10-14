package httpr

import (
	"context"
	"errors"
	"net/http"
)

type Interceptor interface {
	Handle(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error)
}

type HandleFunc func(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error)

func (f HandleFunc) Handle(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error) {
	return f(ctx, req, next)
}

func Chain(interceptors ...Interceptor) Interceptor {
	return HandleFunc(func(ctx context.Context, req *http.Request, _ Interceptor) (*http.Response, error) {
		if len(interceptors) == 0 {
			return nil, errors.New("no interceptors in chain")
		}

		var next Interceptor
		for i := len(interceptors) - 1; i >= 0; i-- {
			current := interceptors[i]
			if next == nil {
				next = HandleFunc(func(ctx context.Context, req *http.Request, _ Interceptor) (*http.Response, error) {
					return current.Handle(ctx, req, nil)
				})
			} else {
				nextCopy := next
				next = HandleFunc(func(ctx context.Context, req *http.Request, _ Interceptor) (*http.Response, error) {
					return current.Handle(ctx, req, nextCopy)
				})
			}
		}

		return next.Handle(ctx, req, nil)
	})
}
