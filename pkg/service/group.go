package service

import (
	"context"
	json "encoding/json"
	"errors"
	//"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/group/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
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

type SpaceTreeNode struct {
	NodeInfo *map[string]interface{}
	SubNode  *map[string]interface{}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	log.Debug("CreateGroup")
	log.Debug("req:", req.Group)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	entityId := GetUUID()
	url := s.httpClient.GetCoreUrl("", tm, "group") + "&id=" + entityId
	log.Debug("core url: ", url)

	//fmt request
	sysField := &pb.GroupEntitySysField{
		XId:        entityId,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
		XOwner:     tm["owner"],
		XSource:    tm["source"],
		XSpacePath: entityId,
	}

	entityInfo := &pb.GroupEntityCoreInfo{
		Group:    req.Group,
		SysField: sysField,
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

	//set spacePath mapper
	err1 := s.httpClient.setSpacePathMapper(tm, entityId, req.Group.ParentId)
	if nil != err1 {
		log.Error("error setSpacePath mapper", err1)
		return nil, err1
	}

	//return
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

	data, err3 := json.Marshal(req.Group)
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
	_, err5 := s.httpClient.CorePatchMethod(ctx, req.GetId(), ma, "sysField.", "replace", "/patch")
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

	//set spacePath mapper
	err1 := s.httpClient.setSpacePathMapper(tm, req.GetId(), req.Group.ParentId)
	if nil != err1 {
		log.Error("error addSpacePath mapper", err1)
		return nil, err1
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

func (s *GroupService) GetGroupTree(ctx context.Context, req *pb.GetGroupTreeRequest) (*pb.GetGroupTreeResponse, error) {
	log.Debug("getGroupTree")
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

	//
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
	listEntityTotalInfo := make(map[string]interface{})
	err3 := json.Unmarshal(res, &listEntityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core", err3)
		return nil, err3
	}
	log.Debug("listEntityTotalInfo = ", listEntityTotalInfo)

	//create space tree
	tree, err4 := s.createSpaceTree(listEntityTotalInfo)
	if err4 != nil {
		log.Error("error parse space tree")
		return nil, err4
	}

	re, err5 := structpb.NewValue(tree)
	if nil != err5 {
		log.Error("convert tree failed ", err5)
		return nil, err5
	}
	out := &pb.GetGroupTreeResponse{
		GroupTree: re,
	}

	return out, nil
}

func (s *GroupService) ListGroupItems(ctx context.Context, req *pb.ListGroupItemsRequest) (*pb.ListGroupItemsResponse, error) {
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
	out := &pb.ListGroupItemsResponse{
		ListEntityInfo: listEntityTotalInfo,
	}

	return out, nil
}

func (s *GroupService) AddGroupExt(ctx context.Context, req *pb.AddGroupExtRequest) (*emptypb.Empty, error) {
	log.Debug("AddGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.httpClient.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace", "/patch")
	default:
		return nil, errors.New("error params")
	}
}

func (s *GroupService) UpdateGroupExt(ctx context.Context, req *pb.UpdateGroupExtRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateGroupExt")
	log.Debug("req:", req.Kvs.AsInterface())
	switch kv := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		return s.httpClient.CorePatchMethod(ctx, req.Id, kv, "group.ext.", "replace", "/patch")
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
	return s.httpClient.CorePatchMethod(ctx, req.Id, delKeysMap, "group.ext.", "remove", "/patch")
}

//abstraction

type CorePatch struct {
	Path     string      `json:"path"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

/*func (s *GroupService) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string) (*emptypb.Empty, error) {
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
}*/

/*func (s *GroupService) setSpacePathMapper(tm map[string]string, Id string, parentId string) error {

	log.Debug("setSpacePathMapper")
	//check ParentId
	if parentId == "" {
		return nil
	}

	//get url
	midUrl := "/" + Id + "/mappers"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
	log.Debug("mapper url = ", url)

	//fmt request
	data := make(map[string]string)
	data["name"] = "mapper_space_path"
	data["tql"] = "insert into " + Id + " select " + parentId + ".sysField._spacePath + '/" + Id + "'  as " + "sysField._spacePath"
	log.Debug("data = ", data)

	send, err := json.Marshal(data)
	if nil != err {
		return err
	}

	// do it
	_, err1 := s.httpClient.Post(url, send)
	if nil != err1 {
		log.Error("error core return")
		return err1
	}

	return nil
}*/

func (s *GroupService) createSpaceTree(listEntityTotalInfo map[string]interface{}) (map[string]interface{}, error) {
	log.Debug("createSpaceTree")
	tree := make(map[string]interface{})

	items, ok := listEntityTotalInfo["items"].([]interface{})
	if !ok {
		return nil, errors.New("error parse items")
	}
	spacePath := ""
	log.Debug("items len = ", len(items))
	for _, v := range items {

		//get spacePath
		if s, ok := v.(map[string]interface{}); ok == true {
			if prop, ok1 := s["properties"]; ok1 == true {
				if prop2, ok2 := prop.(map[string]interface{}); ok2 == true {
					if sysField1, ok3 := prop2["sysField"]; ok3 == true {
						if sysField2, ok4 := sysField1.(map[string]interface{}); ok4 == true {
							if spacePath1, ok5 := sysField2["_spacePath"]; ok5 == true {
								if spacePath2, ok6 := spacePath1.(string); ok6 == true {
									spacePath = spacePath2
								}
							}
						}
					}
				}
			}
		}

		if spacePath == "" {
			continue
		}
		log.Debug("spacePath = ", spacePath)
		str := strings.Split(spacePath, "/")
		log.Debug("spacePath  /= ", str)

		//create tree

		tempTree := tree
		for _, p := range str {
			_, ok := tempTree[p]
			if !ok {
				tempTree[p] = make(map[string]interface{})
			}
			tempTree = tempTree[p].(map[string]interface{})
		}
		/*tempTree := tree
		        lastNode := tree
				for _, p := range str {
					_, ok := tempTree[p]
					if !ok {
						tempTree[p] = make(map[string]interface{})
		                tempTree[p].(map[string]interface{})["nodeInfo"] = make(map[string]interface{})
		                tempTree[p].(map[string]interface{})["subNode"] = make(map[string]interface{})
					}
		            tempTree = tempTree[p].(map[string]interface{})["subNode"].(map[string]interface{})
		            lastNode = tempTree[p].(map[string]interface{})["nodeInfo"].(map[string]interface{})
		        }
		        lastNode.(map[string]interface{}) = v.(map[string]interface{})
		        log.Debug(lastNode)*/
	}
	log.Debug("tree = ", tree)

	return tree, nil
}
