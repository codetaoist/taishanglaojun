# 简化的测试数据添加脚本
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYmY4NTYyMDYtYjIwMC00MTk1LWE3NGMtMjY2ODE0OTcyNTkxIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2NjIzOTQsIm5iZiI6MTc2MDU3NTk5NCwiaWF0IjoxNzYwNTc1OTk0fQ.Z8Lia-sFKx-wvAko3PSzrIyBGcxc1aphnV4aRKTTE9w"
$baseUrl = "http://localhost:8080/api/v1"
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "Adding test permissions..." -ForegroundColor Yellow

# 添加权限
$permissions = @(
    @{ name = "content_read"; description = "View content"; resource = "content"; action = "read" },
    @{ name = "content_write"; description = "Edit content"; resource = "content"; action = "write" },
    @{ name = "content_publish"; description = "Publish content"; resource = "content"; action = "publish" },
    @{ name = "user_read"; description = "View user info"; resource = "user"; action = "read" },
    @{ name = "user_write"; description = "Edit user info"; resource = "user"; action = "write" },
    @{ name = "analytics_read"; description = "View analytics"; resource = "analytics"; action = "read" },
    @{ name = "analytics_write"; description = "Edit analytics"; resource = "analytics"; action = "write" },
    @{ name = "report_generate"; description = "Generate reports"; resource = "report"; action = "generate" },
    @{ name = "ticket_read"; description = "View tickets"; resource = "ticket"; action = "read" },
    @{ name = "ticket_write"; description = "Handle tickets"; resource = "ticket"; action = "write" }
)

foreach ($permission in $permissions) {
    try {
        $body = $permission | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/permissions" -Method POST -Headers $headers -Body $body
        Write-Host "✓ Added permission: $($permission.name)" -ForegroundColor Green
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "⚠ Permission already exists: $($permission.name)" -ForegroundColor Yellow
        }
        else {
            Write-Host "✗ Failed to add permission: $($permission.name) - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 200
}

Write-Host "`nAdding test roles..." -ForegroundColor Yellow

# 添加角色
$roles = @(
    @{
        name = "Content Manager"
        code = "content_manager"
        description = "Manages website content"
        type = "functional"
        level = 3
        is_active = $true
        permissions = @("content_read", "content_write", "content_publish")
    },
    @{
        name = "User Operator"
        code = "user_operator"
        description = "Manages users"
        type = "functional"
        level = 2
        is_active = $true
        permissions = @("user_read", "user_write", "analytics_read")
    },
    @{
        name = "Data Analyst"
        code = "data_analyst"
        description = "Analyzes data and generates reports"
        type = "data"
        level = 4
        is_active = $true
        permissions = @("analytics_read", "analytics_write", "report_generate")
    },
    @{
        name = "Customer Service"
        code = "customer_service"
        description = "Handles customer support"
        type = "functional"
        level = 1
        is_active = $true
        permissions = @("user_read", "ticket_read", "ticket_write")
    },
    @{
        name = "QA Engineer"
        code = "qa_engineer"
        description = "Quality assurance and testing"
        type = "functional"
        level = 3
        is_active = $true
        permissions = @("content_read", "user_read")
    }
)

foreach ($role in $roles) {
    try {
        $body = $role | ConvertTo-Json -Depth 10
        $response = Invoke-RestMethod -Uri "$baseUrl/roles" -Method POST -Headers $headers -Body $body
        Write-Host "✓ Added role: $($role.name)" -ForegroundColor Green
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "⚠ Role already exists: $($role.name)" -ForegroundColor Yellow
        }
        else {
            Write-Host "✗ Failed to add role: $($role.name) - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 200
}

Write-Host "`nGetting data statistics..." -ForegroundColor Yellow

try {
    $rolesResponse = Invoke-RestMethod -Uri "$baseUrl/roles" -Headers $headers
    $permissionsResponse = Invoke-RestMethod -Uri "$baseUrl/permissions" -Headers $headers
    
    Write-Host "📊 Data Statistics:" -ForegroundColor Cyan
    Write-Host "   - Total roles: $($rolesResponse.data.Count)" -ForegroundColor White
    Write-Host "   - Total permissions: $($permissionsResponse.data.Count)" -ForegroundColor White
    
    $activeRoles = ($rolesResponse.data | Where-Object { $_.is_active -eq $true }).Count
    Write-Host "   - Active roles: $activeRoles" -ForegroundColor White
}
catch {
    Write-Host "Failed to get statistics: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🎉 Test data setup completed!" -ForegroundColor Green
Write-Host "You can now visit http://localhost:5173/admin/roles to view the role management page" -ForegroundColor Cyan