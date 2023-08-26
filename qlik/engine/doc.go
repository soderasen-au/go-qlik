package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/loggers"
	"reflect"
	"sync"

	"github.com/soderasen-au/go-common/util"
)

func GetCurrentSelection(doc *enigma.Doc, stateName string) (*enigma.SelectionObject, *util.Result) {
	siProp := enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{Type: "SessionLists"},
		SelectionObjectDef: &enigma.SelectionObjectDef{
			StateName: stateName,
		},
	}
	curSeleObj, err := doc.CreateSessionObject(ConnCtx, &siProp)
	if err != nil {
		return nil, util.Error("CreateSessionObject", err)
	}
	curSeleObjLayoutBuf, err := curSeleObj.GetLayoutRaw(ConnCtx)
	if err != nil {
		return nil, util.Error("GetLayoutRaw", err)
	}

	var sessObjLayout SessionObjectLayout
	err = json.Unmarshal(curSeleObjLayoutBuf, &sessObjLayout)
	if err != nil {
		return nil, util.Error("ParseSelectionLayout", err)
	}

	return sessObjLayout.SelectionObject, nil
}

func GetListObject(doc *enigma.Doc, stateName, fieldName string) (*enigma.ListObject, *util.Result) {
	loProp := enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{Type: "ListObject"},
		ListObjectDef: &enigma.ListObjectDef{
			StateName: stateName,
			Def: &enigma.NxInlineDimensionDef{
				FieldDefs: []string{fieldName},
			},
		},
	}
	listObj, err := doc.CreateSessionObject(ConnCtx, &loProp)
	if err != nil {
		return nil, util.Error("CreateSessionObject", err)
	}
	listObjLayoutBuf, err := listObj.GetLayoutRaw(ConnCtx)
	if err != nil {
		return nil, util.Error("GetLayoutRaw", err)
	}

	var sessObjLayout SessionObjectLayout
	err = json.Unmarshal(listObjLayoutBuf, &sessObjLayout)
	if err != nil {
		return nil, util.Error("ParseSelectionLayout", err)
	}

	return sessObjLayout.ListObject, nil
}

func SetVariable(doc *enigma.Doc, name string, value string) error {
	obj, err := doc.GetVariableByName(ConnCtx, name)
	if err != nil {
		return fmt.Errorf("get variable failed: %s", err.Error())
	}
	prop, err := obj.GetProperties(ConnCtx)
	if err != nil {
		return fmt.Errorf("get var properties failed: %s", err.Error())
	}
	prop.Definition = value
	err = obj.SetProperties(ConnCtx, prop)
	if err != nil {
		return fmt.Errorf("set var properties failed: %s", err.Error())
	}
	return nil
}

func SetStringVariable(doc *enigma.Doc, name string, value string) error {
	obj, err := doc.GetVariableByName(ConnCtx, name)
	if err != nil {
		return fmt.Errorf("get variable failed: %s", err.Error())
	}
	err = obj.SetStringValue(ConnCtx, value)
	if err != nil {
		return fmt.Errorf("SetStringValue: %s", err.Error())
	}

	return nil
}

func GetObject(doc *enigma.Doc, qtype string, qid string) (reflect.Value, error) {
	method, ok := GetObjMethods[qtype]
	if !ok {
		method = "GetObject"
	}
	obj, err := Invoke1Res1Err(doc, method, ConnCtx, qid)
	if err != nil {
		return reflect.ValueOf(nil), errors.New("Invoke1Res1Err failed: " + err.Error())
	}

	objVal := obj
	//fmt.Printf("obj kind: %s, Type: %s, value: %s\n", objVal.Kind(), objVal.Type(), objVal)
	if obj.Kind() == reflect.Interface && !obj.IsNil() {
		elm := obj.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			objVal = elm
		}
	}
	if objVal.Kind() == reflect.Ptr {
		objVal = objVal.Elem()
	}

	remoteObject := objVal.FieldByName("RemoteObject").Interface().(*enigma.RemoteObject)
	if remoteObject.Handle == 0 || len(remoteObject.Type) == 0 {
		return reflect.ValueOf(nil), errors.New("doesn't have remote object")
	}

	return obj, nil
}

