-- Initial sample data for Release Control System

-- 1. Insert sample applications
INSERT OR IGNORE INTO application (name, image_name, git_repo, build_type, created_at, updated_at) VALUES 
('api-service', 'api-service', 'git@github.com:company/api-service.git', 'docker', datetime('now'), datetime('now')),
('web-ui', 'web-ui', 'git@github.com:company/web-ui.git', 'docker', datetime('now'), datetime('now')),
('data-processor', 'data-processor', 'git@github.com:company/data-processor.git', 'docker', datetime('now'), datetime('now'));

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

-- 4. Insert deployment targets (app-env-cluster mappings)
INSERT OR IGNORE INTO deployment_target (app_id, env_id, cluster_id, k8s_namespace, k8s_deployment, container_name, created_at, updated_at) VALUES 
-- api-service
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'api-service', 'api', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'api-service', 'api', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'api-service', 'api-prod', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn2'), 'production', 'api-service', 'api-prod', datetime('now'), datetime('now')),
-- web-ui
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'web-ui', 'web', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'web-ui', 'web', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'web-ui', 'web-prod', datetime('now'), datetime('now')),
-- data-processor
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'development', 'data-processor', 'processor', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'staging', 'data-processor', 'processor', datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'production', 'data-processor', 'processor-prod', datetime('now'), datetime('now'));

-- 5. Insert sample release records
INSERT OR IGNORE INTO release_record (app_id, env_id, cluster_id, image, status, previous_image, triggered_by, started_at, completed_at, created_at, updated_at) VALUES 
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'api-service:v1.2.3', 'success', 'api-service:v1.2.2', 'user@example.com', datetime('now', '-3 days'), datetime('now', '-3 days', '+5 minutes'), datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='staging'), (SELECT id FROM cluster WHERE name='k8s-staging'), 'api-service:v1.2.3', 'success', 'api-service:v1.2.2', 'CI/CD Pipeline', datetime('now', '-2 days'), datetime('now', '-2 days', '+5 minutes'), datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='api-service'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'api-service:v1.2.3', 'deploying', 'api-service:v1.2.1', 'admin@example.com', datetime('now', '-1 days'), NULL, datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='development'), (SELECT id FROM cluster WHERE name='k8s-dev'), 'web-ui:v2.1.0', 'success', 'web-ui:v2.0.9', 'developer@example.com', datetime('now', '-2 days'), datetime('now', '-2 days', '+5 minutes'), datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='web-ui'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'web-ui:v2.1.0', 'failed', 'web-ui:v2.0.8', 'CI/CD Pipeline', datetime('now', '-1 days'), datetime('now', '-1 days', '+10 minutes'), datetime('now'), datetime('now')),
((SELECT id FROM application WHERE name='data-processor'), (SELECT id FROM environment WHERE name='production'), (SELECT id FROM cluster WHERE name='k8s-prod-cn1'), 'data-processor:v3.5.0', 'success', 'data-processor:v3.4.9', 'scheduler@system', datetime('now', '-3 hours'), datetime('now', '-3 hours', '+8 minutes'), datetime('now'), datetime('now'));

-- 6. Insert release events
INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'started', 'Release deployment initialized', datetime(started_at, '+1 minute') FROM release_record WHERE status IN ('success', 'deploying', 'failed') LIMIT 5;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'deploying', 'Pulling image from registry', datetime(started_at, '+2 minutes') FROM release_record WHERE status IN ('success', 'deploying', 'failed') LIMIT 5;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'deploying', 'Applying Kubernetes manifests', datetime(started_at, '+3 minutes') FROM release_record WHERE status IN ('success', 'deploying') LIMIT 4;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'success', 'Deployment completed successfully', datetime(started_at, '+5 minutes') FROM release_record WHERE status = 'success' LIMIT 4;

INSERT INTO release_event (release_id, type, message, created_at) 
SELECT id, 'failed', 'Image pull failed', datetime(started_at, '+2 minutes') FROM release_record WHERE status = 'failed' LIMIT 1;

-- 7. Insert audit logs
INSERT OR IGNORE INTO audit_log (user_id, operation, resource_type, resource_id, created_at) VALUES 
('admin@example.com', 'CREATE', 'application', (SELECT id FROM application WHERE name='api-service'), datetime('now', '-3 days')),
('admin@example.com', 'CREATE', 'environment', (SELECT id FROM environment WHERE name='production'), datetime('now', '-2.5 days')),
('developer@example.com', 'DEPLOY', 'release', 1, datetime('now', '-2 days')),
('CI/CD Pipeline', 'AUTO_DEPLOY', 'release', 2, datetime('now', '-1.5 days')),
('admin@example.com', 'UPDATE', 'deployment_target', 1, datetime('now', '-1 days')),
('developer@example.com', 'VIEW', 'release_history', 1, datetime('now', '-12 hours')),
('admin@example.com', 'DELETE', 'application', 0, datetime('now', '-6 hours')),
('scheduler@system', 'AUTO_DEPLOY', 'release', 6, datetime('now', '-3 hours'));
