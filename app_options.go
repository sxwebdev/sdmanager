package sdmanager

type AppOptions struct {
	serviceName string
}

type AppOption func(*AppOptions)

func WithServiceName(serviceName string) AppOption {
	return func(o *AppOptions) {
		o.serviceName = serviceName
	}
}
