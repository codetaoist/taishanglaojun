package vision

import (
	"context"
	"encoding/json"
	"image"
	"time"

	"github.com/google/uuid"
)

// ImageService 图像服务接口
type ImageService interface {
	// 图像识别
	RecognizeObjects(ctx context.Context, input ImageInput) (*ObjectRecognitionResult, error)
	RecognizeFaces(ctx context.Context, input ImageInput) (*FaceRecognitionResult, error)
	RecognizeText(ctx context.Context, input ImageInput) (*TextRecognitionResult, error)
	RecognizeScene(ctx context.Context, input ImageInput) (*SceneRecognitionResult, error)
	
	// 图像分析
	AnalyzeImage(ctx context.Context, input ImageInput) (*ImageAnalysisResult, error)
	DetectAnomalies(ctx context.Context, input ImageInput) (*AnomalyDetectionResult, error)
	CompareImages(ctx context.Context, image1, image2 ImageInput) (*ImageComparisonResult, error)
	
	// 图像处理
	ProcessImage(ctx context.Context, input ImageInput, operations []ImageOperation) (*ImageProcessingResult, error)
	EnhanceImage(ctx context.Context, input ImageInput, options EnhancementOptions) (*ImageProcessingResult, error)
	
	// 批量处理
	BatchProcess(ctx context.Context, inputs []ImageInput, operations []ImageOperation) (*BatchProcessingResult, error)
	
	// 配置管理
	UpdateConfig(config ImageConfig) error
	GetSupportedFormats() []ImageFormat
	GetSupportedOperations() []OperationType
}

// ImageFormat 图像格式
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
	FormatGIF  ImageFormat = "gif"
	FormatBMP  ImageFormat = "bmp"
	FormatTIFF ImageFormat = "tiff"
	FormatWEBP ImageFormat = "webp"
	FormatSVG  ImageFormat = "svg"
	FormatRAW  ImageFormat = "raw"
)

// ImageInput 图像输入
type ImageInput struct {
	ID          string                 `json:"id"`
	Data        []byte                 `json:"data"`
	Format      ImageFormat            `json:"format"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Size        int64                  `json:"size"`
	URL         string                 `json:"url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	Source      string                 `json:"source"`
	Quality     float64                `json:"quality"`
	ColorSpace  string                 `json:"color_space"`
	Orientation int                    `json:"orientation"`
}

// ObjectRecognitionResult 物体识别结果
type ObjectRecognitionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Objects        []DetectedObject       `json:"objects"`
	TotalObjects   int                    `json:"total_objects"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// DetectedObject 检测到的物体
type DetectedObject struct {
	ID         string      `json:"id"`
	Label      string      `json:"label"`
	Confidence float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Category   string      `json:"category"`
	Attributes map[string]interface{} `json:"attributes"`
	SubObjects []DetectedObject       `json:"sub_objects,omitempty"`
}

// BoundingBox 边界框
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// FaceRecognitionResult 人脸识别结果
type FaceRecognitionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Faces          []DetectedFace         `json:"faces"`
	TotalFaces     int                    `json:"total_faces"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// DetectedFace 检测到的人脸
type DetectedFace struct {
	ID          string                 `json:"id"`
	BoundingBox BoundingBox            `json:"bounding_box"`
	Confidence  float64                `json:"confidence"`
	Landmarks   []FaceLandmark         `json:"landmarks"`
	Attributes  FaceAttributes         `json:"attributes"`
	Emotions    map[string]float64     `json:"emotions"`
	Identity    *FaceIdentity          `json:"identity,omitempty"`
	Encoding    []float64              `json:"encoding,omitempty"`
}

// FaceLandmark 人脸关键点
type FaceLandmark struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// FaceAttributes 人脸属性
type FaceAttributes struct {
	Age        *AgeRange `json:"age,omitempty"`
	Gender     string    `json:"gender"`
	Ethnicity  string    `json:"ethnicity"`
	Glasses    string    `json:"glasses"`
	Beard      bool      `json:"beard"`
	Mustache   bool      `json:"mustache"`
	EyesOpen   bool      `json:"eyes_open"`
	MouthOpen  bool      `json:"mouth_open"`
	Smiling    bool      `json:"smiling"`
	HeadPose   HeadPose  `json:"head_pose"`
}

// AgeRange 年龄范围
type AgeRange struct {
	Low  int `json:"low"`
	High int `json:"high"`
}

// HeadPose 头部姿态
type HeadPose struct {
	Pitch float64 `json:"pitch"`
	Roll  float64 `json:"roll"`
	Yaw   float64 `json:"yaw"`
}

