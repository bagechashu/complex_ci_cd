-- Initial sample data for Release Control System
-- This file contains comprehensive test data for all tables in the schema

-- 1. Insert sample applications (15 records)
INSERT OR IGNORE INTO application (name, image_name, git_repo, build_type, created_at, updated_at) VALUES 
('api-service', 'api-service', 'git@github.com:company/api-service.git', 'docker', datetime('now'), datetime('now')),
('web-ui', 'web-ui', 'git@github.com:company/web-ui.git', 'docker', datetime('now'), datetime('now')),
('data-processor', 'data-processor', 'git@github.com:company/data-processor.git', 'docker', datetime('now'), datetime('now')),
('auth-service', 'auth-service', 'git@github.com:company/auth-service.git', 'docker', datetime('now'), datetime('now')),
('notification-service', 'notification-service', 'git@github.com:company/notification-service.git', 'docker', datetime('now'), datetime('now')),
('payment-gateway', 'payment-gateway', 'git@github.com:company/payment-gateway.git', 'docker', datetime('now'), datetime('now')),
('analytics-engine', 'analytics-engine', 'git@github.com:company/analytics-engine.git', 'docker', datetime('now'), datetime('now')),
('cache-service', 'cache-service', 'git@github.com:company/cache-service.git', 'docker', datetime('now'), datetime('now')),
('search-service', 'search-service', 'git@github.com:company/search-service.git', 'docker', datetime('now'), datetime('now')),
('report-generator', 'report-generator', 'git@github.com:company/report-generator.git', 'docker', datetime('now'), datetime('now')),
('logging-service', 'logging-service', 'git@github.com:company/logging-service.git', 'docker', datetime('now'), datetime('now')),
('metric-collector', 'metric-collector', 'git@github.com:company/metric-collector.git', 'docker', datetime('now'), datetime('now')),
('storage-service', 'storage-service', 'git@github.com:company/storage-service.git', 'docker', datetime('now'), datetime('now')),
('migration-service', 'migration-service', 'git@github.com:company/migration-service.git', 'docker', datetime('now'), datetime('now')),
('backup-manager', 'backup-manager', 'git@github.com:company/backup-manager.git', 'docker', datetime('now'), datetime('now'));

-- 2. Insert sample environments
INSERT OR IGNORE INTO environment (name, rank, created_at, updated_at) VALUES 
('development', 1, datetime('now'), datetime('now')),
('staging', 2, datetime('now'), datetime('now')),
('production', 3, datetime('now'), datetime('now'));

-- 3. Insert sample clusters
INSERT OR IGNORE INTO cluster (name, type, environment, registry_prefix, created_at, updated_at) VALUES 
('k8s-dev', 'kubernetes', 'dev', 'docker.io/company', datetime('now'), datetime('now')),
('k8s-staging', 'kubernetes', 'staging', 'docker.io/company', datetime('now'), datetime('now')),
('k8s-prod-cn1', 'kubernetes', 'production', 'docker.io/company', datetime('now'), datetime('now')),
('k8s-prod-cn2', 'kubernetes', 'production', 'docker.io/company', datetime('now'), datetime('now'));

-- 4. Insert workload targets (app-env-cluster mappings)
INSERT OR IGNORE INTO workload_target (app_id, env_id, cluster_id, k8s_namespace, k8s_workload, container_name, workload_type, workload_name, created_at, updated_at) VALUES 
-- api-service
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'api-service', 'api', 'Deployment', 'api-service-deployment', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'api-service', 'api', 'Deployment', 'api-service-deployment', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'api-service', 'api-prod', 'Deployment', 'api-service-prod', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn2'), 'production', 'api-service', 'api-prod', 'Deployment', 'api-service-prod', datetime('now'), datetime('now')),
-- web-ui
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'web-ui', 'web', 'Deployment', 'web-ui-deployment', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'web-ui', 'web', 'Deployment', 'web-ui-deployment', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'web-ui', 'web-prod', 'Deployment', 'web-ui-prod', datetime('now'), datetime('now')),
-- data-processor
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'data-processor', 'processor', 'StatefulSet', 'data-processor-statefulset', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'data-processor', 'processor', 'StatefulSet', 'data-processor-statefulset', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'data-processor', 'processor-prod', 'StatefulSet', 'data-processor-prod', datetime('now'), datetime('now'));

