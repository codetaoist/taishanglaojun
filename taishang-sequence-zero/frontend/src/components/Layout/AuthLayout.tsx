import React from 'react';
import { Layout, Card } from 'antd';
import { useAppSelector } from '../../hooks/redux';

const { Content } = Layout;

interface AuthLayoutProps {
  children: React.ReactNode;
}

const AuthLayout: React.FC<AuthLayoutProps> = ({ children }) => {
  const { theme, language } = useAppSelector(state => state.ui);

  return (
    <Layout className="auth-layout">
      <Content className="auth-content">
        {/* 背景装饰 */}
        <div className="auth-background">
          <div className="bg-decoration bg-decoration-1"></div>
          <div className="bg-decoration bg-decoration-2"></div>
          <div className="bg-decoration bg-decoration-3"></div>
        </div>

        {/* 主要内容区域 */}
        <div className="auth-container">
          {/* Logo和标题 */}
          <div className="auth-header">
            <div className="auth-logo">
              <div className="logo-icon">🧙‍♂️</div>
              <div className="logo-text">
                <h1 className="logo-title">
                  {language === 'en-US' ? 'Taishang Laojun' : '太上老君'}
                </h1>
                <p className="logo-subtitle">
                  {language === 'en-US' ? 'Sequence Zero' : '序列零'}
                </p>
              </div>
            </div>
            <p className="auth-description">
              {language === 'en-US'
                ? 'Explore the fusion of consciousness and cultural wisdom'
                : '探索意识融合与文化智慧的奥秘'
              }
            </p>
          </div>

          {/* 认证表单卡片 */}
          <Card className="auth-card" variant="borderless">
            {children}
          </Card>

          {/* 页脚 */}
          <div className="auth-footer">
            <p className="footer-text">
              {language === 'en-US'
                ? '© 2024 Taishang Laojun Sequence Zero. All rights reserved.'
                : '© 2024 太上老君序列零. 保留所有权利.'
              }
            </p>
          </div>
        </div>
      </Content>

      <style>{`
        .auth-layout {
          min-height: 100vh;
          background: var(--gradient-primary);
          position: relative;
          overflow: hidden;
        }

        .auth-content {
          display: flex;
          align-items: center;
          justify-content: center;
          min-height: 100vh;
          padding: 24px;
          position: relative;
          z-index: 1;
        }

        .auth-background {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          overflow: hidden;
          z-index: 0;
        }

        .bg-decoration {
          position: absolute;
          border-radius: 50%;
          background: rgba(255, 255, 255, 0.1);
          animation: float 6s ease-in-out infinite;
        }

        .bg-decoration-1 {
          width: 200px;
          height: 200px;
          top: 10%;
          left: 10%;
          animation-delay: 0s;
        }

        .bg-decoration-2 {
          width: 150px;
          height: 150px;
          top: 60%;
          right: 15%;
          animation-delay: 2s;
        }

        .bg-decoration-3 {
          width: 100px;
          height: 100px;
          bottom: 20%;
          left: 20%;
          animation-delay: 4s;
        }

        @keyframes float {
          0%, 100% {
            transform: translateY(0px) rotate(0deg);
          }
          50% {
            transform: translateY(-20px) rotate(180deg);
          }
        }

        .auth-container {
          width: 100%;
          max-width: 400px;
          text-align: center;
        }

        .auth-header {
          margin-bottom: 32px;
        }

        .auth-logo {
          display: flex;
          align-items: center;
          justify-content: center;
          margin-bottom: 16px;
          gap: 16px;
        }

        .logo-icon {
          font-size: 48px;
          animation: pulse 2s ease-in-out infinite;
        }

        @keyframes pulse {
          0%, 100% {
            transform: scale(1);
          }
          50% {
            transform: scale(1.1);
          }
        }

        .logo-text {
          text-align: left;
        }

        .logo-title {
          font-size: 28px;
          font-weight: 700;
          color: #ffffff;
          margin: 0;
          line-height: 1.2;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        }

        .logo-subtitle {
          font-size: 14px;
          color: rgba(255, 255, 255, 0.8);
          margin: 0;
          font-weight: 300;
          letter-spacing: 1px;
        }

        .auth-description {
          font-size: 16px;
          color: rgba(255, 255, 255, 0.9);
          margin: 0;
          line-height: 1.5;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
        }

        .auth-card {
          background: rgba(255, 255, 255, 0.95);
          backdrop-filter: blur(10px);
          border-radius: 16px;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
          padding: 32px;
          margin-bottom: 24px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .auth-footer {
          text-align: center;
        }

        .footer-text {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.7);
          margin: 0;
        }

        /* 暗色主题适配 */
        [data-theme='dark'] .auth-card {
          background: rgba(20, 20, 20, 0.95);
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .auth-content {
            padding: 16px;
          }

          .auth-container {
            max-width: 100%;
          }

          .auth-card {
            padding: 24px;
            border-radius: 12px;
          }

          .logo-icon {
            font-size: 36px;
          }

          .logo-title {
            font-size: 24px;
          }

          .auth-description {
            font-size: 14px;
          }

          .bg-decoration {
            display: none;
          }
        }

        @media (max-width: 480px) {
          .auth-logo {
            flex-direction: column;
            gap: 8px;
          }

          .logo-text {
            text-align: center;
          }

          .auth-card {
            padding: 20px;
          }
        }

        /* 高度较小的屏幕适配 */
        @media (max-height: 600px) {
          .auth-content {
            align-items: flex-start;
            padding-top: 40px;
          }

          .auth-header {
            margin-bottom: 20px;
          }

          .logo-icon {
            font-size: 32px;
          }

          .logo-title {
            font-size: 20px;
          }

          .auth-description {
            font-size: 13px;
          }

          .auth-card {
            padding: 20px;
          }
        }

        /* 打印样式 */
        @media print {
          .auth-layout {
            background: #ffffff !important;
          }

          .auth-card {
            background: #ffffff !important;
            box-shadow: none !important;
            border: 1px solid #cccccc !important;
          }

          .bg-decoration {
            display: none !important;
          }
        }

        /* 减少动画模式 */
        @media (prefers-reduced-motion: reduce) {
          .bg-decoration,
          .logo-icon {
            animation: none !important;
          }
        }
      `}</style>
    </Layout>
  );
};

export default AuthLayout;