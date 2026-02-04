import { Link } from "react-router-dom"
import { ThemeToggle } from "@/components/ui/ThemeToggle"
import { LanguageSwitcher } from "@/components/ui/LanguageSwitcher"
import { Button } from "@/components/ui/Button"

interface MainLayoutProps {
    children: React.ReactNode
    className?: string
}

export function MainLayout({ children, className }: MainLayoutProps) {
    return (
        <div className="flex min-h-screen flex-col bg-background text-foreground overflow-hidden">
            {/* Navbar */}
            <header className="fixed top-0 left-0 right-0 z-50 border-b bg-background/80 backdrop-blur-md h-16">
                <div className="container mx-auto flex h-full items-center justify-between px-6">
                    {/* Logo */}
                    <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
                            A
                        </div>
                        <span className="text-xl font-bold tracking-tight">AhaVault</span>
                    </Link>

                    {/* Actions */}
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-1 border-r border-border/50 pr-4 mr-1">
                            <LanguageSwitcher />
                            <ThemeToggle />
                        </div>

                        <div className="flex items-center gap-2">
                            <Link to="/login">
                                <Button variant="ghost" size="sm">Sign In</Button>
                            </Link>
                            <Link to="/register">
                                <Button size="sm">Get Started</Button>
                            </Link>
                        </div>
                    </div>
                </div>
            </header>

            {/* Main Content - Full Height, pushes content down by header height */}
            <main className={`flex-1 pt-16 flex flex-col ${className}`}>
                {children}
            </main>
        </div>
    )
}
