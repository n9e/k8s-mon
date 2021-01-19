## kube-proxy
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| kubeproxy_sync_proxy_rules_duration_seconds_count	| counter|	网络规则同步记录数 |   |  
| kubeproxy_sync_proxy_rules_duration_seconds_sum	| counter|	网络规则同步记录和 |   |  
| kubeproxy_network_programming_duration_seconds_count	| counter|	网络处理记录数 |   |  
| kubeproxy_network_programming_duration_seconds_sum	| counter|	网络处理记录和 |   |  
| rest_client_requests_total	| counter|	kube-api请求速率 |  code <br> method <br> host    |  
| rest_client_request_latency_seconds_count	| counter|	kube-api请求延迟记录数 |  code <br> method <br> host    |  
| rest_client_request_latency_seconds_sum	| counter|	kube-api请求延迟记录和 |  code <br> method <br> host    |