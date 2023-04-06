package gqlgenmetrics

import "go.opentelemetry.io/otel/metric"

// Option applies an option value when creating a Handler.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (f optionFunc) apply(c *config) {
	f(c)
}

type config struct {
	meterProvider       metric.MeterProvider
	instrumentationName string

	instrumentRequestDuration bool
	instrumentRequestCount    bool

	instrumentResolverDuration   bool
	instrumentResolverCount      bool
	instrumentResolverCustomOnly bool
}

// WithInstrumentationName returns an Option to set custom name for metrics scope.
func WithInstrumentationName(name string) Option {
	return optionFunc(func(c *config) {
		c.instrumentationName = name
	})
}

// WithMeterProvider returns an Option to use custom MetricProvider when creating metrics.
func WithMeterProvider(p metric.MeterProvider) Option {
	return optionFunc(func(c *config) {
		c.meterProvider = p
	})
}

// WithInstrumentRequestDuration enable/disable reporting of 'gql.request.duration' metric
// which is a histogram and could results in high cardinality.
// enabled by default.
func WithInstrumentRequestDuration(instrumentRequestDuration bool) Option {
	return optionFunc(func(c *config) {
		c.instrumentRequestDuration = instrumentRequestDuration
	})
}

// WithInstrumentRequestCount enable/disable reporting of 'gql.request.completed' metric
// enabled by default.
func WithInstrumentRequestCount(instrumentRequestCount bool) Option {
	return optionFunc(func(c *config) {
		c.instrumentRequestCount = instrumentRequestCount
	})
}

// WithInstrumentResolverDuration enable/disable reporting of 'gql.resolver.duration' metric
// which is a histogram and could results in high cardinality.
// enabled by default.
func WithInstrumentResolverDuration(instrumentResolverDuration bool) Option {
	return optionFunc(func(c *config) {
		c.instrumentResolverDuration = instrumentResolverDuration
	})
}

// WithInstrumentResolverCount enable/disable reporting of 'gql.resolver.completes' metric.
// enabled by default.
func WithInstrumentResolverCount(instrumentResolverCount bool) Option {
	return optionFunc(func(c *config) {
		c.instrumentResolverCount = instrumentResolverCount
	})
}

// WithInstrumentResolverCustomOnly allows reducing cardinality of the 'gql.resolver.duration' and 'gql.resolver.completed'
// metrics by only reporting custom field resolvers.
// disabled by default.
func WithInstrumentResolverCustomOnly(instrumentResolverCustomOnly bool) Option {
	return optionFunc(func(c *config) {
		c.instrumentResolverCustomOnly = instrumentResolverCustomOnly
	})
}
