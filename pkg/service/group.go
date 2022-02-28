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

	//0. check group name repeated
	errRepeated := s.checkNameRepated(ctx, req.Group.Name)
	if nil != errRepeated {
		log.Debug("err:", errRepeated)
		return nil, errRepeated
	}

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
	//groupObject := &pb.EntityResponse{} // core define
	groupObject := make(map[string]interface{}) // core define
	err5 := json.Unmarshal(res, &groupObject)
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
	re, err7 := structpb.NewValue(groupObject)
	if nil != err7 {
		log.Error("convert  failed ", err7)
		return nil, err7
	}
	out := &pb.CreateGroupResponse{
		GroupObject: re,
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
	updateGroup := &pb.UpdateGroupEntityCoreInfo{}
	updateGroup.Group = req.Group

	data, err3 := json.Marshal(updateGroup)
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
	//groupObject := &pb.EntityResponse{} // core define
	groupObject := make(map[string]interface{})
	err6 := json.Unmarshal(res, &groupObject)
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

	//return
	re, err7 := structpb.NewValue(groupObject)
	if nil != err7 {
		log.Error("convert  failed ", err7)
		return nil, err7
	}
	out := &pb.UpdateGroupResponse{
		GroupObject: re,
	}

	return out, nil

}

func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupResponse, error) {
	log.Debug("DelGroup")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	out := &pb.DeleteGroupResponse{
		FaildDelGroup: make([]*pb.FaildDelGroup, 0),
	}
	for _, id := range req.Ids.GetIds() {
		//check child
		err1 := s.checkChild(ctx, id)
		if err1 != nil {
			fd := &pb.FaildDelGroup{
				Id:     id,
				Reason: err1.Error(),
			}
			out.FaildDelGroup = append(out.FaildDelGroup, fd)
			log.Error("have SubNode", id)
			continue
		}

		//get core url
		midUrl := "/" + id
		url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
		log.Debug("get url :", url)

		// do it
		_, err2 := s.httpClient.Delete(url)
		if nil != err2 {
			fd := &pb.FaildDelGroup{
				Id:     id,
				Reason: err2.Error(),
			}
			out.FaildDelGroup = append(out.FaildDelGroup, fd)
			log.Error("error core return error", id)
			continue
		}
	}
	//fmt response
	return out, nil
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
	//groupObject := &pb.EntityResponse{} // core define
	groupObject := make(map[string]interface{})
	err3 := json.Unmarshal(res, &groupObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}

	//return
	re, err7 := structpb.NewValue(groupObject)
	if nil != err7 {
		log.Error("convert  failed ", err7)
		return nil, err7
	}
	out := &pb.GetGroupResponse{
		GroupObject: re,
	}

	return out, nil
}

func (s *GroupService) GetGroupTree(ctx context.Context, req *pb.GetGroupTreeRequest) (*pb.GetGroupTreeResponse, error) {
	log.Debug("getGroupTree")
	log.Debug("req:", req)

	listEntityTotalInfo, err3 := s.CoreSearchEntity(ctx, req.ListEntityQuery)
	if err3 != nil {
		log.Error("error Unmarshal data from core", err3)
		return nil, err3
	}

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

		/*tempTree := tree
		for _, p := range str {
			_, ok := tempTree[p]
			if !ok {
				tempTree[p] = make(map[string]interface{})
			}
			tempTree = tempTree[p].(map[string]interface{})
		}*/
		tempTree := tree
		lastNode := tree
		for _, p := range str {
			_, ok := tempTree[p]
			if !ok {
				tempTree[p] = make(map[string]interface{})
				tempTree[p].(map[string]interface{})["nodeInfo"] = make(map[string]interface{})
				tempTree[p].(map[string]interface{})["subNode"] = make(map[string]interface{})
			}
			lastNode = tempTree[p].(map[string]interface{})["nodeInfo"].(map[string]interface{})
			tempTree = tempTree[p].(map[string]interface{})["subNode"].(map[string]interface{})
		}
		for k, v := range v.(map[string]interface{}) {
			lastNode[k] = v
		}
	}
	log.Debug("tree = ", tree)

	return tree, nil
}

func (s *GroupService) CoreSearchEntity(ctx context.Context, listEntityQuery *pb.ListEntityQuery) (map[string]interface{}, error) {
	log.Debug("CoreSearchEntity")

	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/search"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "group")
	log.Debug("core url :", url)

	//Data isolation
	user := &pb.Condition{
		Field:    "owner",
		Operator: "$eq",
		Value:    tm["owner"],
	}
	listEntityQuery.Condition = append(listEntityQuery.Condition, user)
	log.Debug("Query:", listEntityQuery)

	//do it
	filter, err1 := json.Marshal(listEntityQuery)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.httpClient.Post(url, filter)
	if nil != err2 {
		return nil, err2
	}

	listObject := make(map[string]interface{})
	err3 := json.Unmarshal(res, &listObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}

	return listObject, nil
}

func (s *GroupService) checkChild(ctx context.Context, id string) error {
	log.Debug("checkChild")
	//create query
	query := &pb.ListEntityQuery{
		PageNum:      1,
		PageSize:     1000,
		OrderBy:      "name",
		IsDescending: false,
		Query:        "",
		Condition:    make([]*pb.Condition, 0),
	}
	condition1 := &pb.Condition{
		Field:    "sysField._spacePath",
		Operator: "$wildcard",
		Value:    id,
	}
	query.Condition = append(query.Condition, condition1)
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    "device",
	}
	query.Condition = append(query.Condition, condition2)

	log.Debug("child q", query)
	//search
	listObject, err := s.CoreSearchEntity(ctx, query)
	if err != nil {
		log.Error("error Core return ", err)
		return err
	}

	//check
	total, ok := listObject["total"]
	if !ok {
		log.Error("error  total field does not exist")
		return errors.New("total field does not exist")
	}

	tl, ok1 := total.(float64)
	if !ok1 {
		return errors.New("total is not int type")
	}

	if tl == 0 {
		return nil
	} else {
		return errors.New("have SubNode")
	}
}

func (s *GroupService) checkNameRepated(ctx context.Context, name string) error {
	log.Debug("checkNameRepated")
	if name == "" {
		return errors.New("name cannot be empty")
	}
	//create query
	query := &pb.ListEntityQuery{
		PageNum:      1,
		PageSize:     0,
		OrderBy:      "name",
		IsDescending: false,
		Query:        "",
		Condition:    make([]*pb.Condition, 0),
	}
	condition1 := &pb.Condition{
		Field:    "group.name",
		Operator: "$eq",
		Value:    name,
	}
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    "group",
	}
	query.Condition = append(query.Condition, condition1)
	query.Condition = append(query.Condition, condition2)
	//search
	listObject, err := s.CoreSearchEntity(ctx, query)
	if err != nil {
		log.Error("error Core return ", err)
		return err
	}

	//check
	total, ok := listObject["total"]
	if !ok {
		log.Error("error  total field does not exist")
		return errors.New("total field does not exist")
	}
    /*switch total.(type){
    case string :
        log.Debug("string")
    case int32 :
        log.Debug("int")
    case uint32 :
        log.Debug("uint")
    case uint64 :
        log.Debug("uint64")
    case interface{} :
        log.Debug("inter")
    }
    log.Debug("type:", reflect.TypeOf(total))*/ 

	tl, ok1 := total.(float64)
	if !ok1 {
		return errors.New("total is not int type")
	}

	if tl == 0 {
		return nil
	} else {
		return errors.New("have repeated")
	}
}
