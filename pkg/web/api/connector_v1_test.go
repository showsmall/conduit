// Copyright © 2022 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/conduitio/conduit/pkg/connector"
	connmock "github.com/conduitio/conduit/pkg/connector/mock"
	"github.com/conduitio/conduit/pkg/foundation/assert"
	"github.com/conduitio/conduit/pkg/foundation/cchan"
	"github.com/conduitio/conduit/pkg/foundation/cerrors"
	"github.com/conduitio/conduit/pkg/foundation/log"
	"github.com/conduitio/conduit/pkg/inspector"
	"github.com/conduitio/conduit/pkg/record"
	apimock "github.com/conduitio/conduit/pkg/web/api/mock"
	"github.com/conduitio/conduit/pkg/web/api/toproto"
	apiv1 "github.com/conduitio/conduit/proto/gen/api/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectorAPIv1_ListConnectors(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	connBuilder := connmock.Builder{Ctrl: ctrl}
	source := newTestSource(connBuilder)
	destination := newTestDestination(connBuilder)
	destination.EXPECT().State().Return(connector.DestinationState{Positions: map[string]record.Position{source.ID(): []byte("irrelevant")}})

	csMock.EXPECT().
		List(ctx).
		Return(map[string]connector.Connector{source.ID(): source, destination.ID(): destination}).
		Times(1)

	now := time.Now()
	want := &apiv1.ListConnectorsResponse{
		Connectors: []*apiv1.Connector{
			{
				Id: source.ID(),
				State: &apiv1.Connector_SourceState_{
					SourceState: &apiv1.Connector_SourceState{Position: []byte("irrelevant")},
				},
				Config: &apiv1.Connector_Config{
					Name:     source.Config().Name,
					Settings: source.Config().Settings,
				},
				Type:         apiv1.Connector_Type(source.Type()),
				Plugin:       source.Config().Plugin,
				PipelineId:   source.Config().PipelineID,
				ProcessorIds: source.Config().ProcessorIDs,
				CreatedAt:    timestamppb.New(now),
				UpdatedAt:    timestamppb.New(now),
			},

			{
				Id: destination.ID(),
				State: &apiv1.Connector_DestinationState_{
					DestinationState: &apiv1.Connector_DestinationState{
						Positions: map[string][]byte{source.ID(): []byte("irrelevant")},
					},
				},
				Config: &apiv1.Connector_Config{
					Name:     destination.Config().Name,
					Settings: destination.Config().Settings,
				},
				Type:         apiv1.Connector_Type(destination.Type()),
				Plugin:       destination.Config().Plugin,
				PipelineId:   destination.Config().PipelineID,
				ProcessorIds: destination.Config().ProcessorIDs,
				CreatedAt:    timestamppb.New(now),
				UpdatedAt:    timestamppb.New(now),
			},
		},
	}

	got, err := api.ListConnectors(
		ctx,
		&apiv1.ListConnectorsRequest{},
	)
	assert.Ok(t, err)

	// copy expected times
	for i, conn := range got.Connectors {
		conn.CreatedAt = want.Connectors[i].CreatedAt
		conn.UpdatedAt = want.Connectors[i].UpdatedAt
	}

	sortConnectors(want.Connectors)
	sortConnectors(got.Connectors)
	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_ListConnectorsByPipeline(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	connBuilder := connmock.Builder{Ctrl: ctrl}
	source := newTestSource(connBuilder)
	destination := newTestDestination(connBuilder)

	csMock.EXPECT().
		List(ctx).
		Return(map[string]connector.Connector{source.ID(): source, destination.ID(): destination}).
		Times(1)

	now := time.Now()
	want := &apiv1.ListConnectorsResponse{
		Connectors: []*apiv1.Connector{
			{
				Id: source.ID(),
				State: &apiv1.Connector_SourceState_{
					SourceState: &apiv1.Connector_SourceState{Position: []byte("irrelevant")},
				},
				Config: &apiv1.Connector_Config{
					Name:     source.Config().Name,
					Settings: source.Config().Settings,
				},
				Type:         apiv1.Connector_Type(source.Type()),
				Plugin:       source.Config().Plugin,
				PipelineId:   source.Config().PipelineID,
				ProcessorIds: source.Config().ProcessorIDs,
				CreatedAt:    timestamppb.New(now),
				UpdatedAt:    timestamppb.New(now),
			},
		},
	}

	got, err := api.ListConnectors(
		ctx,
		&apiv1.ListConnectorsRequest{PipelineId: source.Config().PipelineID},
	)
	assert.Ok(t, err)

	// copy expected time
	for i, conn := range got.Connectors {
		conn.CreatedAt = want.Connectors[i].CreatedAt
		conn.UpdatedAt = want.Connectors[i].UpdatedAt
	}

	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_CreateConnector(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	connBuilder := connmock.Builder{Ctrl: ctrl}
	source := newTestSource(connBuilder)

	csMock.EXPECT().Create(ctx, source.Type(), source.Config()).Return(source, nil).Times(1)

	now := time.Now()
	want := &apiv1.CreateConnectorResponse{Connector: &apiv1.Connector{
		Id: source.ID(),
		State: &apiv1.Connector_SourceState_{
			SourceState: &apiv1.Connector_SourceState{Position: []byte("irrelevant")},
		},
		Config: &apiv1.Connector_Config{
			Name:     source.Config().Name,
			Settings: source.Config().Settings,
		},
		Type:         apiv1.Connector_Type(source.Type()),
		Plugin:       source.Config().Plugin,
		PipelineId:   source.Config().PipelineID,
		ProcessorIds: source.Config().ProcessorIDs,
		UpdatedAt:    timestamppb.New(now),
		CreatedAt:    timestamppb.New(now),
	}}

	got, err := api.CreateConnector(
		ctx,
		&apiv1.CreateConnectorRequest{
			Type:       want.Connector.Type,
			Plugin:     want.Connector.Plugin,
			PipelineId: want.Connector.PipelineId,
			Config:     want.Connector.Config,
		},
	)

	assert.Ok(t, err)

	// copy expected time
	got.Connector.CreatedAt = want.Connector.CreatedAt
	got.Connector.UpdatedAt = want.Connector.UpdatedAt

	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_InspectConnector_SendRecord(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)

	id := uuid.NewString()
	rec := generateTestRecord()
	recProto, err := toproto.Record(rec)
	assert.Ok(t, err)

	ins := inspector.New(log.Nop(), 10)
	session := ins.NewSession(ctx)

	csMock.EXPECT().
		Inspect(ctx, id).
		Return(session, nil).
		Times(1)

	inspectServer := apimock.NewConnectorService_InspectConnectorServer(ctrl)
	inspectServer.EXPECT().Send(gomock.Eq(&apiv1.InspectConnectorResponse{Record: recProto}))
	inspectServer.EXPECT().Context().Return(ctx).AnyTimes()

	go func() {
		_ = api.InspectConnector(
			&apiv1.InspectConnectorRequest{Id: id},
			inspectServer,
		)
	}()
	ins.Send(ctx, rec)

	time.Sleep(100 * time.Millisecond)
}

func TestConnectorAPIv1_InspectConnector_SendErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	id := uuid.NewString()

	ins := inspector.New(log.Nop(), 10)
	session := ins.NewSession(ctx)

	csMock.EXPECT().
		Inspect(ctx, id).
		Return(session, nil).
		Times(1)

	inspectServer := apimock.NewConnectorService_InspectConnectorServer(ctrl)
	inspectServer.EXPECT().Context().Return(ctx).AnyTimes()
	errSend := cerrors.New("I'm sorry, but no.")
	inspectServer.EXPECT().Send(gomock.Any()).Return(errSend)

	errC := make(chan error)
	go func() {
		err := api.InspectConnector(
			&apiv1.InspectConnectorRequest{Id: id},
			inspectServer,
		)
		errC <- err
	}()
	ins.Send(ctx, generateTestRecord())

	err, b, err2 := cchan.Chan[error](errC).RecvTimeout(context.Background(), 100*time.Millisecond)
	assert.Ok(t, err2)
	assert.True(t, b, "expected to receive an error")
	assert.True(t, cerrors.Is(err, errSend), "expected 'I'm sorry, but no.' error")
}

