import { ThemeToggle } from "@/components/ui/ThemeToggle"
import { LanguageSwitcher } from "@/components/ui/LanguageSwitcher"

interface MainLayoutProps {
    children: React.ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
    return (
        <div className="min-h-screen bg-background text-foreground transition-colors duration-300">
            {/* Navbar */}
            <header className="fixed top-0 left-0 right-0 z-50 border-b bg-background/80 backdrop-blur-md">
                <div className="container mx-auto flex h-16 items-center justify-between px-4">
                    {/* Logo */}
                    <div className="flex items-center gap-2">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
                            A
                        </div>
                        <span className="text-xl font-bold tracking-tight">AhaVault</span>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-2">
                        <LanguageSwitcher />
                        <ThemeToggle />
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="container mx-auto px-4 pt-24 pb-12">
                <div className="mx-auto max-w-4xl animate-in fade-in slide-in-from-bottom-4 duration-700">
                    {children}
                </div>
            </main>

            {/* Footer */}
            <footer className="border-t py-6 text-center text-sm text-muted-foreground">
                <div className="container mx-auto px-4">
                    <p>© 2026 AhaVault. Secure & Encrypted.</p>
                </div>
            </footer>
        </div>
    )
}
