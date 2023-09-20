package engine

import (
	"encoding/json"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
)

type ObjectSnapshot struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Properties  json.RawMessage `json:"properties,omitempty"`
	Digest      string          `json:"digest,omitempty"`
}

func RecursiveGetSnapshots(doc *enigma.Doc, _logger *zerolog.Logger) (AppWalkResult[ObjectSnapshot], *util.Result) {
	objSnapshoter := func(doc *enigma.Doc, info, parent *enigma.NxInfo, _logger *zerolog.Logger) (*ObjWalkResult[ObjectSnapshot], *util.Result) {
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

		objShot := ObjWalkResult[ObjectSnapshot]{
			Info: info,
			Result: &ObjectSnapshot{
				Title:       title,
				Description: desc,
				Properties:  buf,
				Digest:      digest,
			},
			ChildResults: make([]*ObjWalkResult[ObjectSnapshot], 0),
		}
		return &objShot, nil
	}

	walkers := make(ListWalkFuncMap[ObjectSnapshot])
	walkers[ANY_LIST] = NewRecurObjWalkFunc(objSnapshoter)
	appSnapshoter := AppWalker[ObjectSnapshot]{
		Walkers: walkers,
		Logger:  _logger,
	}

	return appSnapshoter.Walk(doc)
}
