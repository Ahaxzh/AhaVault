/**
 * @file Home.tsx
 * @description 首页 - 取件码输入与文件取件页
 *
 * 功能说明：
 *  - 展示品牌介绍与核心特性（左侧分屏）
 *  - 提供取件码输入框（右侧分屏）
 *  - 支持取件码验证（8位字符）
 *  - 支持访问密码二次验证
 *  - 展示取件结果（文件列表）
 *  - 支持一键下载文件
 *
 * @author AhaVault Team
 * @created 2026-02-05
 */

import { useState } from 'react'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ArrowRight, ShieldCheck, Zap, Lock, AlertCircle, FileIcon, Download, Loader2, Package, DownloadCloud } from 'lucide-react'
import { shareService, type ShareInfo } from '@/services/shareService'
import { formatFileSize } from '@/utils/format'
import { cn } from '@/lib/utils'

export default function Home() {
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

    return (
        <MainLayout className="h-[calc(100vh-4rem)] relative">
            {/* Background Image with Overlay */}
            <div className="absolute inset-0 z-0 select-none pointer-events-none">
                <img
                    src="/hero-bg.png"
                    alt="Background Pattern"
                    className="w-full h-full object-cover opacity-[0.15] dark:opacity-[0.2]"
                />
                <div className="absolute inset-0 bg-gradient-to-b from-transparent via-background/50 to-background" />
            </div>

            <div className="container relative z-10 mx-auto h-full flex flex-col lg:flex-row items-center justify-center lg:justify-between px-6 lg:px-12 gap-12 lg:gap-24 min-h-[800px] pt-24 lg:pt-0">

                {/* Left Side: Introduction */}
                <div className="flex-1 flex flex-col justify-center space-y-8 animate-in fade-in slide-in-from-left-8 duration-700 pt-10 lg:pt-0">
                    <div className="space-y-6">
                        <div className="inline-flex items-center rounded-full border border-primary/20 bg-primary/10 backdrop-blur-md px-4 py-1.5 text-sm font-semibold text-primary shadow-sm hover:bg-primary/20 transition-colors cursor-default">
                            <ShieldCheck className="mr-2 h-4 w-4" />
                            Global Secure File Exchange
                        </div>

                        <div className="space-y-2">
                            <h1 className="text-5xl font-extrabold tracking-tight sm:text-6xl lg:text-7xl">
                                <span className="block text-foreground drop-shadow-sm">Your Data,</span>
                                <span className="bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent pb-2 block">
                                    Truly Secure.
                                </span>
                            </h1>
                            <p className="text-xl md:text-2xl text-muted-foreground/90 max-w-lg leading-relaxed font-light">
                                End-to-end encrypted file sharing. No ads, no tracking, just pure privacy.
                            </p>
                        </div>
                    </div>

                    <div className="flex flex-col gap-5">
                        {[
                            { icon: Lock, title: "Zero Knowledge Encryption", desc: "Data is encrypted before it leaves your device." },
                            { icon: Zap, title: "Lighting Fast Transfer", desc: "Optimized global CDN for maximum speed." },
                            { icon: ShieldCheck, title: "Ephemeral & Private", desc: "Set expiration times. No logs kept." }
                        ].map((item, i) => (
                            <div key={i} className="flex items-start gap-4 p-3 rounded-xl hover:bg-white/5 transition-colors border border-transparent hover:border-white/10 group">
                                <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-primary-foreground transition-colors shrink-0">
                                    <item.icon className="h-5 w-5" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-foreground">{item.title}</h3>
                                    <p className="text-sm text-muted-foreground">{item.desc}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Right Side: Pickup Form */}
                <div className="flex-1 w-full max-w-md animate-in fade-in slide-in-from-right-8 duration-700 delay-100 pb-10 lg:pb-0">
                    <Card className="border-white/20 bg-white/40 dark:bg-black/40 backdrop-blur-2xl shadow-2xl relative overflow-hidden ring-1 ring-black/5 dark:ring-white/10">
                        {/* Decorative Top Accent */}
                        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-500 to-indigo-500 opacity-80" />

                        {/* Error Banner */}
                        {error && (
                            <div className="absolute top-1 left-0 right-0 bg-red-500/10 text-red-600 dark:text-red-400 text-sm p-2 text-center border-b border-red-500/20 flex items-center justify-center backdrop-blur-sm">
                                <AlertCircle className="w-4 h-4 mr-2" />
                                {error}
                            </div>
                        )}

                        <CardHeader className={cn("space-y-1", error && "pt-12")}>
                            <CardTitle className="text-2xl font-bold flex items-center gap-2">
                                {shareInfo ? (
                                    <>
                                        <div className="p-1.5 bg-blue-100 dark:bg-blue-900/30 rounded-md">
                                            <Package className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                                        </div>
                                        <span>File Found</span>
                                    </>
                                ) : (
                                    <>
                                        <div className="p-1.5 bg-blue-100 dark:bg-blue-900/30 rounded-md">
                                            <DownloadCloud className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                                        </div>
                                        <span>Retrieve File</span>
                                    </>
                                )}
                            </CardTitle>
                            <CardDescription className="text-sm opacity-90">
                                {shareInfo
                                    ? `Expires at: ${new Date(shareInfo.expires_at).toLocaleString()}`
                                    : "Enter your secure 8-digit code to download."
                                }
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="space-y-5">
                            {!shareInfo ? (
                                <>
                                    <div className="relative group">
                                        <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-lg blur opacity-20 group-hover:opacity-50 transition duration-1000 group-hover:duration-200" />
                                        <Input
                                            placeholder="AHA-XXXX-XXXX"
                                            className="relative text-center text-2xl tracking-widest font-mono uppercase h-16 bg-background/80 border-transparent focus:border-primary/50 shadow-inner"
                                            maxLength={13}
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
                                                className="text-center h-10 mt-2 bg-background/50"
                                                value={password}
                                                onChange={(e) => setPassword(e.target.value)}
                                            />
                                        </div>
                                    )}
                                </>
                            ) : (
                                <div className="space-y-3">
                                    {shareInfo.files.map((file) => (
                                        <div key={file.file_id} className="flex items-center justify-between p-4 rounded-xl bg-background/60 border border-white/10 shadow-sm hover:bg-background/80 transition-all group">
                                            <div className="flex items-center space-x-3 overflow-hidden">
                                                <div className="h-10 w-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400">
                                                    <FileIcon className="w-5 h-5" />
                                                </div>
                                                <div className="truncate">
                                                    <p className="font-semibold truncate text-sm text-foreground">{file.filename}</p>
                                                    <p className="text-xs text-muted-foreground font-mono mt-0.5">{formatFileSize(file.size)}</p>
                                                </div>
                                            </div>
                                            <Button
                                                size="sm"
                                                className="shadow-md bg-blue-600 hover:bg-blue-700 text-white border-0"
                                                onClick={() => window.open(`${import.meta.env.VITE_API_URL || '/api'}/public/pickup/${code}/files/${file.file_id}/download`)}
                                            >
                                                <Download className="w-4 h-4 mr-1" /> Save
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>

                        <CardFooter>
                            {!shareInfo ? (
                                <Button
                                    className="w-full h-14 text-lg font-bold shadow-xl shadow-blue-500/20 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 border-0 transition-all hover:scale-[1.02] active:scale-[0.98]"
                                    size="lg"
                                    onClick={handlePickup}
                                    disabled={loading}
                                >
                                    {loading ? (
                                        <>
                                            <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                                            Verifying...
                                        </>
                                    ) : (
                                        <>
                                            Retrieve File <ArrowRight className="ml-2 h-5 w-5" />
                                        </>
                                    )}
                                </Button>
                            ) : (
                                <Button
                                    variant="secondary"
                                    className="w-full h-12 border border-input bg-background/50 hover:bg-background/80 backdrop-blur-sm"
                                    onClick={() => {
                                        setShareInfo(null)
                                        setCode('')
                                        setPassword('')
                                    }}
                                >
                                    Pick up another file
                                </Button>
                            )}
                        </CardFooter>
                    </Card>

                    {/* Trust Badges below card */}
                    <div className="mt-6 flex justify-center gap-6 opacity-70 grayscale hover:grayscale-0 transition-all duration-500">
                        {/* Placeholder for logos if needed, or just keep it clean */}
                    </div>
                </div>
            </div>
        </MainLayout>
    )
}