// FaceIdentity 人脸身份
type FaceIdentity struct {
	PersonID   string  `json:"person_id"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// TextRecognitionResult 文本识别结果
type TextRecognitionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Text           string                 `json:"text"`
	TextBlocks     []TextBlock            `json:"text_blocks"`
	Language       string                 `json:"language"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// TextBlock 文本块
type TextBlock struct {
	ID          string      `json:"id"`
	Text        string      `json:"text"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Confidence  float64     `json:"confidence"`
	Language    string      `json:"language"`
	Words       []Word      `json:"words"`
	Lines       []Line      `json:"lines"`
}

// Word 单词
type Word struct {
	Text        string      `json:"text"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Confidence  float64     `json:"confidence"`
}

// Line 行
type Line struct {
	Text        string      `json:"text"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Confidence  float64     `json:"confidence"`
	Words       []Word      `json:"words"`
}

// SceneRecognitionResult 场景识别结果
type SceneRecognitionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Scene          string                 `json:"scene"`
	Confidence     float64                `json:"confidence"`
	Categories     []SceneCategory        `json:"categories"`
	Tags           []string               `json:"tags"`
	Description    string                 `json:"description"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// SceneCategory 场景类别
type SceneCategory struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Parent     string  `json:"parent,omitempty"`
}

// ImageAnalysisResult 图像分析结果
type ImageAnalysisResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Quality        ImageQuality           `json:"quality"`
	Colors         ColorAnalysis          `json:"colors"`
	Composition    CompositionAnalysis    `json:"composition"`
	Content        ContentAnalysis        `json:"content"`
	Technical      TechnicalAnalysis      `json:"technical"`
	Aesthetic      AestheticAnalysis      `json:"aesthetic"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ImageQuality 图像质量
type ImageQuality struct {
	Overall    float64 `json:"overall"`
	Sharpness  float64 `json:"sharpness"`
	Brightness float64 `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Saturation float64 `json:"saturation"`
	Noise      float64 `json:"noise"`
	Blur       float64 `json:"blur"`
	Exposure   float64 `json:"exposure"`
}

// ColorAnalysis 颜色分析
type ColorAnalysis struct {
	DominantColors []Color `json:"dominant_colors"`
	ColorScheme    string  `json:"color_scheme"`
	Temperature    string  `json:"temperature"`
	Harmony        float64 `json:"harmony"`
	Vibrance       float64 `json:"vibrance"`
}

// Color 颜色
type Color struct {
	RGB        [3]int  `json:"rgb"`
	HSV        [3]float64 `json:"hsv"`
	Hex        string  `json:"hex"`
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
}

// CompositionAnalysis 构图分析
type CompositionAnalysis struct {
	RuleOfThirds   float64     `json:"rule_of_thirds"`
	Symmetry       float64     `json:"symmetry"`
	Balance        float64     `json:"balance"`
	LeadingLines   []Line2D    `json:"leading_lines"`
	FocalPoints    []Point2D   `json:"focal_points"`
	DepthOfField   float64     `json:"depth_of_field"`
}

// Line2D 二维线条
type Line2D struct {
	Start Point2D `json:"start"`
	End   Point2D `json:"end"`
	Angle float64 `json:"angle"`
}

// Point2D 二维点
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ContentAnalysis 内容分析
type ContentAnalysis struct {
	Objects     []string `json:"objects"`
	People      int      `json:"people"`
	Animals     int      `json:"animals"`
	Vehicles    int      `json:"vehicles"`
	Buildings   int      `json:"buildings"`
	Nature      float64  `json:"nature"`
	Indoor      bool     `json:"indoor"`
	Outdoor     bool     `json:"outdoor"`
	TimeOfDay   string   `json:"time_of_day"`
	Weather     string   `json:"weather"`
}

// TechnicalAnalysis 技术分析
type TechnicalAnalysis struct {
	Resolution   Resolution `json:"resolution"`
	AspectRatio  string     `json:"aspect_ratio"`
	FileSize     int64      `json:"file_size"`
	Compression  float64    `json:"compression"`
	ColorDepth   int        `json:"color_depth"`
	HasAlpha     bool       `json:"has_alpha"`
	EXIF         EXIFData   `json:"exif,omitempty"`
}

// Resolution 分辨率
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	DPI    int `json:"dpi"`
}

// EXIFData EXIF数据
type EXIFData struct {
	Camera       string    `json:"camera"`
	Lens         string    `json:"lens"`
	ISO          int       `json:"iso"`
	Aperture     string    `json:"aperture"`
	ShutterSpeed string    `json:"shutter_speed"`
	FocalLength  string    `json:"focal_length"`
	DateTime     time.Time `json:"date_time"`
	GPS          *GPSData  `json:"gps,omitempty"`
}

