package service

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//const coreUrl string = "http://192.168.123.5:30226/v1/plugins/device/entities"
//const coreUrl string = "http://192.168.123.9:32246/v1/plugins/device/entities"
//const authUrl string = "http://192.168.123.5:30707" // /invoke/keel/method
const coreUrl string = "http://192.168.123.9:32701/v1/plugins/device/entities"
const authUrl string = "http://192.168.123.12:30707/apis/security"
const tokenKey string = "Authorization"

type CoreClient struct {
	//entity_create = "id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(**query)
	url 	   string
	id  	   string
	entityType string
	owner      string
	source     string
}

func NewCoreClient() *CoreClient {
	return &CoreClient{}
}

// get core url
func (c *CoreClient) GetCoreUrl(midUrl string, mapUrl map[string]string) string {
	url :=  fmt.Sprintf(coreUrl + midUrl + "?" +  "type=%s&owner=%s&source=%s", mapUrl["type"], mapUrl["owner"], mapUrl["source"])
	return url
}

//get token
func (c *CoreClient) GetTokenMap(ctx context.Context) (map[string]string, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	token, ok := header[tokenKey]
	if !ok {
		return nil, errors.New("invalid Authorization")
	}
	// only use the first one
	return c.parseToken(token[0])
}

func (c *CoreClient) parseToken(token string) (map[string]string, error) {
	url := authUrl + "/v1/oauth/authenticate"
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return nil, err
	}
	req.Header.Add(tokenKey, token)
	resp, err := http.DefaultClient.Do(req)
	resp2, err2 := c.ParseResp(resp, err)
	if nil != err2 {
		log.Error("error parse token, ", err)
		return nil, err2
	}
	var ar interface{}
	if err3 := json.Unmarshal(resp2, &ar); nil != err3 {
		log.Error("resp Unmarshal error", err3)
		return nil, err3
	}
	//log.Debug("Unmarshal res:", ar)
	res, ok := ar.(map[string]interface{})
	if !ok {
		return nil, errors.New("auth error")
	}
	if res["code"].(float64) != 0 {
		return nil, errors.New(res["msg"].(string))

	}
	tokenMap := res["data"].(map[string]interface{})

	// save token, map[entity_id:406c79543e0245a994a742e69ce48e71 entity_type:device tenant_id: token_id:de25624a-1d0a-4ab0-b1f1-5b0db5a12c30 user_id:abc]
	urlMap := map[string]string{
		"owner": 	  tokenMap["user_id"].(string),
		"type": 	  "device",
		"source":     "device",
	}
	return urlMap, nil
}

func (c *CoreClient) Post(url string, data []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Put(url string, data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PUT", url, payload)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) Patch(url string, data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PATCH", url, payload)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) Delete(url string) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", url, nil)
	//req.Header.Add("Authorization", "xxxx")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) ParseResp(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		log.Error("error ", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error ReadAll", err)
		return body, err
	}

	log.Debug("receive resp, ", string(body))
	if resp.StatusCode != 200 {
		log.Error("bad status ", resp.StatusCode)
		return body, errors.New(resp.Status)
	}
	return body, nil
}

func (c *CoreClient) CreatEntityToken(entityType,id,owner string)(string, error) {
	url:= authUrl + fmt.Sprintf("/v1/entity/%s/%s/token?owner=%s", entityType, id, owner)
	resp, err := http.Get(url)

	body, err2 := c.ParseResp(resp, err)
	if nil != err2{
		return "", err2
	}
	var ar interface{}
	if err3 := json.Unmarshal(body, &ar); nil != err3 {
		log.Error("resp Unmarshal error", err3)
		return "resp Unmarshal error", err3
	}
	//log.Debug("Unmarshal res:", ar)
	res, ok := ar.(map[string]interface{})
	if !ok {
		return "error resp type", errors.New("error resp type")
	}
	if res["code"].(float64) != 0 {
		return "error code", errors.New(res["msg"].(string))

	}
	tokenMap := res["data"].(map[string]interface{})
	entityToken, ok2 :=  tokenMap["token"].(string)
	if !ok2{
		return "error token", errors.New("error token")
	}
	return entityToken, nil
}

// generate uuid
func GetUUID() string {
	id := uuid.New()
	return id.String()
}

// get time
func GetTime() int64 {
	return time.Now().UnixNano()
}
