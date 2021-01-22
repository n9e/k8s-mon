package config

const (
	Version                             = "0.1.0"
	COUNTER_TO_GAUGE_METRIC_NAME_SUFFIX = "_rate"
	METRIC_TYPE_GAUGE                   = "GAUGE"
	METRIC_TYPE_COUNTER                 = "COUNTER"
	FUNCNAME_APISERVER                  = "api-server"
	FUNCNAME_KUBESCHEDULER              = "kube-scheduler"
	FUNCNAME_KUBECONTROLLER             = "kube-controller-manager"
	FUNCNAME_COREDNS                    = "coredns"
	FUNCNAME_KUBELET                    = "kubelet"
	FUNCNAME_KUBELET_NODE               = "kubelet_node"
	FUNCNAME_KUBESTATSMETRICS           = "kube-stats-metrics"
	FUNCNAME_KUBEPROXY                  = "kube-proxy"
	APPENDTAG_SERVER_ADDR               = "server_addr"
	APPENDTAG_FUNC_NAME                 = "func_name"
	COLLECT_MODE_CADVISOR_PLUGIN        = "cadvisor_plugin"
	COLLECT_MODE_KUBELET_AGENT          = "kubelet_agent"
	COLLECT_MODE_SERVER_SIDE            = "server_side"
	DEFAULT_N9ENIDLABELNAME             = "N9E_NID"
)