func TestConnectorAPIv1_InspectConnector_Err(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	id := uuid.NewString()
	err := cerrors.New("not found, sorry")

	csMock.EXPECT().
		Inspect(ctx, gomock.Any()).
		Return(nil, err).
		Times(1)

	inspectServer := apimock.NewConnectorService_InspectConnectorServer(ctrl)
	inspectServer.EXPECT().Context().Return(ctx).AnyTimes()

	errAPI := api.InspectConnector(
		&apiv1.InspectConnectorRequest{Id: id},
		inspectServer,
	)
	assert.NotNil(t, errAPI)
	assert.Equal(
		t,
		"rpc error: code = Internal desc = failed to get connector: not found, sorry",
		errAPI.Error(),
	)
}

func generateTestRecord() record.Record {
	return record.Record{
		Position:  record.Position("test-position"),
		Operation: record.OperationCreate,
		Metadata:  record.Metadata{"metadata-key": "metadata-value"},
		Key:       record.RawData{Raw: []byte("test-key")},
		Payload: record.Change{
			After: record.RawData{Raw: []byte("test-payload")},
		},
	}
}

func TestConnectorAPIv1_GetConnector(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	connBuilder := connmock.Builder{Ctrl: ctrl}
	source := newTestSource(connBuilder)

	csMock.EXPECT().Get(ctx, source.ID()).Return(source, nil).Times(1)

	now := time.Now()
	want := &apiv1.GetConnectorResponse{Connector: &apiv1.Connector{
		Id: source.ID(),
		State: &apiv1.Connector_SourceState_{
			SourceState: &apiv1.Connector_SourceState{Position: []byte("irrelevant")},
		},
		Config: &apiv1.Connector_Config{
			Name:     source.Config().Name,
			Settings: source.Config().Settings,
		},
		Type:         apiv1.Connector_Type(source.Type()),
		Plugin:       source.Config().Plugin,
		PipelineId:   source.Config().PipelineID,
		ProcessorIds: source.Config().ProcessorIDs,
		UpdatedAt:    timestamppb.New(now),
		CreatedAt:    timestamppb.New(now),
	}}

	got, err := api.GetConnector(
		ctx,
		&apiv1.GetConnectorRequest{
			Id: want.Connector.Id,
		},
	)

	assert.Ok(t, err)

	// copy expected time
	got.Connector.CreatedAt = want.Connector.CreatedAt
	got.Connector.UpdatedAt = want.Connector.UpdatedAt

	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_UpdateConnector(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)
	connBuilder := connmock.Builder{Ctrl: ctrl}
	oldConfig := connector.Config{
		Name:       "Old name",
		Settings:   map[string]string{"path": "old/path"},
		Plugin:     "path/to/plugin",
		PipelineID: uuid.NewString(),
	}
	newConfig := connector.Config{
		Name:       "A source connector",
		Settings:   map[string]string{"path": "path/to"},
		Plugin:     "path/to/plugin",
		PipelineID: oldConfig.PipelineID,
	}

	before := connBuilder.NewSourceMock(uuid.NewString(), oldConfig)
	after := connBuilder.NewSourceMock(before.ID(), newConfig)
	after.EXPECT().State().Return(connector.SourceState{Position: []byte("irrelevant")})

	csMock.EXPECT().Get(ctx, before.ID()).Return(before, nil).Times(1)
	csMock.EXPECT().Update(ctx, before.ID(), newConfig).Return(after, nil).Times(1)

	now := time.Now()
	want := &apiv1.UpdateConnectorResponse{Connector: &apiv1.Connector{
		Id: after.ID(),
		State: &apiv1.Connector_SourceState_{
			SourceState: &apiv1.Connector_SourceState{Position: []byte("irrelevant")},
		},
		Config: &apiv1.Connector_Config{
			Name:     after.Config().Name,
			Settings: after.Config().Settings,
		},
		Type:         apiv1.Connector_Type(after.Type()),
		Plugin:       after.Config().Plugin,
		PipelineId:   after.Config().PipelineID,
		ProcessorIds: after.Config().ProcessorIDs,
		CreatedAt:    timestamppb.New(now),
		UpdatedAt:    timestamppb.New(now),
	}}

	got, err := api.UpdateConnector(
		ctx,
		&apiv1.UpdateConnectorRequest{
			Id:     want.Connector.Id,
			Config: want.Connector.Config,
		},
	)
	assert.Ok(t, err)

	// copy expected time
	got.Connector.CreatedAt = want.Connector.CreatedAt
	got.Connector.UpdatedAt = want.Connector.UpdatedAt

	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_DeleteConnector(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)

	id := uuid.NewString()

	csMock.EXPECT().Delete(ctx, id).Return(nil).Times(1)

	want := &apiv1.DeleteConnectorResponse{}

	got, err := api.DeleteConnector(
		ctx,
		&apiv1.DeleteConnectorRequest{
			Id: id,
		},
	)

	assert.Ok(t, err)
	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_ValidateConnector(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)

	config := connector.Config{
		Name:     "A source connector",
		Settings: map[string]string{"path": "path/to"},
		Plugin:   "builtin:file",
	}
	ctype := connector.TypeSource

	csMock.EXPECT().Validate(ctx, ctype, config).Return(nil).Times(1)

	want := &apiv1.ValidateConnectorResponse{}

	got, err := api.ValidateConnector(
		ctx,
		&apiv1.ValidateConnectorRequest{
			Type:   apiv1.Connector_Type(ctype),
			Plugin: config.Plugin,
			Config: &apiv1.Connector_Config{
				Name:     config.Name,
				Settings: config.Settings,
			},
		},
	)

	assert.Ok(t, err)
	assert.Equal(t, want, got)
}

