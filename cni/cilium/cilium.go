package cilium

import (
	"bytes"
	"embed"
	"strconv"
	"text/template"
)

//go:embed cilium.yaml
var f embed.FS

type Cluster struct {
	ClusterHost string
	ClusterPort string
}

func Fetch(host string, port int32) (string, error) {
	data, err := f.ReadFile("cilium.yaml")
	if err != nil {
		return "", err
	}
	c := Cluster{
		ClusterHost: host,
		ClusterPort: strconv.FormatInt(int64(port), 10),
	}
	tmpl, err := template.New("cilium").Parse(string(data))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, c); err != nil {
		return "", err
	}
	return out.String(), nil
}
