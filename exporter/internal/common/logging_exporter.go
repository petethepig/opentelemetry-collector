// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// NOTE: If you are making changes to this file, consider whether you want to make similar changes
// to the Debug exporter in /exporter/debugexporter/exporter.go, which has similar logic.
// This is especially important for security issues.

package common // import "go.opentelemetry.io/collector/exporter/internal/common"

import (
	"context"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter/internal/otlptext"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type loggingExporter struct {
	verbosity         configtelemetry.Level
	logger            *zap.Logger
	logsMarshaler     plog.Marshaler
	metricsMarshaler  pmetric.Marshaler
	tracesMarshaler   ptrace.Marshaler
	profilesMarshaler pprofile.Marshaler
}

func (s *loggingExporter) pushTraces(_ context.Context, td ptrace.Traces) error {
	s.logger.Info("TracesExporter",
		zap.Int("resource spans", td.ResourceSpans().Len()),
		zap.Int("spans", td.SpanCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.tracesMarshaler.MarshalTraces(td)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func (s *loggingExporter) pushMetrics(_ context.Context, md pmetric.Metrics) error {
	s.logger.Info("MetricsExporter",
		zap.Int("resource metrics", md.ResourceMetrics().Len()),
		zap.Int("metrics", md.MetricCount()),
		zap.Int("data points", md.DataPointCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func (s *loggingExporter) pushLogs(_ context.Context, ld plog.Logs) error {
	s.logger.Info("LogsExporter",
		zap.Int("resource logs", ld.ResourceLogs().Len()),
		zap.Int("log records", ld.LogRecordCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func (s *loggingExporter) pushProfiles(_ context.Context, ld pprofile.Profiles) error {
	s.logger.Info("ProfilesExporter",
		zap.Int("resource profiles", ld.ResourceProfiles().Len()),
		zap.Int("profile records", ld.ProfileCount()))
	if s.verbosity != configtelemetry.LevelDetailed {
		return nil
	}

	buf, err := s.profilesMarshaler.MarshalProfiles(ld)
	if err != nil {
		return err
	}
	s.logger.Info(string(buf))
	return nil
}

func newLoggingExporter(logger *zap.Logger, verbosity configtelemetry.Level) *loggingExporter {
	return &loggingExporter{
		verbosity:         verbosity,
		logger:            logger,
		logsMarshaler:     otlptext.NewTextLogsMarshaler(),
		metricsMarshaler:  otlptext.NewTextMetricsMarshaler(),
		tracesMarshaler:   otlptext.NewTextTracesMarshaler(),
		profilesMarshaler: otlptext.NewTextProfilesMarshaler(),
	}
}
