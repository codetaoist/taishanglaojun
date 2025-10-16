# 解析JWT Token的PowerShell脚本

# 登录获取token
$loginUrl = "http://localhost:8080/api/v1/auth/login"
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "正在登录获取token..." -ForegroundColor Yellow

try {
    $loginResponse = Invoke-RestMethod -Uri $loginUrl -Method POST -Body $loginBody -Headers $headers
    $token = $loginResponse.token
    Write-Host "登录成功，获取到token" -ForegroundColor Green
    
    # 解析JWT token
    Write-Host "`n解析JWT Token..." -ForegroundColor Yellow
    
    # JWT token由三部分组成，用.分隔
    $tokenParts = $token.Split('.')
    if ($tokenParts.Length -ne 3) {
        Write-Host "无效的JWT token格式" -ForegroundColor Red
        exit 1
    }
    
    # 解码payload (第二部分)
    $payloadBase64 = $tokenParts[1]
    # 添加padding如果需要
    while ($payloadBase64.Length % 4 -ne 0) {
        $payloadBase64 += "="
    }
    
    $payloadBytes = [System.Convert]::FromBase64String($payloadBase64)
    $payloadJson = [System.Text.Encoding]::UTF8.GetString($payloadBytes)
    Write-Host "JWT Payload:" -ForegroundColor Cyan
    Write-Host $payloadJson -ForegroundColor White
    
    # 解析payload为对象
    $payload = $payloadJson | ConvertFrom-Json
    Write-Host "`n解析后的用户信息:" -ForegroundColor Cyan
    Write-Host "用户ID: $($payload.user_id)" -ForegroundColor White
    Write-Host "用户名: $($payload.username)" -ForegroundColor White
    Write-Host "角色: $($payload.role)" -ForegroundColor White
    Write-Host "等级: $($payload.level)" -ForegroundColor White
    
    # 现在使用这个token获取用户信息
    Write-Host "`n使用token获取用户信息..." -ForegroundColor Yellow
    $userInfoUrl = "http://localhost:8080/api/v1/user/me"
    $authHeaders = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $userInfo = Invoke-RestMethod -Uri $userInfoUrl -Method GET -Headers $authHeaders
    Write-Host "API返回的用户信息:" -ForegroundColor Cyan
    Write-Host ($userInfo | ConvertTo-Json -Depth 3) -ForegroundColor White
    
    # 比较JWT中的用户ID和API返回的用户ID
    Write-Host "`n用户ID比较:" -ForegroundColor Yellow
    Write-Host "JWT中的用户ID: $($payload.user_id)" -ForegroundColor White
    Write-Host "API返回的用户ID: $($userInfo.user_id)" -ForegroundColor White
    
    if ($payload.user_id -eq $userInfo.user_id) {
        Write-Host "✓ 用户ID匹配" -ForegroundColor Green
    } else {
        Write-Host "✗ 用户ID不匹配！这是问题的根源。" -ForegroundColor Red
    }
    
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}