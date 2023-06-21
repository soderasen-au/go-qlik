package engine

const (
	MASTER_OBJECT string = "masterobject"
	STORY         string = "story"
	CONNECTION    string = "connection"
)

var (
	GetObjMethods = map[string]string{
		"sheet":     "GetObject",
		"measure":   "GetMeasure",
		"dimension": "GetDimension",
		"variable":  "GetVariableById",
		"bookmark":  "GetBookmark",
	}
	CreateObjMethods = map[string]string{
		"sheet":     "CreateObjectRaw",
		"measure":   "CreateMeasureRaw",
		"dimension": "CreateDimensionRaw",
		"variable":  "CreateVariableExRaw",
		"bookmark":  "CreateBookmarkRaw",
	}
	DestroyObjMethods = map[string]string{
		"sheet":     "DestroyObject",
		"measure":   "DestroyMeasure",
		"dimension": "DestroyDimension",
		"variable":  "DestroyVariableById",
		"bookmark":  "DestroyBookmark",
	}

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

	SessionDimListDefData = []byte(`
	{
		"title": "/qMetaDef/title",
		"tags": "/qMetaDef/tags",
		"grouping": "/qDim/qGrouping",
		"info": "/qDimInfos",
		"labelExpression": "/qDim/qLabelExpression"
	}`)
)
