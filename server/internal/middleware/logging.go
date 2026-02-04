// Package middleware 提供 HTTP 中间件
//
// 本文件实现请求日志中间件，功能包括：
//   - 记录请求方法、路径、耗时
//   - 记录响应状态码
//   - 记录用户 IP 和 User-Agent
//   - 结构化日志输出（JSON 格式）
//   - 性能监控与审计
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Query      string `json:"query,omitempty"`
	StatusCode int    `json:"status_code"`
	Latency    string `json:"latency"`
	ClientIP   string `json:"client_ip"`
	UserAgent  string `json:"user_agent,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	ErrorMsg   string `json:"error,omitempty"`
}

// Logger 请求日志中间件
//
// 该中间件记录所有 HTTP 请求的详细信息：
//  1. 请求开始时间
//  2. 请求方法、路径、查询参数
//  3. 响应状态码和耗时
//  4. 客户端 IP 和 User-Agent
//  5. 认证用户 ID（如果有）
//
// 日志格式为 JSON，便于后续分析和监控。
//
// 示例:
//   router.Use(middleware.Logger())
//
// 返回:
//   - Gin 中间件函数
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(startTime)

		// 构建日志条目
		entry := LogEntry{
			Timestamp:  startTime.Format(time.RFC3339),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      c.Request.URL.RawQuery,
			StatusCode: c.Writer.Status(),
			Latency:    latency.String(),
			ClientIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}

		// 获取用户 ID（如果已认证）
		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID.(string)
		}

		// 获取错误信息（如果有）
		if len(c.Errors) > 0 {
			entry.ErrorMsg = c.Errors.String()
		}

		// 输出 JSON 格式日志
		logJSON, _ := json.Marshal(entry)
		fmt.Println(string(logJSON))
	}
}

// LoggerWithConfig 自定义配置的日志中间件
//
// 参数:
//   - config: 日志配置选项
//
// 返回:
//   - Gin 中间件函数
type LoggerConfig struct {
	// SkipPaths 跳过日志记录的路径（如健康检查）
	SkipPaths []string

	// LogWriter 自定义日志写入函数
	LogWriter func(entry LogEntry)

	// EnableQueryParams 是否记录查询参数
	EnableQueryParams bool

	// EnableUserAgent 是否记录 User-Agent
	EnableUserAgent bool
}

// LoggerWithConfig 创建带配置的日志中间件
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	// 创建跳过路径映射
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		// 检查是否跳过日志
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)

		entry := LogEntry{
			Timestamp:  startTime.Format(time.RFC3339),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			Latency:    latency.String(),
			ClientIP:   c.ClientIP(),
		}

		// 可选字段
		if config.EnableQueryParams {
			entry.Query = c.Request.URL.RawQuery
		}

		if config.EnableUserAgent {
			entry.UserAgent = c.Request.UserAgent()
		}

		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID.(string)
		}

		if len(c.Errors) > 0 {
			entry.ErrorMsg = c.Errors.String()
		}

		// 使用自定义写入函数或默认输出
		if config.LogWriter != nil {
			config.LogWriter(entry)
		} else {
			logJSON, _ := json.Marshal(entry)
			fmt.Println(string(logJSON))
		}
	}
}
