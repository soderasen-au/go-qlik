package qrs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type TableQueryFilter struct {
	Owner      string
	ObjectType string
	Approved   bool
	Published  bool
}

func (f TableQueryFilter) String() string {
	s := fmt.Sprintf("(approved eq %v and published eq %v", f.Approved, f.Published)
	if f.Owner != "" {
		s = s + fmt.Sprintf(" and (owner.userId so '%s' or owner.name so '%s')", f.Owner, f.Owner)
	}
	if f.ObjectType != "" {
		s = s + fmt.Sprintf(" and objectType so '%s'", f.ObjectType)
	}
	s = s + ")"
	return s
}

type TableQuery struct {
	Filter         TableQueryFilter
	OrderAscending bool
	Skip           int
	SortColumn     string
	Take           int
}

func (q TableQuery) GetParams() map[string]string {
	params := make(map[string]string)
	f := q.Filter.String()
	params["filter"] = f
	params["orderAscending"] = fmt.Sprintf("%v", q.OrderAscending)
	params["sortColumn"] = q.SortColumn
	params["take"] = fmt.Sprintf("%v", q.Take)
	return params
}

func NewTableQuery() *TableQuery {
	return &TableQuery{
		OrderAscending: true,
		Skip:           0,
		SortColumn:     "name",
		Take:           500,
	}
}

type Table struct {
	SchemaPath  *string         `json:"schemaPath,omitempty"`
	ColumnNames []string        `json:"columnNames,omitempty"`
	Rows        [][]interface{} `json:"rows,omitempty"`
	Id          *string         `json:"id,omitempty"`
}

type TableDefinition struct {
	Id                 *string                 `json:"id,omitempty"`
	Entity             *string                 `json:"entity,omitempty"`
	CreatedDate        *time.Time              `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time              `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string                 `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string                 `json:"schemaPath,omitempty"`
	Privileges         []string                `json:"privileges,omitempty"`
	Type               *string                 `json:"type,omitempty"`
	Columns            []TableDefinitionColumn `json:"columns,omitempty"`
}

func NewDefaultTableDefinition() *TableDefinition {
	return &TableDefinition{
		Entity: util.Ptr("App.Object"),
		Columns: []TableDefinitionColumn{
			TablePropertyColumnOf("id"),
			TablePropertyColumnOf("app.id"),
			TablePropertyColumnOf("name"),
			TablePropertyColumnOf("objectType"),
			TablePropertyColumnOf("owner"),
			TablePropertyColumnOf("approved"),
			TablePropertyColumnOf("published"),
			TablePropertyColumnOf("app.name"),
			TablePropertyColumnOf("app.stream.name"),
		},
	}
}

type TableCol int

const (
	TABLE_COL_ID TableCol = iota
	TABLE_COL_APP_ID
	TABLE_COL_NAME
	TABLE_COL_OBJTYPE
	TABLE_COL_OWNER
	TABLE_COL_APPROVED
	TABLE_COL_PUBLISHED
	TABLE_COL_APP_NAME
	TABLE_COL_STREAM
)

type TableDefinitionColumn struct {
	Id                 *string                     `json:"id,omitempty"`
	CreatedDate        *time.Time                  `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time                  `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string                     `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string                     `json:"schemaPath,omitempty"`
	Name               *string                     `json:"name,omitempty"`
	ColumnType         *string                     `json:"columnType,omitempty"`
	Definition         *string                     `json:"definition,omitempty"`
	List               []TableDefinitionListColumn `json:"list,omitempty"`
}

func TablePropertyColumnOf(name string) TableDefinitionColumn {
	return TableDefinitionColumn{
		Name:       util.Ptr(name),
		ColumnType: util.Ptr("Property"),
		Definition: util.Ptr(name),
	}
}

type TableDefinitionListColumn struct {
	Id                 *string    `json:"id,omitempty"`
	CreatedDate        *time.Time `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string    `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string    `json:"schemaPath,omitempty"`
	Name               *string    `json:"name,omitempty"`
	ColumnType         *int32     `json:"columnType,omitempty"`
	Definition         *string    `json:"definition,omitempty"`
}

func (c *Client) GetTable(query TableQuery, def TableDefinition) (*Table, *util.Result) {
	params := query.GetParams()
	buf, res := c.Do(http.MethodPost, "/App/Object/table", params, &def)
	if res != nil {
		return nil, res.With("DoRequest")
	}
	table := Table{}
	err := json.Unmarshal(buf, &table)
	if err != nil {
		return nil, util.Error("ParseTable", err)
	}

	return &table, nil
}
