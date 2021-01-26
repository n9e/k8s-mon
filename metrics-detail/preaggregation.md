## 分位值
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_duration_seconds_all_quantile |gauge|  apiserver延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_get_quantile |gauge|  apiserver verb=get的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_list_quantile |gauge|  apiserver verb=list的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_request_duration_seconds_verb_delete_quantile |gauge|  apiserver verb=delete的延迟分位值|按分位值`quantile`分布 50 90 95 99 <br> |
| apiserver_response_sizes_quantile |gauge|  apiserver响应大小分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| workqueue_queue_duration_seconds_quantile |gauge| apiserver workqueue service time    延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| workqueue_work_duration_seconds_quantile |gauge| apiserver workqueue processing time   延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| etcd_request_duration_seconds_quantile |gauge| etcd延迟分位值  延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_e2e_scheduling_duration_seconds_quantile |gauge|  scheduler 端到端调度延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_pod_scheduling_duration_seconds_quantile |gauge|  scheduler  pod调度延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_scheduling_algorithm_duration_seconds_quantile |gauge|  scheduler 调度算法延迟分位值 |按分位值`quantile`分布 50 90 95 99 <br> |
| scheduler_pod_scheduling_attempts_quantile |gauge|  成功调度一个pod 的平均尝试次数分位值 |按分位值`quantile`分布 50 90 95 99 <br> |



## 平均值
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_duration_seconds_avg |gauge|  apiserver延迟平均值| |
| apiserver_response_sizes_avg |gauge|  apiserver响应大小平均值| |
| authentication_duration_seconds_avg |gauge|  apiserver认证平均耗时| |
| workqueue_queue_duration_seconds_avg |gauge|  apiserver workqueue service time    延迟平均值| |
| workqueue_work_duration_seconds_avg |gauge|  apiserver workqueue processing time  延迟平均值| |
| etcd_request_duration_seconds_avg |gauge|   etcd平均延迟| |
| scheduler_e2e_scheduling_duration_seconds_avg |gauge|   scheduler 端到端调度平均延迟| |
| scheduler_pod_scheduling_duration_seconds_avg |gauge|   scheduler pod调度平均延迟| |
| scheduler_scheduling_algorithm_duration_seconds_avg |gauge|   scheduler 调度算法平均延迟| |
| scheduler_pod_scheduling_attempts_avg |gauge|   scheduler 成功调度一个pod 的平均尝试次数| |


## 成功率
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_successful_rate |gauge| apiserver请求成功率 | |



| scheduler 端到端调度平均延迟| scheduler_e2e_scheduling_duration_seconds_sum /scheduler_e2e_scheduling_duration_seconds_count || 
| scheduler pod调度平均延迟| scheduler_pod_scheduling_duration_seconds_sum /scheduler_pod_scheduling_duration_seconds_count || 
| scheduler 调度算法平均延迟| scheduler_scheduling_algorithm_duration_seconds_sum /scheduler_scheduling_algorithm_duration_seconds_count || 
| scheduler 成功调度一个pod 的平均尝试次数| scheduler_pod_scheduling_attempts_sum /scheduler_pod_scheduling_attempts_count || 


| coredns 解析平均延迟| coredns_dns_request_duration_seconds_sum /coredns_dns_request_duration_seconds_count || 
| coredns 解析响应平均大小| coredns_dns_response_size_bytes_sum /coredns_dns_response_size_bytes_count || 