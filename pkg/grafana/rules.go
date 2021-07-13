package grafana

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	ctClient "github.com/grafana/cortex-tools/pkg/client"
	"github.com/grafana/cortex-tools/pkg/rules"
	"github.com/grafana/grizzly/pkg/grizzly"
	"gopkg.in/yaml.v3"
)

const (
	cortexApiKey     = "CORTEX_API_KEY"
	cortexAddress    = "CORTEX_ADDRESS"
	cortexTenantID   = "CORTEX_TENANT_ID"
	backenTypeCortex = "cortex"
)

var client CTClient = CTService{}

type CTClient interface {
	listRules() ([]byte, error)
	writeRules(string, string) error
}

type CTService struct {
}

// getRemoteRuleGrouping retrieves a datasource object from Grafana
func getRemoteRuleGroup(uid string) (*grizzly.Resource, error) {
	parts := strings.SplitN(uid, ".", 2)
	namespace := parts[0]
	name := parts[1]
	out, err := client.listRules()
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

// getRemoteRuleGroupingList retrieves a datasource object from Grafana
func getRemoteRuleGroupList() ([]string, error) {
	out, err := client.listRules()
	if err != nil {
		return nil, err
	}
	groupings := map[string][]PrometheusRuleGroup{}
	err = yaml.Unmarshal(out, &groupings)
	if err != nil {
		return nil, err
	}

	IDs := []string{}
	for namespace, grouping := range groupings {
		for _, group := range grouping {
			uid := fmt.Sprintf("%s.%s", namespace, group.Name)
			IDs = append(IDs, uid)
		}
	}
	return IDs, nil
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
	if err != nil {
		return err
	}
	newGroup := PrometheusRuleGroup{
		Name: resource.Name(),
		//Rules: resource.Spec()["rules"].([]map[string]interface{}),
		Rules: []map[string]interface{}{},
	}
	rules := resource.Spec()["rules"].([]interface{})
	for _, ruleIf := range rules {
		rule := ruleIf.(map[string]interface{})
		newGroup.Rules = append(newGroup.Rules, rule)
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
	err = client.writeRules(grouping.Namespace, tmpfile.Name())
	if err != nil {
		return err
	}
	os.Remove(tmpfile.Name())
	return err
}

func (ct CTService) listRules() ([]byte, error) {
	client, err := newCortexClient()
	if err != nil {
		return nil, err
	}
	rule, err := client.ListRules(context.Background(), "")
	if err != nil {
		if err == ctClient.ErrResourceNotFound {
			return nil, nil
		}
		return nil, err
	}

	encodedRule, err := yaml.Marshal(&rule)
	if err != nil {
		return nil, err
	}
	return encodedRule, nil
}

func (ct CTService) writeRules(namespace, fileName string) error {
	client, err := newCortexClient()
	if err != nil {
		return err
	}
	ruleNamespaces, err := rules.ParseFiles(backenTypeCortex, []string{fileName})
	if err != nil {
		return err
	}
	for _, ruleNamespace := range ruleNamespaces {
		for _, group := range ruleNamespace.Groups {
			err = client.CreateRuleGroup(context.Background(), ruleNamespace.Namespace, group)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newCortexClient() (*ctClient.CortexClient, error) {
	cfg := ctClient.Config{
		Key:             os.Getenv(cortexApiKey),
		Address:         os.Getenv(cortexAddress),
		ID:              os.Getenv(cortexTenantID),
		UseLegacyRoutes: false,
	}
	cortexClient, err := ctClient.New(cfg)
	if err != nil {
		return nil, err
	}
	return cortexClient, nil
}
