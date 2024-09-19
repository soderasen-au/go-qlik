package config

type HubConfig struct {
	BaseURI  string `json:"base_uri,omitempty" yaml:"base_uri,omitempty" bson:"base_uri,omitempty"`
	AuthMode string `json:"auth_mode,omitempty" yaml:"auth_mode,omitempty" bson:"auth_mode,omitempty"`
}