func DestroyObject(doc *enigma.Doc, qtype string, qid string) (reflect.Value, error) {
	method, ok := DestroyObjMethods[qtype]
	if !ok {
		method = "DestroyObject"
	}
	success, err := Invoke1Res1Err(doc, method, ConnCtx, qid)
	if err != nil {
		return reflect.ValueOf(nil), errors.New("Invoke1Res1Err failed: " + err.Error())
	}
	return success, nil
}

func GetObjectList(doc *enigma.Doc, ctx context.Context, objType string) ([]*NxContainerEntry, error) {
	opt := &enigma.NxGetObjectOptions{
		Types: []string{objType},
	}
	buf, err := doc.GetObjectsRaw(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("can't GetObjectsRaw: %s", err.Error())
	}
	entryList := make([]*NxContainerEntry, 0)
	err = json.Unmarshal(buf, &entryList)
	if err != nil {
		return nil, fmt.Errorf("can't Unmarshal object list %s: %s", objType, err.Error())
	}
	return entryList, nil
}

func CreateObject(doc *enigma.Doc, qtype string, prop json.RawMessage) (reflect.Value, error) {
	method, ok := CreateObjMethods[qtype]
	if !ok {
		method = "CreateObjectRaw"
	}

	obj, err := Invoke1Res1Err(doc, method, ConnCtx, prop)
	if err != nil {
		return reflect.ValueOf(nil), errors.New("Invoke1Res1Err failed: " + err.Error())
	}
	return obj, nil
}

func CreateChild(doc *enigma.Doc, parentInfo enigma.NxInfo, prop json.RawMessage) (reflect.Value, error) {
	obj, err := GetObject(doc, parentInfo.Type, parentInfo.Id)
	if err != nil {
		return reflect.ValueOf(nil), fmt.Errorf("get object failed: %s", err.Error())
	}

	child, err := Invoke1Res1ErrOn(obj, "CreateChildRaw", ConnCtx, prop, nil)
	if err != nil {
		return reflect.ValueOf(nil), fmt.Errorf("CreateChildRaw failed: %s", err.Error())
	}
	return child, nil
}

func RecursiveGetProperties(doc *enigma.Doc, entry NxContainerEntry) (*ObjectPropeties, error) {
	qid := entry.Info.Id
	qtype := entry.Info.Type
	logger := loggers.CoreDebugLogger.With().
		Str("mod", "engine").
		Str("func", "RecursiveGetProperties").
		Str("entry", fmt.Sprintf("%s/%s", qtype, qid)).
		Logger()

	logger.Debug().Msg("get object")
	obj, err := GetObject(doc, qtype, qid)
	if err != nil {
		return nil, errors.New("can't get obj " + qid + ": " + err.Error())
	}

	objProp := ObjectPropeties{
		Info: entry.Info,
		Meta: entry.Meta,
	}
	if !HasMethodOn(obj, "GetPropertiesRaw") {
		return nil, fmt.Errorf("obj: %s-%s doesn't have method: %s", qtype, qid, "GetPropertiesRaw")
	}

	logger.Debug().Msg("get properties")
	prop, err := Invoke1Res1ErrOn(obj, "GetPropertiesRaw", ConnCtx)
	if err != nil {
		return nil, errors.New("can't get obj properties " + qid + ": " + err.Error())
	}
	objProp.Properties = prop.Interface().(json.RawMessage)

	if !HasMethodOn(obj, "GetChildInfos") {
		return &objProp, nil
	}
	logger.Debug().Msg("get child infos")
	ret, err := Invoke1Res1ErrOn(obj, "GetChildInfos", ConnCtx)
	if err != nil {
		return nil, errors.New("can't get obj children info " + qid + ": " + err.Error())
	}

	childrenInfos := ret.Interface().([]*enigma.NxInfo)
	childArray := make([]*ObjectPropeties, len(childrenInfos))
	errArray := make([]error, len(childrenInfos))

	var wg sync.WaitGroup
	for i, child := range childrenInfos {
		i := i
		child := child
		wg.Add(1)
		go func(i int, child *enigma.NxInfo) {
			defer wg.Done()
			entry := NxContainerEntry{
				Info: child,
				Meta: &NxMeta{},
			}
			logger.Debug().Msgf("child[%d]: %s/%s start", i, child.Type, child.Id)
			childObjProp, err := RecursiveGetProperties(doc, entry)
			logger.Debug().Msgf("child[%d]: %s/%s finished", i, child.Type, child.Id)
			errArray[i] = err
			childArray[i] = childObjProp
		}(i, child)
	}
	wg.Wait()

	for i, err := range errArray {
		if err != nil {
			return nil, util.Error(fmt.Sprintf("child[%d]: (%s: %s)", i, childrenInfos[i].Id, childrenInfos[i].Type), err)
		}
	}

	objProp.ChildInfos = childArray
	return &objProp, nil
}

