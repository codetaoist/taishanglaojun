# 太上老君序列0服务启动脚本

Write-Host "启动太上老君序列0开发环境..." -ForegroundColor Green

# 启动基础服务
Write-Host "启动数据库和中间件服务..." -ForegroundColor Yellow
docker-compose up -d postgres redis neo4j prometheus grafana

# 等待服务启动
Write-Host "等待服务启动完成..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# 检查服务状态
Write-Host "检查服务状态..." -ForegroundColor Yellow
docker-compose ps

Write-Host "开发环境启动完成！" -ForegroundColor Green
Write-Host "访问地址:" -ForegroundColor Cyan
Write-Host "  - Grafana监控: http://localhost:3000 (admin/taishang123)" -ForegroundColor Cyan
Write-Host "  - Prometheus: http://localhost:9090" -ForegroundColor Cyan
Write-Host "  - Neo4j浏览器: http://localhost:7474 (neo4j/taishang123)" -ForegroundColor Cyan