import React, { useState, useEffect } from 'react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { TrendingUp, TrendingDown, Clock, Target, Award, BookOpen } from 'lucide-react';

interface LearningReport {
  learner_id: string;
  report_period: ReportPeriod;
  overall_progress: OverallProgress;
  content_progress: ContentProgressSummary[];
  skill_development: SkillProgress[];
  learning_patterns: LearningPatternAnalysis;
  performance_metrics: PerformanceMetrics;
  recommendations: RecommendationItem[];
  goals: GoalProgress[];
  achievements: Achievement[];
  generated_at: string;
}

interface OverallProgress {
  completion_rate: number;
  total_time_spent: number; // 秒
  content_completed: number;
  skills_acquired: number;
  current_streak: number;
  weekly_goal_progress: number;
  monthly_goal_progress: number;
}

interface ContentProgressSummary {
  content_id: string;
  title: string;
  type: string;
  progress: number;
  time_spent: number; // 秒
  completed_at?: string;
  performance_score: number;
  difficulty: string;
}

interface SkillProgress {
  skill_name: string;
  previous_level: number;
  current_level: number;
  improvement: number;
  last_updated: string;
  related_content: string[];
}

interface LearningPatternAnalysis {
  optimal_study_time: TimeSlotAnalysis[];
  preferred_content_types: Record<string, number>;
  learning_velocity: number;
  retention_rate: number;
  engagement_patterns: EngagementPattern[];
  dropoff_points: DropoffAnalysis[];
}

interface TimeSlotAnalysis {
  hour: number;
  performance_score: number;
  engagement_level: number;
  completion_rate: number;
}

interface EngagementPattern {
  pattern: string;
  frequency: number;
  impact: number;
  description: string;
}

interface DropoffAnalysis {
  content_type: string;
  position: number;
  frequency: number;
  reasons: string[];
}

interface PerformanceMetrics {
  average_score: number;
  improvement_rate: number;
  consistency_score: number;
  efficiency_score: number;
  engagement_score: number;
  retention_score: number;
}

interface RecommendationItem {
  type: string;
  priority: number;
  title: string;
  description: string;
  action_items: string[];
  expected_impact: string;
}

interface GoalProgress {
  goal_id: string;
  description: string;
  target_date: string;
  current_progress: number;
  is_on_track: boolean;
  days_remaining: number;
  recommendations: string[];
}

interface Achievement {
  id: string;
  type: string;
  title: string;
  description: string;
  points: number;
  unlocked_at: string;
}

interface ReportPeriod {
  start_date: string;
  end_date: string;
  type: string;
}

