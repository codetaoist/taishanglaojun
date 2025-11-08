package ci

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codetaoist/services/api/internal/middleware"
)

// BuildConfig contains CI/CD build configuration
type BuildConfig struct {
	BuildDir      string            `yaml:"build_dir"`
	ArtifactDir   string            `yaml:"artifact_dir"`
	Dockerfile    string            `yaml:"dockerfile"`
	ImageName     string            `yaml:"image_name"`
	ImageTag      string            `yaml:"image_tag"`
	Registry      string            `yaml:"registry"`
	BuildArgs     map[string]string `yaml:"build_args"`
	Environment   map[string]string `yaml:"environment"`
	Timeout       time.Duration     `yaml:"timeout"`
	CacheEnabled  bool              `yaml:"cache_enabled"`
	CacheDir      string            `yaml:"cache_dir"`
}

// TestConfig contains CI/CD test configuration
type TestConfig struct {
	TestDir       string            `yaml:"test_dir"`
	TestPattern   string            `yaml:"test_pattern"`
	Coverage      bool              `yaml:"coverage"`
	CoverageFile  string            `yaml:"coverage_file"`
	Environment   map[string]string `yaml:"environment"`
	Timeout       time.Duration     `yaml:"timeout"`
	Parallel      bool              `yaml:"parallel"`
	Verbose       bool              `yaml:"verbose"`
}

// DeployConfig contains CI/CD deploy configuration
type DeployConfig struct {
	Environment   string            `yaml:"environment"`
	Namespace     string            `yaml:"namespace"`
	KubeConfig    string            `yaml:"kube_config"`
	Manifests     []string          `yaml:"manifests"`
	HelmChart     string            `yaml:"helm_chart"`
	HelmValues    string            `yaml:"helm_values"`
	HelmRelease   string            `yaml:"helm_release"`
	Wait          bool              `yaml:"wait"`
	Timeout       time.Duration     `yaml:"timeout"`
	Variables     map[string]string `yaml:"variables"`
}

// CIConfig contains CI/CD configuration
type CIConfig struct {
	Build   BuildConfig `yaml:"build"`
	Test    TestConfig  `yaml:"test"`
	Deploy  DeployConfig `yaml:"deploy"`
	Enabled bool        `yaml:"enabled"`
}

// BuildStatus represents the status of a build
type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusSuccess   BuildStatus = "success"
	BuildStatusFailure   BuildStatus = "failure"
	BuildStatusCancelled BuildStatus = "cancelled"
)

