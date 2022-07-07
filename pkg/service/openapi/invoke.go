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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	"github.com/tkeel-io/tkeel/pkg/client"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"net/http"
)

func (c *DaprClient) SchemaChangeAddons(ctx context.Context, tenantId string, objectId string, eventType EventType, templateData *pb.UpdateTemplateResponse) error {
	sendToPluginID := "keel"
	methodEndpoint := fmt.Sprintf("/apis/addons/%s", DEVICE_SCHEMA_CHANGE)
	body := map[string]interface{}{
		"objectId": objectId,
		"tenantId": tenantId,
	}

	optionTable := map[EventType]Option{
		EventTemplateDelete: func(m *map[string]interface{}) {
			(*m)["type"] = SchemaTemp
			(*m)["status"] = OpDelete
		},
		EventDeviceDelete: func(m *map[string]interface{}) {
			(*m)["type"] = SchemaDevice
			(*m)["status"] = OpDelete
		},
		EventTelemetryDelete: func(m *map[string]interface{}) {
			(*m)["type"] = SchemaTelemetry
			(*m)["status"] = OpDelete
		},
		EventTelemetryEdit: func(m *map[string]interface{}) {
			(*m)["type"] = SchemaTelemetry
			(*m)["status"] = OpEdit
		},
		EventUnknown: func(m *map[string]interface{}) {
			(*m)["type"] = SchemaUnknown
			(*m)["status"] = OpDefault
		},
	}
	if option, ok := optionTable[eventType]; ok {
		option(&body)
	} else {
		optionTable[EventUnknown](&body)
	}
	bodyVal, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.CallAddons(ctx, sendToPluginID, methodEndpoint, bodyVal, templateData)
}

func (c *DaprClient) CallAddons(ctx context.Context, sendToPluginID, methodEndpoint string, body []byte, templateData *pb.UpdateTemplateResponse) error {
	resp := &structpb.Value{}
	// return fmt.Sprintf(daprInvokeURLTemplate, c.httpAddr, req.ID, req.Method)
	byt, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     methodEndpoint,
		Verb:       http.MethodPost,
		Header:     transportHTTP.HeaderFromContext(ctx).Clone(),
		QueryValue: nil,
		Body:       body,
	}, nil, resp)
	if err != nil {
		log.Error(fmt.Sprintf("CallAddons:\nID:%v \nmethodEndpoint:%v \nbody:%v \nerr: %v\n",
			sendToPluginID,
			methodEndpoint,
			string(body),
      err)
		return errors.Wrapf(err, "dapr invoke plugin(%s) identify", sendToPluginID)
	}
	log.Info(fmt.Sprintf("CallAddons:\nID:%v \nmethodEndpoint:%v \nbyt:%v \nbody:%v\n",
		sendToPluginID,
		methodEndpoint,
		string(byt),
		string(body)))
	return nil
}
