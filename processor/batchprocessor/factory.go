// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package batchprocessor // import "go.opentelemetry.io/collector/processor/batchprocessor"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor/internal/metadata"
)

const (
	defaultSendBatchSize = uint32(8192)
	defaultTimeout       = 200 * time.Millisecond

	// defaultMetadataCardinalityLimit should be set to the number
	// of metadata configurations the user expects to submit to
	// the collector.
	defaultMetadataCardinalityLimit = 1000
)

// NewFactory returns a new factory for the Batch processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTraces, metadata.TracesStability),
		processor.WithMetrics(createMetrics, metadata.MetricsStability),
		processor.WithLogs(createLogs, metadata.LogsStability),
		processor.WithProfiles(createProfiles, metadata.ProfilesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		SendBatchSize:            defaultSendBatchSize,
		Timeout:                  defaultTimeout,
		MetadataCardinalityLimit: defaultMetadataCardinalityLimit,
	}
}

func createTraces(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	return newBatchTracesProcessor(set, nextConsumer, cfg.(*Config))
}

func createMetrics(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	return newBatchMetricsProcessor(set, nextConsumer, cfg.(*Config))
}

func createLogs(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	return newBatchLogsProcessor(set, nextConsumer, cfg.(*Config))
}

func createProfiles(
	_ context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Profiles,
) (processor.Profiles, error) {
	return newBatchProfilesProcessor(set, nextConsumer, cfg.(*Config))
}
