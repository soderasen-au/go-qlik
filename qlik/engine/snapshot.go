package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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

func RecursiveGetSnapshots(doc *enigma.Doc, cfg MixedConfig, opts *WalkOptions, _logger *zerolog.Logger) (AppWalkResult[ObjectSnapshot], *util.Result) {
	walkers := make(ListWalkFuncMap[ObjectSnapshot])
	walkers[ANY_LIST] = NewRecurObjWalkFunc(ObjSnapshoter)
	appSnapshoter := AppWalker[ObjectSnapshot]{
		Walkers: walkers,
		Logger:  _logger,
	}

	return appSnapshoter.Walk(doc, cfg, opts)
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

	diff, res = HyperCubeDiff("HyperCube", from.HyperCube, to.HyperCube)
	if res != nil {
		return diff, res.With("HyperCube Diff")
	}
	if diff != nil {
		return diff, nil
	}
	return nil, nil
}

func HyperCubeDiff(path string, from, to *enigma.HyperCube) (*util.Diff, *util.Result) {
	d := util.Diff{
		Path: path,
	}

	if from == nil && to == nil {
		d.Type = util.DiffEq
		d.Diff = "DiffEq"
		return &d, nil
	}
	if from == nil {
		d.Type = util.DiffAdd
		d.Diff = "DiffAdd"
		return &d, nil
	}
	if to == nil {
		d.Type = util.DiffDel
		d.Diff = "DiffDel"
		return &d, nil
	}

	d.Type = util.DiffMod
	builder := strings.Builder{}

	fsz := util.MaybeNil(from.Size)
	tsz := util.MaybeNil(to.Size)
	if fsz.Cy != tsz.Cy || fsz.Cx != tsz.Cx {
		builder.WriteString(fmt.Sprintf("Size:\n\tLeft: (%d, %d)\n\tRight: (%d, %d)\n", fsz.Cy, fsz.Cx, tsz.Cy, tsz.Cx))
		d.Diff = builder.String()
		return &d, nil
	}
	if from.Mode != to.Mode {
		builder.WriteString(fmt.Sprintf("Mode:\n\tLeft: %s\n\tRight: %s\n", from.Mode, to.Mode))
		d.Diff = builder.String()
		return &d, nil
	}

	if from.Mode == "P" || from.Mode == "K" {
		frompages := from.PivotDataPages
		topages := to.PivotDataPages
		if len(frompages) != len(topages) {
			builder.WriteString(fmt.Sprintf("Len(Pages):\n\tLeft: %d\n\tRight: %d\n", len(frompages), len(topages)))
			d.Diff = builder.String()
			return &d, nil
		}

		for i := 0; i < len(frompages); i++ {
			fpage := frompages[i]
			tpage := topages[i]
			if len(fpage.Data) != len(tpage.Data) {
				builder.WriteString(fmt.Sprintf("Len(Page[%d]):\n\tLeft: %d\n\tRight: %d\n", i, len(fpage.Data), len(tpage.Data)))
				d.Diff = builder.String()
				return &d, nil
			}
			for r := 0; r < len(fpage.Data); r++ {
				frow := fpage.Data[r]
				trow := tpage.Data[r]
				if len(frow) != len(trow) {
					builder.WriteString(fmt.Sprintf("Len(Page[%d]Row[%d]):\n\tLeft: %d\n\tRight: %d\n", i, r, len(frow), len(trow)))
					d.Diff = builder.String()
					return &d, nil
				}
				for c := 0; c < len(frow); c++ {
					if frow[c].Text != trow[c].Text {
						builder.WriteString(fmt.Sprintf("Page[%d] (%d, %d):\n\tLeft: %s\n\tRight: %s\n", i, r, c, frow[c].Text, trow[c].Text))
					}
				}
			}
		}
	} else {
		frompages := from.DataPages
		topages := to.DataPages
		if len(frompages) != len(topages) {
			builder.WriteString(fmt.Sprintf("Len(Pages):\n\tLeft: %d\n\tRight: %d\n", len(frompages), len(topages)))
			d.Diff = builder.String()
			return &d, nil
		}

		for i := 0; i < len(frompages); i++ {
			fpage := frompages[i]
			tpage := topages[i]
			if len(fpage.Matrix) != len(tpage.Matrix) {
				builder.WriteString(fmt.Sprintf("Len(Page[%d]):\n\tLeft: %d\n\tRight: %d\n", i, len(fpage.Matrix), len(tpage.Matrix)))
				d.Diff = builder.String()
				return &d, nil
			}
			for r := 0; r < len(fpage.Matrix); r++ {
				frow := fpage.Matrix[r]
				trow := tpage.Matrix[r]
				if len(frow) != len(trow) {
					builder.WriteString(fmt.Sprintf("Len(Page[%d]Row[%d]):\n\tLeft: %d\n\tRight: %d\n", i, r, len(frow), len(trow)))
					d.Diff = builder.String()
					return &d, nil
				}
				for c := 0; c < len(frow); c++ {
					if frow[c].Text != trow[c].Text {
						builder.WriteString(fmt.Sprintf("Page[%d] (%d, %d):\n\tLeft: %s\n\tRight: %s\n", i, r, c, frow[c].Text, trow[c].Text))
					}
				}
			}
		}
	}

	d.Diff = builder.String()
	if d.Diff == "" {
		return nil, nil
	}
	return &d, nil
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
						SheetId:     fromObj.SheetId,
						SheetName:   fromObj.SheetName,
						ObjectTitle: fromObj.ObjectTitle,
						Info:        fromObj.Info,
						Meta:        fromObj.Meta,
						Result:      objDiff,
					}
				}
				common[fromObjId] = true
			} else {
				fromPath := fmt.Sprintf("%s/%s", fromObj.Info.Type, fromObj.Info.Id)
				objDiff, _ := util.NewJsonDiff(fromPath, &fromPath, nil)
				diff[fromListName][fromObjId] = &ObjWalkResult[util.Diff]{
					SheetId:     fromObj.SheetId,
					SheetName:   fromObj.SheetName,
					ObjectTitle: fromObj.ObjectTitle,
					Info:        fromObj.Info,
					Meta:        fromObj.Meta,
					Result:      objDiff,
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
					SheetId:     obj.SheetId,
					SheetName:   obj.SheetName,
					ObjectTitle: obj.ObjectTitle,
					Info:        obj.Info,
					Meta:        obj.Meta,
					Result:      objDiff,
				}
			}
		}
	}

	return diff, nil
}
