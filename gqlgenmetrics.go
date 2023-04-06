package gqlgenmetrics

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

const (
	// DefaultInstrumentationName is the default used when creating meters.
	DefaultInstrumentationName = "github.com/mahboubii/gqlgenmetrics"
)

type middleware struct {
	requestsCompleted instrument.Int64Counter
	requestsDuration  instrument.Float64Histogram

	resolversCompleted instrument.Int64Counter
	resolverDuration   instrument.Float64Histogram

	customResolverOnly bool
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = middleware{}

func Middleware(options ...Option) middleware { //nolint:revive
	c := config{
		instrumentRequestCount:       true,
		instrumentRequestDuration:    true,
		instrumentResolverDuration:   true,
		instrumentResolverCount:      true,
		instrumentResolverCustomOnly: false,
		instrumentationName:          DefaultInstrumentationName,
		meterProvider:                global.MeterProvider(),
	}

	for _, o := range options {
		o.apply(&c)
	}

	meter := c.meterProvider.Meter(c.instrumentationName)

	var err error

	m := middleware{
		customResolverOnly: c.instrumentResolverCustomOnly,
	}

	if c.instrumentRequestDuration {
		m.requestsDuration, err = meter.Float64Histogram("gql.request.duration", instrument.WithUnit("ms"), instrument.WithDescription("The time taken for server to process the request."))
		if err != nil {
			panic(err)
		}
	}

	if c.instrumentRequestCount {
		m.requestsCompleted, err = meter.Int64Counter("gql.request.completed", instrument.WithUnit("1"), instrument.WithDescription("Total number of requests completed."))
		if err != nil {
			panic(err)
		}
	}

	if c.instrumentResolverDuration {
		m.resolverDuration, err = meter.Float64Histogram("gql.resolver.duration", instrument.WithUnit("ms"), instrument.WithDescription("The time taken for server to resolve a resolver."))
		if err != nil {
			panic(err)
		}
	}

	if c.instrumentResolverCount {
		m.resolversCompleted, err = meter.Int64Counter("gql.resolver.completed", instrument.WithUnit("1"), instrument.WithDescription("Total number of resolvers completed."))
		if err != nil {
			panic(err)
		}
	}

	return m
}

func (m middleware) ExtensionName() string {
	return "gqlgenmetrics"
}

func (m middleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (m middleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	errs := graphql.GetErrors(ctx)

	opName := oc.OperationName
	if opName == "" {
		opName = "nameless-op"
	}

	if m.requestsDuration != nil {
		m.requestsDuration.Record(ctx, float64(time.Since(oc.Stats.OperationStart).Milliseconds()),
			attribute.Key("gql.request.name").String(opName),
		)
	}

	if m.requestsCompleted != nil {
		m.requestsCompleted.Add(ctx, 1,
			attribute.Key("gql.request.name").String(opName),
			attribute.Key("gql.request.error").Bool(len(errs) != 0),
		)
	}

	return next(ctx)
}

func (m middleware) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	begin := time.Now()
	res, err := next(ctx)

	fc := graphql.GetFieldContext(ctx)

	if m.resolverDuration != nil && (!m.customResolverOnly || fc.IsResolver) {
		m.resolverDuration.Record(ctx, float64(time.Since(begin).Milliseconds()),
			attribute.Key("gql.resolver.object").String(fc.Object),
			attribute.Key("gql.resolver.field").String(fc.Field.Name),
		)
	}

	if m.resolversCompleted != nil && (!m.customResolverOnly || fc.IsResolver) {
		m.resolversCompleted.Add(ctx, 1,
			attribute.Key("gql.resolver.object").String(fc.Object),
			attribute.Key("gql.resolver.field").String(fc.Field.Name),
			attribute.Key("gql.resolver.error").Bool(err != nil),
		)
	}

	return res, err
}
