package service

import (
	//"bytes"
	"context"
	"encoding/base64"
	json "encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/tdtl"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"

	//"github.com/tkeel-io/tdtl"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	pbg "github.com/tkeel-io/tkeel-device/api/group/v1"
	"github.com/tkeel-io/tkeel-device/pkg/service/openapi"
	pb_auth "github.com/tkeel-io/tkeel/api/authentication/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	coreUrl string = "http://localhost:3500/v1.0/invoke/keel/method/apis/core/v1/entities"
	authUrl string = "http://localhost:3500/v1.0/invoke/keel/method/apis/security"
	subUrl  string = "http://localhost:3500/v1.0/invoke/keel/method/apis/core-broker/v1/entities/%s"
	ruleUrl string = "http://localhost:3500/v1.0/invoke/keel/method/apis/rule-manager/v1/devices/%s"

	//coreUrl string = "http://192.168.100.5:31874/v1/entities"
	//authUrl string = "http://192.168.100.5:30707/apis/security"
	//coreUrl string = "http://192.168.123.9:30535/v1/entities"
	//authUrl string = "http://192.168.123.9:30707/apis/security"

	tokenKey string = "Authorization"

	// default header key
	tkeelAuthHeader = `x-tKeel-auth`
	defaultTenant   = `_tKeel_system`
	defaultUser     = `_tKeel_admin`
	defultRole      = `admin`
)

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
func (c *CoreClient) GetCoreUrl(midUrl string, mapUrl map[string]string, entityType string) string {
	url := fmt.Sprintf(coreUrl+midUrl+"?"+"type=%s&owner=%s&source=%s", entityType, mapUrl["owner"], mapUrl["source"])
	return url
}

// get subscribe delete entity url
func (c *CoreClient) GetDeleleEntityFromSubUrl(entityId string) string {
	return fmt.Sprintf(subUrl, entityId)
}

// get rule delete entity url
func (c *CoreClient) GetDeleleEntityFromRuleUrl(entityId string) string {
	return fmt.Sprintf(ruleUrl, entityId)
}

//get token
func (c *CoreClient) GetTokenMap(ctx context.Context) (map[string]string, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	token, ok := header[tokenKey]
	if !ok && len(token) < 1 {
		return nil, errors.New("invalid authorization")
	}
	if token[0] == "" {
		return nil, errors.New("empty authorization token")
	}
	userToken := token[0]

	// only use the first one
	url := authUrl + "/v1/oauth/authenticate"
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return nil, err
	}
	req.Header.Add(tokenKey, userToken)
	resp, err := http.DefaultClient.Do(req)

	resp2, err2 := c.ParseResp(resp, err)
	if nil != err2 {
		log.Error("error parse token, ", err)
		return nil, err2
	}

	//cc := tdtl.New(resp)
	//owner := cc.Get("user_id").String()
	//device := cc.Get("device").String()

	// save token, map[entity_id:406c79543e0245a994a742e69ce48e71 entity_type:device tenant_id: token_id:de25624a-1d0a-4ab0-b1f1-5b0db5a12c30 user_id:abc]
	tokenMap, ok := resp2.(map[string]interface{})
	if !ok {
		return nil, errors.New("auth trans error")
	}

	urlMap := map[string]string{
		"owner":     tokenMap["user_id"].(string),
		"source":    "device",
		"userToken": userToken,
		"tenantId":  tokenMap["tenant_id"].(string),
	}
	return urlMap, nil
}

