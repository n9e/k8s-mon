package collect

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func getLabelMapByDockerSdk(nidLabelName string) (map[string]map[string]string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	insM := make(map[string]map[string]string)
	whiteKeyM := map[string]struct{}{
		nidLabelName: {},
	}
	for _, container := range containers {
		podName := container.Labels["io.kubernetes.pod.name"]

		if podName == "" {
			continue
		}

		lastM, loaded := insM[podName]
		if !loaded {
			lastM = make(map[string]string)
		}

		for k, v := range container.Labels {

			if _, found := whiteKeyM[k]; !found {
				continue
			}
			lastM[k] = v
		}

		insM[podName] = lastM

	}
	return insM, nil
}
