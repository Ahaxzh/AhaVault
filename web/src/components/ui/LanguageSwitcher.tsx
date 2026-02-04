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
            size="sm"
            onClick={toggleLanguage}
            title="Switch language"
            className="w-12 px-0"
        >
            <span className="font-bold text-xs">
                {i18n.language === "zh" ? "中" : "EN"}
            </span>
        </Button>
    )
}
