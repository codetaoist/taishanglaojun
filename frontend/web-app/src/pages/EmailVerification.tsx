import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Card, Button, Result, Spin, message, Typography } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, MailOutlined } from '@ant-design/icons';
import { authAPI } from '../services/api';

const { Title, Paragraph } = Typography;

interface VerificationState {
  status: 'loading' | 'success' | 'error' | 'expired' | 'invalid';
  message: string;
}

const EmailVerification: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [verificationState, setVerificationState] = useState<VerificationState>({
    status: 'loading',
    message: '正在验证您的邮箱...'
  });
  const [resending, setResending] = useState(false);

  const token = searchParams.get('token');
  const email = searchParams.get('email');

  useEffect(() => {
    if (token) {
      verifyEmail(token);
    } else {
      setVerificationState({
        status: 'invalid',
        message: '验证链接无效，请检查您的邮箱或重新注册。'
      });
    }
  }, [token]);

  const verifyEmail = async (verificationToken: string) => {
    try {
      const response = await authAPI.verifyEmail({ token: verificationToken });
      
      if (response.success) {
        setVerificationState({
          status: 'success',
          message: '邮箱验证成功！您现在可以正常使用所有功能。'
        });
        
        // 3秒后自动跳转到登录页面
        setTimeout(() => {
          navigate('/login', { 
            state: { 
              message: '邮箱验证成功，请登录您的账户',
              type: 'success'
            }
          });
        }, 3000);
      } else {
        setVerificationState({
          status: 'error',
          message: response.message || '验证失败，请重试。'
        });
      }
    } catch (error: any) {
      console.error('Email verification error:', error);
      
      if (error.response?.status === 400) {
        const errorMessage = error.response.data?.message || '';
        if (errorMessage.includes('expired') || errorMessage.includes('过期')) {
          setVerificationState({
            status: 'expired',
            message: '验证链接已过期，请重新发送验证邮件。'
          });
        } else if (errorMessage.includes('invalid') || errorMessage.includes('无效')) {
          setVerificationState({
            status: 'invalid',
            message: '验证链接无效，请检查您的邮箱或重新注册。'
          });
        } else {
          setVerificationState({
            status: 'error',
            message: errorMessage || '验证失败，请重试。'
          });
        }
      } else {
        setVerificationState({
          status: 'error',
          message: '网络错误，请检查您的网络连接后重试。'
        });
      }
    }
  };

  const handleResendVerification = async () => {
    if (!email) {
      message.error('无法获取邮箱地址，请重新注册。');
      return;
    }

    setResending(true);
    try {
      const response = await authAPI.resendVerification({ email });
      
      if (response.success) {
        message.success('验证邮件已重新发送，请查收您的邮箱。');
        setVerificationState({
          status: 'loading',
          message: '新的验证邮件已发送，请查收并点击验证链接。'
        });
      } else {
        message.error(response.message || '发送失败，请重试。');
      }
    } catch (error: any) {
      console.error('Resend verification error:', error);
      message.error('发送失败，请重试。');
    } finally {
      setResending(false);
    }
  };

  const renderResult = () => {
    switch (verificationState.status) {
      case 'loading':
        return (
          <div className="text-center py-8">
            <Spin size="large" />
            <div className="mt-4">
              <Title level={3}>正在验证邮箱</Title>
              <Paragraph className="text-gray-600">
                {verificationState.message}
              </Paragraph>
            </div>
          </div>
        );

      case 'success':
        return (
          <Result
            icon={<CheckCircleOutlined className="text-green-500" />}
            title="邮箱验证成功！"
            subTitle={verificationState.message}
            extra={[
              <Button 
                type="primary" 
                key="login"
                onClick={() => navigate('/login')}
              >
                立即登录
              </Button>
            ]}
          />
        );

      case 'expired':
        return (
          <Result
            icon={<CloseCircleOutlined className="text-orange-500" />}
            title="验证链接已过期"
            subTitle={verificationState.message}
            extra={[
              <Button 
                type="primary" 
                icon={<MailOutlined />}
                loading={resending}
                onClick={handleResendVerification}
                key="resend"
              >
                重新发送验证邮件
              </Button>,
              <Button 
                key="register"
                onClick={() => navigate('/login')}
              >
                返回注册
              </Button>
            ]}
          />
        );

      case 'invalid':
        return (
          <Result
            icon={<CloseCircleOutlined className="text-red-500" />}
            title="验证链接无效"
            subTitle={verificationState.message}
            extra={[
              <Button 
                type="primary"
                key="register"
                onClick={() => navigate('/login')}
              >
                重新注册
              </Button>
            ]}
          />
        );

      case 'error':
      default:
        return (
          <Result
            icon={<CloseCircleOutlined className="text-red-500" />}
            title="验证失败"
            subTitle={verificationState.message}
            extra={[
              <Button 
                type="primary" 
                key="retry"
                onClick={() => token && verifyEmail(token)}
              >
                重试验证
              </Button>,
              <Button 
                icon={<MailOutlined />}
                loading={resending}
                onClick={handleResendVerification}
                key="resend"
              >
                重新发送验证邮件
              </Button>
            ]}
          />
        );
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <Card className="shadow-lg">
          {renderResult()}
        </Card>
      </div>
    </div>
  );
};

export default EmailVerification;