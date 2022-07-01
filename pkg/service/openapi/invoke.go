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
	"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tkeel-io/tkeel/pkg/client"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
)

func (c *DaprClient) SchemaChangeAddons(ctx context.Context, templateData *pb.UpdateTemplateResponse) error {
	sendToPluginID := "keel"
	methodEndpoint := fmt.Sprintf("apis/addons/%s", DEVICE_SCHEMA_CHANGE)
	return c.CallAddons(ctx, sendToPluginID, methodEndpoint, templateData)
}

func (c *DaprClient) CallAddons(ctx context.Context, sendToPluginID, methodEndpoint string, templateData *pb.UpdateTemplateResponse) error {
	resp := &structpb.Value{}
	// return fmt.Sprintf(daprInvokeURLTemplate, c.httpAddr, req.ID, req.Method)
	byt, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     methodEndpoint,
		Verb:       http.MethodPost,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, templateData, resp)
	if err != nil {
		return errors.Wrapf(err, "dapr invoke plugin(%s) identify", sendToPluginID)
	}
	log.Info(fmt.Sprintf("CallAddons: ID:%v\n methodEndpoint:%v\n templateData:%v\n byt:%v\n",
		sendToPluginID,
		methodEndpoint,
		templateData,
		string(byt)))
	return nil
}