// GPSData GPS数据
type GPSData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// AestheticAnalysis 美学分析
type AestheticAnalysis struct {
	Beauty      float64 `json:"beauty"`
	Interest    float64 `json:"interest"`
	Emotion     string  `json:"emotion"`
	Mood        string  `json:"mood"`
	Style       string  `json:"style"`
	Artistic    float64 `json:"artistic"`
}

// AnomalyDetectionResult 异常检测结果
type AnomalyDetectionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	HasAnomalies   bool                   `json:"has_anomalies"`
	AnomalyScore   float64                `json:"anomaly_score"`
	Anomalies      []DetectedAnomaly      `json:"anomalies"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// DetectedAnomaly 检测到的异常
type DetectedAnomaly struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Confidence  float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Description string      `json:"description"`
}

// ImageComparisonResult 图像比较结果
type ImageComparisonResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Similarity     float64                `json:"similarity"`
	Differences    []ImageDifference      `json:"differences"`
	MatchedRegions []MatchedRegion        `json:"matched_regions"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ImageDifference 图像差异
type ImageDifference struct {
	Type        string      `json:"type"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Severity    float64     `json:"severity"`
	Description string      `json:"description"`
}

// MatchedRegion 匹配区域
type MatchedRegion struct {
	Region1     BoundingBox `json:"region1"`
	Region2     BoundingBox `json:"region2"`
	Similarity  float64     `json:"similarity"`
	Description string      `json:"description"`
}

// ImageOperation 图像操作
type ImageOperation struct {
	Type       OperationType          `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Order      int                    `json:"order"`
}

// OperationType 操作类型
type OperationType string

const (
	OpResize     OperationType = "resize"
	OpCrop       OperationType = "crop"
	OpRotate     OperationType = "rotate"
	OpFlip       OperationType = "flip"
	OpBrightness OperationType = "brightness"
	OpContrast   OperationType = "contrast"
	OpSaturation OperationType = "saturation"
	OpBlur       OperationType = "blur"
	OpSharpen    OperationType = "sharpen"
	OpNoise      OperationType = "noise"
	OpFilter     OperationType = "filter"
	OpWatermark  OperationType = "watermark"
	OpCompress   OperationType = "compress"
	OpFormat     OperationType = "format"
)

// EnhancementOptions 增强选项
type EnhancementOptions struct {
	AutoAdjust    bool    `json:"auto_adjust"`
	Denoise       bool    `json:"denoise"`
	Sharpen       bool    `json:"sharpen"`
	ColorCorrect  bool    `json:"color_correct"`
	Upscale       bool    `json:"upscale"`
	UpscaleFactor float64 `json:"upscale_factor"`
	Quality       float64 `json:"quality"`
}

// ImageProcessingResult 图像处理结果
type ImageProcessingResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	ProcessedImage ImageOutput            `json:"processed_image"`
	Operations     []ImageOperation       `json:"operations"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ImageOutput 图像输出
type ImageOutput struct {
	ID       string      `json:"id"`
	Data     []byte      `json:"data"`
	Format   ImageFormat `json:"format"`
	Width    int         `json:"width"`
	Height   int         `json:"height"`
	Size     int64       `json:"size"`
	Quality  float64     `json:"quality"`
	Metadata map[string]interface{} `json:"metadata"`
}

// BatchProcessingResult 批量处理结果
type BatchProcessingResult struct {
	ID             string                   `json:"id"`
	TotalImages    int                      `json:"total_images"`
	ProcessedImages int                     `json:"processed_images"`
	FailedImages   int                      `json:"failed_images"`
	Results        []ImageProcessingResult  `json:"results"`
	Errors         []BatchProcessingError   `json:"errors"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Timestamp      time.Time                `json:"timestamp"`
}

// BatchProcessingError 批量处理错误
type BatchProcessingError struct {
	ImageID string `json:"image_id"`
	Error   string `json:"error"`
}

// ImageConfig 图像配置
type ImageConfig struct {
	// 通用配置
	MaxImageSize    int64         `json:"max_image_size" yaml:"max_image_size"`
	MaxBatchSize    int           `json:"max_batch_size" yaml:"max_batch_size"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryCount      int           `json:"retry_count" yaml:"retry_count"`
	
	// 处理配置
	DefaultQuality  float64       `json:"default_quality" yaml:"default_quality"`
	EnableCache     bool          `json:"enable_cache" yaml:"enable_cache"`
	CacheExpiry     time.Duration `json:"cache_expiry" yaml:"cache_expiry"`
	
	// 提供商配置
	Providers       map[string]ProviderConfig `json:"providers" yaml:"providers"`
	
	// 特定功能配置
	ObjectDetection ObjectDetectionConfig `json:"object_detection" yaml:"object_detection"`
	FaceRecognition FaceRecognitionConfig `json:"face_recognition" yaml:"face_recognition"`
	TextRecognition TextRecognitionConfig `json:"text_recognition" yaml:"text_recognition"`
	ImageProcessing ImageProcessingConfig `json:"image_processing" yaml:"image_processing"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name     string                 `json:"name" yaml:"name"`
	Endpoint string                 `json:"endpoint" yaml:"endpoint"`
	APIKey   string                 `json:"api_key" yaml:"api_key"`
	Region   string                 `json:"region" yaml:"region"`
	Params   map[string]interface{} `json:"params" yaml:"params"`
}

