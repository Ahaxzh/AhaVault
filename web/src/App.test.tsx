import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import App from './App'

import { ThemeProvider } from './providers/ThemeProvider'
import { shareService } from './services/shareService'

// Mock dependencies
vi.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string) => key,
        i18n: {
            changeLanguage: () => new Promise(() => { }),
        },
    }),
}))

vi.mock('./services/shareService', () => ({
    shareService: {
        getShareByCode: vi.fn(),
    },
}))

describe('App - Pickup Flow', () => {

    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('renders landing page correctly', () => {
        render(
            <ThemeProvider>
                <App />
            </ThemeProvider>
        )
        expect(screen.getByText('home.title')).toBeInTheDocument()
        expect(screen.getByPlaceholderText('AHA-XXXX-XXXX')).toBeInTheDocument()
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

        render(
            <ThemeProvider>
                <App />
            </ThemeProvider>
        )

        // Type code
        const input = screen.getByPlaceholderText('AHA-XXXX-XXXX')
        fireEvent.change(input, { target: { value: 'ABCDEF12' } })

        // Click button
        const button = screen.getByText('home.pickup_button')
        fireEvent.click(button)

        // Check loading state
        expect(screen.getByText('Fetching...')).toBeInTheDocument()

        // Wait for result
        await waitFor(() => {
            expect(screen.getByText('Files Ready for Pickup')).toBeInTheDocument()
            expect(screen.getByText('test.png')).toBeInTheDocument()
        })
    })

    it('handles validation error (too short)', async () => {
        render(
            <ThemeProvider>
                <App />
            </ThemeProvider>
        )

        const input = screen.getByPlaceholderText('AHA-XXXX-XXXX')
        fireEvent.change(input, { target: { value: 'SHORT' } })

        const button = screen.getByText('home.pickup_button')
        fireEvent.click(button)

        expect(screen.getByText('Please enter a valid 8-digit code')).toBeInTheDocument()
        expect(shareService.getShareByCode).not.toHaveBeenCalled()
    })
})
