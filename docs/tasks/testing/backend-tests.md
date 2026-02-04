# 后端测试补充任务清单

**模块名称**: 后端测试补充
**负责人**: Claude AI
**最后更新**: 2026-02-04
**当前进度**: 30%
**状态**: 🔥 紧急（P0）

---

## 🎯 目标

将后端测试覆盖率从当前的 **~30%** 提升到 **≥70%**

---

## 📊 当前测试覆盖率

| 模块 | 当前覆盖率 | 目标覆盖率 | 差距 | 优先级 |
|------|-----------|-----------|------|--------|
| config | 80.6% | ≥80% | ✅ 达标 | - |
| crypto | 75.0% | ≥80% | -5% | P1 |
| storage | 65.6% | ≥80% | -14.4% | P1 |
| database | 60.0% | ≥70% | -10% | P1 |
| **services** | **~30%** | **≥70%** | **-40%** | **P0** 🔥 |
| **middleware** | **0%** | **≥60%** | **-60%** | **P0** 🔥 |
| **handlers** | **0%** | **≥60%** | **-60%** | **P0** 🔥 |

**总体覆盖率**: ~30% → 目标 ≥70%

---

## 📋 任务清单

### 🔥 P0 紧急任务

#### Services 层测试

- [ ] `internal/services/file_service_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 4 小时
  - 目标覆盖率: 70%
  - 测试用例:
    - ✅ 秒传检测（文件已存在 vs 不存在）
    - ✅ 配额检查（超限场景）
    - ✅ 引用计数管理（上传、删除、多用户）
    - ✅ 软删除功能
    - ✅ 文件列表分页

- [ ] `internal/services/share_service_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 3 小时
  - 目标覆盖率: 70%
  - 测试用例:
    - ✅ 创建分享（单文件、多文件）
    - ✅ 取件码唯一性
    - ✅ 密码保护分享
    - ✅ 下载次数限制（1 次、N 次、无限）
    - ✅ 过期时间检查

- [ ] `internal/services/user_service_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 3 小时
  - 目标覆盖率: 70%
  - 测试用例:
    - ✅ 用户注册（成功、邮箱重复）
    - ✅ 用户登录（成功、密码错误、用户不存在）
    - ✅ JWT 生成与解析
    - ✅ bcrypt 密码加密

#### Handlers 层测试

- [ ] `internal/api/handlers/auth_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 3 小时
  - 目标覆盖率: 60%
  - 测试用例:
    - HTTP POST /api/auth/register
    - HTTP POST /api/auth/login
    - HTTP POST /api/auth/logout
    - 错误场景（400, 401, 409）

- [ ] `internal/api/handlers/file_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 3 小时
  - 目标覆盖率: 60%
  - 测试用例:
    - HTTP POST /api/files/check
    - HTTP GET /api/files
    - HTTP DELETE /api/files/:id
    - 认证检查（401）

- [ ] `internal/api/handlers/share_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 3 小时
  - 目标覆盖率: 60%
  - 测试用例:
    - HTTP POST /api/shares
    - HTTP GET /api/shares/:code
    - 密码验证（403）

#### Middleware 层测试

- [ ] `internal/middleware/auth_test.go`
  - 优先级: **P0** 🔥
  - 预计工作量: 2 小时
  - 目标覆盖率: 70%
  - 测试用例:
    - 有效 Token 验证
    - 无效 Token 验证
    - 过期 Token 验证
    - 缺失 Token 验证

- [ ] `internal/middleware/ratelimit_test.go`
  - 优先级: P0（待实现后）
  - 依赖: ratelimit.go 完成

- [ ] `internal/middleware/logging_test.go`
  - 优先级: P0（待实现后）
  - 依赖: logging.go 完成

### ⚪ P1 重要任务

#### 提升现有覆盖率

- [ ] 提升 crypto 模块覆盖率：75% → 80%
  - 补充边界条件测试
  - 补充错误处理测试

- [ ] 提升 storage 模块覆盖率：65.6% → 80%
  - 补充并发测试
  - 补充错误场景测试

- [ ] 提升 database 模块覆盖率：60% → 70%
  - 补充连接池测试
  - 补充错误恢复测试

---

## 🧪 测试最佳实践

### 1. 使用 testify 断言库

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFileServiceCheck(t *testing.T) {
    // require: 失败立即停止
    db := setupTestDB(t)
    require.NotNil(t, db)

    // assert: 失败继续执行
    exists, err := fileService.Check(hash)
    assert.NoError(t, err)
    assert.True(t, exists)
}
```

### 2. 表驱动测试

```go
func TestUserRegister(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"duplicate email", "exist@example.com", true},
        {"invalid email", "invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := userService.Register(tt.email, "password")
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 3. HTTP 测试

```go
import (
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

func TestLoginHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/auth/login", handlers.Login)

    req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"test@example.com","password":"pass"}`))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

### 4. Mock 和隔离

```go
// 使用接口隔离依赖
type FileStorage interface {
    Put(hash string, reader io.Reader) error
    Get(hash string) (io.ReadCloser, error)
}

// 在测试中使用 mock
type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) Put(hash string, reader io.Reader) error {
    args := m.Called(hash, reader)
    return args.Error(0)
}
```

---

## 📈 测试覆盖率追踪

### 运行测试

```bash
cd server

# 运行所有测试
go test ./...

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 查看详细覆盖率
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
```

### 每日更新覆盖率

完成测试后，更新以下文档：
1. `docs/memory-bank/progress.md` - 更新测试覆盖率表格
2. `docs/progress/coverage.md` - 详细覆盖率报告（自动生成）

---

## 🎯 预期成果

完成所有 P0 任务后：

| 模块 | 当前 | 预期 | 提升 |
|------|------|------|------|
| services | 30% | 70% | +40% |
| middleware | 0% | 60% | +60% |
| handlers | 0% | 60% | +60% |
| **总体** | **30%** | **≥70%** | **+40%** |

---

## 🔗 相关文档

- **Services 任务**: [../backend/04-services.md](../backend/04-services.md)
- **API 任务**: [../backend/05-api.md](../backend/05-api.md)
- **当前进度**: [../../memory-bank/progress.md](../../memory-bank/progress.md)

---

## 📅 更新日志

### 2026-02-04
- 📋 创建后端测试补充任务清单
- 📊 梳理当前测试覆盖率
- 🎯 定义目标和优先级

---

**维护者**: Claude AI
**下一步**: 立即开始补充 Services 层测试（P0）
**预计完成**: 2026-02-06
