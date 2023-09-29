package engine

import (
	"encoding/json"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
)

type ObjectSnapshot struct {
	Title       *string           `json:"title,omitempty"`
	Description *string           `json:"description,omitempty"`
	Properties  json.RawMessage   `json:"properties,omitempty"`
	HyperCube   *enigma.HyperCube `json:"hypercube,omitempty"`
	Digest      string            `json:"digest,omitempty"`
}

func ObjSnapshoter(doc *enigma.Doc, info, parent *enigma.NxInfo, _logger *zerolog.Logger) (*ObjWalkResult[ObjectSnapshot], *util.Result) {
	_logger.Trace().Msgf("snapshot [%s/%s]", info.Type, info.Id)

	obj, err := GetObject(doc, info.Type, info.Id)
	if err != nil {
		return nil, util.Error("GetObject", err)
	}
	if !HasMethodOn(obj, "GetPropertiesRaw") {
		return nil, util.LogMsgError(_logger, "HasMethodOn", "GetPropertiesRaw")
	}
	prop, err := Invoke1Res1ErrOn(obj, "GetPropertiesRaw", ConnCtx)
	if err != nil {
		return nil, util.Error("Invoke1Res1ErrOn::GetPropertiesRaw", err)
	}
	buf := prop.Interface().(json.RawMessage)
	digest, res := crypto.SHA2656Hex(buf)
	if res != nil {
		return nil, res.With("SHA2656Hex")
	}
	objProp := ObjectPropeties{
		Info:       info,
		Properties: buf,
	}
	title, desc := GetTitle(nil, &objProp, _logger)
	_logger.Trace().Msgf(" - Title: %s, Description: %s", util.MaybeNil(title), util.MaybeNil(desc))

	var cube *enigma.HyperCube
	genericObj, ok := obj.Interface().(*enigma.GenericObject)
	if ok && genericObj != nil {
		cube, res = GetHyperCube(genericObj, enigma.Size{Cx: -1, Cy: 100})
		if res != nil {
			return nil, res.With("GetHyperCube")
		}
		if cube != nil && cube.Size != nil {
			_logger.Trace().Msgf(" - [%s/%s] size: (%d, %d)", util.MaybeNil(title), util.MaybeNil(desc), cube.Size.Cy, cube.Size.Cx)
		}
	}

	objShot := ObjWalkResult[ObjectSnapshot]{
		Info: info,
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
