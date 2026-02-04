import { useTranslation } from 'react-i18next'
import { MainLayout } from '@/components/layout/MainLayout'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ArrowRight, ShieldCheck, Zap, Lock } from 'lucide-react'

function App() {
  const { t } = useTranslation()

  return (
    <MainLayout>
      <div className="flex flex-col items-center justify-center space-y-12 py-12 md:py-24">

        {/* Hero Section */}
        <div className="text-center space-y-4 max-w-2xl">
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

        {/* Pickup Form (Glass Card) */}
        <Card className="w-full max-w-md border-primary/20 bg-surface/40 backdrop-blur-xl shadow-2xl transition-all hover:shadow-primary/10">
          <CardHeader>
            <CardTitle>Retrieve File</CardTitle>
            <CardDescription>Enter your secure 8-digit pickup code</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="relative">
              <Input
                placeholder="AHA-XXXX-XXXX"
                className="text-center text-lg tracking-widest font-mono uppercase h-14 bg-background/50 border-primary/20 focus:border-primary/50"
                maxLength={8}
              />
            </div>
          </CardContent>
          <CardFooter>
            <Button className="w-full h-12 text-lg shadow-lg shadow-primary/20" size="lg">
              {t('home.pickup_button')} <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </CardFooter>
        </Card>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full mt-12">
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

      </div>
    </MainLayout>
  )
}

export default App
