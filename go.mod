module github.com/grafana/grizzly

go 1.16

require (
	github.com/centrifugal/centrifuge-go v0.6.2
	github.com/fatih/color v1.9.0
	github.com/gdamore/tcell v1.3.0
	github.com/go-clix/cli v0.2.0
	github.com/gobwas/glob v0.2.3
	github.com/google/go-jsonnet v0.17.0
	github.com/grafana/synthetic-monitoring-agent v0.0.23
	github.com/grafana/synthetic-monitoring-api-go-client v0.0.2
	github.com/grafana/tanka v0.14.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/rivo/tview v0.0.0-20200818120338-53d50e499bf9
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace k8s.io/client-go => k8s.io/client-go v0.18.3