// ObjectDetectionConfig 物体检测配置
type ObjectDetectionConfig struct {
	Provider          string  `json:"provider" yaml:"provider"`
	Model             string  `json:"model" yaml:"model"`
	ConfidenceThreshold float64 `json:"confidence_threshold" yaml:"confidence_threshold"`
	MaxObjects        int     `json:"max_objects" yaml:"max_objects"`
	EnableSubObjects  bool    `json:"enable_sub_objects" yaml:"enable_sub_objects"`
}

// FaceRecognitionConfig 人脸识别配置
type FaceRecognitionConfig struct {
	Provider            string  `json:"provider" yaml:"provider"`
	Model               string  `json:"model" yaml:"model"`
	ConfidenceThreshold float64 `json:"confidence_threshold" yaml:"confidence_threshold"`
	EnableLandmarks     bool    `json:"enable_landmarks" yaml:"enable_landmarks"`
	EnableAttributes    bool    `json:"enable_attributes" yaml:"enable_attributes"`
	EnableEmotions      bool    `json:"enable_emotions" yaml:"enable_emotions"`
	EnableIdentity      bool    `json:"enable_identity" yaml:"enable_identity"`
}

// TextRecognitionConfig 文本识别配置
type TextRecognitionConfig struct {
	Provider    string   `json:"provider" yaml:"provider"`
	Model       string   `json:"model" yaml:"model"`
	Languages   []string `json:"languages" yaml:"languages"`
	EnableWords bool     `json:"enable_words" yaml:"enable_words"`
	EnableLines bool     `json:"enable_lines" yaml:"enable_lines"`
}

// ImageProcessingConfig 图像处理配置
type ImageProcessingConfig struct {
	Provider       string  `json:"provider" yaml:"provider"`
	DefaultFormat  ImageFormat `json:"default_format" yaml:"default_format"`
	DefaultQuality float64 `json:"default_quality" yaml:"default_quality"`
	MaxWidth       int     `json:"max_width" yaml:"max_width"`
	MaxHeight      int     `json:"max_height" yaml:"max_height"`
	EnableGPU      bool    `json:"enable_gpu" yaml:"enable_gpu"`
}

// ImageProvider 图像提供商接口
type ImageProvider interface {
	// 识别功能
	RecognizeObjects(ctx context.Context, input ImageInput) (*ObjectRecognitionResult, error)
	RecognizeFaces(ctx context.Context, input ImageInput) (*FaceRecognitionResult, error)
	RecognizeText(ctx context.Context, input ImageInput) (*TextRecognitionResult, error)
	RecognizeScene(ctx context.Context, input ImageInput) (*SceneRecognitionResult, error)
	
	// 分析功能
	AnalyzeImage(ctx context.Context, input ImageInput) (*ImageAnalysisResult, error)
	
	// 处理功能
	ProcessImage(ctx context.Context, input ImageInput, operations []ImageOperation) (*ImageProcessingResult, error)
	
	// 配置和状态
	GetSupportedFormats() []ImageFormat
	GetSupportedOperations() []OperationType
	HealthCheck(ctx context.Context) error
}

// CreateImageInput 创建图像输入
func CreateImageInput(data []byte, format ImageFormat, userID, sessionID string) ImageInput {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	width, height := 0, 0
	if err == nil {
		width = img.Width
		height = img.Height
	}

	return ImageInput{
		ID:        uuid.New().String(),
		Data:      data,
		Format:    format,
		Width:     width,
		Height:    height,
		Size:      int64(len(data)),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		UserID:    userID,
		SessionID: sessionID,
		Quality:   1.0,
	}
}

// CreateImageOperation 创建图像操作
func CreateImageOperation(opType OperationType, params map[string]interface{}, order int) ImageOperation {
	return ImageOperation{
		Type:       opType,
		Parameters: params,
		Order:      order,
	}
}

// ToJSON 转换为JSON
func (r *ObjectRecognitionResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON 转换为JSON
func (r *FaceRecognitionResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON 转换为JSON
func (r *TextRecognitionResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSON 转换为JSON
func (r *ImageAnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}