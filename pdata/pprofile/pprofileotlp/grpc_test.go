// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pprofileotlp

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"go.opentelemetry.io/collector/pdata/pprofile"
)

func TestGrpc(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	RegisterGRPCServer(s, &fakeProfilesServer{t: t})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, s.Serve(lis))
	}()
	t.Cleanup(func() {
		s.Stop()
		wg.Wait()
	})

	cc, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, cc.Close())
	})

	profileClient := NewGRPCClient(cc)

	resp, err := profileClient.Export(context.Background(), generateProfilesRequest())
	assert.NoError(t, err)
	// TODO(@petethepig): fix this
	assert.Equal(t, NewExportResponse(), resp)
}

func TestGrpcError(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	RegisterGRPCServer(s, &fakeProfilesServer{t: t, err: errors.New("my error")})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, s.Serve(lis))
	}()
	t.Cleanup(func() {
		s.Stop()
		wg.Wait()
	})

	cc, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, cc.Close())
	})

	profileClient := NewGRPCClient(cc)
	resp, err := profileClient.Export(context.Background(), generateProfilesRequest())
	require.Error(t, err)
	st, okSt := status.FromError(err)
	require.True(t, okSt)
	assert.Equal(t, "my error", st.Message())
	assert.Equal(t, codes.Unknown, st.Code())
	assert.Equal(t, ExportResponse{}, resp)
}

type fakeProfilesServer struct {
	UnimplementedGRPCServer
	t   *testing.T
	err error
}

func (f fakeProfilesServer) Export(_ context.Context, request ExportRequest) (ExportResponse, error) {
	assert.Equal(f.t, generateProfilesRequest(), request)
	return NewExportResponse(), f.err
}

func generateProfilesRequest() ExportRequest {
	ld := pprofile.NewProfiles()
	ld.ResourceProfiles().AppendEmpty().ScopeProfiles().AppendEmpty().Profiles().AppendEmpty()
	return NewExportRequestFromProfiles(ld)
}
