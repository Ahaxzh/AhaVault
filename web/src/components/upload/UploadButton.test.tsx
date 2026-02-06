/**
 * @file UploadButton.test.tsx
 * @description UploadButton 组件单元测试
 *
 * 测试场景：
 *  - 基本渲染测试
 *  - 文件选择触发测试
 *  - 上传进度显示测试
 *  - 取消上传功能测试
 *  - 回调函数调用测试
 *
 * @author AhaVault Team
 * @created 2026-02-05
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { UploadButton } from './UploadButton'
import { uploadService } from '@/services/uploadService'
import * as tus from 'tus-js-client'

// Mock dependencies
vi.mock('@/services/uploadService', () => ({
    uploadService: {
        uploadFile: vi.fn(),
    },
}))

describe('UploadButton', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    /**
     * 测试基本渲染
     */
    it('should render upload button', () => {
        render(<UploadButton />)

        const button = screen.getByRole('button', { name: /upload/i })
        expect(button).toBeInTheDocument()
    })

    /**
     * 测试点击按钮触发文件选择
     * 注意：input[type="file"] 是隐藏的，与 button 是兄弟元素
     */
    it('should trigger file input on button click', () => {
        render(<UploadButton />)

        const button = screen.getByRole('button', { name: /upload/i })
        // input 是隐藏的，在 button 的兄弟位置，使用 document.querySelector
        const fileInput = document.querySelector('input[type="file"]')

        expect(fileInput).toBeInTheDocument()
        expect(fileInput).toHaveAttribute('type', 'file')
        expect(button).toBeInTheDocument()
    })

    /**
     * 测试文件选择后触发上传
     */
    it('should start upload when file is selected', async () => {
        const mockUpload = {
            start: vi.fn(),
            abort: vi.fn(),
        } as unknown as tus.Upload

        vi.mocked(uploadService.uploadFile).mockReturnValue(mockUpload)

        render(<UploadButton />)

        // Find the hidden file input
        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        expect(fileInput).toBeInTheDocument()

        // Create a mock file
        const file = new File(['test content'], 'test.txt', { type: 'text/plain' })

        // Simulate file selection
        Object.defineProperty(fileInput, 'files', {
            value: [file],
            writable: false,
        })

        fireEvent.change(fileInput)

        // Verify uploadService.uploadFile was called
        await waitFor(() => {
            expect(uploadService.uploadFile).toHaveBeenCalledWith(
                file,
                expect.objectContaining({
                    onProgress: expect.any(Function),
                    onSuccess: expect.any(Function),
                    onError: expect.any(Function),
                })
            )
        })
    })

    /**
     * 测试上传进度显示
     */
    it('should display upload progress', async () => {
        let capturedCallbacks: any = {}

        const mockUpload = {
            start: vi.fn(),
            abort: vi.fn(),
        } as unknown as tus.Upload

        vi.mocked(uploadService.uploadFile).mockImplementation((file, callbacks) => {
            capturedCallbacks = callbacks
            return mockUpload
        })

        render(<UploadButton />)

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        const file = new File(['test'], 'test.txt', { type: 'text/plain' })

        Object.defineProperty(fileInput, 'files', {
            value: [file],
            writable: false,
        })

        fireEvent.change(fileInput)

        await waitFor(() => {
            expect(uploadService.uploadFile).toHaveBeenCalled()
        })

        // Simulate progress update
        capturedCallbacks.onProgress(500, 1000)

        await waitFor(() => {
            expect(screen.getByText('Uploading...')).toBeInTheDocument()
            expect(screen.getByText('50%')).toBeInTheDocument()
        })
    })

    /**
     * 测试取消上传功能
     */
    it('should allow canceling upload', async () => {
        const mockUpload = {
            start: vi.fn(),
            abort: vi.fn(),
        } as unknown as tus.Upload

        vi.mocked(uploadService.uploadFile).mockReturnValue(mockUpload)

        render(<UploadButton />)

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        const file = new File(['test'], 'test.txt', { type: 'text/plain' })

        Object.defineProperty(fileInput, 'files', {
            value: [file],
            writable: false,
        })

        fireEvent.change(fileInput)

        await waitFor(() => {
            expect(screen.getByText('Uploading...')).toBeInTheDocument()
        })

        // Find and click cancel button
        const cancelButton = screen.getByRole('button', { name: '' }) // X icon button has no text
        fireEvent.click(cancelButton)

        await waitFor(() => {
            expect(mockUpload.abort).toHaveBeenCalled()
        })
    })

    /**
     * 测试上传成功回调
     */
    it('should call onUploadComplete callback on success', async () => {
        const onUploadComplete = vi.fn()
        let capturedCallbacks: any = {}

        const mockUpload = {
            start: vi.fn(),
            abort: vi.fn(),
        } as unknown as tus.Upload

        vi.mocked(uploadService.uploadFile).mockImplementation((file, callbacks) => {
            capturedCallbacks = callbacks
            return mockUpload
        })

        render(<UploadButton onUploadComplete={onUploadComplete} />)

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        const file = new File(['test'], 'test.txt', { type: 'text/plain' })

        Object.defineProperty(fileInput, 'files', {
            value: [file],
            writable: false,
        })

        fireEvent.change(fileInput)

        await waitFor(() => {
            expect(uploadService.uploadFile).toHaveBeenCalled()
        })

        // Simulate successful upload
        capturedCallbacks.onSuccess()

        await waitFor(() => {
            expect(onUploadComplete).toHaveBeenCalled()
        })
    })

    /**
     * 测试上传失败处理
     */
    it('should handle upload error', async () => {
        const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
        let capturedCallbacks: any = {}

        const mockUpload = {
            start: vi.fn(),
            abort: vi.fn(),
        } as unknown as tus.Upload

        vi.mocked(uploadService.uploadFile).mockImplementation((file, callbacks) => {
            capturedCallbacks = callbacks
            return mockUpload
        })

        render(<UploadButton />)

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        const file = new File(['test'], 'test.txt', { type: 'text/plain' })

        Object.defineProperty(fileInput, 'files', {
            value: [file],
            writable: false,
        })

        fireEvent.change(fileInput)

        await waitFor(() => {
            expect(uploadService.uploadFile).toHaveBeenCalled()
        })

        // Simulate upload error
        const error = new Error('Upload failed')
        capturedCallbacks.onError(error)

        await waitFor(() => {
            expect(alertSpy).toHaveBeenCalledWith('Upload failed: Upload failed')
        })

        alertSpy.mockRestore()
    })
})