interface ProgressVisualizationProps {
  learnerId: string;
  reportPeriod?: ReportPeriod;
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

const ProgressVisualization: React.FC<ProgressVisualizationProps> = ({
  learnerId,
  reportPeriod
}) => {
  const [report, setReport] = useState<LearningReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchLearningReport();
  }, [learnerId, reportPeriod]);

  const fetchLearningReport = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        learner_id: learnerId,
        ...(reportPeriod && {
          start_date: reportPeriod.start_date,
          end_date: reportPeriod.end_date,
          type: reportPeriod.type
        })
      });

      const response = await fetch(`/api/v1/progress/report?${params}`);
      if (!response.ok) {
        throw new Error('Failed to fetch learning report');
      }

      const data = await response.json();
      setReport(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatPercentage = (value: number): string => {
    return `${(value * 100).toFixed(1)}%`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!report) {
    return (
      <Alert>
        <AlertDescription>No learning report data available.</AlertDescription>
      </Alert>
    );
  }

  const { overall_progress, content_progress, skill_development, learning_patterns, performance_metrics, recommendations, goals, achievements } = report;

  // 准备图表数据
  const studyTimeData = learning_patterns.optimal_study_time.map(slot => ({
    hour: `${slot.hour}:00`,
    performance: slot.performance_score * 100,
    engagement: slot.engagement_level * 100,
    completion: slot.completion_rate * 100
  }));

  const contentTypeData = Object.entries(learning_patterns.preferred_content_types).map(([type, preference]) => ({
    name: type,
    value: preference * 100
  }));

  const skillData = skill_development.map(skill => ({
    skill: skill.skill_name,
    previous: skill.previous_level * 100,
    current: skill.current_level * 100,
    improvement: skill.improvement * 100
  }));

  const performanceRadarData = [
    { metric: '平均分数', value: performance_metrics.average_score * 100 },
    { metric: '改进率', value: Math.max(0, performance_metrics.improvement_rate * 100) },
    { metric: '一致性', value: performance_metrics.consistency_score * 100 },
    { metric: '效率', value: performance_metrics.efficiency_score * 100 },
    { metric: '参与度', value: performance_metrics.engagement_score * 100 },
    { metric: '保持率', value: performance_metrics.retention_score * 100 }
  ];

  return (
    <div className="space-y-6">
      {/* 总体进度概览 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center space-x-2">
              <BookOpen className="h-5 w-5 text-blue-600" />
              <div>
                <p className="text-sm font-medium text-gray-600">完成率</p>
                <p className="text-2xl font-bold">{formatPercentage(overall_progress.completion_rate)}</p>
              </div>
            </div>
            <Progress value={overall_progress.completion_rate * 100} className="mt-2" />
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center space-x-2">
              <Clock className="h-5 w-5 text-green-600" />
              <div>
                <p className="text-sm font-medium text-gray-600">学习时长</p>
                <p className="text-2xl font-bold">{formatDuration(overall_progress.total_time_spent)}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center space-x-2">
              <Target className="h-5 w-5 text-orange-600" />
              <div>
                <p className="text-sm font-medium text-gray-600">学习连续性</p>
                <p className="text-2xl font-bold">{overall_progress.current_streak} 天</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center space-x-2">
              <Award className="h-5 w-5 text-purple-600" />
              <div>
                <p className="text-sm font-medium text-gray-600">技能获得</p>
                <p className="text-2xl font-bold">{overall_progress.skills_acquired}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="patterns" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="patterns">学习模式</TabsTrigger>
          <TabsTrigger value="performance">性能分析</TabsTrigger>
          <TabsTrigger value="skills">技能发展</TabsTrigger>
          <TabsTrigger value="goals">目标进度</TabsTrigger>
          <TabsTrigger value="recommendations">建议</TabsTrigger>
        </TabsList>

        <TabsContent value="patterns" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* 最佳学习时间 */}
            <Card>
              <CardHeader>
                <CardTitle>最佳学习时间分析</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={studyTimeData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="hour" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="performance" stroke="#8884d8" name="性能分数" />
                    <Line type="monotone" dataKey="engagement" stroke="#82ca9d" name="参与度" />
                    <Line type="monotone" dataKey="completion" stroke="#ffc658" name="完成率" />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* 内容类型偏好 */}
            <Card>
              <CardHeader>
                <CardTitle>内容类型偏好</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={contentTypeData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, percent }) => `${name} ${(percent).toFixed(0)}%`}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {contentTypeData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </div>

          {/* 参与模式 */}
          <Card>
            <CardHeader>
              <CardTitle>学习参与模式</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {learning_patterns.engagement_patterns.map((pattern, index) => (
                  <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="font-medium">{pattern.description}</p>
                      <p className="text-sm text-gray-600">频率: {formatPercentage(pattern.frequency)}</p>
                    </div>
                    <Badge variant={pattern.impact > 0.7 ? "default" : pattern.impact > 0.4 ? "secondary" : "outline"}>
                      影响: {formatPercentage(pattern.impact)}
                    </Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="performance" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* 性能雷达图 */}
            <Card>
              <CardHeader>
                <CardTitle>综合性能分析</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <RadarChart data={performanceRadarData}>
                    <PolarGrid />
                    <PolarAngleAxis dataKey="metric" />
                    <PolarRadiusAxis angle={90} domain={[0, 100]} />
                    <Radar
                      name="性能指标"
                      dataKey="value"
                      stroke="#8884d8"
                      fill="#8884d8"
                      fillOpacity={0.6}
                    />
                    <Tooltip />
                  </RadarChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* 性能指标详情 */}
            <Card>
              <CardHeader>
                <CardTitle>性能指标详情</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span>平均分数</span>
                    <span className="font-bold">{formatPercentage(performance_metrics.average_score)}</span>
                  </div>
                  <Progress value={performance_metrics.average_score * 100} />
                </div>

                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span>改进率</span>
                    <div className="flex items-center space-x-1">
                      {performance_metrics.improvement_rate > 0 ? (
                        <TrendingUp className="h-4 w-4 text-green-600" />
                      ) : (
                        <TrendingDown className="h-4 w-4 text-red-600" />
                      )}
                      <span className="font-bold">
                        {performance_metrics.improvement_rate > 0 ? '+' : ''}
                        {formatPercentage(performance_metrics.improvement_rate)}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span>学习效率</span>
                    <span className="font-bold">{formatPercentage(performance_metrics.efficiency_score)}</span>
                  </div>
                  <Progress value={performance_metrics.efficiency_score * 100} />
                </div>

                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span>参与度</span>
                    <span className="font-bold">{formatPercentage(performance_metrics.engagement_score)}</span>
                  </div>
                  <Progress value={performance_metrics.engagement_score * 100} />
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="skills" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>技能发展进度</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={400}>
                <BarChart data={skillData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="skill" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="previous" fill="#8884d8" name="之前水平" />
                  <Bar dataKey="current" fill="#82ca9d" name="当前水平" />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="goals" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {goals.map((goal) => (
              <Card key={goal.goal_id}>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span className="truncate">{goal.description}</span>
                    <Badge variant={goal.is_on_track ? "default" : "destructive"}>
                      {goal.is_on_track ? "按计划进行" : "需要关注"}
                    </Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>进度</span>
                      <span>{formatPercentage(goal.current_progress)}</span>
                    </div>
                    <Progress value={goal.current_progress * 100} />
                  </div>

                  <div className="text-sm text-gray-600">
                    <p>剩余天数: {goal.days_remaining} 天</p>
                    <p>目标日期: {new Date(goal.target_date).toLocaleDateString()}</p>
                  </div>

                  {goal.recommendations.length > 0 && (
                    <div>
                      <p className="text-sm font-medium mb-2">建议:</p>
                      <ul className="text-sm text-gray-600 space-y-1">
                        {goal.recommendations.map((rec, index) => (
                          <li key={index} className="flex items-start">
                            <span className="mr-2">•</span>
                            <span>{rec}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="recommendations" className="space-y-4">
          <div className="space-y-4">
            {recommendations
              .sort((a, b) => a.priority - b.priority)
              .map((rec, index) => (
                <Card key={index}>
                  <CardHeader>
                    <CardTitle className="flex items-center justify-between">
                      <span>{rec.title}</span>
                      <Badge variant={rec.priority === 1 ? "destructive" : rec.priority === 2 ? "default" : "secondary"}>
                        优先级 {rec.priority}
                      </Badge>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <p className="text-gray-600">{rec.description}</p>
                    
                    <div>
                      <p className="font-medium mb-2">行动建议:</p>
                      <ul className="space-y-1">
                        {rec.action_items.map((item, itemIndex) => (
                          <li key={itemIndex} className="flex items-start text-sm">
                            <span className="mr-2 text-blue-600">✓</span>
                            <span>{item}</span>
                          </li>
                        ))}
                      </ul>
                    </div>

                    <div className="bg-blue-50 p-3 rounded-lg">
                      <p className="text-sm">
                        <span className="font-medium">预期效果:</span> {rec.expected_impact}
                      </p>
                    </div>
                  </CardContent>
                </Card>
              ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default ProgressVisualization;