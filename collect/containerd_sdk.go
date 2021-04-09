package collect

import (
	"context"
	"errors"
	"github.com/containerd/containerd"
)

func getLabelMapByContainerdSdk(nidLabelName string) (map[string]map[string]string, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace("k8s.io"))

	if err != nil {
		return nil, err
	}

	defer client.Close()

	context := context.Background()
	cers, err := client.Containers(context)
	if err != nil {
		return nil, err
	}
	if len(cers) == 0 {
		return nil, errors.New("got zero containers on this node")
	}

	insM := make(map[string]map[string]string)
	whiteKeyM := map[string]struct{}{
		nidLabelName: {},
	}

	for _, c := range cers {
		labels, err := c.Labels(context)
		if err != nil {
			continue
		}

		podName := labels["io.kubernetes.pod.name"]
		if podName == "" {
			continue
		}

		lastM, loaded := insM[podName]
		if !loaded {
			lastM = make(map[string]string)
		}

		for k, v := range labels {

			if _, found := whiteKeyM[k]; !found {
				continue
			}
			lastM[k] = v
		}

		insM[podName] = lastM

	}
	return insM, nil

}
