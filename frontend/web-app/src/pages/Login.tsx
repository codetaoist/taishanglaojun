import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, Space, Divider, Alert } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from "../hooks/useAuth";
import PasswordStrengthIndicator from '../components/PasswordStrengthIndicator';
import { checkPasswordStrength, validatePassword } from '../utils/passwordValidator';
import type { PasswordStrength } from '../utils/passwordValidator';

const { Title } = Typography;

interface LoginForm {
  email: string;
  password: string;
}

interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

const Login: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);
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

  const handleLogin = async (values: LoginForm) => {
    try {
      setIsSubmitting(true);
      clearMessages();
      
      const result = await login(values.email, values.password);
      
      if (result.success) {
        setSuccessMessage('登录成功！正在跳转...');
        // 不需要延迟，直接导航
        navigate('/', { replace: true });
      } else {
        setErrorMessage(result.error || '登录失败，请检查您的邮箱和密码');
        // 保持表单数据，不清空输入
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
        // 检查是否需要邮箱验证
        if (result.data?.message && result.data.message.includes('验证邮件')) {
          setSuccessMessage('注册成功！请查收验证邮件并点击验证链接激活账户。');
          // 不自动跳转，让用户去验证邮箱
        } else {
          setSuccessMessage('注册成功！正在跳转...');
          // 延迟跳转，让用户看到成功消息
          setTimeout(() => {
            navigate('/', { replace: true });
          }, 1000);
        }
      } else {
        setErrorMessage(result.error || '注册失败，请检查输入信息');
        // 保持表单数据，不清空输入
      }
    } catch (error: any) {
      setErrorMessage('网络错误，请稍后重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  // 处理密码输入变化
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value;
    if (password) {
      const strength = checkPasswordStrength(password);
      setPasswordStrength(strength);
    } else {
      setPasswordStrength(null);
    }
  };

  const switchMode = () => {
    setIsLogin(!isLogin);
    loginForm.resetFields();
    registerForm.resetFields();
    clearMessages();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <Card 
        className="w-full max-w-md shadow-2xl border-0"
        style={{ borderRadius: '16px' }}
      >
        <div className="text-center mb-8">
          <Title level={2} className="mb-2 text-gray-800">
            {isLogin ? '欢迎回来' : '创建账户'}
          </Title>
          <span className="text-base text-gray-500">
            {isLogin ? '登录到太上老君系统' : '注册太上老君系统账户'}
          </span>
        </div>

        {/* 错误消息显示 */}
        {errorMessage && (
          <Alert
            message={errorMessage}
            type="error"
            showIcon
            closable
            onClose={clearMessages}
            style={{ marginBottom: '16px' }}
          />
        )}

        {/* 成功消息显示 */}
        {successMessage && (
          <Alert
            message={successMessage}
            type="success"
            showIcon
            icon={<CheckCircleOutlined />}
            style={{ marginBottom: '16px' }}
          />
        )}

        {isLogin ? (
          <Form
            form={loginForm}
            name="login"
            onFinish={handleLogin}
            layout="vertical"
            size="large"
          >
            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input 
                prefix={<MailOutlined />} 
                placeholder="请输入邮箱"
                className="rounded-lg"
                disabled={isSubmitting || !!successMessage}
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
                className="rounded-lg"
                disabled={isSubmitting || !!successMessage}
              />
            </Form.Item>

            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={isSubmitting || isLoading}
                disabled={!!successMessage}
                className="w-full h-12 text-lg font-medium rounded-lg"
                style={{ marginTop: '16px' }}
              >
                {isSubmitting ? '登录中...' : '登录'}
              </Button>
            </Form.Item>
          </Form>
        ) : (
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
                disabled={isSubmitting || !!successMessage}
              />
            </Form.Item>

            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input 
                prefix={<MailOutlined />} 
                placeholder="请输入邮箱"
                className="rounded-lg"
                disabled={isSubmitting || !!successMessage}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 8, message: '密码至少8个字符' },
                {
                  validator: (_, value) => {
                    if (!value) return Promise.resolve();
                    const validation = validatePassword(value);
                    if (!validation.valid) {
                      return Promise.reject(new Error(validation.errors[0]));
                    }
                    return Promise.resolve();
                  }
                }
              ]}
            >
              <Input.Password 
                prefix={<LockOutlined />} 
                placeholder="请输入密码"
                className="rounded-lg"
                disabled={isSubmitting || !!successMessage}
                onChange={handlePasswordChange}
              />
            </Form.Item>

            {/* 密码强度指示器 */}
            {passwordStrength && (
              <Form.Item>
                <PasswordStrengthIndicator 
                  strength={passwordStrength} 
                  showDetails={true}
                />
              </Form.Item>
            )}

            <Form.Item
              name="confirmPassword"
              label="确认密码"
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
                className="rounded-lg"
                disabled={isSubmitting || !!successMessage}
              />
            </Form.Item>

            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={isSubmitting || isLoading}
                disabled={!!successMessage}
                className="w-full h-12 text-lg font-medium rounded-lg"
                style={{ marginTop: '16px' }}
              >
                {isSubmitting ? '注册中...' : '注册'}
              </Button>
            </Form.Item>
          </Form>
        )}

        <Divider className="my-6">
          <span className="text-gray-500">或者</span>
        </Divider>

        <div className="text-center">
          <Space direction="vertical" size="middle" className="w-full">
            <span className="text-gray-500">
              {isLogin ? '还没有账户？' : '已有账户？'}
            </span>
            <Button 
              type="link" 
              onClick={switchMode}
              disabled={isSubmitting || !!successMessage}
              className="text-lg font-medium p-0 h-auto"
            >
              {isLogin ? '立即注册' : '立即登录'}
            </Button>
          </Space>
        </div>

        <div className="text-center mt-6">
          <span className="text-sm text-gray-500">
            登录即表示您同意我们的服务条款和隐私政策
          </span>
        </div>
      </Card>
    </div>
  );
};

export default Login;