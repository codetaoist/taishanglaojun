import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Typography, Space, Divider, Alert, Tabs, Checkbox, Card, Row, Col, Dropdown } from 'antd';
import type { MenuProps } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, AlipayCircleOutlined, TaobaoCircleOutlined, WeiboCircleOutlined, EyeTwoTone, EyeInvisibleOutlined, SafetyCertificateOutlined, ThunderboltOutlined, GlobalOutlined, TranslationOutlined, DownOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthContext } from '../contexts/AuthContext';
import { getNotificationInstance } from '../services/notificationService';
import './Login.css';

interface LoginForm {
  username: string;
  password: string;
  remember?: boolean;
}

interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  agreement?: boolean;
}

// 密码强度检查函数
const checkPasswordStrength = (password: string) => {
  const strength = {
    score: 0,
    level: 'weak' as 'weak' | 'medium' | 'strong',
    feedback: [] as string[]
  };

  if (password.length >= 8) strength.score += 1;
  if (/[a-z]/.test(password)) strength.score += 1;
  if (/[A-Z]/.test(password)) strength.score += 1;
  if (/[0-9]/.test(password)) strength.score += 1;
  if (/[^A-Za-z0-9]/.test(password)) strength.score += 1;

  if (strength.score <= 2) {
    strength.level = 'weak';
    strength.feedback.push('密码强度较弱');
  } else if (strength.score <= 3) {
    strength.level = 'medium';
    strength.feedback.push('密码强度中等');
  } else {
    strength.level = 'strong';
    strength.feedback.push('密码强度较强');
  }

  return strength;
};

