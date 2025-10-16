# 简单的JWT测试脚本

try {
    # 1. 登录获取token
    Write-Host "1. 登录获取token..." -ForegroundColor Yellow
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body '{"username":"admin","password":"admin123"}' -ContentType "application/json"
    
    Write-Host "登录响应:" -ForegroundColor Cyan
    Write-Host ($loginResponse | ConvertTo-Json) -ForegroundColor White
    
    if ($loginResponse.token) {
        $token = $loginResponse.token
        Write-Host "Token获取成功" -ForegroundColor Green
        
        # 2. 解析JWT payload
        Write-Host "`n2. 解析JWT payload..." -ForegroundColor Yellow
        $parts = $token.Split('.')
        if ($parts.Length -eq 3) {
            $payload = $parts[1]
            while ($payload.Length % 4 -ne 0) { $payload += "=" }
            $bytes = [System.Convert]::FromBase64String($payload)
            $json = [System.Text.Encoding]::UTF8.GetString($bytes)
            $data = $json | ConvertFrom-Json
            
            Write-Host "JWT中的用户信息:" -ForegroundColor Cyan
            Write-Host "  用户ID: $($data.user_id)" -ForegroundColor White
            Write-Host "  用户名: $($data.username)" -ForegroundColor White
            Write-Host "  角色: $($data.role)" -ForegroundColor White
            
            # 3. 获取API用户信息
            Write-Host "`n3. 获取API用户信息..." -ForegroundColor Yellow
            $headers = @{ "Authorization" = "Bearer $token" }
            $userInfo = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/user/me" -Headers $headers
            
            Write-Host "API返回的用户信息:" -ForegroundColor Cyan
            Write-Host "  用户ID: $($userInfo.user_id)" -ForegroundColor White
            Write-Host "  用户名: $($userInfo.username)" -ForegroundColor White
            Write-Host "  角色: $($userInfo.current_role)" -ForegroundColor White
            
            # 4. 比较结果
            Write-Host "`n4. 比较结果:" -ForegroundColor Yellow
            if ($data.user_id -eq $userInfo.user_id) {
                Write-Host "  用户ID匹配" -ForegroundColor Green
            } else {
                Write-Host "  用户ID不匹配" -ForegroundColor Red
                Write-Host "  JWT: $($data.user_id)" -ForegroundColor Red
                Write-Host "  API: $($userInfo.user_id)" -ForegroundColor Red
            }
        } else {
            Write-Host "JWT格式错误" -ForegroundColor Red
        }
    } else {
        Write-Host "未获取到token" -ForegroundColor Red
    }
} catch {
    Write-Host "错误: $($_.Exception.Message)" -ForegroundColor Red
}