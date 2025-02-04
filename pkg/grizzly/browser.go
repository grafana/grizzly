package grizzly

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type BrowserInterface struct {
	registry Registry
	port     int
	isDir    bool
}

func NewBrowserInterface(registry Registry, resourcePath string, port int) (*BrowserInterface, error) {
	stat, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	return &BrowserInterface{
		registry: registry,
		isDir:    stat.IsDir(),
		port:     port,
	}, nil
}

func (i BrowserInterface) Open(resources Resources) error {
	path := "/"

	if i.isDir {
		if resources.Len() == 0 {
			return fmt.Errorf("no resources found to proxy")
		} else if resources.Len() == 1 {
			resource := resources.First()
			handler, err := i.registry.GetHandler(resource.Kind())
			if err != nil {
				return err
			}

			proxyConfigProvider, ok := handler.(ProxyConfiguratorProvider)
			if !ok {
				uid, err := handler.GetUID(resource)
				if err != nil {
					return err
				}

				return fmt.Errorf("kind %s (for resource %s) does not support proxying", resource.Kind(), uid)
			}

			proxyConfig := proxyConfigProvider.ProxyConfigurator()
			path = proxyConfig.ProxyURL(resource.Name())
		}
	}

	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}

	url := fmt.Sprintf("http://localhost:%d%s", i.port, path)
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
