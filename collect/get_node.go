package collect

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	kconfig "github.com/n9e/k8s-mon/config"
)

func GetServerAddrByGetNode(logger log.Logger, dataMap *HistoryMap) {
	start := time.Now()
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		level.Error(logger).Log("msg", "create_k8s_InClusterConfig_error", "err:", err)
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		level.Error(logger).Log("msg", "creates_the_clientset_error", "err:", err)
		return
	}
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		level.Error(logger).Log("msg", "list_kube-system_pod_error", "err:", err)
		return
	}

	nodeIps := make([]string, 0)
	if len(nodes.Items) == 0 {
		return
	}
	for _, p := range nodes.Items {
		addr := p.Status.Addresses
		if len(addr) == 0 {
			continue
		}
		for _, a := range addr {
			if a.Type == apiv1.NodeInternalIP {
				nodeIps = append(nodeIps, a.Address)
			}
		}
	}
	level.Info(logger).Log("msg", "server_node_ips_result",
		"num_nodeIps", len(nodeIps),
		"time_took_seconds", time.Since(start).Seconds(),
	)
	if len(nodeIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_KUBELET_NODE, nodeIps)
	}

}
