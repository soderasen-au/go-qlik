package qrs

import (
	"testing"
)

func TestClient_GetTable(t *testing.T) {
	qrsClient, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)

	query := NewTableQuery()
	query.Filter = TableQueryFilter{
		Owner:      "soder",
		ObjectType: "sheet",
		Approved:   false,
		Published:  false,
	}
	table, res := qrsClient.GetTable(*query, *NewDefaultTableDefinition())
	if res != nil {
		t.Errorf("GetTable: %s", res.Error())
		return
	}

	for i, row := range table.Rows {
		aoId := row[TABLE_COL_ID].(string)
		aoName := row[TABLE_COL_NAME].(string)
		aoType := row[TABLE_COL_OBJTYPE].(string)
		obj, res := qrsClient.GetAppObject(aoId)
		if res != nil {
			t.Errorf("row[%d]: id: %s can't find engine object id", i, aoId)
		}
		logger.Trace().Msgf("AppObject %s-%s: %s => EngineObjet %s-%s\n", aoType, aoId, aoName, *obj.EngineObjectType, *obj.EngineObjectId)
	}
}
