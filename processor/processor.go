// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package processor // import "go.opentelemetry.io/collector/processor"

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
)

var (
	errNilNextConsumer = errors.New("nil next Consumer")
)

// Traces is a processor that can consume traces.
type Traces interface {
	component.Component
	consumer.Traces
}

// Metrics is a processor that can consume metrics.
type Metrics interface {
	component.Component
	consumer.Metrics
}

// Logs is a processor that can consume logs.
type Logs interface {
	component.Component
	consumer.Logs
}

// Profiles is a processor that can consume Profiles.
type Profiles interface {
	component.Component
	consumer.Profiles
}

// CreateSettings is passed to Create* functions in Factory.
type CreateSettings struct {
	// ID returns the ID of the component that will be created.
	ID component.ID

	component.TelemetrySettings

	// BuildInfo can be used by components for informational purposes
	BuildInfo component.BuildInfo
}

// Factory is Factory interface for processors.
//
// This interface cannot be directly implemented. Implementations must
// use the NewProcessorFactory to implement it.
type Factory interface {
	component.Factory

	// CreateTracesProcessor creates a TracesProcessor based on this config.
	// If the processor type does not support tracing or if the config is not valid,
	// an error will be returned instead.
	CreateTracesProcessor(ctx context.Context, set CreateSettings, cfg component.Config, nextConsumer consumer.Traces) (Traces, error)

	// TracesProcessorStability gets the stability level of the TracesProcessor.
	TracesProcessorStability() component.StabilityLevel

	// CreateMetricsProcessor creates a MetricsProcessor based on this config.
	// If the processor type does not support metrics or if the config is not valid,
	// an error will be returned instead.
	CreateMetricsProcessor(ctx context.Context, set CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (Metrics, error)

	// MetricsProcessorStability gets the stability level of the MetricsProcessor.
	MetricsProcessorStability() component.StabilityLevel

	// CreateLogsProcessor creates a LogsProcessor based on the config.
	// If the processor type does not support logs or if the config is not valid,
	// an error will be returned instead.
	CreateLogsProcessor(ctx context.Context, set CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (Logs, error)

	// LogsProcessorStability gets the stability level of the LogsProcessor.
	LogsProcessorStability() component.StabilityLevel

	// CreateProfilesProcessor creates a ProfilesProcessor based on the config.
	// If the processor type does not support Profiles or if the config is not valid,
	// an error will be returned instead.
	CreateProfilesProcessor(ctx context.Context, set CreateSettings, cfg component.Config, nextConsumer consumer.Profiles) (Profiles, error)

	// LogsProcessorStability gets the stability level of the LogsProcessor.
	ProfilesProcessorStability() component.StabilityLevel

	unexportedFactoryFunc()
}

// FactoryOption apply changes to Options.
type FactoryOption interface {
	// applyProcessorFactoryOption applies the option.
	applyProcessorFactoryOption(o *factory)
}

var _ FactoryOption = (*factoryOptionFunc)(nil)

// factoryOptionFunc is a FactoryOption created through a function.
type factoryOptionFunc func(*factory)

func (f factoryOptionFunc) applyProcessorFactoryOption(o *factory) {
	f(o)
}

// CreateTracesFunc is the equivalent of Factory.CreateTraces().
type CreateTracesFunc func(context.Context, CreateSettings, component.Config, consumer.Traces) (Traces, error)

// CreateTracesProcessor implements Factory.CreateTracesProcessor().
func (f CreateTracesFunc) CreateTracesProcessor(
	ctx context.Context,
	set CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces) (Traces, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateMetricsFunc is the equivalent of Factory.CreateMetrics().
type CreateMetricsFunc func(context.Context, CreateSettings, component.Config, consumer.Metrics) (Metrics, error)

// CreateMetricsProcessor implements Factory.CreateMetricsProcessor().
func (f CreateMetricsFunc) CreateMetricsProcessor(
	ctx context.Context,
	set CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (Metrics, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateLogsFunc is the equivalent of Factory.CreateLogs().
type CreateLogsFunc func(context.Context, CreateSettings, component.Config, consumer.Logs) (Logs, error)

// CreateLogsProcessor implements Factory.CreateLogsProcessor().
func (f CreateLogsFunc) CreateLogsProcessor(
	ctx context.Context,
	set CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (Logs, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateProfilesFunc is the equivalent of Factory.CreateProfiles().
type CreateProfilesFunc func(context.Context, CreateSettings, component.Config, consumer.Profiles) (Profiles, error)

// CreateProfilesProcessor implements Factory.CreateProfilesProcessor().
func (f CreateProfilesFunc) CreateProfilesProcessor(
	ctx context.Context,
	set CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Profiles,
) (Profiles, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

type factory struct {
	cfgType component.Type
	component.CreateDefaultConfigFunc
	CreateTracesFunc
	tracesStabilityLevel component.StabilityLevel
	CreateMetricsFunc
	metricsStabilityLevel component.StabilityLevel
	CreateLogsFunc
	logsStabilityLevel component.StabilityLevel
	CreateProfilesFunc
	profilesStabilityLevel component.StabilityLevel
}

func (f *factory) Type() component.Type {
	return f.cfgType
}

func (f *factory) unexportedFactoryFunc() {}

func (f factory) TracesProcessorStability() component.StabilityLevel {
	return f.tracesStabilityLevel
}

func (f factory) MetricsProcessorStability() component.StabilityLevel {
	return f.metricsStabilityLevel
}

func (f factory) LogsProcessorStability() component.StabilityLevel {
	return f.logsStabilityLevel
}

func (f factory) ProfilesProcessorStability() component.StabilityLevel {
	return f.profilesStabilityLevel
}

// WithTraces overrides the default "error not supported" implementation for CreateTraces and the default "undefined" stability level.
func WithTraces(createTraces CreateTracesFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.tracesStabilityLevel = sl
		o.CreateTracesFunc = createTraces
	})
}

// WithMetrics overrides the default "error not supported" implementation for CreateMetrics and the default "undefined" stability level.
func WithMetrics(createMetrics CreateMetricsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.metricsStabilityLevel = sl
		o.CreateMetricsFunc = createMetrics
	})
}

// WithLogs overrides the default "error not supported" implementation for CreateLogs and the default "undefined" stability level.
func WithLogs(createLogs CreateLogsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.logsStabilityLevel = sl
		o.CreateLogsFunc = createLogs
	})
}

// WithProfiles overrides the default "error not supported" implementation for CreateProfiles and the default "undefined" stability level.
func WithProfiles(createProfiles CreateProfilesFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.profilesStabilityLevel = sl
		o.CreateProfilesFunc = createProfiles
	})
}

