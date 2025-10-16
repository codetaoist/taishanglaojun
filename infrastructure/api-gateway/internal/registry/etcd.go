package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// etcdRegistry etcd服务注册实现
type etcdRegistry struct {
	client     *clientv3.Client
	logger     logger.Logger
	watchers   map[string][]chan []*ServiceInstance
	mu         sync.RWMutex
	stopCh     chan struct{}
	keyPrefix  string
	leaseTTL   int64
	leaseID    clientv3.LeaseID
}

// newEtcdRegistry 创建etcd注册中心
func newEtcdRegistry(endpoints []string, options map[string]string, log logger.Logger) (Registry, error) {
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	
	// 设置认证信息
	if username, ok := options["username"]; ok {
		config.Username = username
	}
	if password, ok := options["password"]; ok {
		config.Password = password
	}
	
	// 设置TLS配置
	if certFile, ok := options["cert_file"]; ok {
		if keyFile, ok := options["key_file"]; ok {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			
			// 如果提供了CA文件，则验证服务器证书
			if caFile, ok := options["ca_file"]; ok {
				caCert, err := ioutil.ReadFile(caFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read CA certificate: %w", err)
				}
				
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				tlsConfig.RootCAs = caCertPool
			}
			
			config.TLS = tlsConfig
		}
	}
	
	client, err := clientv3.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	
	// 设置key前缀
	keyPrefix := "/services"
	if prefix, ok := options["key_prefix"]; ok {
		keyPrefix = prefix
	}
	
	// 设置租约TTL
	leaseTTL := int64(30)
	if ttl, ok := options["lease_ttl"]; ok {
		if t, err := time.ParseDuration(ttl); err == nil {
			leaseTTL = int64(t.Seconds())
		}
	}
	
	registry := &etcdRegistry{
		client:    client,
		logger:    log,
		watchers:  make(map[string][]chan []*ServiceInstance),
		stopCh:    make(chan struct{}),
		keyPrefix: keyPrefix,
		leaseTTL:  leaseTTL,
	}
	
	// 创建租约
	if err := registry.createLease(); err != nil {
		return nil, fmt.Errorf("failed to create lease: %w", err)
	}
	
	// 启动租约续期
	go registry.keepAliveLease()
	
	return registry, nil
}

// createLease 创建租约
func (r *etcdRegistry) createLease() error {
	resp, err := r.client.Grant(context.Background(), r.leaseTTL)
	if err != nil {
		return err
	}
	
	r.leaseID = resp.ID
	r.logger.Infof("Created etcd lease with ID: %x, TTL: %d seconds", r.leaseID, r.leaseTTL)
	
	return nil
}

// keepAliveLease 保持租约活跃
func (r *etcdRegistry) keepAliveLease() {
	ch, kaerr := r.client.KeepAlive(context.Background(), r.leaseID)
	if kaerr != nil {
		r.logger.Errorf("Failed to keep alive lease: %v", kaerr)
		return
	}
	
	for {
		select {
		case ka := <-ch:
			if ka == nil {
				r.logger.Warn("Lease keep alive channel closed, recreating lease")
				if err := r.createLease(); err != nil {
					r.logger.Errorf("Failed to recreate lease: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				// 重新启动keep alive
				go r.keepAliveLease()
				return
			}
			r.logger.Debugf("Lease keep alive response: %x", ka.ID)
		case <-r.stopCh:
			return
		}
	}
}

// Register 注册服务到etcd
func (r *etcdRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	if instance.ID == "" {
		instance.ID = fmt.Sprintf("%s-%s-%d", instance.Name, instance.Address, instance.Port)
	}
	
	// 构建服务key
	key := r.buildServiceKey(instance.Name, instance.ID)
	
	// 序列化服务实例信息
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal service instance: %w", err)
	}
	
	// 使用租约注册服务
	_, err = r.client.Put(ctx, key, string(data), clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register service to etcd: %w", err)
	}
	
	r.logger.Infof("Service registered to etcd: %s (%s)", instance.Name, instance.ID)
	
	return nil
}

// Deregister 从etcd注销服务
func (r *etcdRegistry) Deregister(ctx context.Context, instanceID string) error {
	// 查找服务实例
	resp, err := r.client.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get services from etcd: %w", err)
	}
	
	var keyToDelete string
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		
		if instance.ID == instanceID {
			keyToDelete = string(kv.Key)
			break
		}
	}
	
	if keyToDelete == "" {
		return fmt.Errorf("service instance not found: %s", instanceID)
	}
	
	// 删除服务实例
	_, err = r.client.Delete(ctx, keyToDelete)
	if err != nil {
		return fmt.Errorf("failed to deregister service from etcd: %w", err)
	}
	
	r.logger.Infof("Service deregistered from etcd: %s", instanceID)
	
	return nil
}

// Discover 从etcd发现服务
func (r *etcdRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	key := r.buildServicePrefix(serviceName)
	
	resp, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover service from etcd: %w", err)
	}
	
	var instances []*ServiceInstance
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.logger.Warnf("Failed to unmarshal service instance: %v", err)
			continue
		}
		
		// 只返回健康的实例
		if instance.Health == HealthStatusHealthy {
			instances = append(instances, &instance)
		}
	}
	
	return instances, nil
}

