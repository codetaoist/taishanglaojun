/**
 * Test data fixtures for E2E tests
 */

export const testUsers = {
  admin: {
    email: 'admin@test.com',
    password: 'Test123456!',
    firstName: 'Admin',
    lastName: 'User',
    role: 'admin',
    region: 'us-east-1',
    locale: 'en-US',
    timezone: 'America/New_York'
  },
  user: {
    email: 'user@test.com',
    password: 'Test123456!',
    firstName: 'Test',
    lastName: 'User',
    role: 'user',
    region: 'us-east-1',
    locale: 'en-US',
    timezone: 'America/New_York'
  },
  euUser: {
    email: 'eu.user@test.com',
    password: 'Test123456!',
    firstName: 'EU',
    lastName: 'User',
    role: 'user',
    region: 'eu-central-1',
    locale: 'de-DE',
    timezone: 'Europe/Berlin'
  },
  asiaUser: {
    email: 'asia.user@test.com',
    password: 'Test123456!',
    firstName: 'Asia',
    lastName: 'User',
    role: 'user',
    region: 'ap-east-1',
    locale: 'zh-CN',
    timezone: 'Asia/Shanghai'
  },
  gdprUser: {
    email: 'gdpr.user@test.com',
    password: 'Test123456!',
    firstName: 'GDPR',
    lastName: 'User',
    role: 'user',
    region: 'eu-central-1',
    locale: 'en-GB',
    timezone: 'Europe/London',
    gdprConsent: true,
    dataProcessingConsent: true
  },
  ccpaUser: {
    email: 'ccpa.user@test.com',
    password: 'Test123456!',
    firstName: 'CCPA',
    lastName: 'User',
    role: 'user',
    region: 'us-west-1',
    locale: 'en-US',
    timezone: 'America/Los_Angeles',
    ccpaOptOut: false
  },
  piplUser: {
    email: 'pipl.user@test.com',
    password: 'Test123456!',
    firstName: 'PIPL',
    lastName: 'User',
    role: 'user',
    region: 'ap-east-1',
    locale: 'zh-CN',
    timezone: 'Asia/Shanghai',
    piplConsent: true
  }
};

export const testRegions = {
  'us-east-1': {
    name: 'US East (N. Virginia)',
    country: 'US',
    currency: 'USD',
    locale: 'en-US',
    timezone: 'America/New_York',
    compliance: ['CCPA'],
    coordinates: { latitude: 39.0458, longitude: -76.6413 }
  },
  'us-west-1': {
    name: 'US West (N. California)',
    country: 'US',
    currency: 'USD',
    locale: 'en-US',
    timezone: 'America/Los_Angeles',
    compliance: ['CCPA'],
    coordinates: { latitude: 37.4419, longitude: -122.1430 }
  },
  'eu-central-1': {
    name: 'Europe (Frankfurt)',
    country: 'DE',
    currency: 'EUR',
    locale: 'de-DE',
    timezone: 'Europe/Berlin',
    compliance: ['GDPR'],
    coordinates: { latitude: 50.1109, longitude: 8.6821 }
  },
  'eu-west-1': {
    name: 'Europe (Ireland)',
    country: 'IE',
    currency: 'EUR',
    locale: 'en-IE',
    timezone: 'Europe/Dublin',
    compliance: ['GDPR'],
    coordinates: { latitude: 53.3498, longitude: -6.2603 }
  },
  'ap-east-1': {
    name: 'Asia Pacific (Hong Kong)',
    country: 'HK',
    currency: 'HKD',
    locale: 'zh-HK',
    timezone: 'Asia/Hong_Kong',
    compliance: ['PIPL'],
    coordinates: { latitude: 22.3193, longitude: 114.1694 }
  },
  'ap-southeast-1': {
    name: 'Asia Pacific (Singapore)',
    country: 'SG',
    currency: 'SGD',
    locale: 'en-SG',
    timezone: 'Asia/Singapore',
    compliance: ['PDPA'],
    coordinates: { latitude: 1.3521, longitude: 103.8198 }
  }
};

