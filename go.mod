module github.com/grafana/grizzly

go 1.16

require (
	github.com/centrifugal/centrifuge-go v0.6.2
	github.com/containerd/containerd v1.5.2 // indirect
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/fatih/color v1.9.0
	github.com/gdamore/tcell v1.3.0
	github.com/go-clix/cli v0.2.0
	github.com/gobwas/glob v0.2.3
	github.com/google/go-jsonnet v0.17.0
	github.com/grafana/cortex-tools v0.10.2
	github.com/grafana/synthetic-monitoring-agent v0.0.23
	github.com/grafana/synthetic-monitoring-api-go-client v0.0.2
	github.com/grafana/tanka v0.14.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/rivo/tview v0.0.0-20200818120338-53d50e499bf9
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// Cortex Overrides
replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

// Keeping this same as Cortex to avoid dependency issues.
replace k8s.io/client-go => k8s.io/client-go v0.20.4

replace k8s.io/api => k8s.io/api v0.20.4

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85

// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab

// Required for Alertmanager
replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.8.1
