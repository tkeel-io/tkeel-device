package service

import (
	"context"
	json "encoding/json"
	"errors"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/group/v1"
)

type GroupService struct {
	pb.UnimplementedGroupServer
	httpClient *CoreClient
}

func NewGroupService() *GroupService {
	return &GroupService{
		httpClient: NewCoreClient(),
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	log.Debug("CreateGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return &pb.CreateGroupResponse{Result: "failed"}, err
	}

	//get core url
	entityId := GetUUID()
	url := s.httpClient.GetCoreUrl("", tm) + "&id=" + entityId
	log.Debug("get url: ", url)

	//fmt request
	sysField := &pb.GroupEntitySysField{
		XId:        entityId,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
	}
	subIds := &pb.GroupEntitySubEntityIds{
		SubEntityId: make(map[string]string),
	}
	entityInfo := &pb.GroupEntityCoreInfo{
		Group:    req.Group,
		SysField: sysField,
		SubIds:   subIds,
	}
	log.Debug("entityinfo : ", entityInfo)
	data, err3 := json.Marshal(entityInfo)
	if nil != err3 {
		return &pb.CreateGroupResponse{Result: "failed"}, err3
	}

	// do it
	res, err4 := s.httpClient.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core", err4)
		return &pb.CreateGroupResponse{Result: "failed"}, err4
	}

	//fmt response
	entityTotalInfo := make(map[string]interface{})
	err5 := json.Unmarshal(res, &entityTotalInfo)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return &pb.CreateGroupResponse{Result: "failed"}, err5 
	}
	properties, ok := entityTotalInfo["properties"]
	if !ok {
		log.Error("error choose data from core")
		return &pb.CreateGroupResponse{Result: "failed"}, errors.New("error choose data from core")
	}

	prop, err6 := json.Marshal(properties)
	if err6 != nil {
		log.Error("error Marshal prop data from core")
		return &pb.CreateGroupResponse{Result: "failed"}, err6
	}
	out := &pb.CreateGroupResponse{}
	err7 := json.Unmarshal(prop, &out.EntityInfo)
	if err7 != nil {
		log.Error("error convert data from core")
		out.Result = "failed"
		return out, err7
	}
	out.Result = "ok"
	return out, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return &pb.CommonResponse{Result: "failed"}, err
	}

	//get core url
	midUrl := "/" + req.GetId()
	url := s.httpClient.GetCoreUrl(midUrl, tm)
	log.Debug("put url :", url)
	log.Debug("body :", req.Group)

	//fmt request
	updateEntityInfo := &pb.UpdateGroupEntityCoreInfo{
		Group:      req.Group,
		XUpdatedAt: GetTime(),
	}
	data, err3 := json.Marshal(updateEntityInfo)
	if nil != err3 {
		return &pb.CommonResponse{Result: "failed"}, err3
	}

	// do it
	_, err4 := s.httpClient.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return &pb.CommonResponse{Result: "failed"}, err4
	}

	//fmt response
	return &pb.CommonResponse{Result: "OK"}, nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.CommonResponse, error) {
	log.Debug("DelGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	for _, id := range req.Ids.GetIds() {
		//get core url
		midUrl := "/" + id
		url := s.httpClient.GetCoreUrl(midUrl, tm)
		log.Debug("get url :", url)

		//fmt request

		// do it
		_, err4 := s.httpClient.Delete(url)
		if nil != err4 {
			log.Error("error post data to core", id)
			return &pb.CommonResponse{Result: "failed"}, err4
		}
	}
	//fmt response
	return &pb.CommonResponse{Result: "OK"}, nil
}
func (s *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupResponse, error) {
	log.Debug("GetGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + req.GetId()
	url := s.httpClient.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url)

	//do it
	res, err2 := s.httpClient.Get(url)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return &pb.GetGroupResponse{Result: "failed"}, err2
	}

	//fmt response
	entityTotalInfo := make(map[string]interface{})
	err3 := json.Unmarshal(res, &entityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return &pb.GetGroupResponse{Result: "failed"}, err3
	}
	properties, ok := entityTotalInfo["properties"]
	if !ok {
		log.Error("error choose data from core")
		return &pb.GetGroupResponse{Result: "failed"}, errors.New("error choose data from core")
	}

	prop, err4 := json.Marshal(properties)
	if err4 != nil {
		log.Error("error Marshal prop data from core")
		return &pb.GetGroupResponse{Result: "failed"}, err4
	}
	out := &pb.GetGroupResponse{}
	err5 := json.Unmarshal(prop, &out.EntityInfo)
	if err5 != nil {
		log.Error("error convert data from core")
		out.Result = "failed"
		return out, err5
	}
	out.Result = "ok"
	return out, nil
}

func (s *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupResponse, error) {
	log.Debug("ListGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/search"
	url := s.httpClient.GetCoreUrl(midUrl, tm)
	log.Debug("url :", url)
	log.Debug("fliter :", req.Filter)

	data, err := json.Marshal(req.Filter)
	if nil != err {
		return &pb.ListGroupResponse{Result: "failed"}, err
	}

	//do it
	res, err2 := s.httpClient.Post(url, data)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return &pb.ListGroupResponse{Result: "failed"}, err2
	}

	//fmt response
	listEntityTotalInfo := &pb.ListEntityResponse{}
	err3 := json.Unmarshal(res, listEntityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return &pb.ListGroupResponse{Result: "failed"}, err3
	}
	out := &pb.ListGroupResponse{
		Result:         "OK",
		ListEntityInfo: listEntityTotalInfo,
	}

	return out, nil
}

