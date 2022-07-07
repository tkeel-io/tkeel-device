/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openapi

import (
	"context"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/client/dapr"
)

var (
	DEVICE_SCHEMA_CHANGE = "device-schema-change"
	tokenKey             = "Authorization"
)

type EventType int

// addons schema type
const (
	SchemaUnknown   = "unknown"
	SchemaTemp      = "temp"
	SchemaDevice    = "device"
	SchemaTelemetry = "telemetry"
)

// addons schema event type
const (
	EventUnknown = iota
	EventTemplateDelete
	EventDeviceDelete
	EventTelemetryDelete
	EventTelemetryEdit
)

// addons schema op
const (
	OpDefault = iota
	OpDelete
	OpEdit
)

type Option func(m *map[string]interface{})

type Client interface {
	SchemaChangeAddons(ctx context.Context, tenantId string, objectId string, eventType EventType, resp *pb.UpdateTemplateResponse) error
}

var _ Client = (*DaprClient)(nil)

type DaprClient struct {
	c      *dapr.HTTPClient
	header http.Header
}

func NewDaprClient(daprHTTPPort string, Tenant, User string) *DaprClient {
	user := new(model.User)
	user.Tenant = Tenant
	user.User = User
	header := http.Header{}
	header.Set(model.XtKeelAuthHeader, user.Base64Encode())
	return NewDaprClientWithConn(dapr.NewHTTPClient(daprHTTPPort), header)
}

func NewDaprClientWithConn(client *dapr.HTTPClient, header http.Header) *DaprClient {
	return &DaprClient{
		header: header,
		c:      client,
	}
}
