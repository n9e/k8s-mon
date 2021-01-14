## pod metrics
|  指标名   | 类型|含义  |
|  ----  | ----  | ---- |
| kube_pod_status_phase  | gauge |pod状态统计: <br>  Pending <br> Succeeded <br> Failed <br> Running <br> Unknown  |
| kube_pod_container_status_waiting	| counter|	pod处于waiting状态，值为1代表waiting | 
| kube_pod_container_status_waiting_reason  |gauge|pod处于waiting状态原因 <br> ContainerCreating <br>CrashLoopBackOff  pod启动崩溃,再次启动然后再次崩溃<br>CreateContainerConfigError<br>ErrImagePull <br>ImagePullBackOff<br>CreateContainerError<br>InvalidImageName<br> |
| kube_pod_container_status_terminated	| gauge|	pod处于terminated状态，值为1代表terminated  | 
| kube_pod_container_status_terminated_reason	| gauge|pod处于terminated状态原因 <br> 	OOMKilled <br> Completed <br> Error <br> ContainerCannotRun <br> DeadlineExceeded <br> Evicted <br>  | 
| kube_pod_container_status_restarts_total	| counter|	pod中的容器重启次数 |
| kube_pod_container_resource_requests_cpu_cores	| gauge|	pod容器cpu limit|
| kube_pod_container_resource_requests_memory_bytes	| gauge|	pod容器mem limit(单位:字节)|


## deployment  metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_deployment_status_replicas  | gauge|dep中的pod num | 
| kube_deployment_status_replicas_available  | gauge|dep中的 可用pod num | 
| kube_deployment_status_replicas_unavailable  | gauge|dep中的 不可用pod num | 

## daemonSet  metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_daemonset_status_number_available  | gauge| ds 可用数| 
| kube_daemonset_status_number_unavailable  | gauge| ds 不可用数| 
| kube_daemonset_status_number_ready  | gauge| ds ready数| 
| kube_daemonset_status_number_misscheduled  | gauge| 未经过调度运行ds的节点数| 
| kube_daemonset_status_current_number_scheduled  | gauge|ds目前运行节点数 | 
| kube_daemonset_status_desired_number_scheduled  | gauge|应该运行ds的节点数 | 


## statefulSet  metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_statefulset_status_replicas  | gauge| ss副本总数| 
| kube_statefulset_status_replicas_current  | gauge| ss当前副本数| 
| kube_statefulset_status_replicas_updated  | gauge| ss已更新副本数| 
| kube_statefulset_replicas  | gauge| ss目标副本数| 

## Job   metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_job_status_active  | gauge| job running pod数| 
| kube_job_status_succeeded  | gauge| job 成功 pod数| 
| kube_job_status_failed  | gauge| job 失败 pod数| 
| kube_job_complete  | gauge| job 是否完成| 
| kube_job_failed  | gauge| job 是否失败| 


## CronJob   metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_cronjob_status_active  | gauge| job running pod数| 
| kube_cronjob_spec_suspend   | gauge| =1代表 job 被挂起| 
| kube_cronjob_next_schedule_time   | gauge| job 下次调度时间| 
| kube_cronjob_status_last_schedule_time   | gauge| job 下次调度时间| 

## PersistentVolume    metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_persistentvolume_capacity_bytes  | gauge| pv申请大小| 
| kube_persistentvolume_status_phase   | gauge| pv状态: <br> Pending <br> Available <br> Bound <br> Released <br> Failed <br> | 

## PersistentVolumeClaim     metrics
|  指标名   |类型| 含义  |
|  ----  | ----  | ---- |
| kube_persistentvolumeclaim_resource_requests_storage_bytes  | gauge| pvc request大小| 
| kube_persistentvolumeclaim_status_phase   | gauge| pvc状态: <br> Lost <br> Bound <br> Pending  | 




# node  metrics
|  指标名   | 类型  |含义|
|  ----  | ----  | ---- |
| kube_node_status_condition  | gauge | condition: <br> NetworkUnavailable <br> MemoryPressure <br> DiskPressure <br> PIDPressure <br> Ready|
| kube_node_status_allocatable_cpu_cores  | gauge | 节点可以分配cpu核数|
| kube_node_status_allocatable_memory_bytes  | gauge | 节点可以分配内存总量(单位：字节)|
| kube_node_spec_taint  | gauge | 节点污点情况|
| kube_node_status_capacity_memory_bytes  | gauge | 节点内存总量(单位：字节)|
| kube_node_status_capacity_cpu_cores  | gauge | 节点cpu核数|
| kube_node_status_capacity_pods  | gauge | 节点可运行的pod总数|


