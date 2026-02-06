/**
 * @file format.ts
 * @description 通用格式化工具函数
 *
 * 功能说明：
 *  - 文件大小格式化（B/KB/MB/GB/TB）
 *  - 日期时间格式化
 *  - 数字格式化
 *
 * @author AhaVault Team
 * @created 2026-02-05
 */

/**
 * 格式化文件大小
 *
 * @param bytes - 文件字节大小
 * @returns 格式化后的文件大小字符串（如 "1.5 MB"）
 *
 * @example
 * ```typescript
 * formatFileSize(0)          // "0 B"
 * formatFileSize(1024)       // "1 KB"
 * formatFileSize(1536)       // "1.5 KB"
 * formatFileSize(1048576)    // "1 MB"
 * formatFileSize(1073741824) // "1 GB"
 * ```
 */
export function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B'

    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 格式化日期时间为本地字符串
 *
 * @param date - Date 对象或 ISO 字符串
 * @returns 格式化后的本地日期时间字符串
 *
 * @example
 * ```typescript
 * formatDateTime(new Date())              // "2026-02-05 11:30:45"
 * formatDateTime("2026-02-05T03:30:45Z")  // "2026-02-05 11:30:45" (本地时间)
 * ```
 */
export function formatDateTime(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date
    return d.toLocaleString()
}

/**
 * 格式化数字为千分位格式
 *
 * @param num - 数字
 * @returns 千分位格式的字符串
 *
 * @example
 * ```typescript
 * formatNumber(1000)     // "1,000"
 * formatNumber(1234567)  // "1,234,567"
 * ```
 */
export function formatNumber(num: number): string {
    return num.toLocaleString()
}
