import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ArrowRight, ShieldCheck, Zap, Lock, AlertCircle, FileIcon, Download, Loader2 } from 'lucide-react'
import { shareService, ShareInfo } from '@/services/shareService'
import { cn } from '@/lib/utils'

function App() {
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

      // Check for special codes (like password required) - usually handled by error interceptor logic 
      // but let's assume successful response returns data
      setShareInfo(response.data)

    } catch (err: any) {
      console.error(err)
      // Handle password required error (code 4040 based on API.md)
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
    <MainLayout>
      <div className="flex flex-col items-center justify-center space-y-12 py-12 md:py-24">

        {/* Hero Section */}
        {!shareInfo ? (
          <div className="text-center space-y-4 max-w-2xl animate-in fade-in slide-in-from-bottom-8 duration-500">
            <div className="inline-flex items-center rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-sm font-medium text-primary mb-4">
              <ShieldCheck className="mr-1 h-3.5 w-3.5" />
              End-to-End Encrypted
            </div>
            <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl lg:text-7xl bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
              {t('home.title')}
            </h1>
            <p className="text-muted-foreground text-lg md:text-xl max-w-[600px] mx-auto">
              {t('home.subtitle')}
            </p>
          </div>
        ) : (
          <div className="text-center space-y-4 animate-in fade-in zoom-in duration-300">
            <h1 className="text-3xl font-bold">Files Ready for Pickup</h1>
            <p className="text-muted-foreground">Expires at: {new Date(shareInfo.expires_at).toLocaleString()}</p>
          </div>
        )}

        {/* Pickup Form or Result (Glass Card) */}
        <Card className="w-full max-w-md border-primary/20 bg-surface/40 backdrop-blur-xl shadow-2xl transition-all hover:shadow-primary/10 relative overflow-hidden">

          {/* Error Banner */}
          {error && (
            <div className="absolute top-0 left-0 right-0 bg-destructive/10 text-destructive text-sm p-2 text-center border-b border-destructive/20 flex items-center justify-center">
              <AlertCircle className="w-4 h-4 mr-2" />
              {error}
            </div>
          )}

          <CardHeader className={cn(error && "pt-10")}>
            <CardTitle>{shareInfo ? "Your Files" : "Retrieve File"}</CardTitle>
            <CardDescription>{shareInfo ? "Click download to save to your device" : "Enter your secure pickup code"}</CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            {!shareInfo ? (
              <>
                <div className="relative">
                  <Input
                    placeholder="AHA-XXXX-XXXX"
                    className="text-center text-lg tracking-widest font-mono uppercase h-14 bg-background/50 border-primary/20 focus:border-primary/50"
                    maxLength={8}
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

        {/* Features Grid (Only show when not viewing files) */}
        {!shareInfo && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full mt-12 animate-in fade-in slide-in-from-bottom-8 duration-700 delay-100">
            {[
              { icon: Lock, title: "Zero Knowledge", desc: "We can't read your files even if we wanted to." },
              { icon: Zap, title: "Lightning Fast", desc: "Up to 1GB/s transfer speeds with global CDN." },
              { icon: ShieldCheck, title: "Audited Security", desc: "Verified by top security firms." },
            ].map((feature, i) => (
              <Card key={i} className="bg-surface/30 border-transparent hover:border-primary/10 transition-colors">
                <CardHeader>
                  <feature.icon className="h-8 w-8 text-primary mb-2" />
                  <CardTitle className="text-lg">{feature.title}</CardTitle>
                  <CardDescription>{feature.desc}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        )}

      </div>
    </MainLayout>
  )
}

export default App
