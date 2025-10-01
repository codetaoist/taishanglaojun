import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Alert, Result, Steps } from 'antd';
import { MailOutlined, CheckCircleOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { clearError } from '../../store/slices/authSlice';

interface ForgotPasswordFormData {
  email: string;
}

interface ResetPasswordFormData {
  code: string;
  newPassword: string;
  confirmPassword: string;
}

const { Step } = Steps;

const ForgotPasswordPage: React.FC = () => {
  const [emailForm] = Form.useForm();
  const [resetForm] = Form.useForm();
  const dispatch = useAppDispatch();
  
  const { loading, error } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  
  const [currentStep, setCurrentStep] = useState(0);
  const [email, setEmail] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [resetSuccess, setResetSuccess] = useState(false);

  // 清除错误信息
  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);

  // 倒计时效果
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (countdown > 0) {
      timer = setTimeout(() => setCountdown(countdown - 1), 1000);
    }
    return () => clearTimeout(timer);
  }, [countdown]);

  // 处理发送重置邮件
  const handleSendResetEmail = async (values: ForgotPasswordFormData) => {
    try {
      // TODO: 实现忘记密码功能
      console.log('Forgot password for:', values.email);
      setEmail(values.email);
      setCurrentStep(1);
      setCountdown(60); // 60秒倒计时
    } catch (error) {
      console.error('Send reset email failed:', error);
    }
  };

  // 处理重新发送邮件
  const handleResendEmail = async () => {
    try {
      // TODO: 实现重新发送邮件功能
      console.log('Resend email for:', email);
      setCountdown(60);
    } catch (error) {
      console.error('Resend email failed:', error);
    }
  };

  // 处理密码重置
  const handleResetPassword = async (values: ResetPasswordFormData) => {
    try {
      // TODO: 实现重置密码功能
      console.log('Reset password with code:', values.code);
      setResetSuccess(true);
      setCurrentStep(2);
    } catch (error) {
      console.error('Reset password failed:', error);
    }
  };

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 步骤配置
  const steps = [
    {
      title: getText('输入邮箱', 'Enter Email'),
      description: getText('输入注册邮箱地址', 'Enter your registered email'),
    },
    {
      title: getText('验证邮箱', 'Verify Email'),
      description: getText('输入验证码', 'Enter verification code'),
    },
    {
      title: getText('重置密码', 'Reset Password'),
      description: getText('设置新密码', 'Set new password'),
    },
  ];

  // 渲染邮箱输入步骤
  const renderEmailStep = () => (
    <div className="step-content">
      <div className="step-header">
        <h3>{getText('找回密码', 'Forgot Password')}</h3>
        <p>{getText('请输入您的注册邮箱，我们将发送重置密码的验证码', 'Enter your registered email and we will send you a verification code')}</p>
      </div>

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
        form={emailForm}
        onFinish={handleSendResetEmail}
        size="large"
      >
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
            placeholder={getText('注册邮箱地址', 'Registered email address')}
            autoComplete="email"
          />
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            className="action-button"
          >
            {loading 
              ? getText('发送中...', 'Sending...') 
              : getText('发送验证码', 'Send Verification Code')
            }
          </Button>
        </Form.Item>
      </Form>
    </div>
  );

  // 渲染验证码输入步骤
  const renderVerificationStep = () => (
    <div className="step-content">
      <div className="step-header">
        <h3>{getText('验证邮箱', 'Verify Email')}</h3>
        <p>
          {getText(
            `验证码已发送至 ${email}，请查收邮件并输入验证码`,
            `Verification code has been sent to ${email}, please check your email`
          )}
        </p>
      </div>

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
        form={resetForm}
        onFinish={handleResetPassword}
        size="large"
      >
        <Form.Item
          name="code"
          rules={[
            {
              required: true,
              message: getText('请输入验证码', 'Please enter verification code'),
            },
            {
              len: 6,
              message: getText('验证码为6位数字', 'Verification code is 6 digits'),
            },
          ]}
        >
          <Input
            placeholder={getText('6位验证码', '6-digit verification code')}
            maxLength={6}
            autoComplete="one-time-code"
          />
        </Form.Item>

        <Form.Item
          name="newPassword"
          rules={[
            {
              required: true,
              message: getText('请输入新密码', 'Please enter new password'),
            },
            {
              min: 6,
              message: getText('密码至少6位字符', 'Password must be at least 6 characters'),
            },
          ]}
        >
          <Input.Password
            placeholder={getText('新密码', 'New password')}
            autoComplete="new-password"
          />
        </Form.Item>

        <Form.Item
          name="confirmPassword"
          dependencies={['newPassword']}
          rules={[
            {
              required: true,
              message: getText('请确认新密码', 'Please confirm new password'),
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error(getText('两次输入的密码不一致', 'Passwords do not match')));
              },
            }),
          ]}
        >
          <Input.Password
            placeholder={getText('确认新密码', 'Confirm new password')}
            autoComplete="new-password"
          />
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            className="action-button"
          >
            {loading 
              ? getText('重置中...', 'Resetting...') 
              : getText('重置密码', 'Reset Password')
            }
          </Button>
        </Form.Item>
      </Form>

      <div className="resend-section">
        {countdown > 0 ? (
          <span className="countdown-text">
            {getText(`${countdown}秒后可重新发送`, `Resend in ${countdown}s`)}
          </span>
        ) : (
          <Button
            type="link"
            onClick={handleResendEmail}
            loading={loading}
            className="resend-button"
          >
            {getText('重新发送验证码', 'Resend verification code')}
          </Button>
        )}
      </div>
    </div>
  );

  // 渲染成功步骤
  const renderSuccessStep = () => (
    <div className="step-content">
      <Result
        icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
        title={getText('密码重置成功', 'Password Reset Successful')}
        subTitle={getText(
          '您的密码已成功重置，请使用新密码登录',
          'Your password has been successfully reset, please login with your new password'
        )}
        extra={[
          <Button type="primary" key="login">
            <Link to="/login">
              {getText('立即登录', 'Login Now')}
            </Link>
          </Button>,
        ]}
      />
    </div>
  );

  return (
    <div className="forgot-password-page">
      <div className="page-header">
        <Link to="/login" className="back-link">
          <ArrowLeftOutlined />
          <span>{getText('返回登录', 'Back to Login')}</span>
        </Link>
      </div>

      <div className="steps-container">
        <Steps current={currentStep} size="small">
          {steps.map((step, index) => (
            <Step
              key={index}
              title={step.title}
              description={step.description}
            />
          ))}
        </Steps>
      </div>

      <div className="content-container">
        {currentStep === 0 && renderEmailStep()}
        {currentStep === 1 && renderVerificationStep()}
        {currentStep === 2 && renderSuccessStep()}
      </div>

      <style>{`
        .forgot-password-page {
          width: 100%;
          max-width: 480px;
        }

        .page-header {
          margin-bottom: 24px;
        }

        .back-link {
          display: inline-flex;
          align-items: center;
          color: var(--text-secondary);
          text-decoration: none;
          font-size: 14px;
          transition: color 0.2s;
        }

        .back-link:hover {
          color: var(--primary-color);
        }

        .back-link span {
          margin-left: 8px;
        }

        .steps-container {
          margin-bottom: 32px;
          padding: 0 16px;
        }

        .content-container {
          background: var(--bg-secondary);
          border-radius: 8px;
          padding: 32px;
          border: 1px solid var(--border-light);
        }

        .step-content {
          width: 100%;
        }

        .step-header {
          text-align: center;
          margin-bottom: 24px;
        }

        .step-header h3 {
          font-size: 20px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
        }

        .step-header p {
          font-size: 14px;
          color: var(--text-secondary);
          margin: 0;
          line-height: 1.5;
        }

        .action-button {
          height: 44px;
          font-size: 16px;
          font-weight: 500;
        }

        .resend-section {
          text-align: center;
          margin-top: 16px;
        }

        .countdown-text {
          font-size: 14px;
          color: var(--text-tertiary);
        }

        .resend-button {
          font-size: 14px;
          padding: 0;
          height: auto;
        }

        /* 响应式设计 */
        @media (max-width: 480px) {
          .forgot-password-page {
            max-width: 100%;
          }

          .content-container {
            padding: 24px 20px;
          }

          .steps-container {
            padding: 0 8px;
          }

          .step-header h3 {
            font-size: 18px;
          }

          .step-header p {
            font-size: 13px;
          }
        }
      `}</style>
    </div>
  );
};

export default ForgotPasswordPage;