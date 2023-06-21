package qrs

import (
	"encoding/json"
	"fmt"

	"github.com/Click-CI/common/util"
)

type (
	AuditMatrixParametersResourceRefCondensed struct {
		SchemaPath     string `json:"schemaPath,omitempty"`
		ResourceFilter string `json:"resourceFilter,omitempty"`
		Selection      string `json:"selection,omitempty"`
	}

	AuditMatrixParameters struct {
		SchemaPath              string                                    `json:"schemaPath,omitempty"`
		ResourceRef             AuditMatrixParametersResourceRefCondensed `json:"resourceRef"`
		SubjectRef              AuditMatrixParametersResourceRefCondensed `json:"subjectRef"`
		ResourceType            string                                    `json:"resourceType,omitempty"`
		Actions                 int64                                     `json:"actions,omitempty"`
		EnvironmentAttributes   string                                    `json:"environmentAttributes,omitempty"`
		AuditLimit              int32                                     `json:"auditLimit,omitempty"`
		SubjectStart            string                                    `json:"subjectStart,omitempty"`
		SubjectProperties       []string                                  `json:"subjectProperties,omitempty"`
		ResourceProperties      []string                                  `json:"resourceProperties,omitempty"`
		OutputObjectsPrivileges int64                                     `json:"outputObjectsPrivileges,omitempty"`
	}

	AuditMatrixElementCondensed struct {
		SchemaPath string           `json:"schemaPath,omitempty"`
		SubjectId  string           `json:"subjectId,omitempty"`
		ResourceId string           `json:"resourceId,omitempty"`
		Audit      map[string]int64 `json:"audit,omitempty"`
	}

	AuditMatrixResourceCondensed struct {
		SchemaPath         string            `json:"schemaPath,omitempty"`
		ResourceProperties map[string]string `json:"resourceProperties,omitempty"`
		Privileges         []string          `json:"privileges,omitempty"`
	}

	AuditMatrixSubjectCondensed struct {
		SchemaPath        string            `json:"schemaPath,omitempty"`
		SubjectProperties map[string]string `json:"subjectProperties,omitempty"`
		Privileges        []string          `json:"privileges,omitempty"`
	}

	AuditMatrix struct {
		SchemaPath string                                  `json:"schemaPath,omitempty"`
		Matrix     []AuditMatrixElementCondensed           `json:"matrix,omitempty"`
		Resources  map[string]AuditMatrixResourceCondensed `json:"resources,omitempty"`
		Subjects   map[string]AuditMatrixSubjectCondensed  `json:"subjects,omitempty"`
	}
)

func (c *Client) GetHubApps() ([]App, *util.Result) {
	matrixParas := AuditMatrixParameters{
		Actions:                 46,
		AuditLimit:              1000,
		EnvironmentAttributes:   "context=AppAccess;",
		OutputObjectsPrivileges: 4,
		ResourceProperties:      []string{"name"},
		ResourceType:            "App",
		SubjectProperties:       []string{"id", "name", "userId", "userDirectory"},
		SubjectRef: AuditMatrixParametersResourceRefCondensed{
			ResourceFilter: fmt.Sprintf("((((userDirectory eq '%s' and userId eq '%s'))))", c.Cfg.Auth.User.Directory, c.Cfg.Auth.User.Id),
		},
	}
	resp, res := c.Post("/SystemRule/Security/audit/matrix", &matrixParas)
	if res != nil {
		return nil, res.With("CreateMatrix")
	}

	var matrix AuditMatrix
	err := json.Unmarshal(resp, &matrix)
	if err != nil {
		return nil, util.Error("ParseMatrix", err)
	}

	apps, res := c.GetAppList()
	if res != nil {
		return nil, res.With("GetAppList")
	}
	c.Logger().Debug().Msgf("audit matrix: %d, total apps: %d", len(matrix.Resources), len(apps))

	ret := make([]App, 0)
	if matrix.Resources == nil {
		c.Logger().Warn().Msg("no resource is returned")
		return ret, nil
	}

	noNeedToCheck := false
	if len(matrix.Resources) == 0 && len(apps) > 0 {
		c.Logger().Warn().Msg("no resource is returned, this user is not admin, no need to check hub/qmc scope")
		noNeedToCheck = true
	}
	for _, app := range apps {
		if _, ok := matrix.Resources[app.ID]; ok || noNeedToCheck {
			ret = append(ret, app)
		} else {
			c.Logger().Debug().Msgf("app %s is not in Hub view", app.ID)
		}
	}
	return ret, nil
}
