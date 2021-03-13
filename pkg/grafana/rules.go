package grafana

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/grizzly/pkg/manifests"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"gopkg.in/yaml.v3"
)

// getRemoteRuleGrouping retrieves a datasource object from Grafana
func getRemoteRuleGroup(uid string) (*manifest.Manifest, error) {
	parts := manifests.SplitUID(uid)
	namespace := parts[0]
	name := parts[1]

	out, err := cortexTool("rules", "print", "--disable-color")
	if err != nil {
		return nil, err
	}
	groupings := map[string][]RuleGroup{}
	err = yaml.Unmarshal(out, &groupings)
	if err != nil {
		return nil, err
	}
	for key, grouping := range groupings {
		if key == namespace {
			for _, group := range grouping {
				if group.Name == name {
					m, err := manifests.New("PrometheusRuleGroup",
						name,
						nil,
						map[string]interface{}{
							"rules": group.Rules,
						})
					if err != nil {
						return nil, err
					}
					m = manifests.SetMetadata(m, "namespace", namespace)
					return m, nil
				}
			}
		}
	}
	return nil, grizzly.ErrNotFound
}

// RuleGroup encapsulates a list of rules
type RuleGroup struct {
	Namespace string                   `yaml:"-"`
	Name      string                   `yaml:"name"`
	Rules     []map[string]interface{} `yaml:"rules"`
}

// UID retrieves the UID from a rule group
func (g *RuleGroup) UID() string {
	return fmt.Sprintf("%s-%s", g.Namespace, g.Name)
}

// toYAML returns YAML for a rule group
func (g *RuleGroup) toYAML() (string, error) {
	y, err := yaml.Marshal(g)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// RuleGrouping encapsulates a set of named rule groups
type RuleGrouping struct {
	Namespace string      `json:"namespace"`
	Groups    []RuleGroup `json:"groups"`
}

func writeRuleGroup(m manifest.Manifest) error {
	tmpfile, err := ioutil.TempFile("", "cortextool-*")
	spec, _ := m["spec"].(map[string]interface{})
	newGroup := RuleGroup{
		Name:  m.Metadata().Name(),
		Rules: spec["rules"].([]map[string]interface{}),
	}
	grouping := RuleGrouping{
		Namespace: m.Metadata().Namespace(),
		Groups:    []RuleGroup{newGroup},
	}
	out, err := yaml.Marshal(grouping)
	if err != nil {
		return err
	}
	ioutil.WriteFile(tmpfile.Name(), out, 0644)
	output, err := cortexTool("rules", "load", tmpfile.Name())
	if err != nil {
		log.Println("OUTPUT", output)
		return err
	}
	os.Remove(tmpfile.Name())
	return err
}

func cortexTool(args ...string) ([]byte, error) {
	path := os.Getenv("CORTEXTOOL_PATH")
	if path == "" {
		var err error
		path, err = exec.LookPath("cortextool")
		if err != nil {
			return nil, err
		} else if path == "" {
			return nil, fmt.Errorf("cortextool not found")
		}
	}
	return exec.Command(path, args...).Output()
}
