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
)
