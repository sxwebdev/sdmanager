package sdmanager

import "context"

type AppOptions struct {
	serviceName string
	ctx         context.Context
}

type AppOption func(*AppOptions)

func WithServiceName(serviceName string) AppOption {
	return func(o *AppOptions) {
		o.serviceName = serviceName
	}
}

func WithContext(ctx context.Context) AppOption {
	return func(o *AppOptions) {
		o.ctx = ctx
	}
}
