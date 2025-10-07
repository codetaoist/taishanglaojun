package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	"github.com/fsnotify/fsnotify"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

// FileConfigSource 文件配置源
type FileConfigSource struct {
	filePath string
	logger   logger.Logger
	watcher  *fsnotify.Watcher
	mu       sync.RWMutex
	stopCh   chan struct{}
}

// NewFileConfigSource 创建文件配置源
func NewFileConfigSource(filePath string, log logger.Logger) (*FileConfigSource, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileConfigSource{
		filePath: filePath,
		logger:   log,
		watcher:  watcher,
		stopCh:   make(chan struct{}),
	}, nil
}

// GetRoutes 从文件获取路由配置
func (fcs *FileConfigSource) GetRoutes(ctx context.Context) ([]*RouteConfig, error) {
	data, err := ioutil.ReadFile(fcs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var routes []*RouteConfig
	ext := filepath.Ext(fcs.filePath)
	
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &routes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &routes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	fcs.logger.WithFields(map[string]interface{}{
		"file":  fcs.filePath,
		"count": len(routes),
	}).Info("Routes loaded from file")

	return routes, nil
}

// Watch 监听文件变化
func (fcs *FileConfigSource) Watch(ctx context.Context) (<-chan []*RouteConfig, error) {
	routeCh := make(chan []*RouteConfig, 1)

	// 添加文件监听
	if err := fcs.watcher.Add(fcs.filePath); err != nil {
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	go func() {
		defer close(routeCh)
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-fcs.stopCh:
				return
			case event := <-fcs.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					fcs.logger.WithFields(map[string]interface{}{
						"file": event.Name,
					}).Info("Config file changed, reloading routes")

					// 延迟一下，确保文件写入完成
					time.Sleep(100 * time.Millisecond)

					routes, err := fcs.GetRoutes(ctx)
					if err != nil {
						fcs.logger.Errorf("Failed to reload routes from file: %v", err)
						continue
					}

					select {
					case routeCh <- routes:
					case <-ctx.Done():
						return
					case <-fcs.stopCh:
						return
					}
				}
			case err := <-fcs.watcher.Errors:
				fcs.logger.Errorf("File watcher error: %v", err)
			}
		}
	}()

	return routeCh, nil
}

// Close 关闭文件配置源
func (fcs *FileConfigSource) Close() error {
	close(fcs.stopCh)
	return fcs.watcher.Close()
}

// EtcdConfigSource etcd配置源
type EtcdConfigSource struct {
	client   *clientv3.Client
	keyPrefix string
	logger   logger.Logger
	mu       sync.RWMutex
	stopCh   chan struct{}
}

// NewEtcdConfigSource 创建etcd配置源
func NewEtcdConfigSource(endpoints []string, keyPrefix string, log logger.Logger) (*EtcdConfigSource, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdConfigSource{
		client:    client,
		keyPrefix: keyPrefix,
		logger:    log,
		stopCh:    make(chan struct{}),
	}, nil
}

// GetRoutes 从etcd获取路由配置
func (ecs *EtcdConfigSource) GetRoutes(ctx context.Context) ([]*RouteConfig, error) {
	resp, err := ecs.client.Get(ctx, ecs.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get routes from etcd: %w", err)
	}

	var routes []*RouteConfig
	for _, kv := range resp.Kvs {
		var route RouteConfig
		if err := json.Unmarshal(kv.Value, &route); err != nil {
			ecs.logger.Errorf("Failed to unmarshal route config from key %s: %v", string(kv.Key), err)
			continue
		}
		routes = append(routes, &route)
	}

	ecs.logger.WithFields(map[string]interface{}{
		"prefix": ecs.keyPrefix,
		"count":  len(routes),
	}).Info("Routes loaded from etcd")

	return routes, nil
}

// Watch 监听etcd变化
func (ecs *EtcdConfigSource) Watch(ctx context.Context) (<-chan []*RouteConfig, error) {
	routeCh := make(chan []*RouteConfig, 1)

	go func() {
		defer close(routeCh)

		watchCh := ecs.client.Watch(ctx, ecs.keyPrefix, clientv3.WithPrefix())
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ecs.stopCh:
				return
			case watchResp := <-watchCh:
				if watchResp.Err() != nil {
					ecs.logger.Errorf("Etcd watch error: %v", watchResp.Err())
					continue
				}

				// 有变化时重新加载所有路由
				routes, err := ecs.GetRoutes(ctx)
				if err != nil {
					ecs.logger.Errorf("Failed to reload routes from etcd: %v", err)
					continue
				}

				select {
				case routeCh <- routes:
				case <-ctx.Done():
					return
				case <-ecs.stopCh:
					return
				}
			}
		}
	}()

	return routeCh, nil
}

// Close 关闭etcd配置源
func (ecs *EtcdConfigSource) Close() error {
	close(ecs.stopCh)
	return ecs.client.Close()
}

// DatabaseConfigSource 数据库配置源
type DatabaseConfigSource struct {
	// TODO: 实现数据库配置源
	logger logger.Logger
	stopCh chan struct{}
}

// NewDatabaseConfigSource 创建数据库配置源
func NewDatabaseConfigSource(log logger.Logger) *DatabaseConfigSource {
	return &DatabaseConfigSource{
		logger: log,
		stopCh: make(chan struct{}),
	}
}

// GetRoutes 从数据库获取路由配置
func (dcs *DatabaseConfigSource) GetRoutes(ctx context.Context) ([]*RouteConfig, error) {
	// TODO: 实现从数据库获取路由配置
	dcs.logger.Info("Getting routes from database (not implemented)")
	return []*RouteConfig{}, nil
}