// BuildInfo contains information about a build
type BuildInfo struct {
	ID          string                 `json:"id"`
	Status      BuildStatus            `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Log         string                 `json:"log"`
	Artifact    string                 `json:"artifact,omitempty"`
	Image       string                 `json:"image,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedBy   string                 `json:"created_by"`
	TriggeredBy string                 `json:"triggered_by"`
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Repository  string                 `json:"repository"`
	Branch      string                 `json:"branch"`
	Commit      string                 `json:"commit"`
	Config      CIConfig               `json:"config"`
	Builds      []*BuildInfo           `json:"builds"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CIPipeline manages CI/CD pipelines
type CIPipeline struct {
	pipelines map[string]*Pipeline
	mu        sync.RWMutex
	logger    *middleware.Logger
	config    *CIConfig
}

// NewCIPipeline creates a new CI/CD pipeline manager
func NewCIPipeline(config *CIConfig, logger *middleware.Logger) *CIPipeline {
	return &CIPipeline{
		pipelines: make(map[string]*Pipeline),
		logger:    logger,
		config:    config,
	}
}

// CreatePipeline creates a new pipeline
func (cp *CIPipeline) CreatePipeline(ctx context.Context, name, repository, branch, commit string, config CIConfig) (*Pipeline, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	id := generateID(name)
	
	pipeline := &Pipeline{
		ID:          id,
		Name:        name,
		Repository:  repository,
		Branch:      branch,
		Commit:      commit,
		Config:      config,
		Builds:      make([]*BuildInfo, 0),
		Enabled:     true,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	cp.pipelines[id] = pipeline
	cp.logger.Infof("Created pipeline %s (ID: %s)", name, id)
	
	return pipeline, nil
}

// GetPipeline returns a pipeline with the given ID
func (cp *CIPipeline) GetPipeline(id string) (*Pipeline, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	pipeline, exists := cp.pipelines[id]
	return pipeline, exists
}

// ListPipelines returns a list of all pipelines
func (cp *CIPipeline) ListPipelines() []*Pipeline {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	var pipelines []*Pipeline
	for _, pipeline := range cp.pipelines {
		pipelines = append(pipelines, pipeline)
	}

	return pipelines
}

// DeletePipeline deletes a pipeline with the given ID
func (cp *CIPipeline) DeletePipeline(ctx context.Context, id string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, exists := cp.pipelines[id]; !exists {
		return fmt.Errorf("pipeline with ID %s not found", id)
	}

	delete(cp.pipelines, id)
	cp.logger.Infof("Deleted pipeline with ID %s", id)
	
	return nil
}

// RunBuild runs a build for the given pipeline
func (cp *CIPipeline) RunBuild(ctx context.Context, pipelineID, triggeredBy string) (*BuildInfo, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	pipeline, exists := cp.pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline with ID %s not found", pipelineID)
	}

	buildID := generateID(fmt.Sprintf("%s-build", pipelineID))
	
	build := &BuildInfo{
		ID:          buildID,
		Status:      BuildStatusPending,
		StartTime:   time.Now(),
		Log:         "",
		Metadata:    make(map[string]interface{}),
		CreatedBy:   "system",
		TriggeredBy: triggeredBy,
	}

	pipeline.Builds = append(pipeline.Builds, build)
	pipeline.UpdatedAt = time.Now()

	// Run the build in a goroutine
	go cp.runBuild(ctx, pipeline, build)

	return build, nil
}

// runBuild executes the build process
func (cp *CIPipeline) runBuild(ctx context.Context, pipeline *Pipeline, build *BuildInfo) {
	// Update status to running
	build.Status = BuildStatusRunning
	
	// Create a context with timeout
	buildCtx, cancel := context.WithTimeout(ctx, pipeline.Config.Build.Timeout)
	defer cancel()

	// Create a log buffer
	var logBuffer strings.Builder
	
	// Run the build steps
	if err := cp.runBuildSteps(buildCtx, pipeline, build, &logBuffer); err != nil {
		build.Status = BuildStatusFailure
		build.Log = logBuffer.String()
		build.EndTime = &time.Time{}
		*build.EndTime = time.Now()
		build.Duration = build.EndTime.Sub(build.StartTime)
		
		cp.logger.Errorf("Build %s failed: %v", build.ID, err)
		return
	}

	// Update status to success
	build.Status = BuildStatusSuccess
	build.Log = logBuffer.String()
	build.EndTime = &time.Time{}
	*build.EndTime = time.Now()
	build.Duration = build.EndTime.Sub(build.StartTime)
	
	cp.logger.Infof("Build %s completed successfully", build.ID)
}

// runBuildSteps executes the build steps
func (cp *CIPipeline) runBuildSteps(ctx context.Context, pipeline *Pipeline, build *BuildInfo, logBuffer *strings.Builder) error {
	// Step 1: Checkout code
	if err := cp.checkoutCode(ctx, pipeline, logBuffer); err != nil {
		return fmt.Errorf("checkout failed: %v", err)
	}

	// Step 2: Run tests
	if err := cp.runTests(ctx, pipeline, logBuffer); err != nil {
		return fmt.Errorf("tests failed: %v", err)
	}

	// Step 3: Build application
	if err := cp.buildApplication(ctx, pipeline, logBuffer); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	// Step 4: Build Docker image
	if err := cp.buildDockerImage(ctx, pipeline, build, logBuffer); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	// Step 5: Push Docker image
	if err := cp.pushDockerImage(ctx, pipeline, build, logBuffer); err != nil {
		return fmt.Errorf("docker push failed: %v", err)
	}

	return nil
}

// checkoutCode checks out the code from the repository
func (cp *CIPipeline) checkoutCode(ctx context.Context, pipeline *Pipeline, logBuffer *strings.Builder) error {
	logBuffer.WriteString("Checking out code...\n")
	
	// Create build directory
	buildDir := filepath.Join(pipeline.Config.Build.BuildDir, pipeline.ID)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %v", err)
	}

	// Clone repository
	cmd := exec.CommandContext(ctx, "git", "clone", "--branch", pipeline.Branch, pipeline.Repository, buildDir)
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Checkout specific commit
	cmd = exec.CommandContext(ctx, "git", "checkout", pipeline.Commit)
	cmd.Dir = buildDir
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout commit: %v", err)
	}

	logBuffer.WriteString("Code checked out successfully\n")
	return nil
}

// runTests runs the test suite
func (cp *CIPipeline) runTests(ctx context.Context, pipeline *Pipeline, logBuffer *strings.Builder) error {
	logBuffer.WriteString("Running tests...\n")
	
	buildDir := filepath.Join(pipeline.Config.Build.BuildDir, pipeline.ID)
	testDir := filepath.Join(buildDir, pipeline.Config.Test.TestDir)
	
	// Set environment variables
	env := os.Environ()
	for k, v := range pipeline.Config.Test.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	
	// Build test command
	var args []string
	args = append(args, "test")
	
	if pipeline.Config.Test.Verbose {
		args = append(args, "-v")
	}
	
	if pipeline.Config.Test.Coverage {
		args = append(args, "-cover")
		args = append(args, fmt.Sprintf("-coverprofile=%s", pipeline.Config.Test.CoverageFile))
	}
	
	if pipeline.Config.Test.Parallel {
		args = append(args, "-parallel", "4")
	}
	
	args = append(args, pipeline.Config.Test.TestPattern)
	
	// Run tests
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = testDir
	cmd.Env = env
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %v", err)
	}

	logBuffer.WriteString("Tests passed successfully\n")
	return nil
}

// buildApplication builds the application
func (cp *CIPipeline) buildApplication(ctx context.Context, pipeline *Pipeline, logBuffer *strings.Builder) error {
	logBuffer.WriteString("Building application...\n")
	
	buildDir := filepath.Join(pipeline.Config.Build.BuildDir, pipeline.ID)
	
	// Set environment variables
	env := os.Environ()
	for k, v := range pipeline.Config.Build.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	
	// Build application
	cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(pipeline.Config.Build.ArtifactDir, pipeline.ID), ".")
	cmd.Dir = buildDir
	cmd.Env = env
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	logBuffer.WriteString("Application built successfully\n")
	return nil
}

// buildDockerImage builds the Docker image
func (cp *CIPipeline) buildDockerImage(ctx context.Context, pipeline *Pipeline, build *BuildInfo, logBuffer *strings.Builder) error {
	logBuffer.WriteString("Building Docker image...\n")
	
	buildDir := filepath.Join(pipeline.Config.Build.BuildDir, pipeline.ID)
	dockerfile := filepath.Join(buildDir, pipeline.Config.Build.Dockerfile)
	
	// Generate image tag
	imageTag := pipeline.Config.Build.ImageTag
	if imageTag == "" {
		imageTag = "latest"
	}
	
	imageName := fmt.Sprintf("%s/%s:%s", pipeline.Config.Build.Registry, pipeline.Config.Build.ImageName, imageTag)
	build.Image = imageName
	
	// Build Docker image
	args := []string{"build", "-f", dockerfile, "-t", imageName}
	
	// Add build arguments
	for k, v := range pipeline.Config.Build.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	
	// Add cache directory if enabled
	if pipeline.Config.Build.CacheEnabled {
		args = append(args, "--cache-from", fmt.Sprintf("%s/%s:cache", pipeline.Config.Build.Registry, pipeline.Config.Build.ImageName))
	}
	
	args = append(args, buildDir)
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	logBuffer.WriteString(fmt.Sprintf("Docker image %s built successfully\n", imageName))
	return nil
}

// pushDockerImage pushes the Docker image to the registry
func (cp *CIPipeline) pushDockerImage(ctx context.Context, pipeline *Pipeline, build *BuildInfo, logBuffer *strings.Builder) error {
	logBuffer.WriteString("Pushing Docker image...\n")
	
	imageName := build.Image
	
	// Push Docker image
	cmd := exec.CommandContext(ctx, "docker", "push", imageName)
	cmd.Stdout = logBuffer
	cmd.Stderr = logBuffer
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %v", err)
	}

	logBuffer.WriteString(fmt.Sprintf("Docker image %s pushed successfully\n", imageName))
	return nil
}

// generateID generates a unique ID
func generateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}