package prometheus

/*
import (
	rulefmt "github.com/cortexproject/cortex/pkg/ruler/legacy_rulefmt"
	"gopkg.in/yaml.v2"
)

type Rules struct {
	Groups []Group `json:"groups"`
}

type Mixin struct {
	Dashboards Boards `json:"grafanaDashboards"`
	Rules      Rules  `json:"prometheusRules"`
	Alerts     Rules  `json:"prometheusAlerts"`
}
type Group rulefmt.RuleGroup

func (g Group) Kind() string {
	return "Group"
}

func (g Group) UID() string {
	return g.Name
}

// UnmarshalJSON uses the YAML parser for this specific type, because the
// embedded prometheus types require this.
func (g *Group) UnmarshalJSON(data []byte) error {
	return yaml.Unmarshal(data, g)
}
*/
