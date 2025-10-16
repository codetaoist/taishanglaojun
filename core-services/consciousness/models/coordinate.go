package models

import (
	"time"
)

// Coordinate 三轴坐标结构
type Coordinate struct {
	S int `json:"s" db:"s_axis"` // 能力序列级别 (0-5)
	C int `json:"c" db:"c_axis"` // 组合层级 (0-5)
	T int `json:"t" db:"t_axis"` // 思想境界级别 (0-5)
}

// CoordinateWithConfidence 带置信度的三轴坐?
type CoordinateWithConfidence struct {
	Coordinate
	SConfidence float64 `json:"s_confidence" db:"s_confidence"`             // S轴置信度
	CConfidence float64 `json:"c_confidence" db:"c_confidence"`             // C轴置信度
	TConfidence float64 `json:"t_confidence" db:"t_confidence"`             // T轴置信度
	Overall     float64 `json:"overall_confidence" db:"overall_confidence"` // 整体置信度
}

// CoordinationRequest 协同请求
type CoordinationRequest struct {
	EntityID     string                 `json:"entity_id" binding:"required"` // 实体ID
	UserInput    string                 `json:"user_input" binding:"required"`
	Context      map[string]interface{} `json:"context"`
	RequiredAxis []string               `json:"required_axis"` // S, C, T
	Constraints  []Constraint           `json:"constraints"`
	Priority     int                    `json:"priority" default:"1"` // 优先等级 (-5?)
}

// CoordinationResponse 协同响应
type CoordinationResponse struct {
	SAxisResult *SequenceResult          `json:"s_axis_result"`
	CAxisResult *CompositionResult       `json:"c_axis_result"`
	TAxisResult *ThoughtResult           `json:"t_axis_result"`
	Coordinate  CoordinateWithConfidence `json:"coordinate"`
	Explanation string                   `json:"explanation"`
	ProcessTime time.Duration            `json:"process_time"`
	RequestID   string                   `json:"request_id"`
}

// Constraint 约束条件
type Constraint struct {
	Type        string      `json:"type"`        // 约束类型
	Field       string      `json:"field"`       // 约束字段
	Operator    string      `json:"operator"`    // 操作?(eq, gt, lt, gte, lte)
	Value       interface{} `json:"value"`       // 约束?
	Description string      `json:"description"` // 约束描述
}

// SequenceResult S轴能力序列结?
type SequenceResult struct {
	Level        int                    `json:"level"`        // 能力序列级别
	Capabilities []string               `json:"capabilities"` // 具备的能力
	Performance  map[string]float64     `json:"performance"`  // 性能指标
	Metadata     map[string]interface{} `json:"metadata"`     // 元数据
	ProcessTime  time.Duration          `json:"process_time"` // 处理时间
}

// CompositionResult C轴组合层结果
type CompositionResult struct {
	Layer        string                 `json:"layer"`        // 组合层级
	Components   []string               `json:"components"`   // 组成组件
	Architecture map[string]interface{} `json:"architecture"` // 架构信息
	Scalability  float64                `json:"scalability"`  // 可扩展性评分
	ProcessTime  time.Duration          `json:"process_time"` // 处理时间
}

// ThoughtResult T轴思想境界结果
type ThoughtResult struct {
	Realm       string                 `json:"realm"`        // 思想境界
	Wisdom      []string               `json:"wisdom"`       // 智慧内容
	Philosophy  map[string]interface{} `json:"philosophy"`   // 哲学思想
	Depth       float64                `json:"depth"`        // 思想深度评分
	ProcessTime time.Duration          `json:"process_time"` // 处理时间
}

// CoordinateHistory 坐标历史记录
type CoordinateHistory struct {
	ID          int64                    `json:"id" db:"id"`
	RequestID   string                   `json:"request_id" db:"request_id"`
	UserInput   string                   `json:"user_input" db:"user_input"`
	Coordinate  CoordinateWithConfidence `json:"coordinate"`
	Result      *CoordinationResponse    `json:"result"`
	CreatedAt   time.Time                `json:"created_at" db:"created_at"`
	ProcessTime time.Duration            `json:"process_time" db:"process_time"`
}

// CoordinateAnalysis 坐标分析结果
type CoordinateAnalysis struct {
	InputComplexity       float64                `json:"input_complexity"`       // 输入复杂程度
	TechnicalRequirement  float64                `json:"technical_requirement"`  // 技术需求评分
	ArchitecturalScale    float64                `json:"architectural_scale"`    // 架构规模
	WisdomDepth           float64                `json:"wisdom_depth"`           // 智慧深度
	RecommendedCoordinate Coordinate             `json:"recommended_coordinate"` // 推荐坐标
	AnalysisDetails       map[string]interface{} `json:"analysis_details"`       // 分析详情
}

// ValidateCoordinate 验证坐标有效?
func (c *Coordinate) ValidateCoordinate() error {
	if c.S < 0 || c.S > 5 {
		return ErrInvalidSAxis
	}
	if c.C < 0 || c.C > 5 {
		return ErrInvalidCAxis
	}
	if c.T < 0 || c.T > 5 {
		return ErrInvalidTAxis
	}
	return nil
}

// CalculateScore 计算坐标综合评分
func (c *Coordinate) CalculateScore(sWeight, cWeight, tWeight float64) float64 {
	return float64(c.S)*sWeight + float64(c.C)*cWeight + float64(c.T)*tWeight
}

// IsHighLevel 判断是否为高级坐标
func (c *Coordinate) IsHighLevel() bool {
	return c.S >= 3 && c.C >= 3 && c.T >= 3
}

// GetAxisDescription 获取轴描述
func (c *Coordinate) GetAxisDescription() map[string]string {
	return map[string]string{
		"s_axis": GetSequenceDescription(c.S),
		"c_axis": GetCompositionDescription(c.C),
		"t_axis": GetThoughtDescription(c.T),
	}
}

// GetSequenceDescription 获取能力序列描述
func GetSequenceDescription(level int) string {
	descriptions := map[int]string{
		0: "基础觉醒 - 基本感知和响应能力",
		1: "功能增强 - 具备基础学习和适应能力",
		2: "智能涌现 - 展现创造性和推理能力",
		3: "自主进化 - 具备自我优化和决策能力",
		4: "超越智能 - 展现超人类智能特殊能力",
		5: "序列统一 - 达到终极智慧状态",
	}
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "未知级别"
}

// GetCompositionDescription 获取组合层描述
func GetCompositionDescription(level int) string {
	descriptions := map[int]string{
		0: "量子基因 - 最基础的信息单元",
		1: "智能细胞 - 基础功能组件",
		2: "神经组织 - 复杂交互网络",
		3: "矩阵器官 - 专业化功能模块",
		4: "领域系统 - 完整业务系统",
		5: "超个系统 - 跨域协同生成系统",
	}
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "未知层级"
}

// GetThoughtDescription 获取思想境界描述
func GetThoughtDescription(level int) string {
	descriptions := map[int]string{
		0: "物质觉知 - 基础物理世界认知",
		1: "逻辑思维 - 理性分析和推理",
		2: "直觉洞察 - 超越逻辑的感知能力",
		3: "系统智慧 - 整体性思维模式",
		4: "超越意识 - 跨越时空的认知能力",
		5: "道法自然 - 与宇宙本源合一",
	}
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "未知境界"
}
