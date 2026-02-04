import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { AlertCircle, Loader2, UserPlus } from 'lucide-react'
import { authService } from '@/services/authService'
import { useAuthStore } from '@/store/useAuthStore'

export default function Register() {
    const { t } = useTranslation()
    const navigate = useNavigate()
    const setAuth = useAuthStore((state) => state.setAuth)

    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [inviteCode, setInviteCode] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleRegister = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!email || !password) {
            setError("Please fill in required fields")
            return
        }

        if (password.length < 8) {
            setError("Password must be at least 8 characters")
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await authService.register({ email, password, invite_code: inviteCode })

            const { user_id, email: userEmail, role, token } = response.data

            // Update global store
            setAuth({ user_id, email: userEmail, role: role || 'user' }, token)

            // Redirect to cabinet
            navigate('/cabinet')

        } catch (err: any) {
            console.error(err)
            setError(err.message || "Registration failed")
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
                            <UserPlus className="w-6 h-6 mr-2 text-primary" />
                            Create Account
                        </CardTitle>
                        <CardDescription>
                            Start using your secure personal vault today
                        </CardDescription>
                    </CardHeader>

                    <form onSubmit={handleRegister}>
                        <CardContent className="space-y-4">
                            {error && (
                                <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md flex items-center">
                                    <AlertCircle className="w-4 h-4 mr-2" />
                                    {error}
                                </div>
                            )}

                            <div className="space-y-2">
                                <label className="text-sm font-medium leading-none">
                                    Email <span className="text-destructive">*</span>
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
                                <label className="text-sm font-medium leading-none">
                                    Password <span className="text-destructive">*</span>
                                </label>
                                <Input
                                    type="password"
                                    placeholder="At least 8 characters"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    disabled={loading}
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium leading-none text-muted-foreground">
                                    Invite Code (Optional)
                                </label>
                                <Input
                                    placeholder="AHA-2026"
                                    value={inviteCode}
                                    onChange={(e) => setInviteCode(e.target.value)}
                                    disabled={loading}
                                />
                            </div>

                        </CardContent>

                        <CardFooter className="flex flex-col space-y-4">
                            <Button className="w-full" type="submit" disabled={loading}>
                                {loading ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Creating account...
                                    </>
                                ) : (
                                    "Create Account"
                                )}
                            </Button>

                            <div className="text-center text-sm text-muted-foreground">
                                Already have an account?{" "}
                                <Link to="/login" className="text-primary hover:underline underline-offset-4 font-medium">
                                    Sign in
                                </Link>
                            </div>
                        </CardFooter>
                    </form>
                </Card>
            </div>
        </MainLayout>
    )
}
