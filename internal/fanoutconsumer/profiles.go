// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package fanoutconsumer contains implementations of Traces/Metrics/Profiles consumers
// that fan out the data to multiple other consumers.
package fanoutconsumer // import "go.opentelemetry.io/collector/internal/fanoutconsumer"

import (
	"context"

	"go.uber.org/multierr"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

// NewProfiles wraps multiple log consumers in a single one.
// It fanouts the incoming data to all the consumers, and does smart routing:
//   - Clones only to the consumer that needs to mutate the data.
//   - If all consumers needs to mutate the data one will get the original mutable data.
func NewProfiles(lcs []consumer.Profiles) consumer.Profiles {
	// Don't wrap if there is only one non-mutating consumer.
	if len(lcs) == 1 && !lcs[0].Capabilities().MutatesData {
		return lcs[0]
	}

	lc := &profilesConsumer{}
	for i := 0; i < len(lcs); i++ {
		if lcs[i].Capabilities().MutatesData {
			lc.mutable = append(lc.mutable, lcs[i])
		} else {
			lc.readonly = append(lc.readonly, lcs[i])
		}
	}
	return lc
}

type profilesConsumer struct {
	mutable  []consumer.Profiles
	readonly []consumer.Profiles
}

func (lsc *profilesConsumer) Capabilities() consumer.Capabilities {
	// If all consumers are mutating, then the original data will be passed to one of them.
	return consumer.Capabilities{MutatesData: len(lsc.mutable) > 0 && len(lsc.readonly) == 0}
}

// ConsumeProfiles exports the pprofile.Profiles to all consumers wrapped by the current one.
func (lsc *profilesConsumer) ConsumeProfiles(ctx context.Context, pd pprofile.Profiles) error {
	var errs error

	if len(lsc.mutable) > 0 {
		// Clone the data before sending to all mutating consumers except the last one.
		for i := 0; i < len(lsc.mutable)-1; i++ {
			errs = multierr.Append(errs, lsc.mutable[i].ConsumeProfiles(ctx, cloneProfiles(pd)))
		}
		// Send data as is to the last mutating consumer only if there are no other non-mutating consumers and the
		// data is mutable. Never share the same data between a mutating and a non-mutating consumer since the
		// non-mutating consumer may process data async and the mutating consumer may change the data before that.
		lastConsumer := lsc.mutable[len(lsc.mutable)-1]
		if len(lsc.readonly) == 0 && !pd.IsReadOnly() {
			errs = multierr.Append(errs, lastConsumer.ConsumeProfiles(ctx, pd))
		} else {
			errs = multierr.Append(errs, lastConsumer.ConsumeProfiles(ctx, cloneProfiles(pd)))
		}
	}

	// Mark the data as read-only if it will be sent to more than one read-only consumer.
	if len(lsc.readonly) > 1 && !pd.IsReadOnly() {
		pd.MarkReadOnly()
	}
	for _, lc := range lsc.readonly {
		errs = multierr.Append(errs, lc.ConsumeProfiles(ctx, pd))
	}

	return errs
}

func cloneProfiles(pd pprofile.Profiles) pprofile.Profiles {
	clonedProfiles := pprofile.NewProfiles()
	pd.CopyTo(clonedProfiles)
	return clonedProfiles
}
