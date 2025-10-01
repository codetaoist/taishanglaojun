import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Checkbox, Alert, Divider, Space } from 'antd';
import { UserOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { loginUser, clearError } from '../../store/slices/authSlice';

interface LoginFormData {
  email: string;
  password: string;
  remember: boolean;
}

const LoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useAppDispatch();
  
  const { loading, error, isAuthenticated } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  
  const [showPassword, setShowPassword] = useState(false);

  // 从路由状态获取重定向路径
  const from = (location.state as any)?.from?.pathname || '/dashboard';

  // 如果已经认证，重定向到目标页面
  useEffect(() => {
    if (isAuthenticated) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, from]);

  // 清除错误信息
  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);



  // 处理表单提交
  const handleSubmit = async (values: LoginFormData) => {
    try {
      await dispatch(loginUser(values)).unwrap();
      
      // 登录成功后会通过useEffect重定向
    } catch (error) {
      // 错误已经在Redux中处理
      console.error('Login failed:', error);
    }
  };

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };



  return (
    <div className="login-page">
      <div className="login-header">
        <h2 className="login-title">
          {getText('登录', 'Sign In')}
        </h2>
        <p className="login-subtitle">
          {getText('欢迎回到太上老君序列零', 'Welcome back to Taishang Sequence Zero')}
        </p>
      </div>

      {/* 错误提示 */}
      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => dispatch(clearError())}
          style={{ marginBottom: 16 }}
        />
      )}



      <Form
        form={form}
        name="login"
        onFinish={handleSubmit}
        autoComplete="off"
        size="large"

      >
        <Form.Item
          name="email"
          rules={[
            {
              required: true,
              message: getText('请输入邮箱地址', 'Please enter your email'),
            },
            {
              type: 'email',
              message: getText('请输入有效的邮箱地址', 'Please enter a valid email'),
            },
          ]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder={getText('邮箱地址', 'Email address')}
            autoComplete="email"
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[
            {
              required: true,
              message: getText('请输入密码', 'Please enter your password'),
            },
            {
              min: 6,
              message: getText('密码至少6位字符', 'Password must be at least 6 characters'),
            },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder={getText('密码', 'Password')}
            autoComplete="current-password"
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
          />
        </Form.Item>

        <Form.Item>
          <div className="login-options">
            <Form.Item name="remember" valuePropName="checked" noStyle>
              <Checkbox>
                {getText('记住我', 'Remember me')}
              </Checkbox>
            </Form.Item>
            <Link to="/forgot-password" className="forgot-link">
              {getText('忘记密码？', 'Forgot password?')}
            </Link>
          </div>
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}

            block
            className="login-button"
          >
            {loading 
              ? getText('登录中...', 'Signing in...') 
              : getText('登录', 'Sign In')
            }
          </Button>
        </Form.Item>
      </Form>

      <Divider>
        {getText('或', 'Or')}
      </Divider>

      <div className="register-link">
        <span>{getText('还没有账户？', "Don't have an account?")}</span>
        <Link to="/register">
          {getText('立即注册', 'Sign up now')}
        </Link>
      </div>

      {/* 演示账户信息 */}
      <div className="demo-info">
        <Divider orientation="left" orientationMargin="0">
          <span className="demo-title">
            {getText('演示账户', 'Demo Accounts')}
          </span>
        </Divider>
        <Space direction="vertical" size="small" style={{ width: '100%' }}>
          <div className="demo-account">
            <strong>{getText('管理员', 'Admin')}:</strong> admin@example.com / admin123
          </div>
          <div className="demo-account">
            <strong>{getText('用户', 'User')}:</strong> user@example.com / user123
          </div>
        </Space>
      </div>

      <style>{`
        .login-page {
          width: 100%;
          max-width: 400px;
        }

        .login-header {
          text-align: center;
          margin-bottom: 32px;
        }

        .login-title {
          font-size: 24px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
        }

        .login-subtitle {
          font-size: 14px;
          color: var(--text-secondary);
          margin: 0;
        }

        .login-options {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .forgot-link {
          color: var(--primary-color);
          text-decoration: none;
          font-size: 14px;
        }

        .forgot-link:hover {
          color: var(--primary-hover);
          text-decoration: underline;
        }

        .login-button {
          height: 44px;
          font-size: 16px;
          font-weight: 500;
        }

        .register-link {
          text-align: center;
          font-size: 14px;
          color: var(--text-secondary);
        }

        .register-link a {
          color: var(--primary-color);
          text-decoration: none;
          margin-left: 8px;
          font-weight: 500;
        }

        .register-link a:hover {
          color: var(--primary-hover);
          text-decoration: underline;
        }

        .demo-info {
          margin-top: 24px;
          padding: 16px;
          background: var(--bg-tertiary);
          border-radius: 8px;
          border: 1px solid var(--border-light);
        }

        .demo-title {
          font-size: 12px;
          color: var(--text-tertiary);
          font-weight: 500;
        }

        .demo-account {
          font-size: 12px;
          color: var(--text-secondary);
          font-family: 'Courier New', monospace;
          background: var(--bg-primary);
          padding: 8px;
          border-radius: 4px;
          border: 1px solid var(--border-light);
        }

        .demo-account strong {
          color: var(--text-primary);
        }

        /* 响应式设计 */
        @media (max-width: 480px) {
          .login-page {
            max-width: 100%;
          }

          .login-title {
            font-size: 20px;
          }

          .login-options {
            flex-direction: column;
            align-items: flex-start;
            gap: 8px;
          }
        }
      `}</style>
    </div>
  );
};

export default LoginPage;