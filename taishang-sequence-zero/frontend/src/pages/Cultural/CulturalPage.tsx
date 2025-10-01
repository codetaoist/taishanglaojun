import React, { useEffect, useState } from 'react';
import { Row, Col, Card, List, Avatar, Tag, Button, Input, Select, Tabs, Progress, Modal, Rate, Empty, Spin } from 'antd';
import {
  BookOutlined,
  SearchOutlined,
  FilterOutlined,
  StarOutlined,
  HeartOutlined,
  MessageOutlined,
  TrophyOutlined,
  PlayCircleOutlined,
  QuestionCircleOutlined,
  BulbOutlined,
  HistoryOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import {
  fetchWisdomItems,
  startWisdomDialogue,
  fetchCulturalQuizzes,
  submitQuizAnswers,
  updateLearningProgress,
} from '../../store/slices/culturalSlice';

const { Search } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

interface QuizModalState {
  visible: boolean;
  quiz: any;
  currentQuestion: number;
  answers: Record<string, string>;
  score: number;
  completed: boolean;
}

const CulturalPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { language } = useAppSelector(state => state.ui);
  const {
    wisdomItems,
    learningProgress,
    dialogues,
    quizzes,
    bookmarks,
    favorites,
    loading,
    error,
  } = useAppSelector(state => state.cultural);

  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedDifficulty, setSelectedDifficulty] = useState('all');
  const [activeTab, setActiveTab] = useState('wisdom');
  const [quizModal, setQuizModal] = useState<QuizModalState>({
    visible: false,
    quiz: null,
    currentQuestion: 0,
    answers: {},
    score: 0,
    completed: false,
  });

  // 加载数据
  useEffect(() => {
    dispatch(fetchWisdomItems({ category: selectedCategory as any }));
    dispatch(fetchCulturalQuizzes({ category: selectedCategory as any, level: selectedDifficulty as any }));
  }, [dispatch, selectedCategory, selectedDifficulty]);

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchTerm(value);
    dispatch(fetchWisdomItems({
      category: selectedCategory as any,
      search: value
    }));
  };

  // 处理分类筛选
  const handleCategoryChange = (category: string) => {
    setSelectedCategory(category);
  };

  // 处理难度筛选
  const handleDifficultyChange = (difficulty: string) => {
    setSelectedDifficulty(difficulty);
  };

  // 开始智慧对话
  const handleStartDialogue = async (wisdomId: string) => {
    try {
      await dispatch(startWisdomDialogue(wisdomId)).unwrap();
      // 可以跳转到对话页面或打开对话模态框
    } catch (error) {
      console.error('Failed to start wisdom dialogue:', error);
    }
  };

  // 开始测验
  const handleStartQuiz = (quiz: any) => {
    setQuizModal({
      visible: true,
      quiz,
      currentQuestion: 0,
      answers: {},
      score: 0,
      completed: false,
    });
  };

  // 提交测验答案
  const handleQuizAnswer = (answer: string) => {
    const newAnswers = {
      ...quizModal.answers,
      [quizModal.currentQuestion]: answer,
    };
    
    setQuizModal({
      ...quizModal,
      answers: newAnswers,
    });
  };

  // 下一题或完成测验
  const handleNextQuestion = async () => {
    const { quiz, currentQuestion, answers } = quizModal;
    
    if (currentQuestion < quiz.questions.length - 1) {
      setQuizModal({
        ...quizModal,
        currentQuestion: currentQuestion + 1,
      });
    } else {
      // 完成测验，计算分数
      try {
        const result = await dispatch(submitQuizAnswers({
          quizId: quiz.id,
          answers: Object.entries(answers).map(([questionId, selectedAnswer]) => ({
            questionId,
            selectedAnswer: parseInt(selectedAnswer)
          })),
        })).unwrap();
        
        setQuizModal({
          ...quizModal,
          score: result.data?.score || 0,
          completed: true,
        });
      } catch (error) {
        console.error('Failed to submit quiz:', error);
      }
    }
  };

  // 关闭测验模态框
  const handleCloseQuiz = () => {
    setQuizModal({
      visible: false,
      quiz: null,
      currentQuestion: 0,
      answers: {},
      score: 0,
      completed: false,
    });
  };

  // 切换书签
  const handleToggleBookmark = (wisdomId: string) => {
    // TODO: Implement bookmark functionality
    console.log('Toggle bookmark for:', wisdomId);
  };

  // 切换收藏
  const handleToggleFavorite = (wisdomId: string) => {
    // TODO: Implement favorite functionality
    console.log('Toggle favorite for:', wisdomId);
  };

  // 分类选项
  const categories = [
    { value: 'all', label: getText('全部', 'All') },
    { value: 'taoism', label: getText('道家', 'Taoism') },
    { value: 'confucianism', label: getText('儒家', 'Confucianism') },
    { value: 'buddhism', label: getText('佛家', 'Buddhism') },
    { value: 'philosophy', label: getText('哲学', 'Philosophy') },
    { value: 'literature', label: getText('文学', 'Literature') },
    { value: 'history', label: getText('历史', 'History') },
  ];

  // 难度选项
  const difficulties = [
    { value: 'all', label: getText('全部', 'All') },
    { value: 'beginner', label: getText('初级', 'Beginner') },
    { value: 'intermediate', label: getText('中级', 'Intermediate') },
    { value: 'advanced', label: getText('高级', 'Advanced') },
  ];

  // 过滤智慧条目
  const filteredWisdomItems = wisdomItems?.filter(item => {
    const matchesSearch = !searchTerm || 
      item.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      item.content.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory = selectedCategory === 'all' || item.category === selectedCategory;
    const matchesDifficulty = selectedDifficulty === 'all' || item.level === selectedDifficulty;
    
    return matchesSearch && matchesCategory && matchesDifficulty;
  }) || [];

  return (
    <div className="cultural-page">
      {/* 页面标题 */}
      <div className="page-header">
        <h1 className="page-title">
          <BookOutlined style={{ marginRight: 12, color: '#1890ff' }} />
          {getText('文化智慧', 'Cultural Wisdom')}
        </h1>
        <p className="page-description">
          {getText(
            '探索中华文化的深邃智慧，通过经典学习、智慧对话和文化测验，传承千年文明',
            'Explore the profound wisdom of Chinese culture through classic learning, wisdom dialogues, and cultural quizzes, inheriting millennia of civilization'
          )}
        </p>
      </div>

      {/* 学习进度概览 */}
      <Row gutter={[16, 16]} className="progress-overview">
        <Col xs={12} sm={6}>
          <Card>
            <div className="stat-item">
              <div className="stat-icon">
                <BookOutlined style={{ color: '#1890ff' }} />
              </div>
              <div className="stat-content">
                <div className="stat-number">{(learningProgress as any)?.studiedItems || 0}</div>
                <div className="stat-label">{getText('已学习', 'Studied')}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <div className="stat-item">
              <div className="stat-icon">
                <StarOutlined style={{ color: '#faad14' }} />
              </div>
              <div className="stat-content">
                <div className="stat-number">{(learningProgress as any)?.totalPoints || 0}</div>
                <div className="stat-label">{getText('智慧积分', 'Wisdom Points')}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <div className="stat-item">
              <div className="stat-icon">
                <TrophyOutlined style={{ color: '#52c41a' }} />
              </div>
              <div className="stat-content">
                <div className="stat-number">{(learningProgress as any)?.completedQuests || 0}</div>
                <div className="stat-label">{getText('完成任务', 'Completed')}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <div className="stat-item">
              <div className="stat-icon">
                <HeartOutlined style={{ color: '#ff4d4f' }} />
              </div>
              <div className="stat-content">
                <div className="stat-number">{(learningProgress as any)?.currentStreak || 0}</div>
                <div className="stat-label">{getText('连续天数', 'Day Streak')}</div>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 搜索和筛选 */}
      <Card className="filter-card">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder={getText('搜索智慧内容...', 'Search wisdom content...')}
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              onSearch={handleSearch}
            />
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              value={selectedCategory}
              onChange={handleCategoryChange}
              size="large"
              style={{ width: '100%' }}
              placeholder={getText('分类', 'Category')}
            >
              {categories.map(cat => (
                <Option key={cat.value} value={cat.value}>{cat.label}</Option>
              ))}
            </Select>
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              value={selectedDifficulty}
              onChange={handleDifficultyChange}
              size="large"
              style={{ width: '100%' }}
              placeholder={getText('难度', 'Difficulty')}
            >
              {difficulties.map(diff => (
                <Option key={diff.value} value={diff.value}>{diff.label}</Option>
              ))}
            </Select>
          </Col>
        </Row>
      </Card>

      {/* 主要内容 */}
      <Card className="main-content">
        <Tabs activeKey={activeTab} onChange={setActiveTab} size="large">
          {/* 智慧学习 */}
          <TabPane 
            tab={
              <span>
                <BookOutlined />
                {getText('智慧学习', 'Wisdom Learning')}
              </span>
            } 
            key="wisdom"
          >
            {loading ? (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Spin size="large" />
              </div>
            ) : filteredWisdomItems.length > 0 ? (
              <List
                grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3, xxl: 4 }}
                dataSource={filteredWisdomItems}
                renderItem={item => (
                  <List.Item>
                    <Card
                      hoverable
                      className="wisdom-card"
                      actions={[
                        <Button 
                          type="text" 
                          icon={<MessageOutlined />}
                          onClick={() => handleStartDialogue(item.id)}
                        >
                          {getText('对话', 'Dialogue')}
                        </Button>,
                        <Button 
                          type="text" 
                          icon={bookmarks.includes(item.id) ? <StarOutlined style={{color: '#faad14'}} /> : <StarOutlined />}
                          onClick={() => handleToggleBookmark(item.id)}
                        >
                          {getText('书签', 'Bookmark')}
                        </Button>,
                        <Button 
                          type="text" 
                          icon={favorites.includes(item.id) ? <HeartOutlined style={{color: '#ff4d4f'}} /> : <HeartOutlined />}
                          onClick={() => handleToggleFavorite(item.id)}
                        >
                          {getText('收藏', 'Favorite')}
                        </Button>,
                      ]}
                    >
                      <Card.Meta
                        avatar={<Avatar icon={<BookOutlined />} style={{ backgroundColor: '#1890ff' }} />}
                        title={item.title}
                        description={
                          <div>
                            <p className="wisdom-content">{item.content.substring(0, 100)}...</p>
                            <div className="wisdom-tags">
                              <Tag color="blue">{item.category}</Tag>
                              <Tag color="green">{item.difficulty}</Tag>
                              {item.source && <Tag>{item.source}</Tag>}
                            </div>
                          </div>
                        }
                      />
                    </Card>
                  </List.Item>
                )}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={getText('暂无智慧内容', 'No wisdom content found')}
              />
            )}
          </TabPane>

          {/* 文化测验 */}
          <TabPane 
            tab={
              <span>
                <QuestionCircleOutlined />
                {getText('文化测验', 'Cultural Quiz')}
              </span>
            } 
            key="quiz"
          >
            {quizzes && quizzes.length > 0 ? (
              <List
                grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3 }}
                dataSource={quizzes}
                renderItem={quiz => (
                  <List.Item>
                    <Card
                      hoverable
                      className="quiz-card"
                      actions={[
                        <Button 
                          type="primary" 
                          icon={<PlayCircleOutlined />}
                          onClick={() => handleStartQuiz(quiz)}
                        >
                          {getText('开始测验', 'Start Quiz')}
                        </Button>
                      ]}
                    >
                      <Card.Meta
                        avatar={<Avatar icon={<QuestionCircleOutlined />} style={{ backgroundColor: '#52c41a' }} />}
                        title={quiz.title}
                        description={
                          <div>
                            <p>{quiz.description}</p>
                            <div className="quiz-info">
                              <span>{getText('题目数量', 'Questions')}: {quiz.questions?.length || 0}</span>
                              <span>{getText('预计时间', 'Duration')}: {quiz.timeLimit || 10} {getText('分钟', 'min')}</span>
                            </div>
                            <div className="quiz-tags">
                              <Tag color="blue">{quiz.category}</Tag>
                              <Tag color="orange">{quiz.level}</Tag>
                            </div>
                          </div>
                        }
                      />
                    </Card>
                  </List.Item>
                )}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={getText('暂无文化测验', 'No cultural quizzes available')}
              />
            )}
          </TabPane>

          {/* 学习历史 */}
          <TabPane 
            tab={
              <span>
                <HistoryOutlined />
                {getText('学习历史', 'Learning History')}
              </span>
            } 
            key="history"
          >
            <div className="learning-history">
              <div className="progress-section">
                <h3>{getText('学习进度', 'Learning Progress')}</h3>
                <div className="progress-items">
                  <div className="progress-item">
                    <span>{getText('道家思想', 'Taoist Philosophy')}</span>
                    <Progress percent={75} strokeColor="#1890ff" />
                  </div>
                  <div className="progress-item">
                    <span>{getText('儒家经典', 'Confucian Classics')}</span>
                    <Progress percent={60} strokeColor="#52c41a" />
                  </div>
                  <div className="progress-item">
                    <span>{getText('佛学智慧', 'Buddhist Wisdom')}</span>
                    <Progress percent={45} strokeColor="#faad14" />
                  </div>
                  <div className="progress-item">
                    <span>{getText('古典文学', 'Classical Literature')}</span>
                    <Progress percent={30} strokeColor="#ff4d4f" />
                  </div>
                </div>
              </div>
            </div>
          </TabPane>
        </Tabs>
      </Card>

      {/* 测验模态框 */}
      <Modal
        title={quizModal.quiz?.title}
        open={quizModal.visible}
        onCancel={handleCloseQuiz}
        footer={null}
        width={700}
        className="quiz-modal"
      >
        {quizModal.quiz && !quizModal.completed && (
          <div className="quiz-content">
            <div className="quiz-progress">
              <Progress 
                percent={((quizModal.currentQuestion + 1) / quizModal.quiz.questions.length) * 100}
                format={() => `${quizModal.currentQuestion + 1}/${quizModal.quiz.questions.length}`}
              />
            </div>
            
            <div className="question-content">
              <h3>{quizModal.quiz.questions[quizModal.currentQuestion]?.question}</h3>
              <div className="question-options">
                {quizModal.quiz.questions[quizModal.currentQuestion]?.options.map((option: string, index: number) => (
                  <Button
                    key={index}
                    type={quizModal.answers[quizModal.currentQuestion] === option ? 'primary' : 'default'}
                    block
                    size="large"
                    onClick={() => handleQuizAnswer(option)}
                    className="option-button"
                  >
                    {option}
                  </Button>
                ))}
              </div>
            </div>
            
            <div className="quiz-actions">
              <Button 
                type="primary" 
                size="large"
                onClick={handleNextQuestion}
                disabled={!quizModal.answers[quizModal.currentQuestion]}
              >
                {quizModal.currentQuestion < quizModal.quiz.questions.length - 1 
                  ? getText('下一题', 'Next Question')
                  : getText('完成测验', 'Complete Quiz')
                }
              </Button>
            </div>
          </div>
        )}
        
        {quizModal.completed && (
          <div className="quiz-result">
            <div className="result-header">
              <TrophyOutlined style={{ fontSize: 48, color: '#faad14' }} />
              <h2>{getText('测验完成！', 'Quiz Completed!')}</h2>
            </div>
            <div className="result-score">
              <Rate disabled value={Math.floor(quizModal.score / 20)} />
              <p>{getText('得分', 'Score')}: {quizModal.score}/100</p>
            </div>
            <div className="result-actions">
              <Button type="primary" size="large" onClick={handleCloseQuiz}>
                {getText('确定', 'OK')}
              </Button>
            </div>
          </div>
        )}
      </Modal>

      <style>{`
        .cultural-page {
          padding: 0;
        }

        .page-header {
          margin-bottom: 24px;
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

        .progress-overview {
          margin-bottom: 24px;
        }

        .stat-item {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .stat-icon {
          font-size: 24px;
        }

        .stat-content {
          flex: 1;
        }

        .stat-number {
          font-size: 20px;
          font-weight: 600;
          color: var(--text-primary);
          line-height: 1;
        }

        .stat-label {
          font-size: 12px;
          color: var(--text-secondary);
          margin-top: 4px;
        }

        .filter-card {
          margin-bottom: 24px;
        }

        .main-content {
          min-height: 600px;
        }

        .wisdom-card {
          height: 100%;
        }

        .wisdom-content {
          color: var(--text-secondary);
          margin-bottom: 12px;
          line-height: 1.5;
        }

        .wisdom-tags {
          display: flex;
          gap: 4px;
          flex-wrap: wrap;
        }

        .quiz-card {
          height: 100%;
        }

        .quiz-info {
          display: flex;
          flex-direction: column;
          gap: 4px;
          margin: 12px 0;
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .quiz-tags {
          display: flex;
          gap: 4px;
          flex-wrap: wrap;
          margin-top: 8px;
        }

        .learning-history {
          padding: 20px 0;
        }

        .progress-section h3 {
          margin-bottom: 20px;
          color: var(--text-primary);
        }

        .progress-items {
          display: flex;
          flex-direction: column;
          gap: 20px;
        }

        .progress-item {
          display: flex;
          align-items: center;
          gap: 16px;
        }

        .progress-item span {
          min-width: 120px;
          font-weight: 500;
          color: var(--text-primary);
        }

        .quiz-modal .quiz-content {
          padding: 20px 0;
        }

        .quiz-progress {
          margin-bottom: 32px;
        }

        .question-content h3 {
          font-size: 18px;
          margin-bottom: 24px;
          color: var(--text-primary);
          line-height: 1.5;
        }

        .question-options {
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-bottom: 32px;
        }

        .option-button {
          text-align: left;
          height: auto;
          padding: 12px 16px;
          white-space: normal;
          word-wrap: break-word;
        }

        .quiz-actions {
          text-align: center;
        }

        .quiz-result {
          text-align: center;
          padding: 40px 20px;
        }

        .result-header h2 {
          margin: 16px 0;
          color: var(--text-primary);
        }

        .result-score {
          margin: 32px 0;
        }

        .result-score p {
          font-size: 18px;
          font-weight: 600;
          color: var(--text-primary);
          margin-top: 16px;
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .page-title {
            font-size: 24px;
          }

          .page-description {
            font-size: 14px;
          }

          .stat-number {
            font-size: 18px;
          }

          .progress-item {
            flex-direction: column;
            align-items: flex-start;
            gap: 8px;
          }

          .progress-item span {
            min-width: auto;
          }
        }
      `}</style>
    </div>
  );
};

export default CulturalPage;