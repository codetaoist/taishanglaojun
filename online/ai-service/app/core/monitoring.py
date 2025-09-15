#!/usr/bin/env python3
"""
码道 AI智能体服务监控模块

使用Prometheus收集和暴露应用指标。
"""

import time
from typing import Dict, Any
from functools import wraps
from prometheus_client import Counter, Histogram, Gauge, Info
from loguru import logger

from .config import settings


# 定义Prometheus指标

# 请求计数器
request_count = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status_code']
)

# 请求延迟直方图
request_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration in seconds',
    ['method', 'endpoint']
)

# AI模型调用计数器
ai_model_calls = Counter(
    'ai_model_calls_total',
    'Total AI model API calls',
    ['model', 'status']
)

# AI模型调用延迟
ai_model_duration = Histogram(
    'ai_model_call_duration_seconds',
    'AI model call duration in seconds',
    ['model']
)

# Token使用量计数器
token_usage = Counter(
    'ai_tokens_used_total',
    'Total AI tokens used',
    ['model', 'type']  # type: prompt_tokens, completion_tokens
)

# 活跃用户数量
active_users = Gauge(
    'active_users_count',
    'Number of active users'
)

# 数据库连接池状态
db_connections = Gauge(
    'database_connections_active',
    'Number of active database connections'
)

# Redis连接状态
redis_connections = Gauge(
    'redis_connections_active',
    'Number of active Redis connections'
)

# 错误计数器
error_count = Counter(
    'errors_total',
    'Total number of errors',
    ['error_type', 'service']
)

# 应用信息
app_info = Info(
    'app_info',
    'Application information'
)

# 内存使用量
memory_usage = Gauge(
    'memory_usage_bytes',
    'Memory usage in bytes',
    ['type']  # type: rss, vms, shared
)

# CPU使用率
cpu_usage = Gauge(
    'cpu_usage_percent',
    'CPU usage percentage'
)


class MetricsCollector:
    """指标收集器"""
    
    def __init__(self):
        self.start_time = time.time()
        self.active_requests = 0
    
    def record_request(self, method: str, endpoint: str, status_code: int, duration: float):
        """记录HTTP请求指标"""
        request_count.labels(
            method=method,
            endpoint=endpoint,
            status_code=status_code
        ).inc()
        
        request_duration.labels(
            method=method,
            endpoint=endpoint
        ).observe(duration)
    
    def record_ai_call(self, model: str, status: str, duration: float, tokens: Dict[str, int] = None):
        """记录AI模型调用指标"""
        ai_model_calls.labels(
            model=model,
            status=status
        ).inc()
        
        ai_model_duration.labels(model=model).observe(duration)
        
        if tokens:
            for token_type, count in tokens.items():
                token_usage.labels(
                    model=model,
                    type=token_type
                ).inc(count)
    
    def record_error(self, error_type: str, service: str = "ai-service"):
        """记录错误指标"""
        error_count.labels(
            error_type=error_type,
            service=service
        ).inc()
    
    def update_active_users(self, count: int):
        """更新活跃用户数量"""
        active_users.set(count)
    
    def update_db_connections(self, count: int):
        """更新数据库连接数"""
        db_connections.set(count)
    
    def update_redis_connections(self, count: int):
        """更新Redis连接数"""
        redis_connections.set(count)
    
    def update_system_metrics(self):
        """更新系统指标"""
        try:
            import psutil
            
            # 内存使用情况
            memory = psutil.virtual_memory()
            memory_usage.labels(type='total').set(memory.total)
            memory_usage.labels(type='available').set(memory.available)
            memory_usage.labels(type='used').set(memory.used)
            
            # CPU使用率
            cpu_percent = psutil.cpu_percent(interval=1)
            cpu_usage.set(cpu_percent)
            
        except ImportError:
            logger.warning("psutil未安装，无法收集系统指标")
        except Exception as e:
            logger.error(f"更新系统指标失败: {e}")


# 全局指标收集器实例
metrics_collector = MetricsCollector()