// Watch 监听数据库变化
func (dcs *DatabaseConfigSource) Watch(ctx context.Context) (<-chan []*RouteConfig, error) {
	routeCh := make(chan []*RouteConfig, 1)
	
	go func() {
		defer close(routeCh)
		
		// TODO: 实现数据库变化监听
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-dcs.stopCh:
				return
			case <-ticker.C:
				// 定期检查数据库变化
				routes, err := dcs.GetRoutes(ctx)
				if err != nil {
					dcs.logger.Errorf("Failed to get routes from database: %v", err)
					continue
				}
				
				select {
				case routeCh <- routes:
				case <-ctx.Done():
					return
				case <-dcs.stopCh:
					return
				}
			}
		}
	}()
	
	return routeCh, nil
}

// Close 关闭数据库配置源
func (dcs *DatabaseConfigSource) Close() error {
	close(dcs.stopCh)
	return nil
}

// HTTPConfigSource HTTP API配置源
type HTTPConfigSource struct {
	apiURL string
	logger logger.Logger
	stopCh chan struct{}
}

// NewHTTPConfigSource 创建HTTP API配置源
func NewHTTPConfigSource(apiURL string, log logger.Logger) *HTTPConfigSource {
	return &HTTPConfigSource{
		apiURL: apiURL,
		logger: log,
		stopCh: make(chan struct{}),
	}
}

// GetRoutes 从HTTP API获取路由配置
func (hcs *HTTPConfigSource) GetRoutes(ctx context.Context) ([]*RouteConfig, error) {
	// TODO: 实现从HTTP API获取路由配置
	hcs.logger.WithFields(map[string]interface{}{
		"url": hcs.apiURL,
	}).Info("Getting routes from HTTP API (not implemented)")
	return []*RouteConfig{}, nil
}

// Watch 监听HTTP API变化
func (hcs *HTTPConfigSource) Watch(ctx context.Context) (<-chan []*RouteConfig, error) {
	routeCh := make(chan []*RouteConfig, 1)
	
	go func() {
		defer close(routeCh)
		
		// TODO: 实现HTTP API变化监听（可能通过轮询或WebSocket）
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-hcs.stopCh:
				return
			case <-ticker.C:
				routes, err := hcs.GetRoutes(ctx)
				if err != nil {
					hcs.logger.Errorf("Failed to get routes from HTTP API: %v", err)
					continue
				}
				
				select {
				case routeCh <- routes:
				case <-ctx.Done():
					return
				case <-hcs.stopCh:
					return
				}
			}
		}
	}()
	
	return routeCh, nil
}

// Close 关闭HTTP配置源
func (hcs *HTTPConfigSource) Close() error {
	close(hcs.stopCh)
	return nil
}

// CompositeConfigSource 复合配置源
type CompositeConfigSource struct {
	sources []RouteConfigSource
	logger  logger.Logger
	mu      sync.RWMutex
}

// NewCompositeConfigSource 创建复合配置源
func NewCompositeConfigSource(log logger.Logger) *CompositeConfigSource {
	return &CompositeConfigSource{
		sources: make([]RouteConfigSource, 0),
		logger:  log,
	}
}

// AddSource 添加配置源
func (ccs *CompositeConfigSource) AddSource(source RouteConfigSource) {
	ccs.mu.Lock()
	defer ccs.mu.Unlock()
	
	ccs.sources = append(ccs.sources, source)
	ccs.logger.Info("Config source added to composite source")
}

// GetRoutes 从所有配置源获取路由配置
func (ccs *CompositeConfigSource) GetRoutes(ctx context.Context) ([]*RouteConfig, error) {
	ccs.mu.RLock()
	sources := make([]RouteConfigSource, len(ccs.sources))
	copy(sources, ccs.sources)
	ccs.mu.RUnlock()

	var allRoutes []*RouteConfig
	
	for _, source := range sources {
		routes, err := source.GetRoutes(ctx)
		if err != nil {
			ccs.logger.Errorf("Failed to get routes from config source: %v", err)
			continue
		}
		allRoutes = append(allRoutes, routes...)
	}

	ccs.logger.WithFields(map[string]interface{}{
		"sources": len(sources),
		"routes":  len(allRoutes),
	}).Info("Routes loaded from composite source")

	return allRoutes, nil
}

// Watch 监听所有配置源变化
func (ccs *CompositeConfigSource) Watch(ctx context.Context) (<-chan []*RouteConfig, error) {
	routeCh := make(chan []*RouteConfig, 1)

	ccs.mu.RLock()
	sources := make([]RouteConfigSource, len(ccs.sources))
	copy(sources, ccs.sources)
	ccs.mu.RUnlock()

	// 为每个配置源启动监听
	for _, source := range sources {
		go func(src RouteConfigSource) {
			watchCh, err := src.Watch(ctx)
			if err != nil {
				ccs.logger.Errorf("Failed to watch config source: %v", err)
				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case routes := <-watchCh:
					if routes == nil {
						continue
					}

					// 当任何一个配置源变化时，重新加载所有配置源
					allRoutes, err := ccs.GetRoutes(ctx)
					if err != nil {
						ccs.logger.Errorf("Failed to reload all routes: %v", err)
						continue
					}

					select {
					case routeCh <- allRoutes:
					case <-ctx.Done():
						return
					}
				}
			}
		}(source)
	}

	return routeCh, nil
}

// Close 关闭所有配置源
func (ccs *CompositeConfigSource) Close() error {
	ccs.mu.RLock()
	sources := make([]RouteConfigSource, len(ccs.sources))
	copy(sources, ccs.sources)
	ccs.mu.RUnlock()

	for _, source := range sources {
		if err := source.Close(); err != nil {
			ccs.logger.Errorf("Failed to close config source: %v", err)
		}
	}

	return nil
}