const Login: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'login' | 'register'>('login');
  const [loginForm] = Form.useForm<LoginForm>();
  const [registerForm] = Form.useForm<RegisterForm>();
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState<any>(null);
  const [currentLanguage, setCurrentLanguage] = useState<'zh-CN' | 'en-US'>('zh-CN');
  
  const { login, register, isLoading } = useAuthContext();
  const navigate = useNavigate();
  const location = useLocation();

  // 清除消息
  const clearMessages = () => {
    setErrorMessage('');
    setSuccessMessage('');
  };

  // 切换标签页
  const handleTabChange = (key: string) => {
    setActiveTab(key as 'login' | 'register');
    clearMessages();
  };

  // 语言切换处理
  const handleLanguageChange = (language: 'zh-CN' | 'en-US') => {
    setCurrentLanguage(language);
    localStorage.setItem('language', language);
    const notification = getNotificationInstance();
    notification.success({
      message: language === 'zh-CN' ? '已切换到简体中文' : 'Switched to English',
      duration: 2
    });
  };

  // 语言菜单配置
  const languageMenuItems: MenuProps['items'] = [
    {
      key: 'zh-CN',
      label: (
        <div className="lang-menu-item" onClick={() => handleLanguageChange('zh-CN')}>
          <span className="lang-flag">🇨🇳</span>
          <span className="lang-text">简体中文</span>
        </div>
      ),
    },
    {
      key: 'en-US',
      label: (
        <div className="lang-menu-item" onClick={() => handleLanguageChange('en-US')}>
          <span className="lang-flag">🇺🇸</span>
          <span className="lang-text">English</span>
        </div>
      ),
    },
  ];

  // 处理登录
  const handleLogin = async (values: LoginForm) => {
    setIsSubmitting(true);
    clearMessages();
    
    try {
      const result = await login(values.username, values.password);
      
      if (result.success) {
        // 保存记住登录状态
        if (values.remember) {
          localStorage.setItem('rememberLogin', 'true');
        } else {
          localStorage.removeItem('rememberLogin');
        }
        
        // 获取重定向路径，管理员自动跳转到管理界面
        let redirectPath = (location.state as any)?.from?.pathname || '/';
        
        // 检查用户是否是管理员，如果是则跳转到管理界面
        if (result.user) {
          const user = result.user;
          const isAdmin = user.role === 'admin' || user.role === 'super_admin' || 
                         user.roles?.some((role: string) => role.toLowerCase() === 'admin' || role.toLowerCase() === 'super_admin') ||
                         user.isAdmin === true;
          
          if (isAdmin && redirectPath === '/') {
            redirectPath = '/admin';
          }
        }
        
        navigate(redirectPath, { replace: true });
      } else {
        setErrorMessage(result.error || '登录失败，请重试');
      }
    } catch (error: any) {
      setErrorMessage('网络错误，请稍后重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  // 处理注册
  const handleRegister = async (values: RegisterForm) => {
    setIsSubmitting(true);
    clearMessages();
    
    try {
      await register({
        username: values.username,
        email: values.email,
        password: values.password
      });
      
      setSuccessMessage('注册成功！请查收验证邮件并点击验证链接激活账户。');
      
      // 3秒后切换到登录页面
      setTimeout(() => {
        setActiveTab('login');
        setSuccessMessage('');
      }, 3000);
      
    } catch (error: any) {
      setErrorMessage(error.message || '注册失败，请重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  // 监听注册表单密码变化
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value;
    if (password) {
      const strength = checkPasswordStrength(password);
      setPasswordStrength(strength);
    } else {
      setPasswordStrength(null);
    }
  };

  // 检查是否记住登录状态和语言设置
  useEffect(() => {
    const rememberLogin = localStorage.getItem('rememberLogin');
    if (rememberLogin) {
      loginForm.setFieldsValue({ remember: true });
    }
    
    // 初始化语言设置
    const savedLanguage = localStorage.getItem('language') as 'zh-CN' | 'en-US';
    if (savedLanguage && (savedLanguage === 'zh-CN' || savedLanguage === 'en-US')) {
      setCurrentLanguage(savedLanguage);
    }
  }, [loginForm]);

  return (
    <div className="login-container">
      <div className="login-lang">
        <Dropdown 
          menu={{ items: languageMenuItems }} 
          placement="bottomRight"
          trigger={['click']}
          overlayClassName="lang-dropdown"
        >
          <div className="lang-selector">
            <TranslationOutlined className="lang-icon" />
            <span className="lang-current">
              {currentLanguage === 'zh-CN' ? '🇨🇳 简体中文' : '🇺🇸 English'}
            </span>
            <DownOutlined className="lang-arrow" />
          </div>
        </Dropdown>
      </div>
      
      <div className="login-content">
        <div className="login-top">
          <div className="login-header">
            <img alt="太上老君" className="login-logo" src="/laojun-avatar.svg" />
            <span className="login-title">太上老君智慧平台</span>
          </div>
          <div className="login-desc">
            融合传统文化智慧与现代AI技术，为您提供全方位的智能服务体验
          </div>
          
          {/* 特色功能展示 */}
          <Row gutter={[24, 16]} className="login-features">
            <Col span={8}>
              <div className="feature-item">
                <SafetyCertificateOutlined className="feature-icon" />
                <div className="feature-text">安全可靠</div>
              </div>
            </Col>
            <Col span={8}>
              <div className="feature-item">
                <ThunderboltOutlined className="feature-icon" />
                <div className="feature-text">智能高效</div>
              </div>
            </Col>
            <Col span={8}>
              <div className="feature-item">
                <GlobalOutlined className="feature-icon" />
                <div className="feature-text">全球服务</div>
              </div>
            </Col>
          </Row>
        </div>

        <Card className="login-main-card" variant="borderless">
          {errorMessage && (
            <Alert
              message={errorMessage}
              type="error"
              showIcon
              closable
              onClose={clearMessages}
              style={{ marginBottom: 24 }}
            />
          )}

          {successMessage && (
            <Alert
              message={successMessage}
              type="success"
              showIcon
              style={{ marginBottom: 24 }}
            />
          )}

          <Tabs 
            activeKey={activeTab} 
            onChange={handleTabChange}
            centered
            size="large"
            className="login-tabs"
            items={[
              {
                key: 'login',
                label: (
                  <span className="tab-label">
                    <UserOutlined />
                    账户登录
                  </span>
                ),
                children: (
                  <Form
                    form={loginForm}
                    name="login"
                    onFinish={handleLogin}
                    size="large"
                    className="login-form"
                  >
                    <Form.Item
                      name="username"
                      rules={[
                        { required: true, message: '请输入用户名或邮箱！' }
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined className="input-prefix-icon" />}
                        placeholder="用户名/邮箱"
                        className="login-input"
                      />
                    </Form.Item>

                    <Form.Item
                      name="password"
                      rules={[{ required: true, message: '请输入密码！' }]}
                    >
                      <Input.Password
                        prefix={<LockOutlined className="input-prefix-icon" />}
                        placeholder="密码"
                        className="login-input"
                        iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                      />
                    </Form.Item>

                    <Form.Item>
                      <div className="login-options">
                        <Form.Item name="remember" valuePropName="checked" noStyle>
                          <Checkbox className="auto-login">自动登录</Checkbox>
                        </Form.Item>
                        <Typography.Link className="forgot-password">忘记密码</Typography.Link>
                      </div>
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={isSubmitting || isLoading}
                        block
                        size="large"
                        className="login-button"
                      >
                        登录
                      </Button>
                    </Form.Item>


                  </Form>
                )
              },
              {
                key: 'register',
                label: (
                  <span className="tab-label">
                    <MailOutlined />
                    注册账户
                  </span>
                ),
                children: (
                  <Form
                    form={registerForm}
                    name="register"
                    onFinish={handleRegister}
                    autoComplete="off"
                    size="large"
                    className="register-form"
                  >
                    <Form.Item
                      name="username"
                      rules={[
                        { required: true, message: '请输入用户名!' },
                        { min: 3, message: '用户名至少3个字符' },
                        { max: 20, message: '用户名最多20个字符' }
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined className="input-prefix-icon" />}
                        placeholder="用户名"
                        autoComplete="username"
                        className="login-input"
                      />
                    </Form.Item>

                    <Form.Item
                      name="email"
                      rules={[
                        { required: true, message: '请输入邮箱!' },
                        { type: 'email', message: '请输入有效的邮箱地址!' }
                      ]}
                    >
                      <Input
                        prefix={<MailOutlined className="input-prefix-icon" />}
                        placeholder="邮箱"
                        autoComplete="email"
                        className="login-input"
                      />
                    </Form.Item>

                    <Form.Item
                      name="password"
                      rules={[
                        { required: true, message: '请输入密码!' },
                        { min: 8, message: '密码至少8个字符' }
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined className="input-prefix-icon" />}
                        placeholder="密码"
                        autoComplete="new-password"
                        className="login-input"
                        iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                        onChange={handlePasswordChange}
                      />
                    </Form.Item>

                    {passwordStrength && (
                      <div className="password-strength">
                        <div className="password-strength-label">密码强度:</div>
                        <div className={`password-strength-bar ${passwordStrength.level}`}>
                          <div className="password-strength-fill"></div>
                        </div>
                        <div className="password-strength-text">{passwordStrength.feedback[0]}</div>
                      </div>
                    )}

                    <Form.Item
                      name="confirmPassword"
                      dependencies={['password']}
                      rules={[
                        { required: true, message: '请确认密码!' },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('password') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error('两次输入的密码不一致!'));
                          },
                        }),
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined className="input-prefix-icon" />}
                        placeholder="确认密码"
                        autoComplete="new-password"
                        className="login-input"
                        iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                      />
                    </Form.Item>

                    <Form.Item
                      name="agreement"
                      valuePropName="checked"
                      rules={[
                        {
                          validator: (_, value) =>
                            value ? Promise.resolve() : Promise.reject(new Error('请同意用户协议和隐私政策')),
                        },
                      ]}
                    >
                      <Checkbox className="agreement-checkbox">
                        我已阅读并同意 <a href="#" onClick={(e) => e.preventDefault()} className="agreement-link">用户协议</a> 和 <a href="#" onClick={(e) => e.preventDefault()} className="agreement-link">隐私政策</a>
                      </Checkbox>
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={isSubmitting || isLoading}
                        block
                        size="large"
                        className="register-button"
                      >
                        注册
                      </Button>
                    </Form.Item>
                  </Form>
                )
              }
            ]}
          />

          <Divider plain className="login-divider">其他登录方式</Divider>

          <div className="login-other">
            <div className="other-login-item">
              <AlipayCircleOutlined className="login-other-icon alipay" />
              <span className="other-login-text">支付宝</span>
            </div>
            <div className="other-login-item">
              <TaobaoCircleOutlined className="login-other-icon taobao" />
              <span className="other-login-text">淘宝</span>
            </div>
            <div className="other-login-item">
              <WeiboCircleOutlined className="login-other-icon weibo" />
              <span className="other-login-text">微博</span>
            </div>
          </div>
        </Card>
      </div>

      <div className="login-footer">
        <div className="login-links">
          <Typography.Link>帮助</Typography.Link>
          <Typography.Link>隐私</Typography.Link>
          <Typography.Link>条款</Typography.Link>
        </div>
        <div className="login-copyright">
          Copyright © 2025 太上老君智慧平台 All Rights Reserved.
        </div>
      </div>
    </div>
  );
};

export default Login;