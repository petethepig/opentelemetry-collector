// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pprofile // import "go.opentelemetry.io/collector/pdata/pprofile"

import (
	"go.opentelemetry.io/collector/pdata/internal"
	otlpcollectorprofile "go.opentelemetry.io/collector/pdata/internal/data/protogen/collector/profiles/v1experimental"
)

// Profiles is the top-level struct that is propagated through the profiles pipeline.
// Use NewProfiles to create new instance, zero-initialized instance is not valid for use.
type Profiles internal.Profiles

func newProfiles(orig *otlpcollectorprofile.ExportProfilesServiceRequest) Profiles {
	state := internal.StateMutable
	return Profiles(internal.NewProfiles(orig, &state))
}

func (ms Profiles) getOrig() *otlpcollectorprofile.ExportProfilesServiceRequest {
	return internal.GetOrigProfiles(internal.Profiles(ms))
}

func (ms Profiles) getState() *internal.State {
	return internal.GetProfilesState(internal.Profiles(ms))
}

// NewProfiles creates a new Profiles struct.
func NewProfiles() Profiles {
	return newProfiles(&otlpcollectorprofile.ExportProfilesServiceRequest{})
}

// IsReadOnly returns true if this Logs instance is read-only.
func (ms Profiles) IsReadOnly() bool {
	return *ms.getState() == internal.StateReadOnly
}

// CopyTo copies the Profiles instance overriding the destination.
func (ms Profiles) CopyTo(dest Profiles) {
	ms.ResourceProfiles().CopyTo(dest.ResourceProfiles())
}

// ProfileCount calculates the total number of profile records.
func (ms Profiles) ProfileCount() int {
	profileCount := 0
	rss := ms.ResourceProfiles()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		ill := rs.ScopeProfiles()
		for i := 0; i < ill.Len(); i++ {
			profiles := ill.At(i)
			profileCount += profiles.Profiles().Len()
		}
	}
	return profileCount
}

// ResourceProfiles returns the ResourceProfilesSlice associated with this Profiles.
func (ms Profiles) ResourceProfiles() ResourceProfilesSlice {
	return newResourceProfilesSlice(&ms.getOrig().ResourceProfiles, internal.GetProfilesState(internal.Profiles(ms)))
}

// MarkReadOnly marks the Logs as shared so that no further modifications can be done on it.
func (ms Profiles) MarkReadOnly() {
	internal.SetProfilesState(internal.Profiles(ms), internal.StateReadOnly)
}
