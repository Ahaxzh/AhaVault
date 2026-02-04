import { Languages } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/Button"

export function LanguageSwitcher() {
    const { i18n } = useTranslation()

    const toggleLanguage = () => {
        const nextLang = i18n.language === "zh" ? "en" : "zh"
        i18n.changeLanguage(nextLang)
    }

    return (
        <Button
            variant="ghost"
            size="icon"
            onClick={toggleLanguage}
            title="Switch language"
        >
            <Languages className="h-[1.2rem] w-[1.2rem]" />
            <span className="absolute -top-1 -right-1 flex h-3 w-3 items-center justify-center rounded-full bg-primary text-[8px] text-primary-foreground">
                {i18n.language === "zh" ? "中" : "En"}
            </span>
        </Button>
    )
}