def monitor_requests(func):
    """监控HTTP请求的装饰器"""
    @wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = time.time()
        metrics_collector.active_requests += 1
        
        try:
            result = await func(*args, **kwargs)
            status_code = getattr(result, 'status_code', 200)
            
            # 记录成功请求
            duration = time.time() - start_time
            metrics_collector.record_request(
                method="POST",  # 大多数AI API都是POST
                endpoint=func.__name__,
                status_code=status_code,
                duration=duration
            )
            
            return result
            
        except Exception as e:
            # 记录错误请求
            duration = time.time() - start_time
            metrics_collector.record_request(
                method="POST",
                endpoint=func.__name__,
                status_code=500,
                duration=duration
            )
            
            # 记录错误
            metrics_collector.record_error(
                error_type=type(e).__name__
            )
            
            raise
        
        finally:
            metrics_collector.active_requests -= 1
    
    return wrapper


def monitor_ai_calls(model_name: str):
    """监控AI模型调用的装饰器"""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            
            try:
                result = await func(*args, **kwargs)
                
                # 记录成功调用
                duration = time.time() - start_time
                
                # 提取token使用情况
                tokens = {}
                if hasattr(result, 'usage'):
                    usage = result.usage
                    tokens = {
                        'prompt_tokens': getattr(usage, 'prompt_tokens', 0),
                        'completion_tokens': getattr(usage, 'completion_tokens', 0),
                        'total_tokens': getattr(usage, 'total_tokens', 0)
                    }
                
                metrics_collector.record_ai_call(
                    model=model_name,
                    status="success",
                    duration=duration,
                    tokens=tokens
                )
                
                return result
                
            except Exception as e:
                # 记录失败调用
                duration = time.time() - start_time
                metrics_collector.record_ai_call(
                    model=model_name,
                    status="error",
                    duration=duration
                )
                
                # 记录错误
                metrics_collector.record_error(
                    error_type=type(e).__name__
                )
                
                raise
        
        return wrapper
    return decorator


def setup_monitoring():
    """设置监控系统"""
    try:
        # 设置应用信息
        app_info.info({
            'version': settings.VERSION,
            'name': settings.APP_NAME,
            'environment': 'development' if settings.DEBUG else 'production'
        })
        
        logger.info("✓ 监控系统已初始化")
        
        # 启动系统指标收集（如果需要）
        if settings.ENABLE_METRICS:
            logger.info("✓ 系统指标收集已启用")
        
    except Exception as e:
        logger.error(f"监控系统初始化失败: {e}")


async def collect_runtime_metrics():
    """收集运行时指标"""
    try:
        # 更新系统指标
        metrics_collector.update_system_metrics()
        
        # 这里可以添加更多运行时指标收集
        # 比如从数据库查询活跃用户数等
        
    except Exception as e:
        logger.error(f"收集运行时指标失败: {e}")


class HealthChecker:
    """健康检查器"""
    
    def __init__(self):
        self.checks = {}
    
    def register_check(self, name: str, check_func):
        """注册健康检查函数"""
        self.checks[name] = check_func
    
    async def run_checks(self) -> Dict[str, Any]:
        """运行所有健康检查"""
        results = {
            "status": "healthy",
            "timestamp": time.time(),
            "checks": {}
        }
        
        overall_healthy = True
        
        for name, check_func in self.checks.items():
            try:
                check_result = await check_func()
                results["checks"][name] = {
                    "status": "healthy" if check_result else "unhealthy",
                    "details": check_result
                }
                
                if not check_result:
                    overall_healthy = False
                    
            except Exception as e:
                results["checks"][name] = {
                    "status": "error",
                    "error": str(e)
                }
                overall_healthy = False
        
        results["status"] = "healthy" if overall_healthy else "unhealthy"
        return results


# 全局健康检查器实例
health_checker = HealthChecker()


if __name__ == "__main__":
    # 测试监控功能
    import asyncio
    
    async def test_monitoring():
        logger.info("测试监控功能...")
        
        # 设置监控
        setup_monitoring()
        
        # 模拟一些指标
        metrics_collector.record_request("POST", "/ai/chat", 200, 1.5)
        metrics_collector.record_ai_call("gpt-4", "success", 2.3, {
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "total_tokens": 150
        })
        
        metrics_collector.update_active_users(10)
        
        logger.info("✓ 监控指标已记录")
        
        # 收集运行时指标
        await collect_runtime_metrics()
        
        logger.info("✓ 运行时指标已收集")
    
    asyncio.run(test_monitoring())