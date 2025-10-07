import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Progress } from '@/components/ui/progress';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell
} from 'recharts';
import { 
  Activity, 
  Brain, 
  TrendingUp, 
  AlertTriangle, 
  Target, 
  Clock,
  Zap,
  Eye,
  Heart,
  BookOpen
} from 'lucide-react';

interface RealtimeMetrics {
  learning_velocity: number;
  engagement_score: number;
  performance_score: number;
  focus_score: number;
  efficiency_score: number;
  motivation_level: number;
  cognitive_load: number;
  engagement_trend: string;
  completion_probability: number;
  dropout_risk: number;
  last_updated: string;
}

interface BehaviorPatterns {
  learning_rhythm: string;
  attention_span: number;
  interaction_frequency: Record<string, number>;
  preferred_content_types: string[];
}

interface EngagementState {
  level: string;
  score: number;
  trend: string;
  interaction_quality: number;
  risk_factors: string[];
}

interface PerformanceState {
  current_level: string;
  score: number;
  trend: string;
  strength_areas: string[];
  improvement_areas: string[];
  recent_achievements: string[];
}

interface PredictiveInsights {
  completion_probability: number;
  estimated_completion_time: string;
  risk_of_dropout: number;
  recommended_actions: string[];
  predicted_challenges: string[];
}

interface AlertItem {
  level: string;
  message: string;
  timestamp: string;
  action_items: string[];
}

interface LearningSession {
  session_id: string;
  start_time: string;
  duration: number;
  interaction_count: number;
  progress_made: number;
  engagement_score: number;
  focus_level: number;
  content_items: string[];
  last_activity: string;
}

interface RealtimeAnalyticsProps {
  learnerId: string;
}

