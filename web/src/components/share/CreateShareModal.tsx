import { useState } from 'react'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { shareService } from '@/services/shareService'
import { Loader2, X, Copy, Check } from 'lucide-react'

interface CreateShareModalProps {
    fileIds: string[]
    onClose: () => void
    onSuccess?: () => void
}

export function CreateShareModal({ fileIds, onClose, onSuccess }: CreateShareModalProps) {
    const [loading, setLoading] = useState(false)
    const [step, setStep] = useState<'form' | 'result'>('form')

    // Form State
    const [expiry, setExpiry] = useState(3600 * 24) // Default 24 hours
    const [maxDownloads, setMaxDownloads] = useState(10)
    const [password, setPassword] = useState('')

    // Result State
    const [result, setResult] = useState<{ pickup_code: string, expires_at: string } | null>(null)
    const [copied, setCopied] = useState(false)

    const handleCreate = async () => {
        setLoading(true)
        try {
            const res = await shareService.createShare({
                file_ids: fileIds,
                expires_in: expiry,
                max_downloads: maxDownloads,
                password: password || undefined
            })
            setResult(res.data)
            setStep('result')
            if (onSuccess) onSuccess()
        } catch (error) {
            console.error("Failed to create share", error)
            alert("Failed to create share")
        } finally {
            setLoading(false)
        }
    }

    const copyToClipboard = () => {
        if (!result) return
        navigator.clipboard.writeText(result.pickup_code)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
            <Card className="w-full max-w-md bg-surface border-primary/20 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <div className="flex flex-col space-y-1.5">
                        <CardTitle>{step === 'form' ? 'Create Share' : 'Ready to Share!'}</CardTitle>
                        <CardDescription>
                            {step === 'form'
                                ? `Sharing ${fileIds.length} file(s)`
                                : 'Your files are ready for pickup'}
                        </CardDescription>
                    </div>
                    <Button variant="ghost" size="icon" className="-mr-2" onClick={onClose}>
                        <X className="h-4 w-4" />
                    </Button>
                </CardHeader>

                <CardContent className="space-y-4 pt-4">
                    {step === 'form' ? (
                        <>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">Expires In</label>
                                <select
                                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    value={expiry}
                                    onChange={(e) => setExpiry(Number(e.target.value))}
                                >
                                    <option value={3600}>1 Hour</option>
                                    <option value={86400}>24 Hours</option>
                                    <option value={604800}>7 Days</option>
                                </select>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">Max Downloads</label>
                                <select
                                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                    value={maxDownloads}
                                    onChange={(e) => setMaxDownloads(Number(e.target.value))}
                                >
                                    <option value={1}>1 time (Burn after reading)</option>
                                    <option value={5}>5 times</option>
                                    <option value={10}>10 times</option>
                                    <option value={100}>100 times</option>
                                </select>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium">Password (Optional)</label>
                                <Input
                                    placeholder="Leave empty for no password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                />
                            </div>
                        </>
                    ) : (
                        <div className="flex flex-col items-center space-y-4 py-4">
                            <div className="text-center space-y-2">
                                <p className="text-sm text-muted-foreground">Pickup Code</p>
                                <div className="text-4xl font-mono font-bold tracking-widest text-primary">
                                    {result?.pickup_code}
                                </div>
                            </div>

                            <Button variant="outline" className="w-full gap-2" onClick={copyToClipboard}>
                                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                {copied ? "Copied!" : "Copy Code"}
                            </Button>
                        </div>
                    )}
                </CardContent>

                <CardFooter>
                    {step === 'form' ? (
                        <Button className="w-full" onClick={handleCreate} disabled={loading}>
                            {loading ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Creating...
                                </>
                            ) : (
                                "Create Link"
                            )}
                        </Button>
                    ) : (
                        <Button className="w-full" onClick={onClose}>
                            Done
                        </Button>
                    )}
                </CardFooter>
            </Card>
        </div>
    )
}
