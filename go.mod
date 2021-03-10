module github.com/grafana/grizzly

go 1.13

require (
	github.com/centrifugal/centrifuge-go v0.6.2
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/gdamore/tcell v1.3.0
	github.com/go-clix/cli v0.1.0
	github.com/go-test/deep v1.0.7
	github.com/gobwas/glob v0.2.3
	github.com/google/go-jsonnet v0.17.0
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/kr/pretty v0.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mitchellh/gox v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.3.3
	github.com/nsf/jsondiff v0.0.0-20210303162244-6ea32392771e
	github.com/pmezard/go-difflib v1.0.0
	github.com/rivo/tview v0.0.0-20200818120338-53d50e499bf9
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/crypto v0.0.0-20200422194213-44a606286825
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8
)

replace k8s.io/client-go => k8s.io/client-go v0.18.3
