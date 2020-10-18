module github.com/grafana/grizzly

go 1.13

require (
	github.com/cortexproject/cortex v1.2.0
	github.com/fatih/color v1.9.0
	github.com/gdamore/tcell v1.3.0
	github.com/go-clix/cli v0.1.0
	github.com/google/go-jsonnet v0.15.1-0.20200331184325-4f4aa80dd785
	github.com/kr/pretty v0.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/malcolmholmes/grizzly v0.0.1
	github.com/mitchellh/mapstructure v1.3.3
	github.com/prometheus/prometheus v1.8.2-0.20200622142935-153f859b7499
	github.com/rivo/tview v0.0.0-20200818120338-53d50e499bf9
	golang.org/x/crypto v0.0.0-20200422194213-44a606286825
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8
)

replace k8s.io/client-go => k8s.io/client-go v0.18.3
