package service

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tkeel-io/kit/log"
	"io/ioutil"
	"net/http"
	"strings"
)

const CoreUrl string = "http://192.168.123.5:30707/v0.1.0/core/plugins/pluginA/entities?"

type CoreClient struct {
	//entity_create = "id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(**query)
	url string
}

func NewCoreClient() *CoreClient {
	var (
		id         = GetUUID()
		entityType = "device"
		owner      = "abc"
		source     = "source"
	)
	url := CoreUrl + fmt.Sprintf("id=%s&type=%s&owner=%s&source=%s", id, entityType, owner, source)
	return &CoreClient{
		url: url,
	}
}

func (c *CoreClient) Post(data []byte) ([]byte, error) {
	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(data))

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Get(id string) ([]byte, error) {
	log.Debug("fixme id", id)
	resp, err := http.Get(c.url)

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Put(data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PUT", c.url, payload)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) Delete(data []byte) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", c.url, nil)
	//req.Header.Add("Authorization", "xxxx")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) ParseResp(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		log.Error("error get", err, c.url)
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
		log.Error("bad status", resp.StatusCode)
		return body, errors.New(resp.Status)
	}
	return body, nil
}

// generate uuid
func GetUUID() string {
	id := uuid.New()
	return id.String()
}