export const testLocales = {
  'zh-CN': {
    name: '简体中文',
    nativeName: '简体中文',
    direction: 'ltr',
    currency: 'CNY',
    dateFormat: 'YYYY年MM月DD日',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234.56',
    region: 'ap-east-1'
  },
  'zh-TW': {
    name: '繁體中文',
    nativeName: '繁體中文',
    direction: 'ltr',
    currency: 'TWD',
    dateFormat: 'YYYY年MM月DD日',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234.56',
    region: 'ap-east-1'
  },
  'en-US': {
    name: 'English (US)',
    nativeName: 'English (US)',
    direction: 'ltr',
    currency: 'USD',
    dateFormat: 'MM/DD/YYYY',
    timeFormat: 'h:mm:ss A',
    numberFormat: '1,234.56',
    region: 'us-east-1'
  },
  'en-GB': {
    name: 'English (UK)',
    nativeName: 'English (UK)',
    direction: 'ltr',
    currency: 'GBP',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234.56',
    region: 'eu-west-1'
  },
  'de-DE': {
    name: 'Deutsch',
    nativeName: 'Deutsch',
    direction: 'ltr',
    currency: 'EUR',
    dateFormat: 'DD.MM.YYYY',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1.234,56',
    region: 'eu-central-1'
  },
  'fr-FR': {
    name: 'Français',
    nativeName: 'Français',
    direction: 'ltr',
    currency: 'EUR',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1 234,56',
    region: 'eu-central-1'
  },
  'ja-JP': {
    name: 'Japanese',
    nativeName: '日本語',
    direction: 'ltr',
    currency: 'JPY',
    dateFormat: 'YYYY年MM月DD日',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234',
    region: 'ap-northeast-1'
  },
  'ko-KR': {
    name: 'Korean',
    nativeName: '한국어',
    direction: 'ltr',
    currency: 'KRW',
    dateFormat: 'YYYY년 MM월 DD일',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234',
    region: 'ap-northeast-2'
  },
  'ar-SA': {
    name: 'Arabic',
    nativeName: 'العربية',
    direction: 'rtl',
    currency: 'SAR',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm:ss',
    numberFormat: '1,234.56',
    region: 'me-south-1'
  }
};

export const testContent = {
  'zh-CN': {
    welcome: '欢迎使用太上老君AI平台',
    login: '登录',
    register: '注册',
    dashboard: '仪表板',
    profile: '个人资料',
    settings: '设置',
    logout: '退出登录',
    createContent: '创建内容',
    aiAssistant: 'AI助手',
    dataExport: '数据导出',
    privacyPolicy: '隐私政策',
    termsOfService: '服务条款',
    cookieConsent: '我们使用Cookie来改善您的体验',
    acceptCookies: '接受Cookie',
    declineCookies: '拒绝Cookie',
    dataRetention: '数据保留期：30天',
    contactSupport: '联系支持'
  },
  'en-US': {
    welcome: 'Welcome to Taishang Laojun AI Platform',
    login: 'Login',
    register: 'Register',
    dashboard: 'Dashboard',
    profile: 'Profile',
    settings: 'Settings',
    logout: 'Logout',
    createContent: 'Create Content',
    aiAssistant: 'AI Assistant',
    dataExport: 'Data Export',
    privacyPolicy: 'Privacy Policy',
    termsOfService: 'Terms of Service',
    cookieConsent: 'We use cookies to improve your experience',
    acceptCookies: 'Accept Cookies',
    declineCookies: 'Decline Cookies',
    dataRetention: 'Data retention period: 30 days',
    contactSupport: 'Contact Support'
  },
  'de-DE': {
    welcome: 'Willkommen bei der Taishang Laojun AI-Plattform',
    login: 'Anmelden',
    register: 'Registrieren',
    dashboard: 'Dashboard',
    profile: 'Profil',
    settings: 'Einstellungen',
    logout: 'Abmelden',
    createContent: 'Inhalt erstellen',
    aiAssistant: 'KI-Assistent',
    dataExport: 'Datenexport',
    privacyPolicy: 'Datenschutzrichtlinie',
    termsOfService: 'Nutzungsbedingungen',
    cookieConsent: 'Wir verwenden Cookies, um Ihre Erfahrung zu verbessern',
    acceptCookies: 'Cookies akzeptieren',
    declineCookies: 'Cookies ablehnen',
    dataRetention: 'Datenaufbewahrungsdauer: 30 Tage',
    contactSupport: 'Support kontaktieren'
  }
};

export const testApiEndpoints = {
  auth: {
    login: '/api/auth/login',
    register: '/api/auth/register',
    logout: '/api/auth/logout',
    refresh: '/api/auth/refresh',
    verify: '/api/auth/verify'
  },
  user: {
    profile: '/api/user/profile',
    settings: '/api/user/settings',
    preferences: '/api/user/preferences',
    delete: '/api/user/delete'
  },
  content: {
    create: '/api/content/create',
    list: '/api/content/list',
    get: '/api/content/:id',
    update: '/api/content/:id',
    delete: '/api/content/:id'
  },
  ai: {
    generate: '/api/ai/generate',
    chat: '/api/ai/chat',
    analyze: '/api/ai/analyze'
  },
  compliance: {
    gdpr: {
      consent: '/api/compliance/gdpr/consent',
      export: '/api/compliance/gdpr/export',
      delete: '/api/compliance/gdpr/delete',
      withdraw: '/api/compliance/gdpr/withdraw'
    },
    ccpa: {
      notice: '/api/compliance/ccpa/notice',
      optout: '/api/compliance/ccpa/optout',
      delete: '/api/compliance/ccpa/delete'
    },
    pipl: {
      consent: '/api/compliance/pipl/consent',
      notice: '/api/compliance/pipl/notice'
    }
  },
  localization: {
    translations: '/api/localization/translations',
    currencies: '/api/localization/currencies',
    timezones: '/api/localization/timezones',
    regions: '/api/localization/regions'
  }
};

