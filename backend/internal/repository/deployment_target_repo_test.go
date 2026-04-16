package repository
package repository

import (
	"testing"
	"time"

	"built-and-deploy/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *DeploymentTargetRepository {
	// 这里应该使用 SQLite 在内存数据库
	// 对于完整的测试，需要实际的数据库设置
	t.Helper()
	
	// 使用 sqlite3 包创建内存数据库
	db, err := setupInMemorySQLite()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	
	return NewDeploymentTargetRepository(db)
}

// TestDeploymentTargetRepository_Create 测试创建部署目标
func TestDeploymentTargetRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	
	target := &models.DeploymentTarget{
		AppID:          1,
		EnvID:          1,
		ClusterID:      1,
		K8sNamespace:   "default",
		K8sDeployment:  "api-service",
		ContainerName:  "api",
		RegistryDomain: "docker.io",
		ImageRepo:      "company/api-service",
	}
	
	result, err := repo.Create(target)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.ID, 0)
	assert.Equal(t, target.AppID, result.AppID)
	assert.Equal(t, target.K8sNamespace, result.K8sNamespace)
}

// TestDeploymentTargetRepository_GetByID 测试获取单个部署目标
func TestDeploymentTargetRepository_GetByID(t *testing.T) {
	repo := setupTestDB(t)
	
	// 先创建
	target := &models.DeploymentTarget{
		AppID:         1,
		EnvID:         1,
		ClusterID:     1,
		K8sNamespace:  "production",
		K8sDeployment: "web-service",
		ContainerName: "web",
		RegistryDomain: "harbor.io",
		ImageRepo:     "company/web",
	}
	
	created, err := repo.Create(target)
	require.NoError(t, err)
	
	// 获取
	retrieved, err := repo.GetByID(created.ID)
	
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, target.AppID, retrieved.AppID)
	assert.Equal(t, target.K8sNamespace, retrieved.K8sNamespace)
}

// TestDeploymentTargetRepository_GetByApp 测试按应用获取部署目标
func TestDeploymentTargetRepository_GetByApp(t *testing.T) {
	repo := setupTestDB(t)
	
	// 创建多个部署目标
	targets := []*models.DeploymentTarget{
		{
			AppID:         1,
			EnvID:         1,
			ClusterID:     1,
			K8sNamespace:  "prod-1",
			K8sDeployment: "app-v1",
			ContainerName: "app",
			RegistryDomain: "registry.io",
			ImageRepo:     "myapp",
		},
		{
			AppID:         1,
			EnvID:         2,
			ClusterID:     2,
			K8sNamespace:  "staging",
			K8sDeployment: "app-staging",
			ContainerName: "app",
			RegistryDomain: "registry.io",
			ImageRepo:     "myapp",
		},
		{
			AppID:         2,
			EnvID:         1,
			ClusterID:     1,
			K8sNamespace:  "prod-1",
			K8sDeployment: "other-app",
			ContainerName: "other",
			RegistryDomain: "registry.io",
			ImageRepo:     "other-app",
		},
	}
	
	for _, t := range targets {
		_, err := repo.Create(t)
		require.NoError(t, err)
	}
	
	// 获取应用 1 的所有部署目标
	results, err := repo.GetByApp(1)
	
	require.NoError(t, err)
	assert.Len(t, results, 2)
	for _, result := range results {
		assert.Equal(t, 1, result.AppID)
	}
}

// TestDeploymentTargetRepository_Update 测试更新部署目标
func TestDeploymentTargetRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	
	// 创建
	target := &models.DeploymentTarget{
		AppID:         1,
		EnvID:         1,
		ClusterID:     1,
		K8sNamespace:  "default",
		K8sDeployment: "service",
		ContainerName: "container",
		RegistryDomain: "docker.io",
		ImageRepo:     "image",
	}
	
	created, err := repo.Create(target)
	require.NoError(t, err)
	
	// 更新
	created.K8sNamespace = "updated-namespace"
	err = repo.Update(created)
	
	require.NoError(t, err)
	
	// 验证
	retrieved, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated-namespace", retrieved.K8sNamespace)
}

// TestDeploymentTargetRepository_Delete 测试删除部署目标
func TestDeploymentTargetRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	
	// 创建
	target := &models.DeploymentTarget{
		AppID:         1,
		EnvID:         1,
		ClusterID:     1,
		K8sNamespace:  "default",
		K8sDeployment: "service",
		ContainerName: "container",
		RegistryDomain: "docker.io",
		ImageRepo:     "image",
	}
	
	created, err := repo.Create(target)
	require.NoError(t, err)
	
	// 删除
	err = repo.Delete(created.ID)
	require.NoError(t, err)
	
	// 验证已删除
	_, err = repo.GetByID(created.ID)
	assert.Error(t, err)
}

// TestDeploymentTargetRepository_List 测试列表查询
func TestDeploymentTargetRepository_List(t *testing.T) {
	repo := setupTestDB(t)
	
	// 创建5个部署目标
	for i := 0; i < 5; i++ {
		target := &models.DeploymentTarget{
			AppID:         1,
			EnvID:         1,
			ClusterID:     i + 1,
			K8sNamespace:  "ns-" + string(rune(i)),
			K8sDeployment: "deploy-" + string(rune(i)),
			ContainerName: "container",
			RegistryDomain: "docker.io",
			ImageRepo:     "image",
		}
		_, err := repo.Create(target)
		require.NoError(t, err)
	}
	
	// 测试分页
	results, err := repo.List(2, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	results, err = repo.List(10, 2)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

// TestDeploymentTargetRepository_GetByAppEnvCluster 测试按应用、环境、集群查询
func TestDeploymentTargetRepository_GetByAppEnvCluster(t *testing.T) {
	repo := setupTestDB(t)
	
	// 创建
	target := &models.DeploymentTarget{
		AppID:         1,
		EnvID:         2,
		ClusterID:     3,
		K8sNamespace:  "prod",
		K8sDeployment: "service",
		ContainerName: "app",
		RegistryDomain: "registry.io",
		ImageRepo:     "app-image",
	}
	
	created, err := repo.Create(target)
	require.NoError(t, err)
	
	// 查询
	retrieved, err := repo.GetByAppEnvCluster(1, 2, 3)
	
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
}

// 辅助函数：设置内存 SQLite
func setupInMemorySQLite() (interface{}, error) {
	// 这里需要实现 SQLite 内存数据库的设置
	// 这是一个简化版本，实际实现需要完整的数据库初始化
	return nil, nil
}
