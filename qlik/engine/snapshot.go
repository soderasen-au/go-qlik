package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
)

type ObjectSnapshot struct {
	Title       *string           `json:"title,omitempty"`
	Description *string           `json:"description,omitempty"`
	Properties  json.RawMessage   `json:"properties,omitempty"`
	HyperCube   *enigma.HyperCube `json:"hypercube,omitempty"`
	Digest      string            `json:"digest,omitempty"`
}

func ObjSnapshoter(e ObjWalkEntry) (*ObjWalkResult[ObjectSnapshot], *util.Result) {
	if e.Logger == nil {
		e.Logger = loggers.NullLogger
	}
	e.Logger.Trace().Msgf("snapshot at %s[%s/%s]", util.MaybeNil(e.AppId), e.Info.Type, e.Info.Id)

	obj, err := GetObject(e.Doc, e.Info.Type, e.Info.Id)
	if err != nil {
		return nil, util.Error("GetObject", err)
	}
	if !HasMethodOn(obj, "GetPropertiesRaw") {
		return nil, util.LogMsgError(e.Logger, "HasMethodOn", "GetPropertiesRaw")
	}
	prop, err := Invoke1Res1ErrOn(obj, "GetPropertiesRaw", ConnCtx)
	if err != nil {
		return nil, util.Error("Invoke1Res1ErrOn::GetPropertiesRaw", err)
	}
	buf := prop.Interface().(json.RawMessage)
	buf = bytes.ReplaceAll(buf, []byte(util.MaybeNil(e.AppId)), []byte("__appid__"))
	digest, res := crypto.SHA2656Hex(buf)
	if res != nil {
		return nil, res.With("SHA2656Hex")
	}
	objProp := ObjectPropeties{
		Info:       e.Info,
		Properties: buf,
	}
	title, desc := GetTitle(e.Parent, &objProp, e.Logger)
	e.Logger.Trace().Msgf(" - Title: %s, Description: %s", util.MaybeNil(title), util.MaybeNil(desc))

	var cube *enigma.HyperCube
	genericObj, ok := obj.Interface().(*enigma.GenericObject)
	if !ok || genericObj == nil {
		return nil, util.MsgError("GetHyperCube", "invalid object object")
	}
	cube, res = GetHyperCube(genericObj, enigma.Size{Cx: -1, Cy: 100})
	if res != nil {
		return nil, res.With("GetHyperCube")
	}
	if cube != nil && cube.Size != nil {
		e.Logger.Trace().Msgf(" - [%s/%s] size: (%d, %d)", util.MaybeNil(title), util.MaybeNil(desc), cube.Size.Cy, cube.Size.Cx)
	}

	objShot := ObjWalkResult[ObjectSnapshot]{
		Info: e.Info,
		Result: &ObjectSnapshot{
			Title:       title,
			Description: desc,
			Properties:  buf,
			HyperCube:   cube,
			Digest:      digest,
		},
		ChildResults: make([]*ObjWalkResult[ObjectSnapshot], 0),
	}
	return &objShot, nil
}

func RecursiveGetSnapshots(doc *enigma.Doc, _logger *zerolog.Logger) (AppWalkResult[ObjectSnapshot], *util.Result) {
	walkers := make(ListWalkFuncMap[ObjectSnapshot])
	walkers[ANY_LIST] = NewRecurObjWalkFunc(ObjSnapshoter)
	appSnapshoter := AppWalker[ObjectSnapshot]{
		Walkers: walkers,
		Logger:  _logger,
	}

	return appSnapshoter.Walk(doc)
}

func (from *ObjectSnapshot) Diff(to *ObjectSnapshot) (*util.Diff, *util.Result) {
	diff, res := util.NewJsonDiff("Title", from.Title, to.Title)
	if res != nil {
		return diff, res.With("Title Diff")
	}
	if diff != nil {
		return diff, nil
	}
	diff, res = util.NewJsonDiff("Description", from.Description, to.Description)
	if res != nil {
		return diff, res.With("Description Diff")
	}
	if diff != nil {
		return diff, nil
	}

	diff, res = util.NewJsonDiff("Properties", &from.Properties, &to.Properties)
	if res != nil {
		return diff, res.With("Properties Diff")
	}
	if diff != nil {
		return diff, nil
	}

	diff, res = util.NewJsonDiff("HyperCube", from.HyperCube, to.HyperCube)
	if res != nil {
		return diff, res.With("HyperCube Diff")
	}
	if diff != nil {
		return diff, nil
	}
	return nil, nil
}

func AppSnapshotDiff(from, to AppWalkResult[ObjectSnapshot]) (AppWalkResult[util.Diff], *util.Result) {
	diff := make(AppWalkResult[util.Diff])
	common := map[string]bool{}

	for fromListName, fromList := range from {
		diff[fromListName] = make(ListWalkResult[util.Diff])
		flattenedList := FlattenList(fromList)

		toFlattened := make(ListWalkResult[ObjectSnapshot])
		if toList, toListExists := to[fromListName]; toListExists {
			toFlattened = FlattenList(toList)
		}
		for fromObjId, fromObj := range flattenedList {
			if toObj, toObjExists := toFlattened[fromObjId]; toObjExists {
				objDiff, res := fromObj.Result.Diff(toObj.Result)
				if res != nil {
					return nil, res.With(fmt.Sprintf("Obj[%s].Diff", fromObjId))
				}
				if objDiff != nil {
					diff[fromListName][fromObjId] = &ObjWalkResult[util.Diff]{
						Info:   fromObj.Info,
						Meta:   fromObj.Meta,
						Result: objDiff,
					}
				}
				common[fromObjId] = true
			} else {
				fromPath := fmt.Sprintf("%s/%s", fromObj.Info.Type, fromObj.Info.Id)
				objDiff, _ := util.NewJsonDiff(fromPath, &fromPath, nil)
				diff[fromListName][fromObjId] = &ObjWalkResult[util.Diff]{
					Info:   fromObj.Info,
					Meta:   fromObj.Meta,
					Result: objDiff,
				}
			}
		}
	}

	for toListName, toList := range to {
		for objId, obj := range toList {
			if _, ok := common[objId]; !ok {
				if _, dok := diff[toListName]; !dok {
					diff[toListName] = make(ListWalkResult[util.Diff])
				}
				toPath := fmt.Sprintf("%s/%s", obj.Info.Type, obj.Info.Id)
				objDiff, _ := util.NewJsonDiff(toPath, nil, &toPath)
				diff[toListName][objId] = &ObjWalkResult[util.Diff]{
					Info:   obj.Info,
					Meta:   obj.Meta,
					Result: objDiff,
				}
			}
		}
	}

	return diff, nil
}
