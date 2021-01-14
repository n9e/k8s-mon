## kube-controller-manager
- 参考dashboard [grafana kube-controller-manager](https://grafana.com/grafana/dashboards/12122)

|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| workqueue_adds_total	| counter|	工作队列新增速率 |   |  
| workqueue_depth	| gauge|	当前工作队列深度 |   |  
| workqueue_queue_duration_seconds_sum	| gauge|	wq中等待延迟记录和 |     |  
| workqueue_queue_duration_seconds_count	| gauge|	wq中等待延迟记录数 |     |  
| rest_client_requests_total	| counter|	kube-api请求速率 |  code <br> method <br> host    |  
| rest_client_request_duration_seconds_sum	| counter|	请求延迟记录和 |     |  
| rest_client_request_duration_seconds_count	| counter|	请求延迟记录数 |     |  