export const testFiles = {
  images: {
    avatar: 'test-files/avatar.jpg',
    logo: 'test-files/logo.png',
    banner: 'test-files/banner.webp'
  },
  documents: {
    pdf: 'test-files/document.pdf',
    word: 'test-files/document.docx',
    excel: 'test-files/spreadsheet.xlsx'
  },
  malicious: {
    script: 'test-files/malicious.js',
    executable: 'test-files/malicious.exe',
    virus: 'test-files/virus.txt'
  }
};

export const testPayments = {
  validCard: {
    number: '4242424242424242',
    expiry: '12/25',
    cvc: '123',
    name: 'Test User'
  },
  invalidCard: {
    number: '4000000000000002',
    expiry: '12/25',
    cvc: '123',
    name: 'Test User'
  },
  declinedCard: {
    number: '4000000000000341',
    expiry: '12/25',
    cvc: '123',
    name: 'Test User'
  }
};

export const testPerformanceBudgets = {
  fcp: 1800, // First Contentful Paint (ms)
  lcp: 2500, // Largest Contentful Paint (ms)
  fid: 100,  // First Input Delay (ms)
  cls: 0.1,  // Cumulative Layout Shift
  ttfb: 600, // Time to First Byte (ms)
  loadTime: 3000, // Total load time (ms)
  bundleSize: 1024 * 1024, // 1MB
  imageSize: 500 * 1024 // 500KB
};

export const testSecurityHeaders = {
  'X-Frame-Options': 'DENY',
  'X-Content-Type-Options': 'nosniff',
  'X-XSS-Protection': '1; mode=block',
  'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
  'Content-Security-Policy': "default-src 'self'",
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'Permissions-Policy': 'geolocation=(), microphone=(), camera=()'
};

export const testErrorMessages = {
  'zh-CN': {
    invalidEmail: '请输入有效的邮箱地址',
    passwordTooShort: '密码至少需要8个字符',
    networkError: '网络连接错误，请稍后重试',
    unauthorized: '您没有权限访问此资源',
    notFound: '页面未找到',
    serverError: '服务器内部错误'
  },
  'en-US': {
    invalidEmail: 'Please enter a valid email address',
    passwordTooShort: 'Password must be at least 8 characters',
    networkError: 'Network connection error, please try again later',
    unauthorized: 'You do not have permission to access this resource',
    notFound: 'Page not found',
    serverError: 'Internal server error'
  },
  'de-DE': {
    invalidEmail: 'Bitte geben Sie eine gültige E-Mail-Adresse ein',
    passwordTooShort: 'Das Passwort muss mindestens 8 Zeichen lang sein',
    networkError: 'Netzwerkverbindungsfehler, bitte versuchen Sie es später erneut',
    unauthorized: 'Sie haben keine Berechtigung, auf diese Ressource zuzugreifen',
    notFound: 'Seite nicht gefunden',
    serverError: 'Interner Serverfehler'
  }
};

export const testDevices = {
  mobile: {
    'iPhone 12': { width: 390, height: 844 },
    'iPhone 12 Pro': { width: 390, height: 844 },
    'iPhone 13': { width: 390, height: 844 },
    'Samsung Galaxy S21': { width: 384, height: 854 },
    'Google Pixel 5': { width: 393, height: 851 }
  },
  tablet: {
    'iPad': { width: 768, height: 1024 },
    'iPad Pro': { width: 1024, height: 1366 },
    'Samsung Galaxy Tab': { width: 800, height: 1280 }
  },
  desktop: {
    'Small Desktop': { width: 1366, height: 768 },
    'Medium Desktop': { width: 1920, height: 1080 },
    'Large Desktop': { width: 2560, height: 1440 },
    'Ultra Wide': { width: 3440, height: 1440 }
  }
};

export const testBrowsers = [
  'chromium',
  'firefox',
  'webkit',
  'chrome',
  'edge'
];

export const testNetworkConditions = {
  slow3G: {
    downloadThroughput: 500 * 1024 / 8,
    uploadThroughput: 500 * 1024 / 8,
    latency: 400
  },
  fast3G: {
    downloadThroughput: 1.6 * 1024 * 1024 / 8,
    uploadThroughput: 750 * 1024 / 8,
    latency: 150
  },
  offline: {
    downloadThroughput: 0,
    uploadThroughput: 0,
    latency: 0
  }
};