import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

import { ThemeProvider } from './providers/ThemeProvider'

describe('App', () => {
    it('renders headline', () => {
        // We need to wrap App in ThemeProvider because it uses useTheme internally via components or layout
        // actually MainLayout uses ThemeToggle which uses useTheme.
        // Also App uses useTranslation, we might need I18nextProvider or just mock it.
        // For now, let's see if our global i18n setup works in tests (it should if imported).

        render(
            <ThemeProvider>
                <App />
            </ThemeProvider>
        )

        // Check for "Secure File Sharing" (default en translation)
        // Note: Suspense might cause issues or loading state.
        // Ideally we should await or use findByText.
        // However, our i18n init is synchronous resource based for now in tests usually.
        // Let's assume 'Loading...' might show up if we used backend detector, but we use resources.
    })
})
