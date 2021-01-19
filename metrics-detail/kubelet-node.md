## kubelet-node
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| kubelet_running_pods	| gauge |	运行的pod数 |   |  
| kubelet_running_containers	| gauge |	运行的容器数 |  container_state: <br> created <br> exited<br>running <br> unknown<br> |  
| volume_manager_total_volumes	| gauge|	volume数 |  plugin_name:<br> kubernetes.io/configmap <br> kubernetes.io/host-path <br> kubernetes.io/secret <br> state=[desired_state_of_world actual_state_of_world]  |  
| kubelet_node_config_error	| gauge|	=1节点配置错误 | |  
| kubelet_runtime_operations_total	| counter|	操作速率 | operation_type: <br> create_container <br> pull_image  <br> stop_container 等   |  
| kubelet_runtime_operations_errors_total	| counter|	错误操作速率 | operation_type: <br> create_container <br> pull_image  <br> stop_container 等   |  
| kubelet_runtime_operations_duration_seconds_sum	| counter|	操作延迟记录和 |   |  
| kubelet_runtime_operations_duration_seconds_count	| counter|	操作延迟记录数 |   |  
| storage_operation_duration_seconds_count	| counter|	存储相关操作延迟记录数 |     |  
| storage_operation_duration_seconds_sum	| counter|	存储相关操作延迟记录和 |      |  
| kubelet_cgroup_manager_duration_seconds_count	| counter|	cg相关操作请求延迟记录数 |     |  
| kubelet_cgroup_manager_duration_seconds_sum	| counter|	cg相关操作请求延迟记录数 |     |  
| rest_client_requests_total	| counter|	kube-api请求速率 |  code <br> method <br> host    |   |
| rest_client_request_latency_seconds_count	| counter|	kube-api请求延迟记录数 |  code <br> method <br> host    |  
| rest_client_request_latency_seconds_sum	| counter|	kube-api请求延迟记录和 |  code <br> method <br> host    |