import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Alert, Divider, Progress, Checkbox } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { registerUser, clearError } from '../../store/slices/authSlice';

interface RegisterFormData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  agreement: boolean;
}

// 密码强度检查
const checkPasswordStrength = (password: string) => {
  let score = 0;
  const checks = {
    length: password.length >= 8,
    lowercase: /[a-z]/.test(password),
    uppercase: /[A-Z]/.test(password),
    number: /\d/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
  };
  
  Object.values(checks).forEach(check => {
    if (check) score += 20;
  });
  
  return { score, checks };
};

const RegisterPage: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  
  const { loading, error, isAuthenticated } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  
  const [passwordStrength, setPasswordStrength] = useState({ score: 0, checks: {} });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // 如果已经认证，重定向到仪表板
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // 清除错误信息
  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);

  // 处理密码变化
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value;
    setPasswordStrength(checkPasswordStrength(password));
  };

  // 处理表单提交
  const handleSubmit = async (values: RegisterFormData) => {
    try {
      await dispatch(registerUser({
        username: values.username,
        email: values.email,
        password: values.password,
        confirmPassword: values.confirmPassword,
        agreeToTerms: true,
      })).unwrap();
      
      // 注册成功后会通过useEffect重定向
    } catch (error) {
      // 错误已经在Redux中处理
      console.error('Registration failed:', error);
    }
  };

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 获取密码强度文本和颜色
  const getPasswordStrengthInfo = () => {
    const { score } = passwordStrength;
    if (score < 40) {
      return {
        text: getText('弱', 'Weak'),
        color: '#ff4d4f',
        status: 'exception' as const,
      };
    } else if (score < 80) {
      return {
        text: getText('中等', 'Medium'),
        color: '#faad14',
        status: 'active' as const,
      };
    } else {
      return {
        text: getText('强', 'Strong'),
        color: '#52c41a',
        status: 'success' as const,
      };
    }
  };

  const strengthInfo = getPasswordStrengthInfo();

  return (
    <div className="register-page">
      <div className="register-header">
        <h2 className="register-title">
          {getText('注册账户', 'Create Account')}
        </h2>
        <p className="register-subtitle">
          {getText('加入太上老君序列零，开启智慧之旅', 'Join Taishang Sequence Zero and start your wisdom journey')}
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
        name="register"
        onFinish={handleSubmit}
        autoComplete="off"
        size="large"
        scrollToFirstError
      >
        <Form.Item
          name="username"
          rules={[
            {
              required: true,
              message: getText('请输入用户名', 'Please enter username'),
            },
            {
              min: 3,
              message: getText('用户名至少3位字符', 'Username must be at least 3 characters'),
            },
            {
              max: 20,
              message: getText('用户名最多20位字符', 'Username must be at most 20 characters'),
            },
            {
              pattern: /^[a-zA-Z0-9_]+$/,
              message: getText('用户名只能包含字母、数字和下划线', 'Username can only contain letters, numbers and underscores'),
            },
          ]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder={getText('用户名', 'Username')}
            autoComplete="username"
          />
        </Form.Item>

        <Form.Item
          name="email"
          rules={[
            {
              required: true,
              message: getText('请输入邮箱地址', 'Please enter email address'),
            },
            {
              type: 'email',
              message: getText('请输入有效的邮箱地址', 'Please enter a valid email address'),
            },
          ]}
        >
          <Input
            prefix={<MailOutlined />}
            placeholder={getText('邮箱地址', 'Email address')}
            autoComplete="email"
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[
            {
              required: true,
              message: getText('请输入密码', 'Please enter password'),
            },
            {
              min: 6,
              message: getText('密码至少6位字符', 'Password must be at least 6 characters'),
            },
            {
              validator: (_, value) => {
                if (!value) return Promise.resolve();
                const { score } = checkPasswordStrength(value);
                if (score < 40) {
                  return Promise.reject(new Error(getText('密码强度太弱', 'Password is too weak')));
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder={getText('密码', 'Password')}
            autoComplete="new-password"
            onChange={handlePasswordChange}
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
          />
        </Form.Item>

        {/* 密码强度指示器 */}
        {form.getFieldValue('password') && (
          <div className="password-strength">
            <div className="strength-header">
              <span className="strength-label">
                {getText('密码强度', 'Password Strength')}
              </span>
              <span className="strength-text" style={{ color: strengthInfo.color }}>
                {strengthInfo.text}
              </span>
            </div>
            <Progress
              percent={passwordStrength.score}
              strokeColor={strengthInfo.color}
              status={strengthInfo.status}
              showInfo={false}
              size="small"
            />
            <div className="strength-tips">
              {getText(
                '建议包含大小写字母、数字和特殊字符',
                'Include uppercase, lowercase, numbers and special characters'
              )}
            </div>
          </div>
        )}

        <Form.Item
          name="confirmPassword"
          dependencies={['password']}
          rules={[
            {
              required: true,
              message: getText('请确认密码', 'Please confirm password'),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error(getText('两次输入的密码不一致', 'Passwords do not match')));
              },
            }),
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder={getText('确认密码', 'Confirm password')}
            autoComplete="new-password"
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
          />
        </Form.Item>

        <Form.Item
          name="agreement"
          valuePropName="checked"
          rules={[
            {
              validator: (_, value) =>
                value
                  ? Promise.resolve()
                  : Promise.reject(new Error(getText('请同意用户协议', 'Please agree to the terms'))),
            },
          ]}
        >
          <Checkbox>
            {getText('我已阅读并同意', 'I have read and agree to the')}
            <Link to="/terms" target="_blank" style={{ marginLeft: 4 }}>
              {getText('用户协议', 'Terms of Service')}
            </Link>
            {getText('和', ' and ')}
            <Link to="/privacy" target="_blank">
              {getText('隐私政策', 'Privacy Policy')}
            </Link>
          </Checkbox>
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            className="register-button"
          >
            {loading 
              ? getText('注册中...', 'Creating account...') 
              : getText('注册', 'Sign Up')
            }
          </Button>
        </Form.Item>
      </Form>

      <Divider>
        {getText('或', 'Or')}
      </Divider>

      <div className="login-link">
        <span>{getText('已有账户？', 'Already have an account?')}</span>
        <Link to="/login">
          {getText('立即登录', 'Sign in now')}
        </Link>
      </div>

      <style>{`
        .register-page {
          width: 100%;
          max-width: 400px;
        }

        .register-header {
          text-align: center;
          margin-bottom: 32px;
        }

        .register-title {
          font-size: 24px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
        }

        .register-subtitle {
          font-size: 14px;
          color: var(--text-secondary);
          margin: 0;
          line-height: 1.5;
        }

        .password-strength {
          margin-bottom: 16px;
          padding: 12px;
          background: var(--bg-tertiary);
          border-radius: 6px;
          border: 1px solid var(--border-light);
        }

        .strength-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
        }

        .strength-label {
          font-size: 12px;
          color: var(--text-secondary);
          font-weight: 500;
        }

        .strength-text {
          font-size: 12px;
          font-weight: 600;
        }

        .strength-tips {
          font-size: 11px;
          color: var(--text-tertiary);
          margin-top: 4px;
          line-height: 1.4;
        }

        .register-button {
          height: 44px;
          font-size: 16px;
          font-weight: 500;
        }

        .login-link {
          text-align: center;
          font-size: 14px;
          color: var(--text-secondary);
        }

        .login-link a {
          color: var(--primary-color);
          text-decoration: none;
          margin-left: 8px;
          font-weight: 500;
        }

        .login-link a:hover {
          color: var(--primary-hover);
          text-decoration: underline;
        }

        /* 响应式设计 */
        @media (max-width: 480px) {
          .register-page {
            max-width: 100%;
          }

          .register-title {
            font-size: 20px;
          }

          .register-subtitle {
            font-size: 13px;
          }
        }
      `}</style>
    </div>
  );
};

export default RegisterPage;