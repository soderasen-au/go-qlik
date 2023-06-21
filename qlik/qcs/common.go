package qcs

type Link struct {
	Href *string `json:"href,omitempty"`
}

type Links struct {
	Self Link  `json:"self"`
	Prev *Link `json:"prev,omitempty"`
	Next *Link `json:"next,omitempty"`
}

type Meta struct {
	Fault     *bool `json:"fault,omitempty"`
	Temporary *bool `json:"temporary,omitempty"`
	Timeout   *bool `json:"timeout,omitempty"`
}

type ResponseBase struct {
	ID       *string `json:"id,omitempty"`
	AppID    *string `json:"appId,omitempty"`
	TenantID *string `json:"tenantId,omitempty"`
	UserID   *string `json:"userId,omitempty"`
	Type     *string `json:"type,omitempty"`
}
