import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        debug: true,
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false, // not needed for react as it escapes by default
        },
        resources: {
            en: {
                translation: {
                    common: {
                        loading: "Loading...",
                        theme: {
                            light: "Light",
                            dark: "Dark",
                            system: "System"
                        }
                    },
                    home: {
                        title: "Secure File Sharing",
                        subtitle: "End-to-end encrypted, ephemeral, and simple.",
                        pickup_placeholder: "Enter 8-digit Pickup Code",
                        pickup_button: "Retrieve File",
                        upload_drag: "Drag & drop files here, or click to select",
                        footer: "Designed for privacy. No logs. No tracking."
                    }
                }
            },
            zh: {
                translation: {
                    common: {
                        loading: "加载中...",
                        theme: {
                            light: "浅色",
                            dark: "深色",
                            system: "跟随系统"
                        }
                    },
                    home: {
                        title: "安全文件传输",
                        subtitle: "端到端加密，阅后即焚，极简体验。",
                        pickup_placeholder: "输入 8 位取件码",
                        pickup_button: "立即取件",
                        upload_drag: "拖拽文件到此处，或点击选择",
                        footer: "为隐私而生。无日志。无追踪。"
                    }
                }
            }
        }
    });

export default i18n;
