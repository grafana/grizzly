package grafana

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"io"
	"net/http"
	"os"
	"time"
)

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func initClient(ctx context.Context) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "grafana/grafana:8.0.4", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)
	return cli, err
}

func startContainer(err error, cli *client.Client, ctx context.Context) string {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "grafana/grafana:8.0.4",
		Env: []string{
			"GF_PATHS_CONFIG=" + getCurrentDir() + "/testdata/custom.ini",
			"GF_PATHS_PROVISIONING=" + getCurrentDir() + "/grafana/provisioning",
		},
		ExposedPorts: nat.PortSet{"3000/tcp": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{"3000/tcp": {{HostIP: "0.0.0.0", HostPort: "3000"}}},
		AutoRemove:   true,
		NetworkMode:  "host",
	}, nil, nil, "grafana")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	return resp.ID
}

func pingLocalhost() *time.Ticker {
	ticker := time.NewTicker(1 * time.Second)
	timeoutExceeded := time.After(120 * time.Second)

	success := false
	for !success {
		select {
		case <-timeoutExceeded:
			fmt.Println("failed")

		case <-ticker.C:
			resp, _ := http.Get("http://0.0.0.0:3000/")
			if resp != nil {
				success = true
				break
			}
		}
	}
	return ticker
}

func removeContainer(cli *client.Client, ctx context.Context, containerID string) {
	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: false, Force: true}); err != nil {
		fmt.Printf("fail to kill container, err: %q", err)
		os.Exit(1)
	}
}

func printContainerLogs(cli *client.Client, ctx context.Context, containerID string) {
	out, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
