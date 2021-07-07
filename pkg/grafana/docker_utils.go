package grafana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
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

func startContainer(cli *client.Client, ctx context.Context) string {
	exposedPorts, portBindings, _ := nat.ParsePortSpecs([]string{"3000:3000"})
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "grafana/grafana:8.0.4",
		Env: []string{
			"GF_PATHS_CONFIG=" + getCurrentDir() + "/testdata/custom.ini",
			"GF_PATHS_PROVISIONING=" + getCurrentDir() + "/grafana/provisioning",
		},
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   true,
		NetworkMode:  "host",
	}, nil, nil, "grafana")
	if err != nil {
		panic(err)
	}
	fmt.Println("Container ID:", resp.ID)

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
			panic("Unable to connect to grizzly-grafana:3000")

		case <-ticker.C:
			resp, _ := http.Get("http://grizzly-grafana:3000/")
			fmt.Println("Response:", resp)
			if resp != nil {
				success = true
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