// ListServices 获取所有服务
func (r *etcdRegistry) ListServices(ctx context.Context) (map[string][]*ServiceInstance, error) {
	resp, err := r.client.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list services from etcd: %w", err)
	}
	
	result := make(map[string][]*ServiceInstance)
	
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.logger.Warnf("Failed to unmarshal service instance: %v", err)
			continue
		}
		
		if result[instance.Name] == nil {
			result[instance.Name] = make([]*ServiceInstance, 0)
		}
		result[instance.Name] = append(result[instance.Name], &instance)
	}
	
	return result, nil
}

// UpdateHealth 更新服务健康状态
func (r *etcdRegistry) UpdateHealth(ctx context.Context, instanceID string, status HealthStatus) error {
	// 查找服务实例
	resp, err := r.client.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get services from etcd: %w", err)
	}
	
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		
		if instance.ID == instanceID {
			// 更新健康状态
			instance.Health = status
			instance.LastSeen = time.Now()
			
			// 重新序列化并保存
			data, err := json.Marshal(instance)
			if err != nil {
				return fmt.Errorf("failed to marshal service instance: %w", err)
			}
			
			_, err = r.client.Put(ctx, string(kv.Key), string(data), clientv3.WithLease(r.leaseID))
			if err != nil {
				return fmt.Errorf("failed to update health status in etcd: %w", err)
			}
			
			r.logger.Debugf("Service health updated in etcd: %s -> %s", instanceID, status)
			
			// 通知监听者
			r.notifyWatchers(instance.Name)
			
			return nil
		}
	}
	
	return fmt.Errorf("service instance not found: %s", instanceID)
}

// Watch 监听服务变化
func (r *etcdRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	ch := make(chan []*ServiceInstance, 10)
	
	if r.watchers[serviceName] == nil {
		r.watchers[serviceName] = make([]chan []*ServiceInstance, 0)
	}
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)
	
	// 发送当前状态
	instances, err := r.Discover(ctx, serviceName)
	if err == nil {
		select {
		case ch <- instances:
		default:
		}
	}
	
	// 启动监听协程
	go r.watchService(ctx, serviceName, ch)
	
	return ch, nil
}

// watchService 监听单个服务的变化
func (r *etcdRegistry) watchService(ctx context.Context, serviceName string, ch chan []*ServiceInstance) {
	defer func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		
		// 移除监听器
		watchers := r.watchers[serviceName]
		for i, watcher := range watchers {
			if watcher == ch {
				r.watchers[serviceName] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		
		close(ch)
	}()
	
	key := r.buildServicePrefix(serviceName)
	watchCh := r.client.Watch(ctx, key, clientv3.WithPrefix())
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case watchResp := <-watchCh:
			if watchResp.Err() != nil {
				r.logger.Errorf("Watch error for service %s: %v", serviceName, watchResp.Err())
				continue
			}
			
			// 获取最新的服务实例
			instances, err := r.Discover(context.Background(), serviceName)
			if err != nil {
				r.logger.Errorf("Failed to discover service %s: %v", serviceName, err)
				continue
			}
			
			// 发送更新
			select {
			case ch <- instances:
			default:
				// 如果通道满了，跳过这次更新
			}
		}
	}
}

// notifyWatchers 通知监听者
func (r *etcdRegistry) notifyWatchers(serviceName string) {
	r.mu.RLock()
	watchers := r.watchers[serviceName]
	r.mu.RUnlock()
	
	if len(watchers) == 0 {
		return
	}
	
	// 获取最新的服务实例
	instances, err := r.Discover(context.Background(), serviceName)
	if err != nil {
		r.logger.Errorf("Failed to discover service %s for notification: %v", serviceName, err)
		return
	}
	
	// 通知所有监听者
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
			// 如果通道满了，跳过这次通知
		}
	}
}

// buildServiceKey 构建服务实例key
func (r *etcdRegistry) buildServiceKey(serviceName, instanceID string) string {
	return path.Join(r.keyPrefix, serviceName, instanceID)
}

// buildServicePrefix 构建服务前缀
func (r *etcdRegistry) buildServicePrefix(serviceName string) string {
	return path.Join(r.keyPrefix, serviceName) + "/"
}

// Close 关闭etcd注册中心
func (r *etcdRegistry) Close() error {
	close(r.stopCh)
	
	// 撤销租约
	if r.leaseID != 0 {
		_, err := r.client.Revoke(context.Background(), r.leaseID)
		if err != nil {
			r.logger.Errorf("Failed to revoke lease: %v", err)
		}
	}
	
	// 关闭客户端
	if err := r.client.Close(); err != nil {
		r.logger.Errorf("Failed to close etcd client: %v", err)
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 关闭所有监听器
	for _, watchers := range r.watchers {
		for _, watcher := range watchers {
			close(watcher)
		}
	}
	
	r.watchers = make(map[string][]chan []*ServiceInstance)
	
	r.logger.Info("etcd registry closed")
	
	return nil
}