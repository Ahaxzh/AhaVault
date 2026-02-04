import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ArrowRight, ShieldCheck, Zap, Lock, AlertCircle, FileIcon, Download, Loader2 } from 'lucide-react'
import { shareService, type ShareInfo } from '@/services/shareService'
import { cn } from '@/lib/utils'

export default function Home() {
    const { t } = useTranslation()
    const [code, setCode] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [shareInfo, setShareInfo] = useState<ShareInfo | null>(null)

    // Need password state?
    const [password, setPassword] = useState('')
    const [requiresPassword, setRequiresPassword] = useState(false)

    const handlePickup = async () => {
        if (!code || code.length < 8) {
            setError("Please enter a valid 8-digit code")
            return
        }

        setLoading(true)
        setError(null)
        setShareInfo(null)
        setRequiresPassword(false)

        try {
            const response = await shareService.getShareByCode(code, password)
            setShareInfo(response.data)
        } catch (err: any) {
            console.error(err)
            if (err.code === 4040) {
                setRequiresPassword(true)
                setError("Password required")
            } else {
                setError(err.message || "Failed to retrieve file")
            }
        } finally {
            setLoading(false)
        }
    }

    const formatSize = (bytes: number) => {
        if (bytes === 0) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    return (
        <MainLayout className="h-[calc(100vh-4rem)]">
            <div className="container mx-auto h-full flex flex-col lg:flex-row items-center justify-between px-6 gap-12">

                {/* Left Side: Introduction */}
                <div className="flex-1 flex flex-col justify-center space-y-8 animate-in fade-in slide-in-from-left-8 duration-700">
                    <div className="space-y-4">
                        <div className="inline-flex items-center rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
                            <ShieldCheck className="mr-1 h-3.5 w-3.5" />
                            End-to-End Encrypted
                        </div>
                        <h1 className="text-5xl font-extrabold tracking-tight sm:text-6xl lg:text-7xl">
                            {t('home.title')}
                        </h1>
                        <p className="text-xl text-muted-foreground max-w-lg">
                            {t('home.subtitle')}
                        </p>
                    </div>

                    <div className="flex flex-col gap-4 text-sm text-muted-foreground">
                        <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                <Lock className="h-4 w-4" />
                            </div>
                            <span>Zero Knowledge Encryption</span>
                        </div>
                        <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                <Zap className="h-4 w-4" />
                            </div>
                            <span>Lightning Fast Global CDN</span>
                        </div>
                        <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                <ShieldCheck className="h-4 w-4" />
                            </div>
                            <span>Audited & Secure Infrastructure</span>
                        </div>
                    </div>
                </div>

                {/* Right Side: Pickup Form */}
                <div className="flex-1 w-full max-w-md animate-in fade-in slide-in-from-right-8 duration-700 delay-100">
                    <Card className="border-primary/20 bg-surface/40 backdrop-blur-xl shadow-2xl relative overflow-hidden">

                        {/* Error Banner */}
                        {error && (
                            <div className="absolute top-0 left-0 right-0 bg-destructive/10 text-destructive text-sm p-2 text-center border-b border-destructive/20 flex items-center justify-center">
                                <AlertCircle className="w-4 h-4 mr-2" />
                                {error}
                            </div>
                        )}

                        <CardHeader className={cn("space-y-1", error && "pt-10")}>
                            <CardTitle className="text-2xl">{shareInfo ? "Your Files" : "Retrieve File"}</CardTitle>
                            <CardDescription>
                                {shareInfo
                                    ? `Expires at: ${new Date(shareInfo.expires_at).toLocaleString()}`
                                    : "Enter your 8-digit secure pickup code"
                                }
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="space-y-4">
                            {!shareInfo ? (
                                <>
                                    <div className="relative">
                                        <Input
                                            placeholder="AHA-XXXX-XXXX"
                                            className="text-center text-2xl tracking-widest font-mono uppercase h-16 bg-background/50 border-primary/20 focus:border-primary/50"
                                            maxLength={13} // AHA-XXXX-XXXX = 13 chars
                                            value={code}
                                            onChange={(e) => setCode(e.target.value.toUpperCase())}
                                            onKeyDown={(e) => e.key === 'Enter' && handlePickup()}
                                            disabled={loading}
                                        />
                                    </div>
                                    {requiresPassword && (
                                        <div className="animate-in fade-in slide-in-from-top-2">
                                            <Input
                                                type="password"
                                                placeholder="Enter Access Password"
                                                className="text-center h-10 mt-2"
                                                value={password}
                                                onChange={(e) => setPassword(e.target.value)}
                                            />
                                        </div>
                                    )}
                                </>
                            ) : (
                                <div className="space-y-2">
                                    {shareInfo.files.map((file) => (
                                        <div key={file.file_id} className="flex items-center justify-between p-3 rounded-lg bg-background/50 border border-border/50">
                                            <div className="flex items-center space-x-3 overflow-hidden">
                                                <FileIcon className="w-8 h-8 text-primary/80 flex-shrink-0" />
                                                <div className="truncate">
                                                    <p className="font-medium truncate text-sm">{file.filename}</p>
                                                    <p className="text-xs text-muted-foreground">{formatSize(file.size)}</p>
                                                </div>
                                            </div>
                                            <Button
                                                size="sm"
                                                variant="ghost"
                                                className="text-primary hover:text-primary hover:bg-primary/10"
                                                onClick={() => window.open(`${import.meta.env.VITE_API_URL || '/api'}/public/pickup/${code}/files/${file.file_id}/download`)}
                                            >
                                                <Download className="w-4 h-4" />
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>

                        <CardFooter>
                            {!shareInfo ? (
                                <Button
                                    className="w-full h-12 text-lg shadow-lg shadow-primary/20"
                                    size="lg"
                                    onClick={handlePickup}
                                    disabled={loading}
                                >
                                    {loading ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Fetching...
                                        </>
                                    ) : (
                                        <>
                                            {t('home.pickup_button')} <ArrowRight className="ml-2 h-4 w-4" />
                                        </>
                                    )}
                                </Button>
                            ) : (
                                <Button
                                    variant="outline"
                                    className="w-full"
                                    onClick={() => {
                                        setShareInfo(null)
                                        setCode('')
                                        setPassword('')
                                    }}
                                >
                                    Pick up another
                                </Button>
                            )}
                        </CardFooter>
                    </Card>
                </div>
            </div>
        </MainLayout>
    )
}
