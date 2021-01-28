## 分位值
- 分位值没有数据的是因为没有请求，对应的bucket没有增长
- 如prometheus实时计算也会没有值   `histogram_quantile(0.90, sum(rate(scheduler_e2e_scheduling_duration_seconds_bucket[5m])) by (le))`
- 分位值metric_name为 将_bucket替换为_quantile
- 平均值metric_name为 将_bucket替换为_avg



|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_duration_seconds_all_quantile |gauge|  apiserver延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_get_quantile |gauge|  apiserver verb=get的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_list_quantile |gauge|  apiserver verb=list的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_delete_quantile |gauge|  apiserver verb=delete的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_response_sizes_quantile |gauge|  apiserver响应大小分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_workqueue_queue_duration_seconds_quantile |gauge| apiserver workqueue 项目在队列中等待延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_workqueue_work_duration_seconds_quantile |gauge| apiserver workqueue 消费一个项目延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_rest_client_request_duration_seconds_quantile |gauge| 请求apiserver 的延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_workqueue_queue_duration_seconds_quantile |gauge| scheduler workqueue 项目在队列中等待延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_workqueue_work_duration_seconds_quantile |gauge| scheduler workqueue 消费一个项目延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_rest_client_request_duration_seconds_quantile |gauge| scheduler请求apiserver 的延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_e2e_scheduling_duration_seconds_quantile |gauge|  scheduler 端到端调度延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_pod_scheduling_duration_seconds_quantile |gauge|  scheduler  pod调度延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_scheduling_algorithm_duration_seconds_quantile |gauge|  scheduler 调度算法延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_pod_scheduling_attempts_quantile |gauge|  成功调度一个pod 的平均尝试次数分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| controller_manager_workqueue_queue_duration_seconds_quantile |gauge| controller_manager workqueue 项目在队列中等待延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| controller_manager_workqueue_work_duration_seconds_quantile |gauge| controller_manager workqueue 消费一个项目延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| controller_manager_rest_client_request_duration_seconds_quantile |gauge| controller_manager请求apiserver 的延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| etcd_request_duration_seconds_quantile |gauge| apiserver 请求etcd延迟分位值  延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| coredns_dns_request_duration_seconds_quantile |gauge|  coredns 解析平均延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| coredns_dns_response_size_bytes_quantile |gauge|  coredns 解析响应大小分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| etcd_disk_wal_fsync_duration_seconds_quantile |gauge| wal fsync延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| etcd_disk_backend_commit_duration_seconds_quantile |gauge| db sync延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| etcd_debugging_snap_save_total_duration_seconds_quantile |gauge| dsnapshot save延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |





## 平均值
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_duration_seconds_avg |gauge|  apiserver延迟平均值| |
| apiserver_response_sizes_avg |gauge|  apiserver响应大小平均值| |
| authentication_duration_seconds_avg |gauge|  apiserver认证平均耗时| |
| rest_client_request_duration_seconds_avg |gauge|   请求apiserver的延迟平均值| 按func_name分布|
| workqueue_queue_duration_seconds_avg |gauge|   项目在队列中等待延迟平均值| 按func_name分布|
| workqueue_work_duration_seconds_avg |gauge|   消费一个项目延迟平均值| 按func_name分布|
| etcd_request_duration_seconds_avg |gauge|   apiserver请求etcd平均延迟| |
| scheduler_e2e_scheduling_duration_seconds_avg |gauge|   scheduler 端到端调度平均延迟| |
| scheduler_pod_scheduling_duration_seconds_avg |gauge|   scheduler pod调度平均延迟| |
| scheduler_scheduling_algorithm_duration_seconds_avg |gauge|   scheduler 调度算法平均延迟| |
| scheduler_pod_scheduling_attempts_avg |gauge|   scheduler 成功调度一个pod 的平均尝试次数| |
| coredns_dns_request_duration_seconds_avg |gauge|   coredns 解析平均延迟平均值| |
| coredns_dns_response_size_bytes_avg |gauge|   coredns  coredns 解析响应大小平均值| |
| etcd_disk_wal_fsync_duration_seconds_avg |gauge|   wal fsync延迟平均值| |
| etcd_disk_backend_commit_duration_seconds_avg |gauge|  db sync延迟平均值| |
| etcd_debugging_snap_save_total_duration_seconds_avg |gauge|  dsnapshot save延迟平均值| |



## 节点资源汇总
- 节点cpu请求核数 `sum(kube_pod_container_resource_requests_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"}) by (node)`
- 节点cpu 请求率 `sum(kube_pod_container_resource_requests_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node) / sum(kube_node_status_allocatable_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node)`
- 节点cpu限制 `sum(kube_pod_container_resource_limits_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"}) by (node)`
- 节点cpu限制率 `sum(kube_pod_container_resource_limits_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node) / sum(kube_node_status_allocatable_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node)`

- 节点内存请求 `sum(kube_pod_container_resource_requests_memory_bytes{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"}) by (node)`
- 节点内存请求% `sum(kube_pod_container_resource_requests_memory_bytes{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node) / sum(kube_node_status_capacity_memory_bytes{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node)`
- 节点内存限制 `sum(kube_pod_container_resource_limits_memory_bytes{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"}) by (node)`
- 节点内存限制% `sum(kube_pod_container_resource_limits_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node) / sum(kube_node_status_allocatable_cpu_cores{origin_prometheus=~"$origin_prometheus",node=~"^$Node$"})by (node)`




## 成功率/百分比
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_successful_rate |gauge| apiserver请求成功率 | |
| kube_node_pod_container_cpu_limits_value |gauge|节点cpu限制 | |
| kube_node_pod_container_cpu_limits_percent|gauge|节点cpu限制率 | |
| kube_node_pod_container_cpu_requests_value|gauge|节点cpu 请求| |
| kube_node_pod_container_cpu_requests_percent|gauge|节点cpu 请求率 | `sum(kube_pod_container_resource_requests_cpu_cores)by (node) / sum(kube_node_status_allocatable_cpu_cores)by (node)`|
| kube_node_pod_container_memory_requests_value|gauge|节点内存请求| |
| kube_node_pod_container_memory_requests_percent|gauge|节点内存请求率| |
| kube_node_pod_container_memory_limits_value|gauge|节点内存限制| |
| kube_node_pod_container_memory_limits_percent|gauge|节点内存限制率| `sum(kube_pod_container_resource_limits_cpu_cores)by (node) / sum(kube_node_status_allocatable_cpu_cores)by (node)` |