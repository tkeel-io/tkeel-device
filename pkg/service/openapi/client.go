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
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/model"
)

var (
	DEVICE_SCHEMA_CHANGE = "device-schema-change"
	tokenKey             = "Authorization"
)

type Client interface {
	SchemaChangeAddons(ctx context.Context, resp *pb.UpdateTemplateResponse) error
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
	return &DaprClient{
		header: header,
		c:      dapr.NewHTTPClient(daprHTTPPort),
	}
}
