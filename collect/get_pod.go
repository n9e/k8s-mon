package collect

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kconfig "github.com/n9e/k8s-mon/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

func GetServerAddrAll(logger log.Logger, dataMap *HistoryMap) {
	GetServerAddrByGetPod(logger, dataMap)
	GetServerAddrByGetNode(logger, dataMap)
}

func GetServerAddrByGetPod(logger log.Logger, dataMap *HistoryMap) {
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
	pods, err := clientset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		level.Error(logger).Log("msg", "list_kube-system_pod_error", "err:", err)
		return
	}

	kubeSchedulerIps := make([]string, 0)
	kubeControllerIps := make([]string, 0)
	apiServerIps := make([]string, 0)
	coreDnsIps := make([]string, 0)
	kubeProxyIps := make([]string, 0)
	if len(pods.Items) == 0 {
		return
	}
	for _, p := range pods.Items {
		if p.Labels["tier"] == "control-plane" && p.Labels["component"] == "kube-scheduler" {
			ip := p.Status.PodIP
			if ip != "" {
				kubeSchedulerIps = append(kubeSchedulerIps, p.Status.PodIP)
			}

		}

		if p.Labels["tier"] == "control-plane" && p.Labels["component"] == "kube-controller-manager" {
			ip := p.Status.PodIP
			if ip != "" {
				kubeControllerIps = append(kubeControllerIps, p.Status.PodIP)
			}

		}

		if p.Labels["tier"] == "control-plane" && p.Labels["component"] == "kube-apiserver" {
			ip := p.Status.PodIP
			if ip != "" {
				apiServerIps = append(apiServerIps, p.Status.PodIP)
			}

		}

		if p.Labels["k8s-app"] == "kube-dns" {
			ip := p.Status.PodIP
			if ip != "" {
				coreDnsIps = append(coreDnsIps, p.Status.PodIP)
			}

		}

		if p.Labels["k8s-app"] == "kube-proxy" {
			ip := p.Status.PodIP
			if ip != "" {
				kubeProxyIps = append(kubeProxyIps, p.Status.PodIP)
			}

		}
	}
	level.Info(logger).Log("msg", "server_pod_ips_result",
		"num_kubeSchedulerIps", len(kubeSchedulerIps),
		"num_kubeControllerIps", len(kubeControllerIps),
		"num_apiServerIps", len(apiServerIps),
		"num_coreDnsIps", len(coreDnsIps),
		"num_kubeProxyIps", len(kubeProxyIps),
		"time_took_seconds", time.Since(start).Seconds(),
	)
	if len(coreDnsIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_COREDNS, coreDnsIps)
	}
	if len(apiServerIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_APISERVER, apiServerIps)
	}
	if len(kubeSchedulerIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_KUBESCHEDULER, kubeSchedulerIps)
	}
	if len(kubeControllerIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_KUBECONTROLLER, kubeControllerIps)
	}
	if len(kubeProxyIps) > 0 {
		dataMap.Map.Store(kconfig.FUNCNAME_KUBEPROXY, kubeProxyIps)
	}

}