//get token
func (c *CoreClient) GetUser(ctx context.Context) (*pb_auth.AuthenticateResponse, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	token, ok := header[tokenKey]
	if !ok && len(token) < 1 {
		return nil, errors.New("invalid authorization")
	}
	if token[0] == "" {
		return nil, errors.New("empty authorization token")
	}
	userToken := token[0]

	// only use the first one
	url := authUrl + "/v1/oauth/authenticate"
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return nil, err
	}
	req.Header.Add(tokenKey, userToken)
	resp, err := http.DefaultClient.Do(req)

	cc := tdtl.New(resp)
	user_id := cc.Get("user_id").String()
	tenant_id := cc.Get("tenant_id").String()
	role := cc.Get("role").String()
	destination := cc.Get("destination").String()
	method := cc.Get("method").String()

	// save token, map[entity_id:406c79543e0245a994a742e69ce48e71 entity_type:device tenant_id: token_id:de25624a-1d0a-4ab0-b1f1-5b0db5a12c30 user_id:abc]
	urlMap := &pb_auth.AuthenticateResponse{
		UserId:      user_id,
		TenantId:    tenant_id,
		Role:        role,
		Method:      method,
		Destination: destination,
	}
	return urlMap, nil
}

func (c *CoreClient) Post(url string, data []byte) ([]byte, error) {
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	AddDefaultAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) Get(url string) ([]byte, error) {
	//resp, err := http.Get(url)
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	AddDefaultAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) Put(url string, data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PUT", url, payload)

	req.Header.Add("Content-Type", "application/json")
	AddDefaultAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) Patch(url string, data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PATCH", url, payload)

	req.Header.Add("Content-Type", "application/json")
	AddDefaultAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) Delete(url string) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Content-Type", "application/json")
	AddDefaultAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) DeleteWithCtx(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", url, nil)
	header := transportHTTP.HeaderFromContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(tkeelAuthHeader, header.Get(tkeelAuthHeader))

	resp, err := http.DefaultClient.Do(req)
	//return c.ParseResp(resp, err)
	dt, err1 := c.ParseResp(resp, err)
	if err1 != nil {
		return nil, err1
	}
	dataByte, err2 := json.Marshal(dt)
	if nil != err2 {
		return nil, err2
	}
	return dataByte, nil
}

