import React from 'react';
import PlaceholderPage from '../components/common/PlaceholderPage';

const HealthManagement: React.FC = () => {
  return (
    <PlaceholderPage
      title="健康管理系统"
      description="融合传统中医养生智慧与现代健康管理理念，提供个性化的健康指导、养生建议和生活方式优化方案。"
      status="planned"
      features={[
        '传统养生知识库',
        '个人体质分析',
        '季节养生指导',
        '饮食调理建议',
        '运动健身计划',
        '情志调节方法',
        '健康数据跟踪',
        '中医经络穴位指导',
        '太极八卦养生功法',
        '健康社群交流'
      ]}
      estimatedCompletion="2024年第四季度"
      relatedPages={[
        { title: '文化智慧', path: '/wisdom' },
        { title: '个人中心', path: '/profile' },
        { title: '社区讨论', path: '/community' }
      ]}
    />
  );
};

export default HealthManagement;