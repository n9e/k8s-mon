## coredns
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| coredns_dns_requests_total	| counter|	解析请求数 | A记录 <br>  AAAA记录 <br> other记录  |  
| coredns_dns_responses_total	| counter|	解析响应数 | NOERROR <br> NXDOMAIN <br> REFUSED  |  
| coredns_cache_entries	| gauge|	缓存记录数 | 成功或失败    |  
| coredns_cache_hits_total	| counter|	缓存命中数 | 成功或失败    |  
| coredns_cache_misses_total	| counter|	缓存未命中数 | 成功或失败    |  
| coredns_dns_request_duration_seconds_sum	| gauge|	解析延迟记录和 | 
| coredns_dns_request_duration_seconds_count	| gauge|	解析延迟记录数 | 
| coredns_dns_response_size_bytes_sum	| gauge|	解析响应大小记录和 | 
| coredns_dns_response_size_bytes_count	| gauge|	解析响应大小记录数 | 