const RealtimeAnalytics: React.FC<RealtimeAnalyticsProps> = ({ learnerId }) => {
  const [metrics, setMetrics] = useState<RealtimeMetrics | null>(null);
  const [insights, setInsights] = useState<any>(null);
  const [session, setSession] = useState<LearningSession | null>(null);
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [recommendations, setRecommendations] = useState<any>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [loading, setLoading] = useState(true);

  // 获取实时指标
  const fetchMetrics = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/realtime-analytics/${learnerId}/metrics`);
      if (response.ok) {
        const data = await response.json();
        setMetrics(data.metrics);
      }
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    }
  }, [learnerId]);

  // 获取学习洞察
  const fetchInsights = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/realtime-analytics/${learnerId}/insights`);
      if (response.ok) {
        const data = await response.json();
        setInsights(data.insights);
      }
    } catch (error) {
      console.error('Failed to fetch insights:', error);
    }
  }, [learnerId]);

  // 获取会话数据
  const fetchSession = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/realtime-analytics/${learnerId}/session`);
      if (response.ok) {
        const data = await response.json();
        setSession(data.session);
      }
    } catch (error) {
      console.error('Failed to fetch session:', error);
    }
  }, [learnerId]);

  // 获取警报
  const fetchAlerts = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/realtime-analytics/${learnerId}/alerts`);
      if (response.ok) {
        const data = await response.json();
        setAlerts(data.alerts.all_alerts || []);
      }
    } catch (error) {
      console.error('Failed to fetch alerts:', error);
    }
  }, [learnerId]);

  // 获取推荐
  const fetchRecommendations = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/realtime-analytics/${learnerId}/recommendations`);
      if (response.ok) {
        const data = await response.json();
        setRecommendations(data.recommendations);
      }
    } catch (error) {
      console.error('Failed to fetch recommendations:', error);
    }
  }, [learnerId]);

  // WebSocket连接
  useEffect(() => {
    const ws = new WebSocket(`ws://localhost:8080/api/v1/realtime-analytics/subscribe?subscriber_id=${learnerId}`);
    
    ws.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      const update = JSON.parse(event.data);
      console.log('Received update:', update);
      
      // 根据更新类型刷新相应数据
      if (update.type === 'realtime_analysis') {
        fetchMetrics();
        fetchInsights();
      } else if (update.type === 'prediction_update') {
        fetchRecommendations();
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, [learnerId, fetchMetrics, fetchInsights, fetchRecommendations]);

  // 初始数据加载
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      await Promise.all([
        fetchMetrics(),
        fetchInsights(),
        fetchSession(),
        fetchAlerts(),
        fetchRecommendations()
      ]);
      setLoading(false);
    };

    loadData();
  }, [fetchMetrics, fetchInsights, fetchSession, fetchAlerts, fetchRecommendations]);

  // 定期刷新数据
  useEffect(() => {
    const interval = setInterval(() => {
      fetchMetrics();
      fetchAlerts();
    }, 30000); // 每30秒刷新一次

    return () => clearInterval(interval);
  }, [fetchMetrics, fetchAlerts]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">加载实时分析数据...</p>
        </div>
      </div>
    );
  }

  const getEngagementColor = (score: number) => {
    if (score >= 0.8) return 'text-green-600';
    if (score >= 0.6) return 'text-yellow-600';
    if (score >= 0.3) return 'text-orange-600';
    return 'text-red-600';
  };

  const getPerformanceColor = (level: string) => {
    switch (level) {
      case 'excellent': return 'text-green-600';
      case 'good': return 'text-blue-600';
      case 'average': return 'text-yellow-600';
      case 'below_average': return 'text-orange-600';
      case 'poor': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  const getAlertColor = (level: string) => {
    switch (level) {
      case 'critical': return 'destructive';
      case 'warning': return 'default';
      case 'info': return 'secondary';
      default: return 'default';
    }
  };

  // 雷达图数据
  const radarData = metrics ? [
    { subject: '参与度', A: metrics.engagement_score * 100, fullMark: 100 },
    { subject: '表现', A: metrics.performance_score * 100, fullMark: 100 },
    { subject: '专注度', A: metrics.focus_score * 100, fullMark: 100 },
    { subject: '效率', A: metrics.efficiency_score * 100, fullMark: 100 },
    { subject: '动机', A: metrics.motivation_level * 100, fullMark: 100 },
    { subject: '学习速度', A: Math.min(metrics.learning_velocity * 50, 100), fullMark: 100 },
  ] : [];

  // 交互频率数据
  const interactionData = insights?.behavior_patterns?.interaction_frequency ? 
    Object.entries(insights.behavior_patterns.interaction_frequency).map(([type, count]) => ({
      type,
      count: count as number
    })) : [];

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

  return (
    <div className="space-y-6">
      {/* 连接状态 */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">实时学习分析</h2>
        <div className="flex items-center space-x-2">
          <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className="text-sm text-gray-600">
            {isConnected ? '实时连接' : '连接断开'}
          </span>
        </div>
      </div>

      {/* 警报 */}
      {alerts.length > 0 && (
        <div className="space-y-2">
          {alerts.map((alert, index) => (
            <Alert key={index} variant={getAlertColor(alert.level)}>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>{alert.message}</AlertTitle>
              <AlertDescription>
                {alert.action_items.length > 0 && (
                  <ul className="mt-2 list-disc list-inside">
                    {alert.action_items.map((item, i) => (
                      <li key={i}>{item}</li>
                    ))}
                  </ul>
                )}
              </AlertDescription>
            </Alert>
          ))}
        </div>
      )}

      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="overview">概览</TabsTrigger>
          <TabsTrigger value="performance">表现</TabsTrigger>
          <TabsTrigger value="behavior">行为</TabsTrigger>
          <TabsTrigger value="session">会话</TabsTrigger>
          <TabsTrigger value="predictions">预测</TabsTrigger>
        </TabsList>

        {/* 概览标签页 */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {metrics && (
              <>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">参与度</CardTitle>
                    <Heart className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className={`text-2xl font-bold ${getEngagementColor(metrics.engagement_score)}`}>
                      {(metrics.engagement_score * 100).toFixed(1)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      趋势: {metrics.engagement_trend}
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">学习表现</CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {(metrics.performance_score * 100).toFixed(1)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      等级: {insights?.performance_state?.current_level}
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">学习速度</CardTitle>
                    <Zap className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {metrics.learning_velocity.toFixed(2)}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      单位/小时
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">专注度</CardTitle>
                    <Eye className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {(metrics.focus_score * 100).toFixed(1)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      认知负荷: {(metrics.cognitive_load * 100).toFixed(1)}%
                    </p>
                  </CardContent>
                </Card>
              </>
            )}
          </div>

          {/* 综合雷达图 */}
          <Card>
            <CardHeader>
              <CardTitle>学习能力雷达图</CardTitle>
              <CardDescription>多维度学习能力评估</CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <RadarChart data={radarData}>
                  <PolarGrid />
                  <PolarAngleAxis dataKey="subject" />
                  <PolarRadiusAxis angle={90} domain={[0, 100]} />
                  <Radar
                    name="当前水平"
                    dataKey="A"
                    stroke="#8884d8"
                    fill="#8884d8"
                    fillOpacity={0.6}
                  />
                  <Tooltip />
                </RadarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 表现标签页 */}
        <TabsContent value="performance" className="space-y-6">
          {insights?.performance_state && (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>表现状态</CardTitle>
                  <CardDescription>当前学习表现分析</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <span>当前等级:</span>
                    <Badge className={getPerformanceColor(insights.performance_state.current_level)}>
                      {insights.performance_state.current_level}
                    </Badge>
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span>表现分数</span>
                      <span>{(insights.performance_state.score * 100).toFixed(1)}%</span>
                    </div>
                    <Progress value={insights.performance_state.score * 100} />
                  </div>
                  <div className="space-y-2">
                    <span className="font-medium">优势领域:</span>
                    <div className="flex flex-wrap gap-2">
                      {insights.performance_state.strength_areas?.map((area: string, index: number) => (
                        <Badge key={index} variant="secondary">{area}</Badge>
                      ))}
                    </div>
                  </div>
                  <div className="space-y-2">
                    <span className="font-medium">改进领域:</span>
                    <div className="flex flex-wrap gap-2">
                      {insights.performance_state.improvement_areas?.map((area: string, index: number) => (
                        <Badge key={index} variant="outline">{area}</Badge>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>最近成就</CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-2">
                    {insights.performance_state.recent_achievements?.map((achievement: string, index: number) => (
                      <li key={index} className="flex items-center space-x-2">
                        <Target className="h-4 w-4 text-green-600" />
                        <span>{achievement}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            </>
          )}
        </TabsContent>

        {/* 行为标签页 */}
        <TabsContent value="behavior" className="space-y-6">
          {insights?.behavior_patterns && (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <CardTitle>学习模式</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex justify-between">
                      <span>学习节奏:</span>
                      <Badge>{insights.behavior_patterns.learning_rhythm}</Badge>
                    </div>
                    <div className="flex justify-between">
                      <span>注意力持续时间:</span>
                      <span>{Math.round(insights.behavior_patterns.attention_span / 60000)} 分钟</span>
                    </div>
                    <div className="space-y-2">
                      <span className="font-medium">偏好内容类型:</span>
                      <div className="flex flex-wrap gap-2">
                        {insights.behavior_patterns.preferred_content_types?.map((type: string, index: number) => (
                          <Badge key={index} variant="secondary">{type}</Badge>
                        ))}
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>交互频率</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ResponsiveContainer width="100%" height={200}>
                      <PieChart>
                        <Pie
                          data={interactionData}
                          cx="50%"
                          cy="50%"
                          labelLine={false}
                          label={({ type, percent }) => `${type} ${(percent * 100).toFixed(0)}%`}
                          outerRadius={80}
                          fill="#8884d8"
                          dataKey="count"
                        >
                          {interactionData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip />
                      </PieChart>
                    </ResponsiveContainer>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>参与状态</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex justify-between">
                    <span>参与等级:</span>
                    <Badge className={getEngagementColor(insights.engagement_state.score)}>
                      {insights.engagement_state.level}
                    </Badge>
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span>交互质量</span>
                      <span>{(insights.engagement_state.interaction_quality * 100).toFixed(1)}%</span>
                    </div>
                    <Progress value={insights.engagement_state.interaction_quality * 100} />
                  </div>
                  {insights.engagement_state.risk_factors?.length > 0 && (
                    <div className="space-y-2">
                      <span className="font-medium text-orange-600">风险因素:</span>
                      <ul className="list-disc list-inside space-y-1">
                        {insights.engagement_state.risk_factors.map((factor: string, index: number) => (
                          <li key={index} className="text-sm">{factor}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </CardContent>
              </Card>
            </>
          )}
        </TabsContent>

        {/* 会话标签页 */}
        <TabsContent value="session" className="space-y-6">
          {session && (
            <Card>
              <CardHeader>
                <CardTitle>当前学习会话</CardTitle>
                <CardDescription>会话ID: {session.session_id}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <span className="text-sm text-gray-600">开始时间:</span>
                    <p>{new Date(session.start_time).toLocaleString()}</p>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">持续时间:</span>
                    <p>{Math.round(session.duration / 60)} 分钟</p>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">交互次数:</span>
                    <p>{session.interaction_count}</p>
                  </div>
                  <div>
                    <span className="text-sm text-gray-600">进度:</span>
                    <p>{(session.progress_made * 100).toFixed(1)}%</p>
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span>参与度分数</span>
                    <span>{(session.engagement_score * 100).toFixed(1)}%</span>
                  </div>
                  <Progress value={session.engagement_score * 100} />
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span>专注水平</span>
                    <span>{(session.focus_level * 100).toFixed(1)}%</span>
                  </div>
                  <Progress value={session.focus_level * 100} />
                </div>
                <div>
                  <span className="text-sm text-gray-600">学习内容数量:</span>
                  <p>{session.content_items.length} 项</p>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* 预测标签页 */}
        <TabsContent value="predictions" className="space-y-6">
          {insights?.predictive_insights && recommendations && (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <CardTitle>预测洞察</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span>完成概率</span>
                        <span className="font-bold text-green-600">
                          {(insights.predictive_insights.completion_probability * 100).toFixed(1)}%
                        </span>
                      </div>
                      <Progress value={insights.predictive_insights.completion_probability * 100} />
                    </div>
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span>辍学风险</span>
                        <span className="font-bold text-red-600">
                          {(insights.predictive_insights.risk_of_dropout * 100).toFixed(1)}%
                        </span>
                      </div>
                      <Progress value={insights.predictive_insights.risk_of_dropout * 100} />
                    </div>
                    <div>
                      <span className="text-sm text-gray-600">预计完成时间:</span>
                      <p>{new Date(insights.predictive_insights.estimated_completion_time).toLocaleDateString()}</p>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>学习策略建议</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <span className="font-medium">基于学习节奏:</span>
                      <p className="text-sm text-gray-600">{recommendations.learning_strategy?.based_on_rhythm}</p>
                    </div>
                    <div>
                      <span className="font-medium">注意力持续时间:</span>
                      <p className="text-sm text-gray-600">
                        {Math.round(recommendations.learning_strategy?.attention_span / 60000)} 分钟
                      </p>
                    </div>
                    <div>
                      <span className="font-medium">推荐内容类型:</span>
                      <div className="flex flex-wrap gap-2 mt-1">
                        {recommendations.learning_strategy?.preferred_content?.map((type: string, index: number) => (
                          <Badge key={index} variant="outline">{type}</Badge>
                        ))}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>即时行动建议</CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-2">
                    {recommendations.immediate_actions?.map((action: string, index: number) => (
                      <li key={index} className="flex items-center space-x-2">
                        <Activity className="h-4 w-4 text-blue-600" />
                        <span>{action}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>

              {insights.predictive_insights.predicted_challenges?.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>预测挑战</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ul className="space-y-2">
                      {insights.predictive_insights.predicted_challenges.map((challenge: string, index: number) => (
                        <li key={index} className="flex items-center space-x-2">
                          <AlertTriangle className="h-4 w-4 text-orange-600" />
                          <span>{challenge}</span>
                        </li>
                      ))}
                    </ul>
                  </CardContent>
                </Card>
              )}
            </>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default RealtimeAnalytics;