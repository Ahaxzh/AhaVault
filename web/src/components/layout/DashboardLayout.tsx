import { Link, useNavigate } from 'react-router-dom'
import { ThemeToggle } from "@/components/ui/ThemeToggle"
import { LanguageSwitcher } from "@/components/ui/LanguageSwitcher"
import { Button } from "@/components/ui/Button"
import { useAuthStore } from '@/store/useAuthStore'
import { authService } from '@/services/authService'
import { LogOut, User as UserIcon, HardDrive, Share2 } from 'lucide-react'

interface DashboardLayoutProps {
    children: React.ReactNode
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
    const navigate = useNavigate()
    const { user, clearAuth } = useAuthStore()

    const handleLogout = async () => {
        try {
            await authService.logout()
        } catch (error) {
            console.error('Logout failed', error)
        } finally {
            clearAuth()
            navigate('/login')
        }
    }

    return (
        <div className="min-h-screen bg-background text-foreground transition-colors duration-300">
            {/* Navbar */}
            <header className="fixed top-0 left-0 right-0 z-50 border-b bg-background/80 backdrop-blur-md">
                <div className="container mx-auto flex h-16 items-center justify-between px-4">
                    {/* Logo & Nav */}
                    <div className="flex items-center gap-8">
                        <Link to="/cabinet" className="flex items-center gap-2">
                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
                                A
                            </div>
                            <span className="text-xl font-bold tracking-tight hidden md:inline-block">AhaVault</span>
                        </Link>

                        <nav className="hidden md:flex items-center gap-1">
                            <Link to="/cabinet">
                                <Button variant="ghost" className="gap-2">
                                    <HardDrive className="w-4 h-4" />
                                    Cabinet
                                </Button>
                            </Link>
                            <Link to="/shares">
                                <Button variant="ghost" className="gap-2">
                                    <Share2 className="w-4 h-4" />
                                    My Shares
                                </Button>
                            </Link>
                        </nav>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-2">
                        <div className="hidden md:flex items-center gap-2 mr-4 text-sm text-muted-foreground">
                            <UserIcon className="w-4 h-4" />
                            <span>{user?.email}</span>
                        </div>

                        <LanguageSwitcher />
                        <ThemeToggle />

                        <Button variant="ghost" size="icon" onClick={handleLogout} title="Logout">
                            <LogOut className="w-4 h-4" />
                        </Button>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="container mx-auto px-4 pt-24 pb-12">
                <div className="mx-auto max-w-6xl animate-in fade-in slide-in-from-bottom-4 duration-500">
                    {children}
                </div>
            </main>
        </div>
    )
}
