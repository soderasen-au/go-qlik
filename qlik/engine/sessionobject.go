package engine

import (
	"encoding/json"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/util"
)

var (
	SessionObjDef = []byte(`
	{
		"qInfo": {
			"qId": "",
			"qType": "SessionLists"
		},
		"qAppObjectListDef": {
			"qType": "sheet",
			"qData": {
				"id": "/qInfo/qId"
			}
		},
		"qDimensionListDef": {
			"qType": "dimension",
			"qData": {
				"id": "/qInfo/qId"
			}
		},
		"qMeasureListDef": {
			"qType": "measure",
			"qData": {
				"id": "/qInfo/qId"
			}
		},
		"qBookmarkListDef": {
			"qType": "bookmark",
			"qData": {
				"id": "/qInfo/qId"
			}
		},
		"qVariableListDef": {
			"qType": "variable",
			"qData": {
				"id": "/qInfo/qId"
			}
		}
	}
	`)

	SessionDimListDefData     = []byte(`{"qDim": "/qDim", "qDimInfos": "/qDimInfos"}`)
	SessionMeasureListDefData = []byte(`{"qMeasure": "/qMeasure"}`)
)

type (
	SessionDimensionData struct {
		Dim      *enigma.NxLibraryDimension     `json:"qDim,omitempty"`
		DimInfos []*enigma.GenericDimensionInfo `json:"qDimInfos,omitempty"`
		Title    *string                        `json:"qTitle,omitempty"`
	}
	SessionDimensionLayout struct {
		Info     *enigma.NxInfo                 `json:"qInfo,omitempty"`
		Meta     *NxMeta                        `json:"qMeta,omitempty"`
		Dim      *enigma.NxLibraryDimension     `json:"qDim,omitempty"`
		DimInfos []*enigma.GenericDimensionInfo `json:"qDimInfos,omitempty"`
	}
	SessionMeasureData struct {
		Measure *enigma.NxLibraryMeasure `json:"qMeasure,omitempty"`
		Title   *string                  `json:"qTitle,omitempty"`
	}
	SessionMeasureLayout struct {
		Info    *enigma.NxInfo           `json:"qInfo,omitempty"`
		Measure *enigma.NxLibraryMeasure `json:"qMeasure,omitempty"`
		Meta    *NxMeta                  `json:"qMeta,omitempty"`
	}
)

func GetDimensionList(doc *enigma.Doc) ([]*SessionDimensionLayout, *util.Result) {
	prop := enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{Type: "DimensionList"},
		DimensionListDef: &enigma.DimensionListDef{
			Type: "dimension",
			Data: SessionDimListDefData,
		},
	}
	obj, err := doc.CreateSessionObject(ConnCtx, &prop)
	if err != nil {
		return nil, util.Error("CreateSessionObject", err)
	}
	layoutBuf, err := obj.GetLayoutRaw(ConnCtx)
	if err != nil {
		return nil, util.Error("GetLayoutRaw", err)
	}

	var layout SessionObjectLayout
	err = json.Unmarshal(layoutBuf, &layout)
	if err != nil {
		return nil, util.Error("ParseLayout", err)
	}

	ret := make([]*SessionDimensionLayout, 0)
	if layout.DimensionList != nil && layout.DimensionList.Items != nil {
		for ii, item := range layout.DimensionList.Items {
			var dataLayout SessionDimensionData
			err = json.Unmarshal(item.Data, &dataLayout)
			if err != nil {
				return nil, util.Errorf("ParseMeasureLayout[%d]: %s", ii, err.Error())
			}
			ret = append(ret, &SessionDimensionLayout{
				Info:     item.Info,
				Meta:     item.Meta,
				Dim:      dataLayout.Dim,
				DimInfos: dataLayout.DimInfos,
			})
		}
	}

	return ret, nil
}

func GetMeasureList(doc *enigma.Doc) ([]*SessionMeasureLayout, *util.Result) {
	prop := enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{Type: "MeasureList"},
		MeasureListDef: &enigma.MeasureListDef{
			Type: "measure",
			Data: SessionMeasureListDefData,
		},
	}
	obj, err := doc.CreateSessionObject(ConnCtx, &prop)
	if err != nil {
		return nil, util.Error("CreateSessionObject", err)
	}
	layoutBuf, err := obj.GetLayoutRaw(ConnCtx)
	if err != nil {
		return nil, util.Error("GetLayoutRaw", err)
	}

	var layout SessionObjectLayout
	err = json.Unmarshal(layoutBuf, &layout)
	if err != nil {
		return nil, util.Error("ParseLayout", err)
	}

	ret := make([]*SessionMeasureLayout, 0)
	if layout.MeasureList != nil && layout.MeasureList.Items != nil {
		for ii, item := range layout.MeasureList.Items {
			var data SessionMeasureData
			err = json.Unmarshal(item.Data, &data)
			if err != nil {
				return nil, util.Errorf("ParseMeasureLayout[%d]: %s", ii, err.Error())
			}
			ret = append(ret, &SessionMeasureLayout{
				Info:    item.Info,
				Meta:    item.Meta,
				Measure: data.Measure,
			})
		}
	}

	return ret, nil
}