func GetSheetsObjectProperties(doc *enigma.Doc) ([]*ObjectPropeties, *util.Result) {
	sessionObj, res := GetSessionObjectLayout(doc)
	if res != nil {
		return nil, res.With("GetSessionObjectLayout")
	}
	if sessionObj.AppObjectList == nil {
		return make([]*ObjectPropeties, 0), nil
	}
	items := sessionObj.AppObjectList.Items

	sheetsProperties := make([]*ObjectPropeties, len(items))
	errArray := make([]error, len(items))
	var wg sync.WaitGroup
	for i, item := range items {
		i := i
		item := item
		wg.Add(1)

		go func() {
			defer wg.Done()
			if item.Meta == nil {
				item.Meta = &NxMeta{}
			}
			prop, err := RecursiveGetProperties(doc, *item)
			errArray[i] = err
			sheetsProperties[i] = prop
		}()
	}
	wg.Wait()

	for i, err := range errArray {
		if err != nil {
			return nil, util.Error(fmt.Sprintf("sheet[%d]: %s", i, items[i].Info.Id), err)
		}
	}

	return sheetsProperties, nil
}

func RecursiveCreateObject(doc *enigma.Doc, prop ObjectPropeties, parent *enigma.NxInfo) (*enigma.GenericObject, *util.Result) {
	qid := prop.Info.Id
	qtype := prop.Info.Type
	logger := loggers.CoreDebugLogger.With().
		Str("mod", "engine").
		Str("func", "RecursiveCreateObject").
		Str("entry", fmt.Sprintf("%s/%s", qtype, qid)).
		Logger()

	var objRet *enigma.GenericObject
	if parent == nil {
		logger.Debug().Msg("create object")
		objVal, err := CreateObject(doc, qtype, prop.Properties)
		if err != nil {
			return nil, util.Error("CreateObject", err)
		}
		objRet = objVal.Interface().(*enigma.GenericObject)
	} else {
		logger.Debug().Msgf("create child object for %s-%s", parent.Type, parent.Id)
		objVal, err := CreateChild(doc, *parent, prop.Properties)
		if err != nil {
			return nil, util.Error("CreateChild", err)
		}
		objRet = objVal.Interface().(*enigma.GenericObject)
	}
	obj := *objRet

	//objLayout, err := objRet.GetLayout(ConnCtx)
	//if err != nil {
	//	return nil, util.Error("GetLayout", err)
	//}

	childrenInfos := prop.ChildInfos
	for i, cp := range childrenInfos {
		_, err := RecursiveCreateObject(doc, *cp, prop.Info)
		if err != nil {
			return nil, util.Error(fmt.Sprintf("child[%d]: (%s: %s)", i, childrenInfos[i].Info.Id, childrenInfos[i].Info.Type), err)
		}
	}

	return &obj, nil
}
