// Package middleware 提供 HTTP 中间件
//
// 本文件实现 Panic 恢复中间件，功能包括：
//   - 捕获 panic 异常
//   - 记录完整堆栈信息
//   - 返回 500 错误而非程序崩溃
//   - 优雅降级，保证服务可用性
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery Panic 恢复中间件
//
// 该中间件捕获所有 panic 异常，防止服务崩溃：
//  1. 捕获 panic
//  2. 记录堆栈信息到日志
//  3. 返回统一的 500 错误响应
//  4. 继续处理后续请求
//
// 示例:
//   router.Use(middleware.Recovery())
//
// 返回:
//   - Gin 中间件函数
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 记录 panic 信息
				// TODO: 使用结构化日志库（如 zap 或 logrus）
				fmt.Printf("[PANIC RECOVERED] %v\n", err)
				fmt.Printf("Stack Trace:\n%s\n", stack)

				// 返回 500 错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
					"error":   fmt.Sprintf("%v", err),
				})

				// 终止后续处理
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next()
	}
}

// RecoveryWithWriter 自定义日志输出的 Panic 恢复中间件
//
// 参数:
//   - logFunc: 自定义日志函数，接收 panic 信息和堆栈
//
// 返回:
//   - Gin 中间件函数
func RecoveryWithWriter(logFunc func(err interface{}, stack []byte)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				// 使用自定义日志函数
				if logFunc != nil {
					logFunc(err, stack)
				}

				// 返回 500 错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