func TestConnectorAPIv1_ValidateConnectorError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	csMock := apimock.NewConnectorOrchestrator(ctrl)
	api := NewConnectorAPIv1(csMock)

	config := connector.Config{
		Name:     "A source connector",
		Settings: map[string]string{"path": "path/to"},
		Plugin:   "builtin:file",
	}
	ctype := connector.TypeSource
	err := cerrors.New("validation error")

	csMock.EXPECT().Validate(ctx, ctype, config).Return(err).Times(1)

	_, err = api.ValidateConnector(
		ctx,
		&apiv1.ValidateConnectorRequest{
			Type:   apiv1.Connector_Type(ctype),
			Plugin: config.Plugin,
			Config: &apiv1.Connector_Config{
				Name:     config.Name,
				Settings: config.Settings,
			},
		},
	)

	assert.Error(t, err)
}

func sortConnectors(c []*apiv1.Connector) {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Id < c[j].Id
	})
}

func newTestSource(connBuilder connmock.Builder) *connmock.Source {
	source := connBuilder.NewSourceMock(uuid.NewString(), connector.Config{
		Name:       "A source connector",
		Settings:   map[string]string{"path": "path/to"},
		Plugin:     "path/to/plugin",
		PipelineID: uuid.NewString(),
	})
	source.EXPECT().State().Return(connector.SourceState{Position: []byte("irrelevant")})
	source.EXPECT().CreatedAt().AnyTimes()
	source.EXPECT().UpdatedAt().AnyTimes()
	return source
}

func newTestDestination(connBuilder connmock.Builder) *connmock.Destination {
	destination := connBuilder.NewDestinationMock(uuid.NewString(), connector.Config{
		Name:       "A destination connector",
		Settings:   map[string]string{"path": "path/to"},
		Plugin:     "path/to/plugin",
		PipelineID: uuid.NewString(),
	})
	destination.EXPECT().CreatedAt().AnyTimes()
	destination.EXPECT().UpdatedAt().AnyTimes()
	return destination
}