func (s *GroupService) ListGroupItems(ctx context.Context, req *pb.ListGroupItemsRequest) (*pb.ListGroupItemsResponse, error) {
	log.Debug("ListGroupItems")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return &pb.ListGroupItemsResponse{Result: "failed"}, err
	}

	//get core url
	midUrl := "/" + req.GetId()
	url := s.httpClient.GetCoreUrl(midUrl, tm)
	log.Debug("patch url :", url)

	//fmt request

	// do it
	res, err4 := s.httpClient.Get(url)
	if nil != err4 {
		log.Error("error post data to core", err4)
		return &pb.ListGroupItemsResponse{Result: "failed"}, err4
	}

	//fmt response
	entityTotalInfo := make(map[string]interface{})
	err3 := json.Unmarshal(res, &entityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return &pb.ListGroupItemsResponse{Result: "failed"}, err3
	}

	properties, ok := entityTotalInfo["properties"]
	if !ok {
		log.Error("error choose data from core")
		return &pb.ListGroupItemsResponse{Result: "failed"}, errors.New("error choose data from core")
	}

	prop, err4 := json.Marshal(properties)
	if err4 != nil {
		log.Error("error Marshal prop data from core")
		return &pb.ListGroupItemsResponse{Result: "failed"}, err4
	}
	entityInfo := &pb.GroupEntityCoreInfo{}
	err5 := json.Unmarshal(prop, entityInfo)
	if err5 != nil {
		log.Error("error convert data from core")
		return &pb.ListGroupItemsResponse{Result: "failed"}, err5
	}
	out := &pb.ListGroupItemsResponse{
		Result: "Ok",
		SubIds: entityInfo.SubIds,
	}
	return out, nil
}

func (s *GroupService) AddGroupItems(ctx context.Context, req *pb.AddGroupItemsRequest) (*pb.CommonResponse, error) {
	log.Debug("AddGroupItems")
	log.Debug("req:", req.Ids.Ids)
	idsMap := make(map[string]interface{})
	modifyIdsMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		//add SubEntityId
		idsMap[id] = "subEntity"
		_, err := s.CorePatchMethod(ctx, req.Id, idsMap, "subIds.subEntityId.", "replace")
		if err != nil {
			log.Error("error add SubEntityId")
			return &pb.CommonResponse{Result: "failed"}, err
		}
		//modify SubEntity
		modifyIdsMap["group"] = req.Id
		_, err2 := s.CorePatchMethod(ctx, id, modifyIdsMap, "dev.","replace")
		if err2 != nil {
			log.Error("error modify SubEntity parentId")
			return &pb.CommonResponse{Result: "failed"}, err2
		}
	}
	return &pb.CommonResponse{Result: "Ok"}, nil
}

func (s *GroupService) DelGroupItems(ctx context.Context, req *pb.DelGroupItemsRequest) (*pb.CommonResponse, error) {
	log.Debug("DelGroupItems")
	log.Debug("req:", req.Ids.Ids)
	idsMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		//del SubEntityId
		idsMap[id] = "SubEntity"
		_, err := s.CorePatchMethod(ctx, req.Id, idsMap, "subIds.subEntityId.", "remove")
		if err != nil {
			log.Error("error add SubEntityId")
			return &pb.CommonResponse{Result: "failed"}, err
		}
		//modify SubEntity
		/*idsMap["group"] = req.Id
		        _, err2 := s.CorePatchMethod(ctx, id, idsMap, "dev.","replace")
		        if err2 != nil {
				    log.Error("error modify SubEntity parentId")
				    return &pb.CommonResponse{Result: "failed"}, err2
		        }*/
	}
	return &pb.CommonResponse{Result: "Ok"}, nil
}

func (s *GroupService) AddGroupExt(ctx context.Context, req *pb.AddGroupExtRequest) (*pb.CommonResponse, error) {
	log.Debug("AddGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace")
	default:
		return &pb.CommonResponse{Result: "faild"}, errors.New("error params")
	}
}

func (s *GroupService) UpdateGroupExt(ctx context.Context, req *pb.UpdateGroupExtRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace")
	default:
		return &pb.CommonResponse{Result: "faild"}, errors.New("error params")
	}
}

func (s *GroupService) DelGroupExt(ctx context.Context, req *pb.DelGroupExtRequest) (*pb.CommonResponse, error) {
	log.Debug("DeleteGroupExt")
	log.Debug("req:", req.Keys.Keys)
	delKeysMap := make(map[string]interface{})
	for _, id := range req.Keys.Keys {
		delKeysMap[id] = "del"
	}
	return s.CorePatchMethod(ctx, req.Id, delKeysMap, "group.ext.", "remove")
}

//abstraction

type CorePatch struct {
	Path     string      `json:"path"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func (s *GroupService) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string) (*pb.CommonResponse, error) {
	log.Debug("CorePatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	/*tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err{
		return &pb.CommonResponse{Result: "failed"}, err
	}*/
	tm := make(map[string]string)
	tm["id"] = entityId
	tm["entityType"] = "group"
	tm["owner"] = "tl"
	tm["source"] = "test"

	//get core url
	midUrl := "/" + entityId
	url := s.httpClient.GetCoreUrl(midUrl, tm)
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
		return &pb.CommonResponse{Result: "failed"}, err3
	}

	// do it
	_, err4 := s.httpClient.Patch(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return &pb.CommonResponse{Result: "failed"}, err4
	}

	//fmt response
	return &pb.CommonResponse{Result: "OK"}, nil
}
