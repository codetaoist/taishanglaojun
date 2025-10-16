# 添加角色管理测试数据 - 正确版本
$ErrorActionPreference = "Continue"

# 配置
$baseUrl = "http://localhost:8080/api/v1"
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZGI0NzIwNmItNWIzZi00MjE5LWE3YWMtNzI2MjdkMTMyNjNkIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsInJvbGUiOiJVU0VSIiwibGV2ZWwiOjUsImlzcyI6InRhaXNoYW5nLWxhb2p1biIsImV4cCI6MTc2MDY2MjY0OCwibmJmIjoxNzYwNTc2MjQ4LCJpYXQiOjE3NjA1NzYyNDh9.ZYgqro1CoMXOrNz0dLcafhKeD33L_nhu8-a0zJ7ajgE"
$tenantId = "default"

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "开始添加角色管理测试数据..." -ForegroundColor Green

# 测试权限数据
$testPermissions = @(
    @{
        name = "View Content"
        code = "content.view"
        description = "View content items"
        category = "content"
        resource = "content"
        action = "view"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "Create Content"
        code = "content.create"
        description = "Create new content items"
        category = "content"
        resource = "content"
        action = "create"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "Edit Content"
        code = "content.edit"
        description = "Edit existing content items"
        category = "content"
        resource = "content"
        action = "edit"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "Delete Content"
        code = "content.delete"
        description = "Delete content items"
        category = "content"
        resource = "content"
        action = "delete"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "Manage Users"
        code = "user.manage"
        description = "Manage user accounts"
        category = "user"
        resource = "user"
        action = "manage"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "View Reports"
        code = "report.view"
        description = "View system reports"
        category = "report"
        resource = "report"
        action = "view"
        effect = "allow"
        tenant_id = $tenantId
    },
    @{
        name = "Export Data"
        code = "data.export"
        description = "Export system data"
        category = "data"
        resource = "data"
        action = "export"
        effect = "allow"
        tenant_id = $tenantId
    }
)

# 测试角色数据
$testRoles = @(
    @{
        name = "Content Manager"
        code = "content_manager"
        description = "Manages content creation and editing"
        type = "custom"
        level = 2
        tenant_id = $tenantId
    },
    @{
        name = "User Operator"
        code = "user_operator"
        description = "Manages user accounts and operations"
        type = "custom"
        level = 3
        tenant_id = $tenantId
    },
    @{
        name = "Data Analyst"
        code = "data_analyst"
        description = "Analyzes data and generates reports"
        type = "custom"
        level = 1
        tenant_id = $tenantId
    },
    @{
        name = "Content Editor"
        code = "content_editor"
        description = "Edits and reviews content"
        type = "custom"
        level = 1
        tenant_id = $tenantId
    },
    @{
        name = "System Monitor"
        code = "system_monitor"
        description = "Monitors system performance and health"
        type = "custom"
        level = 2
        tenant_id = $tenantId
    }
)

# 添加权限
Write-Host "`n添加权限数据..." -ForegroundColor Yellow
$addedPermissions = @()

foreach ($permission in $testPermissions) {
    try {
        $body = $permission | ConvertTo-Json -Depth 3
        $response = Invoke-RestMethod -Uri "$baseUrl/permissions" -Method POST -Headers $headers -Body $body
        Write-Host "✅ 权限添加成功: $($permission.name)" -ForegroundColor Green
        $addedPermissions += $response
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "⚠️  权限已存在: $($permission.name)" -ForegroundColor Yellow
        }
        else {
            Write-Host "❌ 权限添加失败: $($permission.name) - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 200
}

# 添加角色
Write-Host "`n添加角色数据..." -ForegroundColor Yellow
$addedRoles = @()

foreach ($role in $testRoles) {
    try {
        $body = $role | ConvertTo-Json -Depth 3
        $response = Invoke-RestMethod -Uri "$baseUrl/roles" -Method POST -Headers $headers -Body $body
        Write-Host "✅ 角色添加成功: $($role.name)" -ForegroundColor Green
        $addedRoles += $response
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "⚠️  角色已存在: $($role.name)" -ForegroundColor Yellow
        }
        else {
            Write-Host "❌ 角色添加失败: $($role.name) - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 200
}

# 获取统计信息
Write-Host "`n获取数据统计..." -ForegroundColor Yellow
try {
    $rolesResponse = Invoke-RestMethod -Uri "$baseUrl/roles" -Method GET -Headers $headers
    $permissionsResponse = Invoke-RestMethod -Uri "$baseUrl/permissions" -Method GET -Headers $headers
    
    Write-Host "`n📊 数据统计:" -ForegroundColor Cyan
    Write-Host "   总角色数: $($rolesResponse.total)" -ForegroundColor White
    Write-Host "   总权限数: $($permissionsResponse.total)" -ForegroundColor White
    Write-Host "   本次添加的权限: $($addedPermissions.Count)" -ForegroundColor White
    Write-Host "   本次添加的角色: $($addedRoles.Count)" -ForegroundColor White
}
catch {
    Write-Host "❌ 获取统计信息失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n✅ 测试数据添加完成!" -ForegroundColor Green