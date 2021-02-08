package prometheus

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// getRemoteRuleGrouping retrieves a datasource object from Grafana
func getRemoteRuleGrouping(namespace string) (*RuleGrouping, error) {
	out, err := cortexTool("rules", "print", "--disable-color")
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	err = yaml.Unmarshal(out, &data)
	for ns, groups := range data {
		if ns == namespace {
			grouping := RuleGrouping{
				Namespace: ns,
			}
			err := mapstructure.Decode(groups, &grouping.Groups)
			if err != nil {
				return nil, err
			}
			return &grouping, nil
		}
	}
	return nil, grizzly.ErrNotFound
}

// RuleGrouping encapsulates a set of named rule groups
type RuleGrouping struct {
	Namespace string      `json:"namespace"`
	Groups    []RuleGroup `json:"groups"`
}

// RuleGroup encapsulates a list of rules
type RuleGroup struct {
	Name  string                   `yaml:"name"`
	Rules []map[string]interface{} `yaml:"rules"`
}

// toYAML returns YAML for a rule group
func (g *RuleGrouping) toYAML() (string, error) {
	y, err := yaml.Marshal(g)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func writeRuleGrouping(grouping RuleGrouping) error {
	tmpfile, err := ioutil.TempFile("", "cortextool-*")
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
