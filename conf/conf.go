package conf

import (
	"github.com/kelseyhightower/envconfig"
	"strings"
)

type Config struct {
	RoleARN  string `envconfig:"ROLE_ARN"`
	Project  string
	Projects []string
}

func ParseConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	c.Projects = strings.Split(c.Project, ",")
	return &c, err
}
