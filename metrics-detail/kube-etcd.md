## etcd
|  指标名   | 类型|含义  | 说明| 
|  ----  | ----  | ---- | ---- |
| etcd_mvcc_db_total_size_in_bytes	| gauge|	当前db文件大小 |   |  
| etcd_mvcc_db_total_size_in_use_in_bytes	| gauge|	逻辑db文件大小 |   |  
| etcd_server_quota_backend_bytes	| gauge|	db quto大小|     |  
| etcd_debugging_mvcc_keys_total	| gauge|	总key数 |     |  
| etcd_server_id	| gauge|	member  |     |  
| etcd_disk_wal_fsync_duration_seconds	| counter|	wal fsync延迟|     |  
| etcd_disk_backend_commit_duration_seconds	| counter|	db sync延迟|     |  
| etcd_server_is_leader	| gauge|	是否是leader|     |  
| etcd_debugging_mvcc_db_compaction_keys_total	| counter|	压实的key | 
| etcd_server_has_leader	| gauge|	是否有leader | 
| etcd_network_client_grpc_received_bytes_total	| counter|	client入流量 | 
| etcd_network_client_grpc_sent_bytes_total	| counter|	client出流量 | 
| etcd_debugging_mvcc_put_total	| counter|	put key qps | 
| etcd_debugging_mvcc_delete_total	| counter|	delete key qps | 
| etcd_server_slow_apply_total	| counter|	慢动作 | 
| etcd_server_slow_read_indexes_total	| counter|	慢索引动作 | 
| etcd_server_leader_changes_seen_total	| counter| 选主 | 
| etcd_server_heartbeat_send_failures_total	| counter| heartbeat  failures | 
| etcd_server_health_failures	| counter| 健康检测失败 | 
| etcd_server_proposals_failed_total	| counter| 失败提案| 
| etcd_server_proposals_pending	| counter| pending提案| 
| etcd_server_proposals_committed_total	| counter|commit提案| 
| etcd_server_proposals_applied_total	| counter|apply提案| 
| grpc_server_started_total	| counter| grpc qps| 
| grpc_server_handled_total	| counter| 已完成grpc qps| 
| etcd_debugging_snap_save_total_duration_seconds	| gauge|	snapshot save延迟 | 