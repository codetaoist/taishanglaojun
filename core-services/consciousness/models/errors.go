package models

import "errors"

// 意识服务相关错误定义
var (
	// 坐标相关错误
	ErrInvalidSAxis      = errors.New("invalid S axis value, must be between 0 and 5")
	ErrInvalidCAxis      = errors.New("invalid C axis value, must be between 0 and 5")
	ErrInvalidTAxis      = errors.New("invalid T axis value, must be between 0 and 5")
	ErrInvalidCoordinate = errors.New("invalid coordinate")

	// 融合引擎错误
	ErrFusionEngineNotReady         = errors.New("fusion engine is not ready")
	ErrFusionEngineBusy             = errors.New("fusion engine is busy")
	ErrFusionSessionNotFound        = errors.New("fusion session not found")
	ErrFusionSessionAlreadyFinished = errors.New("fusion session already finished")
	ErrNoCompatibleFusionStrategy   = errors.New("no compatible fusion strategy found")
	ErrCarbonSiliconImbalance       = errors.New("carbon-silicon fusion imbalance detected")
	ErrFusionProcessFailed          = errors.New("fusion process failed")
	ErrInvalidFusionInput           = errors.New("invalid fusion input")

	// 进化追踪错误
	ErrEvolutionTrackerNotInitialized = errors.New("evolution tracker not initialized")
	ErrSequenceTargetNotSet           = errors.New("sequence target not set")
	ErrEvolutionPathNotFound          = errors.New("evolution path not found")
	ErrEvolutionStagnation            = errors.New("evolution stagnation detected")

	// 量子基因错误
	ErrQuantumGeneNotFound       = errors.New("quantum gene not found")
	ErrQuantumGeneMutationFailed = errors.New("quantum gene mutation failed")
	ErrInvalidMutationRate       = errors.New("invalid mutation rate")
	ErrGeneExpressionFailed      = errors.New("gene expression failed")

	// 三轴协同错误
	ErrAxisCoordinationFailed = errors.New("axis coordination failed")
	ErrInsufficientAxisData   = errors.New("insufficient axis data for coordination")
	ErrAxisImbalance          = errors.New("axis imbalance detected")
	ErrCoordinationTimeout    = errors.New("coordination process timeout")

	// 意识状态错?
	ErrConsciousnessNotActive      = errors.New("consciousness is not active")
	ErrInvalidConsciousnessState   = errors.New("invalid consciousness state")
	ErrConsciousnessAnalysisFailed = errors.New("consciousness analysis failed")
	ErrConsciousnessOverload       = errors.New("consciousness system overload")

	// 数据库相关错?
	ErrDatabaseConnectionFailed = errors.New("database connection failed")
	ErrRecordNotFound           = errors.New("record not found")
	ErrDuplicateRecord          = errors.New("duplicate record")
	ErrDatabaseOperationFailed  = errors.New("database operation failed")

	// 配置相关错误
	ErrInvalidConfiguration    = errors.New("invalid configuration")
	ErrMissingRequiredConfig   = errors.New("missing required configuration")
	ErrConfigurationLoadFailed = errors.New("configuration load failed")

	// 服务相关错误
	ErrServiceNotAvailable   = errors.New("service not available")
	ErrServiceTimeout        = errors.New("service timeout")
	ErrServiceOverloaded     = errors.New("service overloaded")
	ErrInvalidServiceRequest = errors.New("invalid service request")
)

