import React from 'react';
import { Progress, Space, Typography, Tag } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { getPasswordStrengthText } from '../utils/passwordValidator';
import type { PasswordStrength } from '../utils/passwordValidator';

const { Text } = Typography;

interface PasswordStrengthIndicatorProps {
  strength: PasswordStrength;
  showDetails?: boolean;
  className?: string;
}

const PasswordStrengthIndicator: React.FC<PasswordStrengthIndicatorProps> = ({
  strength,
  showDetails = true,
  className = ''
}) => {
  const getProgressStatus = () => {
    if (strength.level === 'weak') return 'exception';
    if (strength.level === 'fair') return 'normal';
    if (strength.level === 'good') return 'active';
    return 'success';
  };

  const getStrengthIcon = () => {
    if (strength.score >= 3.5) {
      return <CheckCircleOutlined style={{ color: strength.color }} />;
    } else if (strength.score >= 2) {
      return <InfoCircleOutlined style={{ color: strength.color }} />;
    } else {
      return <CloseCircleOutlined style={{ color: strength.color }} />;
    }
  };

  return (
    <div className={`password-strength-indicator ${className}`}>
      <Space direction="vertical" size="small" style={{ width: '100%' }}>
        {/* 强度进度条 */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Progress
            percent={strength.percentage}
            strokeColor={strength.color}
            status={getProgressStatus()}
            size="small"
            showInfo={false}
            style={{ flex: 1, minWidth: '100px' }}
          />
          <Space size="small">
            {getStrengthIcon()}
            <Text style={{ color: strength.color, fontWeight: 500, minWidth: '40px' }}>
              {getPasswordStrengthText(strength.level)}
            </Text>
          </Space>
        </div>

        {/* 详细反馈 */}
        {showDetails && strength.feedback.length > 0 && (
          <div style={{ marginTop: '4px' }}>
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              {strength.feedback.map((feedback, index) => (
                <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  {index === 0 ? (
                    // 第一条是总体评价
                    <Tag 
                      color={strength.score >= 3 ? 'success' : strength.score >= 2 ? 'warning' : 'error'}
                      style={{ margin: 0, fontSize: '12px' }}
                    >
                      {feedback}
                    </Tag>
                  ) : (
                    // 其他是具体建议
                    <Text 
                      type="secondary" 
                      style={{ 
                        fontSize: '12px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '4px'
                      }}
                    >
                      <span style={{ 
                        width: '4px', 
                        height: '4px', 
                        borderRadius: '50%', 
                        backgroundColor: '#d9d9d9',
                        display: 'inline-block'
                      }} />
                      {feedback}
                    </Text>
                  )}
                </div>
              ))}
            </Space>
          </div>
        )}
      </Space>

      <style jsx>{`
        .password-strength-indicator {
          padding: 8px 0;
        }
        
        .password-strength-indicator .ant-progress-line {
          margin: 0;
        }
        
        .password-strength-indicator .ant-progress-bg {
          border-radius: 4px;
        }
        
        .password-strength-indicator .ant-tag {
          border-radius: 4px;
          font-size: 12px;
          padding: 2px 6px;
          line-height: 1.2;
        }
      `}</style>
    </div>
  );
};

export default PasswordStrengthIndicator;