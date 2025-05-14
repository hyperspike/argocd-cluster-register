package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigParse(t *testing.T) {
	os.Clearenv()
	env := map[string]string{
		"ROLE_ARN": "testing",
		//	"ROLEARN": "testing",
		"PROJECT": "test1,test2",
	}
	for k, v := range env {
		_ = os.Setenv(k, v)
	}

	c, err := ParseConfig()
	assert.Nil(t, err)

	assert.Equal(t, c.RoleARN, "testing", "RoleARN should be 'testing'")
	assert.Equal(t, c.Projects[0], "test1", "First project should be 'test1'")
	assert.Equal(t, c.Projects[1], "test2", "Second project should be 'test2'")
}
