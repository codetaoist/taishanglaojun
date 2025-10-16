import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 导入语言资源
import zhCN from './locales/zh-CN.json';
import enUS from './locales/en-US.json';
import jaJP from './locales/ja-JP.json';

// 支持的语言列表
export const supportedLanguages = [
  { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', flag: '🇺🇸' },
  { code: 'ja-JP', name: '日本語', flag: '🇯🇵' }
];

// 语言资源
const resources = {
  'zh-CN': {
    translation: zhCN
  },
  'en-US': {
    translation: enUS
  },
  'ja-JP': {
    translation: jaJP
  }
};

// 初始化i18n
i18n
  .use(LanguageDetector) // 自动检测用户语言
  .use(initReactI18next) // 绑定React
  .init({
    resources,
    fallbackLng: 'zh-CN', // 默认语言
    debug: process.env.NODE_ENV === 'development',

    // 语言检测配置
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
      lookupLocalStorage: 'i18nextLng'
    },

    interpolation: {
      escapeValue: false // React已经处理了XSS
    },

    // 命名空间配置
    defaultNS: 'translation',
    ns: ['translation'],

    // 键值分隔符
    keySeparator: '.',
    nsSeparator: ':',

    // 复数规则
    pluralSeparator: '_',
    contextSeparator: '_',

    // 加载配置：使用 currentOnly 以匹配完整地区码（如 ja-JP）
    load: 'currentOnly',
    preload: ['zh-CN', 'en-US', 'ja-JP'],

    // 缓存配置
    saveMissing: process.env.NODE_ENV === 'development',
    missingKeyHandler: (lng, ns, key) => {
      if (process.env.NODE_ENV === 'development') {
        console.warn(`Missing translation key: ${key} for language: ${lng}`);
      }
    }
  });

// 语言切换函数
export const changeLanguage = (language: string) => {
  return i18n.changeLanguage(language);
};

// 获取当前语言
export const getCurrentLanguage = () => {
  return i18n.language || 'zh-CN';
};

// 获取语言信息
export const getLanguageInfo = (code: string) => {
  return supportedLanguages.find(lang => lang.code === code);
};

// 格式化数字
export const formatNumber = (number: number, locale?: string) => {
  const currentLocale = locale || getCurrentLanguage();
  return new Intl.NumberFormat(currentLocale).format(number);
};

// 格式化货币
export const formatCurrency = (amount: number, currency = 'CNY', locale?: string) => {
  const currentLocale = locale || getCurrentLanguage();
  return new Intl.NumberFormat(currentLocale, {
    style: 'currency',
    currency
  }).format(amount);
};

// 格式化日期
export const formatDate = (date: Date | string, options?: Intl.DateTimeFormatOptions, locale?: string) => {
  const currentLocale = locale || getCurrentLanguage();
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  
  const defaultOptions: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  };

  return new Intl.DateTimeFormat(currentLocale, { ...defaultOptions, ...options }).format(dateObj);
};

// 格式化相对时间
export const formatRelativeTime = (date: Date | string, locale?: string) => {
  const currentLocale = locale || getCurrentLanguage();
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - dateObj.getTime()) / 1000);

  const rtf = new Intl.RelativeTimeFormat(currentLocale, { numeric: 'auto' });

  if (diffInSeconds < 60) {
    return rtf.format(-diffInSeconds, 'second');
  } else if (diffInSeconds < 3600) {
    return rtf.format(-Math.floor(diffInSeconds / 60), 'minute');
  } else if (diffInSeconds < 86400) {
    return rtf.format(-Math.floor(diffInSeconds / 3600), 'hour');
  } else if (diffInSeconds < 2592000) {
    return rtf.format(-Math.floor(diffInSeconds / 86400), 'day');
  } else if (diffInSeconds < 31536000) {
    return rtf.format(-Math.floor(diffInSeconds / 2592000), 'month');
  } else {
    return rtf.format(-Math.floor(diffInSeconds / 31536000), 'year');
  }
};

// 文本方向检测
export const getTextDirection = (locale?: string) => {
  const currentLocale = locale || getCurrentLanguage();
  // RTL语言列表
  const rtlLanguages = ['ar', 'he', 'fa', 'ur'];
  const languageCode = currentLocale.split('-')[0];
  return rtlLanguages.includes(languageCode) ? 'rtl' : 'ltr';
};

// 语言变化监听器
export const onLanguageChange = (callback: (language: string) => void) => {
  i18n.on('languageChanged', callback);
  return () => i18n.off('languageChanged', callback);
};

export default i18n;