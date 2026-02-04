import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { AlertCircle, Loader2, LogIn } from 'lucide-react'
import { authService } from '@/services/authService'
import { useAuthStore } from '@/store/useAuthStore'

export default function Login() {
    const { t } = useTranslation()
    const navigate = useNavigate()
    const setAuth = useAuthStore((state) => state.setAuth)

    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!email || !password) {
            setError("Please enter both email and password")
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await authService.login({ email, password })

            const { user_id, email: userEmail, role, token } = response.data

            // Update global store
            setAuth({ user_id, email: userEmail, role: role || 'user' }, token)

            // Redirect to cabinet or home
            navigate('/cabinet')

        } catch (err: any) {
            console.error(err)
            setError(err.message || "Login failed")
        } finally {
            setLoading(false)
        }
    }

    return (
        <MainLayout>
            <div className="flex flex-col items-center justify-center min-h-[calc(100vh-200px)] py-12">
                <Card className="w-full max-w-md border-primary/20 bg-surface/40 backdrop-blur-xl shadow-2xl">
                    <CardHeader className="space-y-1">
                        <CardTitle className="text-2xl font-bold flex items-center">
                            <LogIn className="w-6 h-6 mr-2 text-primary" />
                            Sign In
                        </CardTitle>
                        <CardDescription>
                            Enter your email and password to access your vault
                        </CardDescription>
                    </CardHeader>

                    <form onSubmit={handleLogin}>
                        <CardContent className="space-y-4">
                            {error && (
                                <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md flex items-center">
                                    <AlertCircle className="w-4 h-4 mr-2" />
                                    {error}
                                </div>
                            )}

                            <div className="space-y-2">
                                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                    Email
                                </label>
                                <Input
                                    type="email"
                                    placeholder="m@example.com"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    disabled={loading}
                                />
                            </div>

                            <div className="space-y-2">
                                <div className="flex items-center justify-between">
                                    <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        Password
                                    </label>
                                    <Link
                                        to="/forgot-password"
                                        className="text-sm font-medium text-primary hover:underline underline-offset-4"
                                    >
                                        Forgot password?
                                    </Link>
                                </div>
                                <Input
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    disabled={loading}
                                />
                            </div>
                        </CardContent>

                        <CardFooter className="flex flex-col space-y-4">
                            <Button className="w-full" type="submit" disabled={loading}>
                                {loading ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Signing in...
                                    </>
                                ) : (
                                    "Sign In"
                                )}
                            </Button>

                            <div className="text-center text-sm text-muted-foreground">
                                Don't have an account?{" "}
                                <Link to="/register" className="text-primary hover:underline underline-offset-4 font-medium">
                                    Sign up
                                </Link>
                            </div>
                        </CardFooter>
                    </form>
                </Card>
            </div>
        </MainLayout>
    )
}
