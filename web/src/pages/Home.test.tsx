/**
 * @file Home.test.tsx
 * @description 首页（取件页）单元测试
 *
 * 测试场景：
 *  - 页面正确渲染
 *  - 取件码输入与验证
 *  - 成功取件流程
 *  - 错误处理（验证失败、网络错误）
 *
 * @author AhaVault Team
 * @created 2026-02-05
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import Home from './Home'

import { ThemeProvider } from '../providers/ThemeProvider'
import { shareService } from '../services/shareService'

// Mock dependencies
vi.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string) => key,
        i18n: {
            changeLanguage: () => new Promise(() => { }),
        },
    }),
}))

vi.mock('../services/shareService', () => ({
    shareService: {
        getShareByCode: vi.fn(),
    },
}))

/**
 * 测试辅助函数：渲染 Home 组件并提供必要的上下文
 */
const renderHome = () => {
    return render(
        <BrowserRouter>
            <ThemeProvider>
                <Home />
            </ThemeProvider>
        </BrowserRouter>
    )
}

describe('Home - Pickup Flow', () => {

    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('renders landing page correctly', () => {
        renderHome()

        // Check main title
        expect(screen.getByText('Your Data,')).toBeInTheDocument()
        expect(screen.getByText('Truly Secure.')).toBeInTheDocument()

        // Check pickup code input
        expect(screen.getByPlaceholderText('AHA-XXXX-XXXX')).toBeInTheDocument()

        // Check retrieve button (use role for precision)
        expect(screen.getByRole('button', { name: /Retrieve File/i })).toBeInTheDocument()
    })

    it('handles successful pickup', async () => {
        const mockData = {
            code: 0,
            data: {
                share_id: '123',
                files: [{ file_id: 'f1', filename: 'test.png', size: 1024, mime_type: 'image/png' }],
                expires_at: new Date().toISOString(),
                remaining_downloads: 5
            }
        }

        // Setup mock
        vi.mocked(shareService.getShareByCode).mockResolvedValue(mockData as any)

        renderHome()

        // Type code
        const input = screen.getByPlaceholderText('AHA-XXXX-XXXX')
        fireEvent.change(input, { target: { value: 'ABCDEF12' } })

        // Click button (use role for precision)
        const button = screen.getByRole('button', { name: /Retrieve File/i })
        fireEvent.click(button)

        // Check loading state
        expect(screen.getByText('Verifying...')).toBeInTheDocument()

        // Wait for result
        await waitFor(() => {
            expect(screen.getByText('File Found')).toBeInTheDocument()
            expect(screen.getByText('test.png')).toBeInTheDocument()
        })
    })

    it('handles validation error (too short)', async () => {
        renderHome()

        const input = screen.getByPlaceholderText('AHA-XXXX-XXXX')
        fireEvent.change(input, { target: { value: 'SHORT' } })

        const button = screen.getByRole('button', { name: /Retrieve File/i })
        fireEvent.click(button)

        expect(screen.getByText('Please enter a valid 8-digit code')).toBeInTheDocument()
        expect(shareService.getShareByCode).not.toHaveBeenCalled()
    })
})
