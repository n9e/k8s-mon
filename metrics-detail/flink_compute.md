## flink聚合指标
|  计算指标名   | 计算表达式 | 说明 | 
|  ---  | ---  | --- |
| apiserver 请求成功率| sum(apiserver_request_total{code="200"}) / sum (apiserver_request_total) || 
| apiserver 请求平均延迟| apiserver_request_duration_seconds_sum /apiserver_request_duration_seconds_count || 
| apiserver 响应平均大小| apiserver_response_sizes_sum /apiserver_response_sizes_count || 
| apiserver 认证平均耗时| authentication_duration_seconds_sum /authentication_duration_seconds_count || 
| apiserver 证书平均过期时间| apiserver_client_certificate_expiration_seconds_sum /apiserver_client_certificate_expiration_seconds_count || 
| apiserver wq中等待平均延迟| workqueue_queue_duration_seconds_sum /workqueue_queue_duration_seconds_count || 
| apiserver wq中处理平均延迟| workqueue_work_duration_seconds_sum /workqueue_work_duration_seconds_count || 
| etcd请求平均延迟| etcd_request_duration_seconds_sum /etcd_request_duration_seconds_count || 
| scheduler 端到端调度平均延迟| scheduler_e2e_scheduling_duration_seconds_sum /scheduler_e2e_scheduling_duration_seconds_count || 
| scheduler pod调度平均延迟| scheduler_pod_scheduling_duration_seconds_sum /scheduler_pod_scheduling_duration_seconds_count || 
| scheduler 调度算法平均延迟| scheduler_scheduling_algorithm_duration_seconds_sum /scheduler_scheduling_algorithm_duration_seconds_count || 
| scheduler 成功调度一个pod 的平均尝试次数| scheduler_pod_scheduling_attempts_sum /scheduler_pod_scheduling_attempts_count || 
| coredns 解析平均延迟| coredns_dns_request_duration_seconds_sum /coredns_dns_request_duration_seconds_count || 
| coredns 解析响应平均大小| coredns_dns_response_size_bytes_sum /coredns_dns_response_size_bytes_count || 
| kube-controller-manager 请求apiserver平均延迟| rest_client_request_duration_seconds_sum /rest_client_request_duration_seconds_count || 
| kube-proxy 请求apiserver平均延迟| rest_client_request_latency_seconds_sum /rest_client_request_latency_seconds_count || 
| kube-proxy 网络规则同步平均延迟| kubeproxy_sync_proxy_rules_duration_seconds_sum /kubeproxy_sync_proxy_rules_duration_seconds_count || 
| kube-proxy 网络处理平均延迟| kubeproxy_network_programming_duration_seconds_sum /kubeproxy_network_programming_duration_seconds_count || 
| kubelet-node 操作处理平均延迟| kubelet_runtime_operations_duration_seconds_sum /kubelet_runtime_operations_duration_seconds_count || 
| kubelet-node 存储操作处理平均延迟| storage_operation_duration_seconds_sum /storage_operation_duration_seconds_count || 
| kubelet-node cg操作处理平均延迟| kubelet_cgroup_manager_duration_seconds_sum /kubelet_cgroup_manager_duration_seconds_count || 



