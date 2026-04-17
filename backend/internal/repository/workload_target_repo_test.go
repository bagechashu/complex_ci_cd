package repository

import (
	"testing"

	"built-and-deploy/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *WorkloadTargetRepository {
	// 这里应该使用 SQLite 在内存数据库
	// 对于完整的测试，需要实际的数据库设置
	t.Helper()

	// 使用 sqlite3 包创建内存数据库
	db, err := setupInMemorySQLite()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	return NewWorkloadTargetRepository(db)
}

// TestWorkloadTargetRepository_Create 测试创建部署目标
func TestWorkloadTargetRepository_Create(t *testing.T) {
	repo := setupTestDB(t)

	target := &models.WorkloadTarget{
		AppID:          1,
		EnvID:          1,
		ClusterID:      1,
		K8sNamespace:   "default",
		K8sWorkload:    "api-service",
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

// TestWorkloadTargetRepository_GetByID 测试获取单个部署目标
func TestWorkloadTargetRepository_GetByID(t *testing.T) {
	repo := setupTestDB(t)

	// 先创建
	target := &models.WorkloadTarget{
		AppID:          1,
		EnvID:          1,
		ClusterID:      1,
		K8sNamespace:   "production",
		K8sWorkload:    "web-service",
		ContainerName:  "web",
		RegistryDomain: "harbor.io",
		ImageRepo:      "company/web",
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

// TestWorkloadTargetRepository_GetByApp 测试按应用获取部署目标
func TestWorkloadTargetRepository_GetByApp(t *testing.T) {
	repo := setupTestDB(t)

	// 创建多个部署目标
	targets := []*models.WorkloadTarget{
		{
			AppID:          1,
			EnvID:          1,
			ClusterID:      1,
			K8sNamespace:   "prod-1",
			K8sWorkload:    "app-v1",
			ContainerName:  "app",
			RegistryDomain: "registry.io",
			ImageRepo:      "myapp",
		},
		{
			AppID:          1,
			EnvID:          2,
			ClusterID:      2,
			K8sNamespace:   "staging",
			K8sWorkload:    "app-staging",
			ContainerName:  "app",
			RegistryDomain: "registry.io",
			ImageRepo:      "myapp",
		},
		{
			AppID:          2,
			EnvID:          1,
			ClusterID:      1,
			K8sNamespace:   "prod-1",
			K8sWorkload:    "other-app",
			ContainerName:  "other",
			RegistryDomain: "registry.io",
			ImageRepo:      "other-app",
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

// TestWorkloadTargetRepository_Update 测试更新部署目标
func TestWorkloadTargetRepository_Update(t *testing.T) {
	repo := setupTestDB(t)

	// 创建
	target := &models.WorkloadTarget{
		AppID:          1,
		EnvID:          1,
		ClusterID:      1,
		K8sNamespace:   "default",
		K8sWorkload:    "service",
		ContainerName:  "container",
		RegistryDomain: "docker.io",
		ImageRepo:      "image",
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

// TestWorkloadTargetRepository_Delete 测试删除部署目标
func TestWorkloadTargetRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)

	// 创建
	target := &models.WorkloadTarget{
		AppID:          1,
		EnvID:          1,
		ClusterID:      1,
		K8sNamespace:   "default",
		K8sWorkload:    "service",
		ContainerName:  "container",
		RegistryDomain: "docker.io",
		ImageRepo:      "image",
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

// TestWorkloadTargetRepository_List 测试列表查询
func TestWorkloadTargetRepository_List(t *testing.T) {
	repo := setupTestDB(t)

	// 创建5个部署目标
	for i := 0; i < 5; i++ {
		target := &models.WorkloadTarget{
			AppID:          1,
			EnvID:          1,
			ClusterID:      i + 1,
			K8sNamespace:   "ns-" + string(rune(i)),
			K8sWorkload:    "deploy-" + string(rune(i)),
			ContainerName:  "container",
			RegistryDomain: "docker.io",
			ImageRepo:      "image",
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

// TestWorkloadTargetRepository_GetByAppEnvCluster 测试按应用、环境、集群查询
func TestWorkloadTargetRepository_GetByAppEnvCluster(t *testing.T) {
	repo := setupTestDB(t)

	// 创建
	target := &models.WorkloadTarget{
		AppID:          1,
		EnvID:          2,
		ClusterID:      3,
		K8sNamespace:   "prod",
		K8sWorkload:    "service",
		ContainerName:  "app",
		RegistryDomain: "registry.io",
		ImageRepo:      "app-image",
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