// NewFactory returns a Factory.
func NewFactory(cfgType component.Type, createDefaultConfig component.CreateDefaultConfigFunc, options ...FactoryOption) Factory {
	f := &factory{
		cfgType:                 cfgType,
		CreateDefaultConfigFunc: createDefaultConfig,
	}
	for _, opt := range options {
		opt.applyProcessorFactoryOption(f)
	}
	return f
}

// MakeFactoryMap takes a list of factories and returns a map with Factory type as keys.
// It returns a non-nil error when there are factories with duplicate type.
func MakeFactoryMap(factories ...Factory) (map[component.Type]Factory, error) {
	fMap := map[component.Type]Factory{}
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return fMap, fmt.Errorf("duplicate processor factory %q", f.Type())
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}

// Builder processor is a helper struct that given a set of Configs and Factories helps with creating processors.
type Builder struct {
	cfgs      map[component.ID]component.Config
	factories map[component.Type]Factory
}

// NewBuilder creates a new processor.Builder to help with creating components form a set of configs and factories.
func NewBuilder(cfgs map[component.ID]component.Config, factories map[component.Type]Factory) *Builder {
	return &Builder{cfgs: cfgs, factories: factories}
}

// CreateTraces creates a Traces processor based on the settings and config.
func (b *Builder) CreateTraces(ctx context.Context, set CreateSettings, next consumer.Traces) (Traces, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("processor %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("processor factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.TracesProcessorStability())
	return f.CreateTracesProcessor(ctx, set, cfg, next)
}

// CreateMetrics creates a Metrics processor based on the settings and config.
func (b *Builder) CreateMetrics(ctx context.Context, set CreateSettings, next consumer.Metrics) (Metrics, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("processor %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("processor factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.MetricsProcessorStability())
	return f.CreateMetricsProcessor(ctx, set, cfg, next)
}

// CreateLogs creates a Logs processor based on the settings and config.
func (b *Builder) CreateLogs(ctx context.Context, set CreateSettings, next consumer.Logs) (Logs, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("processor %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("processor factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.LogsProcessorStability())
	return f.CreateLogsProcessor(ctx, set, cfg, next)
}

// CreateProfiles creates a Profiles processor based on the settings and config.
func (b *Builder) CreateProfiles(ctx context.Context, set CreateSettings, next consumer.Profiles) (Profiles, error) {
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("processor %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("processor factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.ProfilesProcessorStability())
	return f.CreateProfilesProcessor(ctx, set, cfg, next)
}

func (b *Builder) Factory(componentType component.Type) component.Factory {
	return b.factories[componentType]
}

// logStabilityLevel logs the stability level of a component. The log level is set to info for
// undefined, unmaintained, deprecated and development. The log level is set to debug
// for alpha, beta and stable.
func logStabilityLevel(logger *zap.Logger, sl component.StabilityLevel) {
	if sl >= component.StabilityLevelAlpha {
		logger.Debug(sl.LogMessage())
	} else {
		logger.Info(sl.LogMessage())
	}
}
