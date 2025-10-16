import React from 'react';
import { Typography, Card } from 'antd';

const { Title, Paragraph } = Typography;

function VideoProcessing() {
  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>视频处理</Title>
      <Paragraph>多模态 AI：视频处理功能占位页面，后续将支持解析与生成。</Paragraph>
      <Card>暂未实现具体功能。请参考文档与菜单配置。</Card>
    </div>
  );
}

export default VideoProcessing;