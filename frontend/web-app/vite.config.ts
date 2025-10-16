import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      // 启用React Fast Refresh
      fastRefresh: true,
      // 优化JSX运行时
      jsxRuntime: 'automatic',
    }),
    VitePWA({
      registerType: 'autoUpdate',
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}']
      }
    })
  ],
  
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@components': resolve(__dirname, 'src/components'),
      '@pages': resolve(__dirname, 'src/pages'),
      '@hooks': resolve(__dirname, 'src/hooks'),
      '@services': resolve(__dirname, 'src/services'),
      '@store': resolve(__dirname, 'src/store'),
      '@utils': resolve(__dirname, 'src/utils'),
      '@types': resolve(__dirname, 'src/types'),
      '@styles': resolve(__dirname, 'src/styles'),
    },
  },
  
  server: {
    port: 5173,
    strictPort: true,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 不需要重写路径，直接代理到后端
      },
      // 代理静态上传文件目录到后端
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  
  build: {
    target: 'es2020',
    outDir: 'dist',
    sourcemap: process.env.NODE_ENV === 'development',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.info', 'console.debug'],
        passes: 2,
      },
      mangle: {
        safari10: true,
      },
      format: {
        comments: false,
      },
    },
    // 启用CSS代码分割
    cssCodeSplit: true,
    // 设置更小的chunk大小警告限制
    chunkSizeWarningLimit: 500,
    // 启用压缩报告
    reportCompressedSize: true,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // React 核心库 - 最高优先级
          if (id.includes('react') || id.includes('react-dom')) {
            return 'react-core';
          }
          
          // React Router - 路由相关
          if (id.includes('react-router')) {
            return 'react-router';
          }
          
          // Ant Design 核心组件
          if (id.includes('antd') && !id.includes('@ant-design/icons') && !id.includes('@ant-design/plots')) {
            return 'antd-core';
          }
          
          // Ant Design 图标
          if (id.includes('@ant-design/icons')) {
            return 'antd-icons';
          }
          
          // 图表库
          if (id.includes('@ant-design/plots') || id.includes('recharts')) {
            return 'charts';
          }
          
          // Redux 状态管理
          if (id.includes('redux') || id.includes('@reduxjs/toolkit')) {
            return 'redux';
          }
          
          // 工具库
          if (id.includes('dayjs') || id.includes('moment') || id.includes('date-fns')) {
            return 'date-utils';
          }
          
          // HTTP 客户端
          if (id.includes('axios')) {
            return 'http-client';
          }
          
          // 国际化
          if (id.includes('i18next')) {
            return 'i18n';
          }
          
          // 样式相关
          if (id.includes('tailwindcss') || id.includes('postcss') || id.includes('autoprefixer')) {
            return 'styles';
          }
          
          // Node modules 中的其他库
          if (id.includes('node_modules')) {
            return 'vendor';
          }
        },
        
        // 文件命名策略
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId
            ? chunkInfo.facadeModuleId.split('/').pop()?.replace(/\.\w+$/, '')
            : 'chunk'
          return `js/[name]-[hash].js`
        },
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name?.split('.') || []
          let extType = info[info.length - 1]
          if (/\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/i.test(assetInfo.name || '')) {
            extType = 'media'
          } else if (/\.(png|jpe?g|gif|svg)(\?.*)?$/i.test(assetInfo.name || '')) {
            extType = 'img'
          } else if (/\.(woff2?|eot|ttf|otf)(\?.*)?$/i.test(assetInfo.name || '')) {
            extType = 'fonts'
          }
          return `${extType}/[name]-[hash].[ext]`
        },
      },
    }
  },
  
  // 优化依赖预构建
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'antd',
      '@ant-design/icons',
      'dayjs',
      'axios',
      '@reduxjs/toolkit',
      'react-redux',
      'redux-persist',
      'i18next',
      'react-i18next',
      'react-error-boundary',
    ],
    exclude: [
      '@monaco-editor/react',
      'vite-plugin-pwa',
      'workbox-window',
    ],
    // 强制预构建某些依赖
    force: true,
  },
  
  // CSS 配置
  css: {
    modules: {
      localsConvention: 'camelCase',
    },
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
        modifyVars: {
          '@primary-color': '#1890ff',
          '@border-radius-base': '6px',
        },
      },
    },
  },
})