-- 5. Insert sample release records (15 records)
INSERT OR IGNORE INTO release_record (app_id, env_id, cluster_id, image, status, previous_image, triggered_by, started_at, completed_at, created_at, updated_at) VALUES 
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'api-service:v1.2.3', 'success', 'api-service:v1.2.2', 'user@example.com', datetime('now', '-15 days'), datetime('now', '-15 days', '+5 minutes'), datetime('now', '-15 days'), datetime('now', '-15 days')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'api-service:v1.2.3', 'success', 'api-service:v1.2.2', 'CI/CD Pipeline', datetime('now', '-14 days'), datetime('now', '-14 days', '+5 minutes'), datetime('now', '-14 days'), datetime('now', '-14 days')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'api-service:v1.2.3', 'deploying', 'api-service:v1.2.1', 'admin@example.com', datetime('now', '-13 days'), NULL, datetime('now', '-13 days'), datetime('now', '-13 days')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'web-ui:v2.1.0', 'success', 'web-ui:v2.0.9', 'developer@example.com', datetime('now', '-12 days'), datetime('now', '-12 days', '+5 minutes'), datetime('now', '-12 days'), datetime('now', '-12 days')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'web-ui:v2.1.0', 'failed', 'web-ui:v2.0.8', 'CI/CD Pipeline', datetime('now', '-11 days'), datetime('now', '-11 days', '+10 minutes'), datetime('now', '-11 days'), datetime('now', '-11 days')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'data-processor:v3.5.0', 'success', 'data-processor:v3.4.9', 'scheduler@system', datetime('now', '-10 days'), datetime('now', '-10 days', '+8 minutes'), datetime('now', '-10 days'), datetime('now', '-10 days')),
((SELECT id FROM application WHERE name='auth-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'auth-service:v1.0.5', 'success', 'auth-service:v1.0.4', 'developer@example.com', datetime('now', '-9 days'), datetime('now', '-9 days', '+3 minutes'), datetime('now', '-9 days'), datetime('now', '-9 days')),
((SELECT id FROM application WHERE name='auth-service'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'auth-service:v1.0.5', 'success', 'auth-service:v1.0.4', 'CI/CD Pipeline', datetime('now', '-8 days'), datetime('now', '-8 days', '+4 minutes'), datetime('now', '-8 days'), datetime('now', '-8 days')),
((SELECT id FROM application WHERE name='notification-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'notification-service:v2.3.1', 'success', 'notification-service:v2.3.0', 'user@example.com', datetime('now', '-7 days'), datetime('now', '-7 days', '+6 minutes'), datetime('now', '-7 days'), datetime('now', '-7 days')),
((SELECT id FROM application WHERE name='payment-gateway'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'payment-gateway:v1.5.0', 'success', 'payment-gateway:v1.4.9', 'admin@example.com', datetime('now', '-6 days'), datetime('now', '-6 days', '+7 minutes'), datetime('now', '-6 days'), datetime('now', '-6 days')),
((SELECT id FROM application WHERE name='analytics-engine'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'analytics-engine:v3.2.1', 'success', 'analytics-engine:v3.2.0', 'developer@example.com', datetime('now', '-5 days'), datetime('now', '-5 days', '+5 minutes'), datetime('now', '-5 days'), datetime('now', '-5 days')),
((SELECT id FROM application WHERE name='cache-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'cache-service:v1.1.0', 'success', 'cache-service:v1.0.9', 'user@example.com', datetime('now', '-4 days'), datetime('now', '-4 days', '+4 minutes'), datetime('now', '-4 days'), datetime('now', '-4 days')),
((SELECT id FROM application WHERE name='search-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn2'), 'search-service:v2.0.0', 'success', 'search-service:v1.9.9', 'CI/CD Pipeline', datetime('now', '-3 days'), datetime('now', '-3 days', '+9 minutes'), datetime('now', '-3 days'), datetime('now', '-3 days')),
((SELECT id FROM application WHERE name='report-generator'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'report-generator:v1.8.2', 'failed', 'report-generator:v1.8.1', 'developer@example.com', datetime('now', '-2 days'), datetime('now', '-2 days', '+12 minutes'), datetime('now', '-2 days'), datetime('now', '-2 days')),
((SELECT id FROM application WHERE name='logging-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'logging-service:v2.1.0', 'success', 'logging-service:v2.0.9', 'scheduler@system', datetime('now', '-1 days'), datetime('now', '-1 days', '+6 minutes'), datetime('now', '-1 days'), datetime('now', '-1 days'));

-- 6. Insert release events
INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'started', 'Release workload initialized', datetime(started_at, '+1 minute') FROM release_record WHERE status IN ('success', 'deploying', 'failed') LIMIT 15;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'deploying', 'Pulling image from registry', datetime(started_at, '+2 minutes') FROM release_record WHERE status IN ('success', 'deploying', 'failed') LIMIT 15;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'deploying', 'Applying Kubernetes manifests', datetime(started_at, '+3 minutes') FROM release_record WHERE status IN ('success', 'deploying') LIMIT 14;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'success', 'Workload completed successfully', datetime(started_at, '+5 minutes') FROM release_record WHERE status = 'success' LIMIT 13;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'failed', 'Image pull failed', datetime(started_at, '+2 minutes') FROM release_record WHERE status = 'failed' LIMIT 2;

-- 7. Insert audit logs
INSERT OR IGNORE INTO audit_log (user_id, operation, resource_type, resource_id, created_at) VALUES 
('admin@example.com', 'CREATE', 'application', (SELECT id FROM application WHERE name='api-service'), datetime('now', '-15 days')),
('admin@example.com', 'CREATE', 'environment', (SELECT id FROM environment WHERE name='production'), datetime('now', '-14 days')),
('developer@example.com', 'DEPLOY', 'release', 1, datetime('now', '-13 days')),
('CI/CD Pipeline', 'AUTO_DEPLOY', 'release', 2, datetime('now', '-12 days')),
('admin@example.com', 'UPDATE', 'workload_target', 1, datetime('now', '-11 days')),
('developer@example.com', 'VIEW', 'release_history', 1, datetime('now', '-10 days')),
('admin@example.com', 'DELETE', 'application', 0, datetime('now', '-9 days')),
('scheduler@system', 'AUTO_DEPLOY', 'release', 6, datetime('now', '-8 days')),
('developer@example.com', 'CREATE', 'workload_target', 5, datetime('now', '-7 days')),
('admin@example.com', 'UPDATE', 'cluster', 1, datetime('now', '-6 days')),
('user@example.com', 'DEPLOY', 'release', 7, datetime('now', '-5 days')),
('CI/CD Pipeline', 'AUTO_DEPLOY', 'release', 10, datetime('now', '-4 days')),
('developer@example.com', 'VIEW', 'workload_history', 2, datetime('now', '-3 days')),
('admin@example.com', 'CREATE', 'application', 5, datetime('now', '-2 days')),
('scheduler@system', 'AUTO_DEPLOY', 'release', 15, datetime('now', '-1 days'));

-- 8. Insert application cluster configurations
INSERT OR IGNORE INTO application_cluster_config (id, application_id, cluster_id, labels, created_at, updated_at) VALUES 
('app-1-cluster-1', (SELECT id FROM application WHERE name='api-service'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'env=dev,tier=backend,owner=backend-team', datetime('now'), datetime('now')),
('app-1-cluster-2', (SELECT id FROM application WHERE name='api-service'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'env=staging,tier=backend,owner=backend-team', datetime('now'), datetime('now')),
('app-1-cluster-3', (SELECT id FROM application WHERE name='api-service'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'env=prod,tier=backend,zone=cn1,critical=true,owner=backend-team', datetime('now'), datetime('now')),
('app-1-cluster-4', (SELECT id FROM application WHERE name='api-service'), (SELECT id FROM cluster WHERE name='k8s-prod-cn2'), 'env=prod,tier=backend,zone=cn2,critical=true,owner=backend-team', datetime('now'), datetime('now')),
('app-2-cluster-1', (SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'env=dev,tier=frontend,owner=frontend-team', datetime('now'), datetime('now')),
('app-2-cluster-2', (SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'env=staging,tier=frontend,owner=frontend-team', datetime('now'), datetime('now')),
('app-2-cluster-3', (SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'env=prod,tier=frontend,zone=cn1,critical=true,owner=frontend-team', datetime('now'), datetime('now')),
('app-3-cluster-1', (SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'env=dev,tier=processing,owner=data-team', datetime('now'), datetime('now')),
('app-3-cluster-2', (SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'env=staging,tier=processing,owner=data-team', datetime('now'), datetime('now')),
('app-3-cluster-3', (SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'env=prod,tier=processing,zone=cn1,critical=true,owner=data-team', datetime('now'), datetime('now'));

-- 9. Insert shell servers (5 servers)
INSERT OR IGNORE INTO shell_server (name, host, port, username, auth_type, password, private_key, status, last_connected, created_at, updated_at) VALUES 
('dev-server-01', '192.168.1.10', 22, 'deploy', 'password', 'dev_password_123', NULL, 'active', datetime('now', '-1 hours'), datetime('now'), datetime('now')),
('staging-server-01', '192.168.2.10', 22, 'deploy', 'key', NULL, '-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2Z2wZ...truncated\n-----END RSA PRIVATE KEY-----', 'active', datetime('now', '-2 hours'), datetime('now'), datetime('now')),
('prod-server-cn1-01', '10.0.1.10', 22, 'deploy', 'key', NULL, '-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA3a3xY...truncated\n-----END RSA PRIVATE KEY-----', 'active', datetime('now', '-30 minutes'), datetime('now'), datetime('now')),
('prod-server-cn1-02', '10.0.1.11', 22, 'deploy', 'key', NULL, '-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4b4yZ...truncated\n-----END RSA PRIVATE KEY-----', 'active', datetime('now', '-45 minutes'), datetime('now'), datetime('now')),
('prod-server-cn2-01', '10.0.2.10', 22, 'deploy', 'key', NULL, '-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA5c5zZ...truncated\n-----END RSA PRIVATE KEY-----', 'active', datetime('now', '-1 hours'), datetime('now'), datetime('now'));

-- 10. Insert shell commands (10 commands)
INSERT OR IGNORE INTO shell_command (server_id, command, description, is_published, created_at, updated_at) VALUES 
((SELECT id FROM shell_server WHERE name='dev-server-01'), 'systemctl status api-service', 'Check api-service status on dev server', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='dev-server-01'), 'docker logs -f api-service', 'View api-service logs in real-time', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='staging-server-01'), 'systemctl restart web-ui', 'Restart web-ui service on staging', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='staging-server-01'), 'kubectl get pods -n staging', 'List all pods in staging namespace', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), 'systemctl status payment-gateway', 'Check payment-gateway status', 0, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), 'kubectl get deployments -n production', 'List all deployments in prod', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn1-02'), 'df -h', 'Check disk usage', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn1-02'), 'free -m', 'Check memory usage', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn2-01'), 'uptime', 'Check server uptime', 1, datetime('now'), datetime('now')),
((SELECT id FROM shell_server WHERE name='prod-server-cn2-01'), 'ps aux | grep java', 'Check running Java processes', 1, datetime('now'), datetime('now'));

-- 11. Insert shell tasks (6 tasks)
INSERT OR IGNORE INTO shell_task (name, description, command_id, execution_method, requires_approval, created_at, updated_at) VALUES 
('health-check-dev', 'Perform health check on dev cluster', (SELECT id FROM shell_command WHERE description LIKE 'Check api-service%' LIMIT 1), 'parallel', 0, datetime('now'), datetime('now')),
('restart-staging-web', 'Restart web-ui on staging server', (SELECT id FROM shell_command WHERE description='Restart web-ui service on staging'), 'serial', 0, datetime('now'), datetime('now')),
('prod-list-deployments', 'List all production deployments', (SELECT id FROM shell_command WHERE description='List all deployments in prod'), 'serial', 1, datetime('now'), datetime('now')),
('monitor-resources', 'Monitor disk and memory resources', (SELECT id FROM shell_command WHERE description='Check disk usage'), 'parallel', 0, datetime('now'), datetime('now')),
('critical-restart-payment', 'Restart payment-gateway (CRITICAL)', (SELECT id FROM shell_command WHERE description='Check payment-gateway status'), 'serial', 1, datetime('now'), datetime('now')),
('check-uptime', 'Check server uptime in production', (SELECT id FROM shell_command WHERE description='Check server uptime'), 'parallel', 0, datetime('now'), datetime('now'));

-- 12. Insert shell task server associations (map tasks to servers)
INSERT OR IGNORE INTO shell_task_server (task_id, server_id, created_at) VALUES 
((SELECT id FROM shell_task WHERE name='health-check-dev'), (SELECT id FROM shell_server WHERE name='dev-server-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='restart-staging-web'), (SELECT id FROM shell_server WHERE name='staging-server-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='prod-list-deployments'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='prod-list-deployments'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-02'), datetime('now')),
((SELECT id FROM shell_task WHERE name='monitor-resources'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='monitor-resources'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-02'), datetime('now')),
((SELECT id FROM shell_task WHERE name='monitor-resources'), (SELECT id FROM shell_server WHERE name='prod-server-cn2-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='critical-restart-payment'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), datetime('now')),
((SELECT id FROM shell_task WHERE name='check-uptime'), (SELECT id FROM shell_server WHERE name='prod-server-cn2-01'), datetime('now'));

-- 13. Insert shell task executions (execution history)
INSERT OR IGNORE INTO shell_task_execution (task_id, server_id, command_id, status, output, error_message, exit_code, started_at, completed_at, created_at, updated_at) VALUES 
((SELECT id FROM shell_task WHERE name='health-check-dev'), (SELECT id FROM shell_server WHERE name='dev-server-01'), (SELECT id FROM shell_command WHERE description LIKE 'Check api-service%' LIMIT 1), 'success', 'api-service is running (PID 12345)', NULL, 0, datetime('now', '-2 hours'), datetime('now', '-2 hours', '+10 seconds'), datetime('now', '-2 hours'), datetime('now', '-2 hours')),
((SELECT id FROM shell_task WHERE name='restart-staging-web'), (SELECT id FROM shell_server WHERE name='staging-server-01'), (SELECT id FROM shell_command WHERE description='Restart web-ui service on staging'), 'success', 'Service restarted successfully', NULL, 0, datetime('now', '-90 minutes'), datetime('now', '-90 minutes', '+15 seconds'), datetime('now', '-90 minutes'), datetime('now', '-90 minutes')),
((SELECT id FROM shell_task WHERE name='monitor-resources'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), (SELECT id FROM shell_command WHERE description='Check disk usage'), 'success', 'Filesystem: 82% used (410GB/500GB)', NULL, 0, datetime('now', '-45 minutes'), datetime('now', '-45 minutes', '+5 seconds'), datetime('now', '-45 minutes'), datetime('now', '-45 minutes')),
((SELECT id FROM shell_task WHERE name='monitor-resources'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-02'), (SELECT id FROM shell_command WHERE description='Check memory usage'), 'success', 'Memory: 78% used (15.6GB/20GB)', NULL, 0, datetime('now', '-45 minutes'), datetime('now', '-45 minutes', '+5 seconds'), datetime('now', '-45 minutes'), datetime('now', '-45 minutes')),
((SELECT id FROM shell_task WHERE name='health-check-dev'), (SELECT id FROM shell_server WHERE name='dev-server-01'), (SELECT id FROM shell_command WHERE description LIKE 'Check api-service%' LIMIT 1), 'success', 'api-service is running (PID 12346)', NULL, 0, datetime('now', '-30 minutes'), datetime('now', '-30 minutes', '+10 seconds'), datetime('now', '-30 minutes'), datetime('now', '-30 minutes')),
((SELECT id FROM shell_task WHERE name='check-uptime'), (SELECT id FROM shell_server WHERE name='prod-server-cn2-01'), (SELECT id FROM shell_command WHERE description='Check server uptime'), 'success', 'up 45 days, 12 hours, 34 minutes', NULL, 0, datetime('now', '-20 minutes'), datetime('now', '-20 minutes', '+5 seconds'), datetime('now', '-20 minutes'), datetime('now', '-20 minutes')),
((SELECT id FROM shell_task WHERE name='prod-list-deployments'), (SELECT id FROM shell_server WHERE name='prod-server-cn1-01'), (SELECT id FROM shell_command WHERE description='List all deployments in prod'), 'failed', NULL, 'Connection timeout after 30s', 1, datetime('now', '-10 minutes'), datetime('now', '-10 minutes', '+30 seconds'), datetime('now', '-10 minutes'), datetime('now', '-10 minutes'));

-- 14. Insert command approvals (for tasks requiring approval)
INSERT OR IGNORE INTO command_approval (id, request_id, approval_status, approved_by, approved_at, created_at, updated_at) VALUES 
('approval-001', 'req-prod-list-001', 'approved', 'admin@example.com', datetime('now', '-1 days'), datetime('now', '-1 days'), datetime('now', '-1 days')),
('approval-002', 'req-critical-restart-001', 'pending', NULL, NULL, datetime('now', '-30 minutes'), datetime('now', '-30 minutes')),
('approval-003', 'req-prod-list-002', 'rejected', 'security@example.com', datetime('now', '-2 days'), datetime('now', '-2 days', '+2 hours'), datetime('now', '-2 days', '+2 hours')),
('approval-004', 'req-critical-restart-002', 'approved', 'admin@example.com', datetime('now', '-12 hours'), datetime('now', '-12 hours', '-30 minutes'), datetime('now', '-12 hours')),
('approval-005', 'req-prod-list-003', 'approved', 'ops-lead@example.com', datetime('now', '-6 hours'), datetime('now', '-6 hours', '-1 hour'), datetime('now', '-6 hours'));
