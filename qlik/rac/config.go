package rac

import (
	"crypto/tls"
	"net/http"

	"github.com/soderasen-au/go-common/util"
)

type ExtraTLSConfig struct {
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty" bson:"insecure_skip_verify,omitempty"`
}

func (c ExtraTLSConfig) Apply(cfg *tls.Config) {
	cfg.InsecureSkipVerify = c.InsecureSkipVerify
}

type Config struct {
	BaseUrl        string            `json:"base_url,omitempty" yaml:"base_url,omitempty" bson:"base_url,omitempty"`
	APIPrefix      *string           `json:"api_prefix,omitempty" yaml:"api_prefix,omitempty" bson:"api_prefix,omitempty"`
	Auth           *AuthConfig       `json:"auth" yaml:"auth,omitempty" bson:"auth,omitempty"`
	IsCloud        *bool             `json:"is_cloud,omitempty" yaml:"is_cloud,omitempty" bson:"is_cloud,omitempty"`
	VirtualProxy   *string           `json:"virtual_proxy,omitempty" yaml:"virtual_proxy,omitempty" bson:"virtual_proxy,omitempty"`
	ExtraTlsConfig *ExtraTLSConfig   `json:"extra_tls_config,omitempty" yaml:"extra_tls_config,omitempty" bson:"extra_tls_config,omitempty"`
	TimeoutSec     *int              `json:"timeout_sec,omitempty" yaml:"timeout_sec,omitempty" bson:"timeout_sec,omitempty"`
	ExtraHeaders   map[string]string `json:"extra_headers,omitempty" yaml:"extra_headers,omitempty" bson:"extra_headers,omitempty"`
	LogFileName    *string           `json:"log_file_name,omitempty" yaml:"log_file_name,omitempty" bson:"log_file_name,omitempty"`
	Cookie         http.CookieJar    `json:"-" yaml:"-" bson:"-"`
}

func (c Config) IsForCloud() bool {
	return util.MaybeNil(c.IsCloud)
}
