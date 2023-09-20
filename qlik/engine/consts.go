package engine

const (
	MASTER_OBJECT string = "masterobject"
	STORY         string = "story"
	CONNECTION    string = "connection"

	ANY_LIST     string = "AnyList"
	SHEET_LIST   string = "AppObjectList"
	BM_LIST      string = "BookmarkList"
	DIM_LIST     string = "DimensionList"
	MEASURE_LIST string = "MeasureList"
	VAR_LIST     string = "VariableList"
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

	ObjectListNameMap = map[string]string{
		"AppObjectList": "AppObject",
		"BookmarkList":  "Bookmark",
		"connection":    "Connection",
		"DimensionList": "Dimension",
		"masterobject":  "MasterObject",
		"MeasureList":   "Measure",
		"story":         "Story",
		"VariableList":  "Variable",
	}
)
