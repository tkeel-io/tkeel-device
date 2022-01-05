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

//const coreUrl string = "http://localhost:3500/v1.0/invoke/core/method/v1/entities"
//const authUrl string = "http://localhost:3500/v1.0/invoke/keel/method/apis/security"
const coreUrl string = "http://192.168.123.9:31438/v1/entities"
const authUrl string = "http://192.168.123.11:30707/apis/security"
const tokenKey string = "Authorization"

type CoreClient struct {
	//entity_create = "id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(**query)
	//url 	   string
	//id  	   string
	//entityType string
	//owner      string
	//source     string
}

func NewCoreClient() *CoreClient {
	return &CoreClient{}
}

// get core url
func (c *CoreClient) GetCoreUrl(midUrl string, mapUrl map[string]string, entityType string ) string {
	url := fmt.Sprintf(coreUrl+midUrl+"?"+"type=%s&owner=%s&source=%s", entityType, mapUrl["owner"], mapUrl["source"])
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
	if res["code"].(float64) != 200 {
		return nil, errors.New(res["msg"].(string))

	}
	tokenMap := res["data"].(map[string]interface{})

	// save token, map[entity_id:406c79543e0245a994a742e69ce48e71 entity_type:device tenant_id: token_id:de25624a-1d0a-4ab0-b1f1-5b0db5a12c30 user_id:abc]
	urlMap := map[string]string{
		"owner":     tokenMap["user_id"].(string),
		//"type":      "device",
		"source":    "device",
		"userToken": token,
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

func (c *CoreClient) CreatEntityToken(entityType, id, owner string, token string) (string, error) {
	//get url and request body
	log.Debug("CreateEntityToken")
	url := authUrl + fmt.Sprintf("/v1/entity/token")
	log.Debug("post auth url: ", url)
	tokenReq := map[string]interface{}{
		"entity_id":   id,
		"entity_type": entityType,
		"owner":       owner,
	}
	tr, err := json.Marshal(tokenReq)
	if nil != err {
		return "marshal error", err
	}

	//do it
	payload := strings.NewReader(string(tr))
	req, err1 := http.NewRequest("POST", url, payload)
	if nil != err1 {
		return "", err1
	}
	req.Header.Add(tokenKey, token)
	resp, er := http.DefaultClient.Do(req)
	body, err2 := c.ParseResp(resp, er)
	if nil != err2 {
		log.Error("error get device token, ", err)
		return "", err2
	}

	//Parse
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
	if res["code"].(float64) != 200 {
		return "error code", errors.New(res["msg"].(string))

	}
	tokenMap := res["data"].(map[string]interface{})
	entityToken, ok2 := tokenMap["token"].(string)
	if !ok2 {
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
