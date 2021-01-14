## kube-apiserver metrics
|  指标名   | 类型|含义  | 说明 | 
|  ---  | ---  | --- | --- | 
| apiserver_request_total |counter|  请求总数 <br> sum_by使用|按状态码`code`分布 2xx 3xx 4xx 5xx 等<br> 按动作`verb`分布 list get watch post delete等 <br> 按资源`resource`分布: pod node endpoint等|
| apiserver_request_duration_seconds_sum  |gauge| 请求延迟记录和|按动作`verb`分布 list get watch post delete等 <br> 按资源`resource`分布: pod node endpoint等|
| apiserver_request_duration_seconds_count |gauge|  请求延迟记录数|计算平均延迟: `apiserver_request_duration_seconds_sum`/`apiserver_request_duration_seconds_count` |
| apiserver_response_sizes_sum |counter|  请求响应大小记录和|
| apiserver_response_sizes_count |counter|  请求响应大小记录数|
| authentication_attempts |counter|  认证尝试数|
| authentication_duration_seconds_sum |counter|  认证耗时记录和|
| authentication_duration_seconds_count |counter|  认证耗时记录数|
| apiserver_tls_handshake_errors_total |counter|  tls握手失败计数|
| apiserver_client_certificate_expiration_seconds_sum |gauge| 证书过期时间总数|
| apiserver_client_certificate_expiration_seconds_count |gauge| 证书过期时间记录个数|
| apiserver_client_certificate_expiration_seconds_bucket |gauge| 证书过期时间分布|
| apiserver_current_inflight_requests |gauge| 该量保存了最后一个窗口中，正在处理的请求数量的高水位线|
| apiserver_current_inqueue_requests  |gauge|是一个表向量， 记录最近排队请求数量的高水位线|[apiserver请求限流](https://kubernetes.io/zh/docs/concepts/cluster-administration/flow-control/) |
| apiserver_flowcontrol_current_executing_requests |gauge| 记录包含执行中（不在队列中等待）请求的瞬时数量| APF api的QOS APIPriorityAndFairness | 
| apiserver_flowcontrol_current_inqueue_requests |gauge|  记录包含排队中的（未执行）请求的瞬时数量|
| workqueue_adds_total |counter|  wq 入队数|
| workqueue_retries_total |counter|  wq retry数|
| workqueue_longest_running_processor_seconds |gauge|  wq中最长运行时间|
| workqueue_queue_duration_seconds_sum |gauge|  wq中等待延迟记录和|
| workqueue_queue_duration_seconds_count |gauge|  wq中等待延迟记录数|
| workqueue_work_duration_seconds_sum |gauge|  wq中处理延迟记录和|
| workqueue_work_duration_seconds_count |gauge|  wq中处理延迟记录数|


## etcd metrics
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- |---| 
| etcd_db_total_size_in_bytes	| gauge|	db物理文件大小 | 
| etcd_object_counts	| gauge|	etcd对象按种类计数| 
| etcd_request_duration_seconds_sum	| gauge|	etcd请求延迟记录和| 
| etcd_request_duration_seconds_count	| gauge|	etcd请求延迟记录数| 



## kube-scheduler
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- |---| 
| scheduler_e2e_scheduling_duration_seconds_sum	| gauge|	端到端调度延迟记录和 | 
| scheduler_e2e_scheduling_duration_seconds_count	| gauge|	端到端调度延迟记录数 | 
| scheduler_pod_scheduling_duration_seconds_sum	| gauge|	调度延迟记录和 | 分析次数|
| scheduler_pod_scheduling_duration_seconds_count	| gauge|	调度延迟记录数 | 
| scheduler_pending_pods	| gauge|	调度队列pending pod数| 
| scheduler_queue_incoming_pods_total	| counter|	进入调度队列pod数| 
| scheduler_scheduling_algorithm_duration_seconds_sum	| gauge|	调度算法延迟记录和| 
| scheduler_scheduling_algorithm_duration_seconds_count	| gauge|	调度算法延迟记录数| 
| scheduler_pod_scheduling_attempts_sum	| gauge|	成功调度一个pod 的尝试次数记录和| 
| scheduler_pod_scheduling_attempts_count	| gauge|	成功调度一个pod 的尝试次数记录数| 




