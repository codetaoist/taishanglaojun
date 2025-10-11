import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, Space, Divider, Alert, Row, Col, Tabs, Checkbox } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, EyeInvisibleOutlined, EyeTwoTone, SafetyCertificateOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from "../hooks/useAuth";
import PasswordStrengthIndicator from '../components/PasswordStrengthIndicator';
import { checkPasswordStrength, validatePassword } from '../utils/passwordValidator';
import type { PasswordStrength } from '../utils/passwordValidator';

const { Title, Text, Paragraph } = Typography;

interface LoginForm {
  email: string;
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

const Login: React.FC = () => {
  const [activeTab, setActiveTab] = useState('login');
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState<PasswordStrength | null>(null);
  const navigate = useNavigate();
  const [loginForm] = Form.useForm();
  const [registerForm] = Form.useForm();
  const { login, register, isLoading } = useAuth();

  // 清除错误和成功消息
  const clearMessages = () => {
    setErrorMessage('');
    setSuccessMessage('');
  };

  // 切换标签页时清除消息
  const handleTabChange = (key: string) => {
    setActiveTab(key);
    clearMessages();
  };

  const handleLogin = async (values: LoginForm) => {
    try {
      setIsSubmitting(true);
      clearMessages();
      
      const result = await login(values.email, values.password);
      
      if (result.success) {
        setSuccessMessage('登录成功！正在跳转...');
        // 记住登录状态
        if (values.remember) {
          localStorage.setItem('rememberLogin', 'true');
        }
        navigate('/', { replace: true });
      } else {
        setErrorMessage(result.error || '登录失败，请检查您的邮箱和密码');
      }
    } catch (error: any) {
      setErrorMessage('网络错误，请稍后重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRegister = async (values: RegisterForm) => {
    try {
      setIsSubmitting(true);
      clearMessages();

      if (values.password !== values.confirmPassword) {
        setErrorMessage('两次输入的密码不一致');
        setIsSubmitting(false);
        return;
      }

      if (!values.agreement) {
        setErrorMessage('请同意用户协议和隐私政策');
        setIsSubmitting(false);
        return;
      }

      // 验证密码强度
      const passwordValidation = validatePassword(values.password);
      if (!passwordValidation.valid) {
        setErrorMessage(`密码不符合要求：${passwordValidation.errors.join('、')}`);
        setIsSubmitting(false);
        return;
      }

      const result = await register({
        username: values.username,
        email: values.email,
        password: values.password
      });
      
      if (result.success) {
        if (result.data?.message && result.data.message.includes('验证邮件')) {
          setSuccessMessage('注册成功！请查收验证邮件并点击验证链接激活账户。');
        } else {
          setSuccessMessage('注册成功！正在跳转...');
          setTimeout(() => {
            navigate('/', { replace: true });
          }, 1000);
        }
      } else {
        setErrorMessage(result.error || '注册失败，请重试');
      }
    } catch (error: any) {
      setErrorMessage('网络错误，请稍后重试');
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

  // 检查是否记住登录状态
  useEffect(() => {
    const rememberLogin = localStorage.getItem('rememberLogin');
    if (rememberLogin) {
      loginForm.setFieldsValue({ remember: true });
    }
  }, [loginForm]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 flex items-center justify-center p-4">
      <Row className="w-full max-w-6xl">
        <Col xs={24} lg={12} className="flex items-center justify-center">
          <div className="text-center p-8">
            <div className="mb-8">
              <SafetyCertificateOutlined className="text-6xl text-blue-600 mb-4" />
              <Title level={1} className="text-gray-800 mb-4">
                太上老君智慧平台
              </Title>
              <Paragraph className="text-lg text-gray-600 max-w-md mx-auto">
                融合传统文化智慧与现代AI技术，为您提供全方位的智能服务体验
              </Paragraph>
            </div>
            
            <div className="grid grid-cols-2 gap-4 max-w-md mx-auto">
              <div className="bg-white/60 backdrop-blur-sm rounded-lg p-4 border border-white/20">
                <div className="text-blue-600 text-2xl mb-2">🧠</div>
                <Text strong>AI智能助手</Text>
                <div className="text-sm text-gray-600 mt-1">智能对话与分析</div>
              </div>
              <div className="bg-white/60 backdrop-blur-sm rounded-lg p-4 border border-white/20">
                <div className="text-green-600 text-2xl mb-2">🛡️</div>
                <Text strong>安全防护</Text>
                <div className="text-sm text-gray-600 mt-1">全方位安全监控</div>
              </div>
              <div className="bg-white/60 backdrop-blur-sm rounded-lg p-4 border border-white/20">
                <div className="text-purple-600 text-2xl mb-2">📚</div>
                <Text strong>文化智慧</Text>
                <div className="text-sm text-gray-600 mt-1">传统文化传承</div>
              </div>
              <div className="bg-white/60 backdrop-blur-sm rounded-lg p-4 border border-white/20">
                <div className="text-orange-600 text-2xl mb-2">💡</div>
                <Text strong>智能学习</Text>
                <div className="text-sm text-gray-600 mt-1">个性化学习体验</div>
              </div>
            </div>
          </div>
        </Col>
        
        <Col xs={24} lg={12} className="flex items-center justify-center">
          <Card 
            className="w-full max-w-md shadow-2xl border-0 bg-white/90 backdrop-blur-sm"
            style={{ borderRadius: '16px' }}
          >
            <div className="text-center mb-6">
              <Title level={3} className="text-gray-800 mb-2">
                {activeTab === 'login' ? '欢迎回来' : '创建账户'}
              </Title>
              <Text type="secondary">
                {activeTab === 'login' ? '登录您的账户以继续' : '注册新账户开始使用'}
              </Text>
            </div>

            {errorMessage && (
              <Alert
                message={errorMessage}
                type="error"
                showIcon
                closable
                onClose={clearMessages}
                className="mb-4"
              />
            )}

            {successMessage && (
              <Alert
                message={successMessage}
                type="success"
                showIcon
                className="mb-4"
              />
            )}

            <Tabs 
              activeKey={activeTab} 
              onChange={handleTabChange}
              centered
              className="mb-4"
              items={[
                {
                  key: 'login',
                  label: '登录',
                  children: (
                    <Form
                      form={loginForm}
                      name="login"
                      onFinish={handleLogin}
                      layout="vertical"
                      size="large"
                    >
                      <Form.Item
                        name="email"
                        label="邮箱地址"
                        rules={[
                          { required: true, message: '请输入邮箱地址' },
                          { type: 'email', message: '请输入有效的邮箱地址' }
                        ]}
                      >
                        <Input
                          prefix={<MailOutlined />}
                          placeholder="请输入邮箱地址"
                          className="rounded-lg"
                        />
                      </Form.Item>

                      <Form.Item
                        name="password"
                        label="密码"
                        rules={[{ required: true, message: '请输入密码' }]}
                      >
                        <Input.Password
                          prefix={<LockOutlined />}
                          placeholder="请输入密码"
                          iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                          className="rounded-lg"
                        />
                      </Form.Item>

                      <Form.Item>
                        <div className="flex justify-between items-center">
                          <Form.Item name="remember" valuePropName="checked" noStyle>
                            <Checkbox>记住我</Checkbox>
                          </Form.Item>
                          <Button type="link" className="p-0">
                            忘记密码？
                          </Button>
                        </div>
                      </Form.Item>

                      <Form.Item>
                        <Button
                          type="primary"
                          htmlType="submit"
                          loading={isSubmitting || isLoading}
                          block
                          size="large"
                          className="rounded-lg h-12 bg-gradient-to-r from-blue-600 to-purple-600 border-0"
                        >
                          登录
                        </Button>
                      </Form.Item>
                    </Form>
                  )
                },
                {
                  key: 'register',
                  label: '注册',
                  children: (
                    <Form
                      form={registerForm}
                      name="register"
                      onFinish={handleRegister}
                      layout="vertical"
                      size="large"
                    >
                      <Form.Item
                        name="username"
                        label="用户名"
                        rules={[
                          { required: true, message: '请输入用户名' },
                          { min: 3, message: '用户名至少3个字符' },
                          { max: 20, message: '用户名最多20个字符' }
                        ]}
                      >
                        <Input
                          prefix={<UserOutlined />}
                          placeholder="请输入用户名"
                          className="rounded-lg"
                        />
                      </Form.Item>

                      <Form.Item
                        name="email"
                        label="邮箱地址"
                        rules={[
                          { required: true, message: '请输入邮箱地址' },
                          { type: 'email', message: '请输入有效的邮箱地址' }
                        ]}
                      >
                        <Input
                          prefix={<MailOutlined />}
                          placeholder="请输入邮箱地址"
                          className="rounded-lg"
                        />
                      </Form.Item>

                      <Form.Item
                        name="password"
                        label="密码"
                        rules={[
                          { required: true, message: '请输入密码' },
                          { min: 8, message: '密码至少8个字符' }
                        ]}
                      >
                        <Input.Password
                          prefix={<LockOutlined />}
                          placeholder="请输入密码"
                          onChange={handlePasswordChange}
                          iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                          className="rounded-lg"
                        />
                      </Form.Item>

                      {passwordStrength && (
                        <div className="mb-4">
                          <PasswordStrengthIndicator strength={passwordStrength} />
                        </div>
                      )}

                      <Form.Item
                        name="confirmPassword"
                        label="确认密码"
                        dependencies={['password']}
                        rules={[
                          { required: true, message: '请确认密码' },
                          ({ getFieldValue }) => ({
                            validator(_, value) {
                              if (!value || getFieldValue('password') === value) {
                                return Promise.resolve();
                              }
                              return Promise.reject(new Error('两次输入的密码不一致'));
                            },
                          }),
                        ]}
                      >
                        <Input.Password
                          prefix={<LockOutlined />}
                          placeholder="请再次输入密码"
                          iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                          className="rounded-lg"
                        />
                      </Form.Item>

                      <Form.Item
                        name="agreement"
                        valuePropName="checked"
                        rules={[
                          { 
                            validator: (_, value) =>
                              value ? Promise.resolve() : Promise.reject(new Error('请同意用户协议'))
                          }
                        ]}
                      >
                        <Checkbox>
                          我已阅读并同意 <Button type="link" className="p-0">用户协议</Button> 和 <Button type="link" className="p-0">隐私政策</Button>
                        </Checkbox>
                      </Form.Item>

                      <Form.Item>
                        <Button
                          type="primary"
                          htmlType="submit"
                          loading={isSubmitting || isLoading}
                          block
                          size="large"
                          className="rounded-lg h-12 bg-gradient-to-r from-green-600 to-blue-600 border-0"
                        >
                          注册
                        </Button>
                      </Form.Item>
                    </Form>
                  )
                }
              ]}
            />

            <Divider>
              <Text type="secondary" className="text-sm">其他登录方式</Text>
            </Divider>

            <div className="flex justify-center space-x-4">
              <Button 
                shape="circle" 
                size="large" 
                className="border-gray-300 hover:border-blue-500"
                title="微信登录"
              >
                💬
              </Button>
              <Button 
                shape="circle" 
                size="large" 
                className="border-gray-300 hover:border-red-500"
                title="QQ登录"
              >
                🐧
              </Button>
              <Button 
                shape="circle" 
                size="large" 
                className="border-gray-300 hover:border-green-500"
                title="支付宝登录"
              >
                💰
              </Button>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Login;