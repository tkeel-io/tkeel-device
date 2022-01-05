package service

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/group/v1"
	"google.golang.org/protobuf/types/known/emptypb"
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
		return nil, err
	}

	//get core url
	entityId := GetUUID()
	url := s.httpClient.GetCoreUrl("", tm, "group") + "&id=" + entityId
	log.Debug("get url: ", url)

	//fmt request
	sysField := &pb.GroupEntitySysField{
		XId:        entityId,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
		XOwner:     tm["owner"],
		XSource:    tm["source"],
	}
	subIds := &pb.GroupEntitySubEntityIds{
		SubEntityId: make(map[string]string),
	}
	subIds.SubEntityId["default"] = "SubEntity"
	entityInfo := &pb.GroupEntityCoreInfo{
		Group:    req.Group,
		SysField: sysField,
		SubIds:   subIds,
	}
	log.Debug("entityinfo : ", entityInfo)
	data, err3 := json.Marshal(entityInfo)
	if nil != err3 {
		return nil, err3
	}

	// do it
	res, err4 := s.httpClient.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core", err4)
		return nil, err4
	}

	//fmt response
	groupObject := &pb.EntityResponse{} // core define
	err5 := json.Unmarshal(res, groupObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.CreateGroupResponse{
		GroupObject: groupObject,
	}

	return out, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupResponse, error) {
	log.Debug("UpdateGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + req.GetId()
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
	log.Debug("put url :", url)
	log.Debug("body :", req.Group)

	//fmt request
	updateEntityInfo := &pb.GroupEntity{}
    updateEntityInfo = req.Group

	data, err3 := json.Marshal(updateEntityInfo)
	if nil != err3 {
		return nil, err3
	}

	// do it 
	    //update basicInfo 
    res, err4 := s.httpClient.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}
        //update updateAt
	ma := make(map[string]interface{})
	ma["_updatedAt"] = GetTime()
    _, err5 := s.CorePatchMethod(ctx, req.GetId(), ma, "sysField.", "replace")
	if nil != err5 {
		log.Error("error patch _updateAt", err5)
		return nil, err5 
	}

	//fmt response
	groupObject := &pb.EntityResponse{} // core define
	err6 := json.Unmarshal(res, groupObject)
	if err6 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err6
	}
	out := &pb.UpdateGroupResponse{
		GroupObject: groupObject,
	}

	return out, nil

}

func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*emptypb.Empty, error) {
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
		url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
		log.Debug("get url :", url)

		//fmt request

		// do it
		_, err1 := s.httpClient.Delete(url)
		if nil != err1 {
			log.Error("error post data to core", id)
			return nil, err1
		}
	}
	//fmt response
	return &emptypb.Empty{}, nil
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
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
	log.Debug("get url :", url)

	//do it
	res, err2 := s.httpClient.Get(url)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return nil, err2
	}

	//fmt response
	groupObject := &pb.EntityResponse{} // core define
	err3 := json.Unmarshal(res, groupObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	out := &pb.GetGroupResponse{
		GroupObject: groupObject,
	}

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
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
	log.Debug("url :", url)

	data, err := json.Marshal(req.ListEntityQuery)
	if nil != err {
		return nil, err
	}

	//do it
	res, err2 := s.httpClient.Post(url, data)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return nil, err2
	}

	//fmt response
	listEntityTotalInfo := &pb.ListEntityResponse{}
	err3 := json.Unmarshal(res, listEntityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	out := &pb.ListGroupResponse{
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
		return nil, err
	}

	//get core url
	midUrl := "/" + req.GetId() + "/properties"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group") + fmt.Sprintf("&pids=%s", "subIds")
	log.Debug("url :", url)

	//fmt request

	// do it
	res, err1 := s.httpClient.Get(url)
	if nil != err1 {
		log.Error("error post data to core", err1)
		return nil, err1 
	}

	//fmt response
    entityInfo := &pb.EntityResponse{}
	if err2 := json.Unmarshal(res, entityInfo); nil != err2 {
		return nil, err2 
	}

	kv := entityInfo.Properties.GetStructValue().Fields
	out := &pb.ListGroupItemsResponse{}
	prop, err3 := json.Marshal(kv)
	if nil != err3 {
		return nil, err3 
	}
	err4 := json.Unmarshal(prop, out)
	if nil != err4 {
		return nil, err4 
	}

	return out, nil
}

func (s *GroupService) AddGroupItems(ctx context.Context, req *pb.AddGroupItemsRequest) (*emptypb.Empty, error) {
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
			return nil, err
		}
		//modify SubEntity
		modifyIdsMap["group"] = req.Id
		_, err2 := s.CorePatchMethod(ctx, id, modifyIdsMap, "basicInfo.", "replace")
		if err2 != nil {
			log.Error("error modify SubEntity parentId")
			return nil, err2
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *GroupService) DelGroupItems(ctx context.Context, req *pb.DelGroupItemsRequest) (*emptypb.Empty, error) {
	log.Debug("DelGroupItems")
	log.Debug("req:", req.Ids.Ids)
	idsMap := make(map[string]interface{})
	modifyIdsMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		//del SubEntityId
		idsMap[id] = "SubEntity"
		_, err := s.CorePatchMethod(ctx, req.Id, idsMap, "subIds.subEntityId.", "remove")
		if err != nil {
			log.Error("error add SubEntityId")
			return nil, err
		}
		//modify SubEntity
		modifyIdsMap["group"] = "root"
		_, err2 := s.CorePatchMethod(ctx, id, modifyIdsMap, "basicInfo.", "replace")
		if err2 != nil {
			log.Error("error modify SubEntity parentId")
			return nil, err2
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *GroupService) AddGroupExt(ctx context.Context, req *pb.AddGroupExtRequest) (*emptypb.Empty, error) {
	log.Debug("AddGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace")
	default:
		return nil, errors.New("error params")
	}
}

func (s *GroupService) UpdateGroupExt(ctx context.Context, req *pb.UpdateGroupExtRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace")
	default:
		return nil, errors.New("error params")
	}
}

func (s *GroupService) DelGroupExt(ctx context.Context, req *pb.DelGroupExtRequest) (*emptypb.Empty, error) {
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

func (s *GroupService) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string) (*emptypb.Empty, error) {
	log.Debug("CorePatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/patch"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
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
	_, err4 := s.httpClient.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}

	//fmt response
	return &emptypb.Empty{}, nil
}