func (c *CoreClient) ParseResp(resp *http.Response, err error) (interface{}, error) {
	if err != nil {
		log.Error("error ", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error ReadAll", err)
		return nil, err
	}

	//log.Debug("receive resp, ", string(body))
	if resp.StatusCode != 200 {
		log.Error("bad status ", resp.StatusCode)
		return nil, errors.New(resp.Status)
	}

	//Parse
	var ar interface{}
	if err3 := json.Unmarshal(body, &ar); nil != err3 {
		log.Error("resp Unmarshal error", err3)
		return nil, err3
	}
	//log.Debug("Unmarshal res:", ar)
	res, ok := ar.(map[string]interface{})
	if !ok {
		return nil, errors.New("error resp type")
	}
	if res["code"].(string) != "io.tkeel.SUCCESS" {
		return "error code", errors.New(res["msg"].(string))
	}
	data, ok := res["data"]
	if !ok {
		return nil, errors.New("error return data")
	}
	return data, nil
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
	log.Debug("data= ", tokenReq)

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
	AddDefaultAuthHeader(req)
	resp, er := http.DefaultClient.Do(req)
	//body, err2 := c.ParseResp(resp, er)
	res, err2 := c.ParseResp(resp, er)
	if nil != err2 {
		log.Error("error return ", err2)
		return "", err2
	}

	//Parse
	/*var ar interface{}
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

	}*/
	//tokenMap := res["data"].(map[string]interface{})
	tokenMap, ok := res.(map[string]interface{})
	if !ok {
		return "", errors.New("error resp trans")
	}
	entityToken, ok2 := tokenMap["token"].(string)
	if !ok2 {
		return "", errors.New("error token")
	}
	return entityToken, nil
}

func AddDefaultAuthHeader(req *http.Request) {
	authString := fmt.Sprintf("tenant=%s&user=%s&role=%s", defaultTenant, defaultUser, defultRole)
	req.Header.Add(tkeelAuthHeader, base64.StdEncoding.EncodeToString([]byte(authString)))
}

// generate uuid
func GetUUID() string {
	id := uuid.New()
	return "iotd-" + id.String()
}

// get time
func GetTime() int64 {
	return time.Now().UnixNano() / 1e6
}

func (c *CoreClient) setMapper(tm map[string]string, mapperName string, curId string, curProp string, targetId string, targetProp string) error {
	log.Debug("setMapper")

	//get url
	midUrl := "/" + curId + "/mappers"
	url := c.GetCoreUrl(midUrl, tm, "group")
	log.Debug("mapper url = ", url)

	//fmt request
	data := make(map[string]string)
	data["name"] = mapperName
	data["tql"] = "insert into " + curId + " select " + targetId + "." + targetProp + "+''  as " + curProp
	log.Debug("data = ", data)

	send, err := json.Marshal(data)
	if nil != err {
		return err
	}

	// do it
	_, err1 := c.Post(url, send)
	if nil != err1 {
		log.Error("error core return")
		return err1
	}

	return nil
}

func (c *CoreClient) setSpacePathMapper(tm map[string]string, Id string, pId string, entityType string) error {

	log.Debug("setSpacePathMapper")
	parentId := pId
	//check ParentId
	log.Debug("parentId = ", parentId)
	if (parentId == "") && entityType == "group" {
		return nil
	}

	//if (parentId == "") && entityType == "device" {
	defaultGroupId := "iotd-" + tm["owner"] + "-defaultGroup"
	if parentId == defaultGroupId {
		log.Debug("check default group")
		exist := c.checkEntityExist(tm, entityType, defaultGroupId)
		if !exist {
			err := c.CreateDevDefaultGroup(tm, defaultGroupId)
			if nil != err {
				return err
			}
		}
		parentId = defaultGroupId
	}
	if Id == parentId {
		return errors.New("error:  parent Id cannot be the same as the Id")
	}

	//get url
	midUrl := "/" + Id + "/mappers"
	url := c.GetCoreUrl(midUrl, tm, "group")
	log.Debug("mapper url = ", url)

	//fmt request
	data := make(map[string]string)
	data["name"] = "mapper_space_path"
	data["tql"] = "insert into " + Id + " select " + parentId + ".sysField._spacePath + '/" + Id + "'  as " + "sysField._spacePath, " + parentId + ".group.name " + "  as " + "basicInfo.parentName"
	log.Debug("data = ", data)

	send, err := json.Marshal(data)
	if nil != err {
		return err
	}

	// do it
	_, err1 := c.Post(url, send)
	if nil != err1 {
		log.Error("error core return")
		return err1
	}

	return nil
}
func (c *CoreClient) setTemplateNameMapper(tm map[string]string, Id string, tId string, entityType string) error {

	log.Debug("setTemplateNameMapper")
	//check templateId
	log.Debug("tId = ", tId)
	if tId == "" {
		return nil
	}

	//get url
	midUrl := "/" + Id + "/mappers"
	url := c.GetCoreUrl(midUrl, tm, "device")
	log.Debug("mapper url = ", url)

	//fmt request
	data := make(map[string]string)
	data["name"] = "mapper_templateName_path"
	data["tql"] = "insert into " + Id + " select " + tId + ".basicInfo.name as basicInfo.templateName"
	log.Debug("data = ", data)

	send, err := json.Marshal(data)
	if nil != err {
		return err
	}

	// do it
	_, err1 := c.Post(url, send)
	if nil != err1 {
		log.Error("error core return")
		return err1
	}

	return nil
}

func (c *CoreClient) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string, pathClassify string) (*emptypb.Empty, error) {
	log.Debug("CoreConfigPatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	tm, err := c.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + pathClassify
	url := c.GetCoreUrl(midUrl, tm, "template")
	log.Debug("patch url :", url)

	//fmt request
	var patchArray []CorePatch

	for k, v := range kv {
		ph := CorePatch{
			Path:     path + k,
			Operator: operator,
			Value:    v,
		}
		patchArray = append(patchArray, ph)
	}
	log.Debug("patch Array :", patchArray)

	data, err3 := json.Marshal(patchArray)
	if nil != err3 {
		return nil, err3
	}

	// do it
	_, err4 := c.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}

	//fmt response
	return &emptypb.Empty{}, nil
}

