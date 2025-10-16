import React from 'react';
import { Layout, Typography } from 'antd';
import { useTheme } from '../../hooks';

const { Footer: AntFooter } = Layout;
const { Text } = Typography;

interface FooterProps {
  className?: string;
}

export const Footer: React.FC<FooterProps> = ({ className }) => {
  const { theme } = useTheme();
  const currentYear = new Date().getFullYear();

  return (
    <AntFooter 
      className={`
        ${className}
        ${theme === 'dark' 
          ? 'bg-gray-800 border-gray-700' 
          : 'bg-gray-50 border-gray-200'
        }
        border-t
      `}
      style={{
        padding: '24px',
        marginTop: 'auto'
      }}
    >
      <div className="max-w-7xl mx-auto text-center">
        <Text className={`text-sm ${theme === 'dark' ? 'text-gray-400' : 'text-gray-600'}`}>
          © {currentYear} 太上老君AI平台. 保留所有权利.
        </Text>
      </div>
    </AntFooter>
  );
};

export default Footer;