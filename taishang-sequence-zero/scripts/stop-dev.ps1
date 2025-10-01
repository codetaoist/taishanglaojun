# 太上老君序列0服务停止脚本

Write-Host "停止太上老君序列0开发环境..." -ForegroundColor Yellow

# 停止所有服务
docker-compose down

Write-Host "开发环境已停止" -ForegroundColor Green