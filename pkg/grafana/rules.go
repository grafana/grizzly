package grafana

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/grizzly/pkg/grizzly"
	"gopkg.in/yaml.v3"
)

// getRemoteRuleGrouping retrieves a datasource object from Grafana
func getRemoteRuleGroup(uid string) (*grizzly.Resource, error) {
	parts := strings.SplitN(uid, ".", 2)
	namespace := parts[0]
	name := parts[1]

	out, err := cortexTool("rules", "print", "--disable-color")
	if err != nil {
		return nil, err
	}
	groupings := map[string][]PrometheusRuleGroup{}
	err = yaml.Unmarshal(out, &groupings)
	if err != nil {
		return nil, err
	}
	for key, grouping := range groupings {
		if key == namespace {
			for _, group := range grouping {
				if group.Name == name {
					spec := map[string]interface{}{
						"rules": group.Rules,
					}
					handler := RuleHandler{}
					resource := grizzly.NewResource(handler.APIVersion(), handler.Kind(), uid, spec)
					resource.SetMetadata("namespace", namespace)
					return &resource, nil
				}
			}
		}
	}
	return nil, grizzly.ErrNotFound
}

// PrometheusRuleGroup encapsulates a list of rules
type PrometheusRuleGroup struct {
	Namespace string                   `yaml:"-"`
	Name      string                   `yaml:"name"`
	Rules     []map[string]interface{} `yaml:"rules"`
}

// PrometheusRuleGrouping encapsulates a set of named rule groups
type PrometheusRuleGrouping struct {
	Namespace string                `json:"namespace"`
	Groups    []PrometheusRuleGroup `json:"groups"`
}

func writeRuleGroup(resource grizzly.Resource) error {
	tmpfile, err := ioutil.TempFile("", "cortextool-*")
	newGroup := PrometheusRuleGroup{
		Name:  resource.Name(),
		Rules: resource.Spec()["rules"].([]map[string]interface{}),
	}
	grouping := PrometheusRuleGrouping{
		Namespace: resource.GetMetadata("namespace"),
		Groups:    []PrometheusRuleGroup{newGroup},
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
