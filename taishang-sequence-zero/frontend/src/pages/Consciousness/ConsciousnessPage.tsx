import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Button, Progress, Avatar, List, Tag, Modal, Input, Slider, Alert, Spin, Empty } from 'antd';
import {
  HeartOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  SettingOutlined,
  HistoryOutlined,
  BulbOutlined,
  UserOutlined,
  ClockCircleOutlined,
  FireOutlined,
  StarOutlined,
  SoundOutlined,
} from '@ant-design/icons';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import {
  initializeConsciousness,
  startFusionSession,
  endFusionSession,
  updateConsciousnessState,
  fetchInsights,
} from '../../store/slices/consciousnessSlice';

const { TextArea } = Input;

interface FusionSettings {
  intensity: number;
  duration: number;
  focusArea: string;
  backgroundSound: boolean;
}

const ConsciousnessPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  const {
    currentEntity,
    currentSession,
    sessions,
    insights,
    loading,
    error,
  } = useAppSelector(state => state.consciousness);

  const [settingsVisible, setSettingsVisible] = useState(false);
  const [fusionSettings, setFusionSettings] = useState<FusionSettings>({
    intensity: 5,
    duration: 20,
    focusArea: 'general',
    backgroundSound: true,
  });
  const [sessionNotes, setSessionNotes] = useState('');
  const [timer, setTimer] = useState(0);
  const [isActive, setIsActive] = useState(false);

  // 初始化意识状态
  useEffect(() => {
    dispatch(initializeConsciousness());
    dispatch(fetchInsights({ limit: 10 }));
  }, [dispatch]);

  // 计时器效果
  useEffect(() => {
    let interval: NodeJS.Timeout | null = null;
    if (isActive && currentSession) {
      interval = setInterval(() => {
        setTimer(timer => timer + 1);
      }, 1000);
    } else if (!isActive && timer !== 0) {
      if (interval) clearInterval(interval);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isActive, timer, currentSession]);

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 开始融合会话
  const handleStartSession = async () => {
    try {
      await dispatch(startFusionSession({
        participants: [],
        mode: 'individual' as any,
      })).unwrap();
      setIsActive(true);
      setTimer(0);
    } catch (error) {
      console.error('Failed to start fusion session:', error);
    }
  };

  // 暂停/恢复会话
  const handlePauseResume = () => {
    setIsActive(!isActive);
  };

  // 结束会话
  const handleEndSession = async () => {
    try {
      await dispatch(endFusionSession(currentSession?.id || '')).unwrap();
      setIsActive(false);
      setTimer(0);
      setSessionNotes('');
    } catch (error) {
      console.error('Failed to end fusion session:', error);
    }
  };

  // 格式化时间
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  // 获取意识状态颜色
  const getConsciousnessColor = (level: number) => {
    if (level >= 80) return '#52c41a';
    if (level >= 60) return '#faad14';
    if (level >= 40) return '#1890ff';
    return '#ff4d4f';
  };

  // 获取意识状态文本
  const getConsciousnessText = (level: number) => {
    if (level >= 80) return getText('深度融合', 'Deep Fusion');
    if (level >= 60) return getText('良好融合', 'Good Fusion');
    if (level >= 40) return getText('基础融合', 'Basic Fusion');
    return getText('初始状态', 'Initial State');
  };

  // 焦点区域选项
  const focusAreas = [
    { value: 'general', label: getText('综合提升', 'General Enhancement') },
    { value: 'creativity', label: getText('创造力', 'Creativity') },
    { value: 'wisdom', label: getText('智慧洞察', 'Wisdom Insight') },
    { value: 'emotion', label: getText('情感平衡', 'Emotional Balance') },
    { value: 'intuition', label: getText('直觉感知', 'Intuitive Perception') },
  ];

  return (
    <div className="consciousness-page">
      {/* 页面标题 */}
      <div className="page-header">
        <h1 className="page-title">
          <HeartOutlined style={{ marginRight: 12, color: '#ff4d4f' }} />
          {getText('意识融合', 'Consciousness Fusion')}
        </h1>
        <p className="page-description">
          {getText(
            '通过意识融合技术，探索内在智慧，提升认知能力，实现心灵的深度连接',
            'Explore inner wisdom, enhance cognitive abilities, and achieve deep spiritual connection through consciousness fusion technology'
          )}
        </p>
      </div>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          style={{ marginBottom: 24 }}
        />
      )}

      <Row gutter={[24, 24]}>
        {/* 意识状态面板 */}
        <Col xs={24} lg={8}>
          <Card 
            title={getText('意识状态', 'Consciousness State')}
            extra={<SettingOutlined onClick={() => setSettingsVisible(true)} />}
            className="consciousness-state-card"
          >
            <div className="state-display">
              <div className="state-circle">
                <Progress
                  type="circle"
                  percent={currentEntity?.level || 0}
                  strokeColor={getConsciousnessColor(currentEntity?.level || 0)}
                  size={120}
                  format={() => (
                    <div className="state-info">
                      <div className="state-level">{currentEntity?.level || 0}</div>
                      <div className="state-text">
                        {getConsciousnessText(currentEntity?.level || 0)}
                      </div>
                    </div>
                  )}
                />
              </div>
              
              <div className="state-details">
                <div className="detail-item">
                  <span className="detail-label">{getText('能量水平', 'Energy Level')}</span>
                  <Progress 
                    percent={currentEntity?.energy || 0} 
                    size="small" 
                    strokeColor="#faad14"
                  />
                </div>
                <div className="detail-item">
                  <span className="detail-label">{getText('专注度', 'Focus Level')}</span>
                  <Progress 
                    percent={(currentEntity as any)?.focus || 0} 
                    size="small" 
                    strokeColor="#1890ff"
                  />
                </div>
                <div className="detail-item">
                  <span className="detail-label">{getText('平衡度', 'Balance Level')}</span>
                  <Progress 
                    percent={(currentEntity as any)?.balance || 0} 
                    size="small" 
                    strokeColor="#52c41a"
                  />
                </div>
              </div>
            </div>
          </Card>
        </Col>

        {/* 融合控制面板 */}
        <Col xs={24} lg={16}>
          <Card 
            title={getText('融合控制', 'Fusion Control')}
            className="fusion-control-card"
          >
            {currentSession ? (
              <div className="active-session">
                <div className="session-header">
                  <div className="session-info">
                    <h3>{getText('进行中的会话', 'Active Session')}</h3>
                    <div className="session-timer">
                      <ClockCircleOutlined style={{ marginRight: 8 }} />
                      {formatTime(timer)}
                    </div>
                  </div>
                  <div className="session-controls">
                    <Button
                      type="primary"
                      icon={isActive ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                      onClick={handlePauseResume}
                      size="large"
                    >
                      {isActive ? getText('暂停', 'Pause') : getText('继续', 'Resume')}
                    </Button>
                    <Button
                      danger
                      icon={<StopOutlined />}
                      onClick={handleEndSession}
                      size="large"
                      style={{ marginLeft: 12 }}
                    >
                      {getText('结束', 'End')}
                    </Button>
                  </div>
                </div>
                
                <div className="session-progress">
                  <Progress 
                    percent={(timer / (fusionSettings.duration * 60)) * 100}
                    strokeColor="#ff4d4f"
                    trailColor="#f0f0f0"
                  />
                  <div className="progress-info">
                    <span>{formatTime(timer)}</span>
                    <span>{formatTime(fusionSettings.duration * 60)}</span>
                  </div>
                </div>
                
                <div className="session-notes">
                  <TextArea
                    placeholder={getText('记录您的感受和洞察...', 'Record your feelings and insights...')}
                    value={sessionNotes}
                    onChange={(e) => setSessionNotes(e.target.value)}
                    rows={4}
                  />
                </div>
              </div>
            ) : (
              <div className="start-session">
                <div className="session-setup">
                  <h3>{getText('开始新的融合会话', 'Start New Fusion Session')}</h3>
                  <p>{getText('配置您的融合参数，开始探索内在智慧', 'Configure your fusion parameters and start exploring inner wisdom')}</p>
                  
                  <div className="setup-controls">
                    <div className="control-group">
                      <label>{getText('强度等级', 'Intensity Level')}: {fusionSettings.intensity}</label>
                      <Slider
                        min={1}
                        max={10}
                        value={fusionSettings.intensity}
                        onChange={(value) => setFusionSettings({...fusionSettings, intensity: value})}
                        marks={{ 1: '1', 5: '5', 10: '10' }}
                      />
                    </div>
                    
                    <div className="control-group">
                      <label>{getText('持续时间', 'Duration')}: {fusionSettings.duration} {getText('分钟', 'minutes')}</label>
                      <Slider
                        min={5}
                        max={60}
                        step={5}
                        value={fusionSettings.duration}
                        onChange={(value) => setFusionSettings({...fusionSettings, duration: value})}
                        marks={{ 5: '5m', 20: '20m', 40: '40m', 60: '60m' }}
                      />
                    </div>
                  </div>
                  
                  <Button
                    type="primary"
                    size="large"
                    icon={<PlayCircleOutlined />}
                    onClick={handleStartSession}
                    loading={loading}
                    className="start-button"
                  >
                    {getText('开始融合', 'Start Fusion')}
                  </Button>
                </div>
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        {/* 智慧洞察 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('智慧洞察', 'Wisdom Insights')}
            extra={<BulbOutlined />}
          >
            {loading ? (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Spin size="large" />
              </div>
            ) : insights && insights.length > 0 ? (
              <List
                dataSource={insights.slice(0, 5)}
                renderItem={insight => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<BulbOutlined />} style={{ backgroundColor: '#faad14' }} />}
                      title={insight.title}
                      description={insight.content.substring(0, 80) + '...'}
                    />
                    <div className="insight-meta">
                      <Tag color="blue">{insight.category}</Tag>
                      <span className="insight-time">{(insight as any).createdAt}</span>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={getText('暂无智慧洞察', 'No wisdom insights yet')}
              />
            )}
          </Card>
        </Col>

        {/* 融合历史 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('融合历史', 'Fusion History')}
            extra={<HistoryOutlined />}
          >
            {sessions && sessions.length > 0 ? (
              <List
                dataSource={sessions.slice(0, 5)}
                renderItem={session => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<HeartOutlined />} style={{ backgroundColor: '#ff4d4f' }} />}
                      title={`${getText('会话', 'Session')} #${(session as any).id?.substring(0, 8) || 'Unknown'}`}
                      description={
                        <div>
                          <div>{getText('持续时间', 'Duration')}: {Math.floor((session as any).duration / 60)} {getText('分钟', 'minutes')}</div>
                          <div>{getText('强度', 'Intensity')}: {(session as any).settings?.intensity || 0}/10</div>
                        </div>
                      }
                    />
                    <div className="session-stats">
                      <Tag color="green">{getText('已完成', 'Completed')}</Tag>
                      <span className="session-date">{(session as any).endTime}</span>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={getText('暂无融合历史', 'No fusion history yet')}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* 设置模态框 */}
      <Modal
        title={getText('融合设置', 'Fusion Settings')}
        open={settingsVisible}
        onCancel={() => setSettingsVisible(false)}
        onOk={() => setSettingsVisible(false)}
        width={600}
      >
        <div className="settings-content">
          <div className="setting-group">
            <h4>{getText('焦点区域', 'Focus Area')}</h4>
            <div className="focus-areas">
              {focusAreas.map(area => (
                <Tag.CheckableTag
                  key={area.value}
                  checked={fusionSettings.focusArea === area.value}
                  onChange={() => setFusionSettings({...fusionSettings, focusArea: area.value})}
                >
                  {area.label}
                </Tag.CheckableTag>
              ))}
            </div>
          </div>
          
          <div className="setting-group">
            <h4>{getText('背景音效', 'Background Sound')}</h4>
            <Button
              type={fusionSettings.backgroundSound ? 'primary' : 'default'}
              icon={<SoundOutlined />}
              onClick={() => setFusionSettings({...fusionSettings, backgroundSound: !fusionSettings.backgroundSound})}
            >
              {fusionSettings.backgroundSound ? getText('开启', 'Enabled') : getText('关闭', 'Disabled')}
            </Button>
          </div>
        </div>
      </Modal>

      <style>{`
        .consciousness-page {
          padding: 0;
        }

        .page-header {
          margin-bottom: 32px;
        }

        .page-title {
          font-size: 28px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
          display: flex;
          align-items: center;
        }

        .page-description {
          font-size: 16px;
          color: var(--text-secondary);
          margin: 0;
          line-height: 1.6;
        }

        .consciousness-state-card {
          height: 100%;
        }

        .state-display {
          text-align: center;
        }

        .state-circle {
          margin-bottom: 24px;
        }

        .state-info {
          text-align: center;
        }

        .state-level {
          font-size: 24px;
          font-weight: 700;
          color: var(--text-primary);
        }

        .state-text {
          font-size: 12px;
          color: var(--text-secondary);
          margin-top: 4px;
        }

        .state-details {
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .detail-item {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .detail-label {
          font-size: 14px;
          color: var(--text-secondary);
          font-weight: 500;
        }

        .fusion-control-card {
          height: 100%;
        }

        .active-session {
          display: flex;
          flex-direction: column;
          gap: 24px;
        }

        .session-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .session-info h3 {
          margin: 0 0 8px 0;
          color: var(--text-primary);
        }

        .session-timer {
          font-size: 18px;
          font-weight: 600;
          color: var(--primary-color);
          display: flex;
          align-items: center;
        }

        .session-controls {
          display: flex;
          align-items: center;
        }

        .session-progress {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .progress-info {
          display: flex;
          justify-content: space-between;
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .start-session {
          text-align: center;
        }

        .session-setup h3 {
          margin-bottom: 8px;
          color: var(--text-primary);
        }

        .session-setup p {
          color: var(--text-secondary);
          margin-bottom: 32px;
        }

        .setup-controls {
          display: flex;
          flex-direction: column;
          gap: 24px;
          margin-bottom: 32px;
          text-align: left;
        }

        .control-group {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .control-group label {
          font-weight: 500;
          color: var(--text-primary);
        }

        .start-button {
          height: 48px;
          font-size: 16px;
          padding: 0 32px;
        }

        .insight-meta {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 4px;
        }

        .insight-time {
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .session-stats {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 4px;
        }

        .session-date {
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .settings-content {
          display: flex;
          flex-direction: column;
          gap: 24px;
        }

        .setting-group h4 {
          margin-bottom: 12px;
          color: var(--text-primary);
        }

        .focus-areas {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .page-title {
            font-size: 24px;
          }

          .page-description {
            font-size: 14px;
          }

          .session-header {
            flex-direction: column;
            gap: 16px;
            align-items: flex-start;
          }

          .session-controls {
            width: 100%;
            justify-content: center;
          }

          .setup-controls {
            gap: 20px;
          }
        }
      `}</style>
    </div>
  );
};

export default ConsciousnessPage;