func (c *CoreClient) checkEntityExist(tm map[string]string, entityType string, id string) bool {

	midUrl := "/" + id
	url := c.GetCoreUrl(midUrl, tm, entityType)
	log.Debug("get url :", url)

	_, err2 := c.Get(url)
	if nil != err2 {
		return false
	}
	return true
	//return
	/*deviceObject := make(map[string]interface{})
	if err3 := json.Unmarshal(res, &deviceObject); nil != err3 {
		return nil, err3
	}*/
}
func (c *CoreClient) GetCoreEntitySpecContent(tm map[string]string, entityId string, entityType string, classify string, pids string) (map[string]interface{}, error) {
	midUrl := "/" + entityId + "/" + classify
	url := c.GetCoreUrl(midUrl, tm, entityType) + fmt.Sprintf("&property_keys=%s", pids)
	//url := c.GetCoreUrl(midUrl, tm, entityType) + fmt.Sprintf("&pids=%s", pids)
	log.Debug("url :", url)

	res, err1 := c.Get(url)
	if nil != err1 {
		log.Error("error post data to core", err1)
		return nil, err1
	}
	xObject := make(map[string]interface{})
	err2 := json.Unmarshal(res, &xObject)
	if err2 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err2
	}
	return xObject, nil
}

func (c *CoreClient) CreateDevDefaultGroup(tm map[string]string, id string) error {
	url := c.GetCoreUrl("", tm, "group") + "&id=" + id
	log.Debug("core url: ", url)

	//fmt request
	sysField := &pbg.GroupEntitySysField{
		XId:        id,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
		XOwner:     tm["owner"],
		XSource:    tm["source"],
		XSpacePath: id,
	}
	groupInfo := &pbg.GroupEntity{
		Name:        "默认分组",
		Description: "系统默认创建",
		ParentId:    "",
		ParentName:  "",
	}
	entityInfo := &pbg.GroupEntityCoreInfo{
		Group:    groupInfo,
		SysField: sysField,
	}

	log.Debug("entityinfo : ", entityInfo)
	data, err3 := json.Marshal(entityInfo)
	if nil != err3 {
		return err3
	}

	// do it
	res, err4 := c.Post(url, data)
	if nil != err4 {
		log.Error("error post return", err4)
		return err4
	}

	//fmt response
	groupObject := make(map[string]interface{})
	err5 := json.Unmarshal(res, &groupObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return err5
	}
	return nil
}

func (c *CoreClient) GetTenantsList() ([]string, error) {
	tenantList := make([]string, 0)
	url := fmt.Sprintf(authUrl+"/v1/tenants"+"?"+"page_num=%d&page_size=%d", 0, 10000)
	//log.Debug("GetTenantsList url", url)

	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return tenantList, err
	}
	resp, err := http.DefaultClient.Do(req)

	resp2, err2 := c.ParseResp(resp, err)
	if nil != err2 {
		log.Error("error get tenants, ", err)
		return tenantList, err2
	}

	tenantsMap, ok := resp2.(map[string]interface{})
	if !ok {
		return tenantList, errors.New("auth trans error")
	}

	tenants, ok1 := tenantsMap["tenants"]
	if !ok1 {
		return tenantList, errors.New("tenants error")
	}
	tenantsArry, ok2 := tenants.([]interface{})
	if !ok2 {
		return tenantList, errors.New("tenants error")
	}
	for _, v := range tenantsArry {
		if v1, ok1 := v.(map[string]interface{}); ok1 == true {
			if v2, ok2 := v1["tenant_id"]; ok2 == true {
				tenantList = append(tenantList, v2.(string))
			}
		}
	}
	return tenantList, nil
}

var _defaultCli = NewCoreClient()

func NewDaprClientFromContext(ctx context.Context, daprHTTPPort string) (*openapi.DaprClient, error) {
	tm, err := _defaultCli.GetUser(ctx)
	if nil != err {
		return nil, err
	}
	cli := openapi.NewDaprClient("3500", tm.TenantId, tm.UserId)
	return cli, nil
}

func NewDaprClientDefault(client *dapr.HTTPClient) *openapi.DaprClient {
	cli := openapi.NewDaprClientWithConn(client, http.Header{})
	return cli
}
