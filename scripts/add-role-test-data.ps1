# 角色管理测试数据添加脚本
# PowerShell版本，支持Windows环境

param(
    [string]$BaseUrl = "http://localhost:8080/api/v1",
    [string]$Username = "admin",
    [string]$Password = "admin123"
)

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# API请求函数
function Invoke-APIRequest {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )
    
    try {
        $requestParams = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            ContentType = "application/json"
        }
        
        if ($Body) {
            $requestParams.Body = ($Body | ConvertTo-Json -Depth 10)
        }
        
        $response = Invoke-RestMethod @requestParams
        return @{
            Success = $true
            Data = $response
            StatusCode = 200
        }
    }
    catch {
        $statusCode = if ($_.Exception.Response) { $_.Exception.Response.StatusCode.value__ } else { 0 }
        $errorMessage = if ($_.Exception.Response) { 
            try {
                $errorStream = $_.Exception.Response.GetResponseStream()
                $reader = New-Object System.IO.StreamReader($errorStream)
                $errorBody = $reader.ReadToEnd()
                $errorObj = $errorBody | ConvertFrom-Json
                $errorObj.message
            }
            catch {
                $_.Exception.Message
            }
        } else {
            $_.Exception.Message
        }
        
        return @{
            Success = $false
            Error = $errorMessage
            StatusCode = $statusCode
        }
    }
}

# 获取授权token
function Get-AuthToken {
    Write-ColorOutput "🔑 获取授权token..." "Yellow"
    
    $createUserUrl = "$BaseUrl/auth/test-user?username=$Username&role=admin&password=$Password"
    $result = Invoke-APIRequest -Url $createUserUrl -Method "POST"
    
    if ($result.Success) {
        $token = $result.Data.data.token
        Write-ColorOutput "✅ 成功获取授权token" "Green"
        return $token
    }
    else {
        Write-ColorOutput "❌ 获取授权token失败: $($result.Error)" "Red"
        return $null
    }
}

# 测试权限数据
$testPermissions = @(
    @{ name = "content_read"; description = "查看内容"; resource = "content"; action = "read" },
    @{ name = "content_write"; description = "编辑内容"; resource = "content"; action = "write" },
    @{ name = "content_publish"; description = "发布内容"; resource = "content"; action = "publish" },
    @{ name = "user_read"; description = "查看用户信息"; resource = "user"; action = "read" },
    @{ name = "user_write"; description = "编辑用户信息"; resource = "user"; action = "write" },
    @{ name = "analytics_read"; description = "查看分析数据"; resource = "analytics"; action = "read" },
    @{ name = "analytics_write"; description = "编辑分析配置"; resource = "analytics"; action = "write" },
    @{ name = "report_generate"; description = "生成报告"; resource = "report"; action = "generate" },
    @{ name = "ticket_read"; description = "查看工单"; resource = "ticket"; action = "read" },
    @{ name = "ticket_write"; description = "处理工单"; resource = "ticket"; action = "write" },
    @{ name = "finance_read"; description = "查看财务数据"; resource = "finance"; action = "read" },
    @{ name = "audit_read"; description = "查看审计日志"; resource = "audit"; action = "read" },
    @{ name = "audit_write"; description = "编辑审计配置"; resource = "audit"; action = "write" },
    @{ name = "product_read"; description = "查看产品信息"; resource = "product"; action = "read" },
    @{ name = "product_write"; description = "编辑产品信息"; resource = "product"; action = "write" },
    @{ name = "test_read"; description = "查看测试结果"; resource = "test"; action = "read" },
    @{ name = "test_write"; description = "执行测试"; resource = "test"; action = "write" },
    @{ name = "bug_report"; description = "报告缺陷"; resource = "bug"; action = "report" },
    @{ name = "marketing_read"; description = "查看营销数据"; resource = "marketing"; action = "read" },
    @{ name = "marketing_write"; description = "编辑营销活动"; resource = "marketing"; action = "write" },
    @{ name = "security_read"; description = "查看安全日志"; resource = "security"; action = "read" },
    @{ name = "security_write"; description = "配置安全策略"; resource = "security"; action = "write" },
    @{ name = "system_monitor"; description = "系统监控"; resource = "system"; action = "monitor" },
    @{ name = "basic_read"; description = "基础查看权限"; resource = "basic"; action = "read" }
)

# 测试角色数据
$testRoles = @(
    @{
        name = "Content Manager"
        code = "content_manager"
        description = "Responsible for creating, editing and publishing website content, managing article categories and tags, reviewing user-submitted content"
        type = "functional"
        level = 3
        is_active = $true
        permissions = @("content_read", "content_write", "content_publish")
    },
    @{
        name = "User Operator"
        code = "user_operator"
        description = "Responsible for user management, user behavior analysis, user feedback processing, maintaining good user experience"
        type = "functional"
        level = 2
        is_active = $true
        permissions = @("user_read", "user_write", "analytics_read")
    },
    @{
        name = "Data Analyst"
        code = "data_analyst"
        description = "Responsible for data collection, analysis and report generation, providing data support for business decisions"
        type = "data"
        level = 4
        is_active = $true
        permissions = @("analytics_read", "analytics_write", "report_generate")
    },
    @{
        name = "Customer Service"
        code = "customer_service"
        description = "Handle user inquiries, complaints and suggestions, provide technical support and problem solutions"
        type = "functional"
        level = 1
        is_active = $true
        permissions = @("user_read", "ticket_read", "ticket_write")
    },
    @{
        name = "Financial Auditor"
        code = "financial_auditor"
        description = "Responsible for financial data audit, compliance check and risk control, ensuring financial security"
        type = "data"
        level = 5
        is_active = $true
        permissions = @("finance_read", "audit_read", "audit_write")
    },
    @{
        name = "Product Manager"
        code = "product_manager"
        description = "Responsible for product planning, requirement analysis and feature design, coordinating development team to achieve product goals"
        type = "custom"
        level = 4
        is_active = $true
        permissions = @("product_read", "product_write", "analytics_read", "user_read")
    },
    @{
        name = "QA Engineer"
        code = "qa_engineer"
        description = "Responsible for software testing, quality assurance and defect tracking, ensuring product quality"
        type = "functional"
        level = 3
        is_active = $true
        permissions = @("test_read", "test_write", "bug_report")
    },
    @{
        name = "Marketing Specialist"
        code = "marketing_specialist"
        description = "Responsible for marketing activity planning, brand promotion and user growth, improving product visibility"
        type = "custom"
        level = 2
        is_active = $true
        permissions = @("marketing_read", "marketing_write", "analytics_read")
    },
    @{
        name = "Security Officer"
        code = "security_officer"
        description = "Responsible for system security monitoring, vulnerability repair and security policy formulation, ensuring system security"
        type = "system"
        level = 5
        is_active = $true
        permissions = @("security_read", "security_write", "audit_read", "system_monitor")
    },
    @{
        name = "Temp Visitor"
        code = "temp_visitor"
        description = "Temporary access permission for short-term partners or external consultants with limited access"
        type = "custom"
        level = 1
        is_active = $false
        permissions = @("basic_read")
    }
)

# 添加权限数据
function Add-Permissions {
    param([string]$Token)
    
    Write-ColorOutput "🔐 开始添加权限数据..." "Yellow"
    
    $headers = @{
        "Authorization" = "Bearer $Token"
    }
    
    $successCount = 0
    $skipCount = 0
    $errorCount = 0
    
    foreach ($permission in $testPermissions) {
        $result = Invoke-APIRequest -Url "$BaseUrl/permissions" -Method "POST" -Headers $headers -Body $permission
        
        if ($result.Success) {
            Write-ColorOutput "✅ 权限添加成功: $($permission.name)" "Green"
            $successCount++
        }
        elseif ($result.StatusCode -eq 409 -or $result.Error -like "*already exists*" -or $result.Error -like "*重复*") {
            Write-ColorOutput "⚠️  权限已存在: $($permission.name)" "Yellow"
            $skipCount++
        }
        else {
            Write-ColorOutput "❌ 权限添加失败: $($permission.name) - $($result.Error)" "Red"
            $errorCount++
        }
        
        # 避免请求过快
        Start-Sleep -Milliseconds 100
    }
    
    Write-ColorOutput "📊 权限数据添加完成 - 成功: $successCount, 跳过: $skipCount, 失败: $errorCount" "Cyan"
}

# 添加角色数据
function Add-Roles {
    param([string]$Token)
    
    Write-ColorOutput "👥 开始添加角色数据..." "Yellow"
    
    $headers = @{
        "Authorization" = "Bearer $Token"
    }
    
    $successCount = 0
    $skipCount = 0
    $errorCount = 0
    
    foreach ($role in $testRoles) {
        $result = Invoke-APIRequest -Url "$BaseUrl/roles" -Method "POST" -Headers $headers -Body $role
        
        if ($result.Success) {
            Write-ColorOutput "✅ 角色添加成功: $($role.name) ($($role.code))" "Green"
            $successCount++
        }
        elseif ($result.StatusCode -eq 409 -or $result.Error -like "*already exists*" -or $result.Error -like "*重复*") {
            Write-ColorOutput "⚠️  角色已存在: $($role.name) ($($role.code))" "Yellow"
            $skipCount++
        }
        else {
            Write-ColorOutput "❌ 角色添加失败: $($role.name) - $($result.Error)" "Red"
            $errorCount++
        }
        
        # 避免请求过快
        Start-Sleep -Milliseconds 100
    }
    
    Write-ColorOutput "📊 角色数据添加完成 - 成功: $successCount, 跳过: $skipCount, 失败: $errorCount" "Cyan"
}

# 获取数据统计
function Get-DataStats {
    param([string]$Token)
    
    Write-ColorOutput "📊 获取数据统计..." "Yellow"
    
    $headers = @{
        "Authorization" = "Bearer $Token"
    }
    
    try {
        $rolesResult = Invoke-APIRequest -Url "$BaseUrl/roles" -Headers $headers
        $permissionsResult = Invoke-APIRequest -Url "$BaseUrl/permissions" -Headers $headers
        
        if ($rolesResult.Success -and $permissionsResult.Success) {
            $roles = $rolesResult.Data.data
            $permissions = $permissionsResult.Data.data
            
            $activeRoles = ($roles | Where-Object { $_.is_active -eq $true }).Count
            $roleTypes = ($roles | Group-Object type).Count
            
            Write-ColorOutput "📈 数据统计信息:" "Cyan"
            Write-ColorOutput "   - 角色总数: $($roles.Count)" "White"
            Write-ColorOutput "   - 权限总数: $($permissions.Count)" "White"
            Write-ColorOutput "   - 启用角色: $activeRoles" "White"
            Write-ColorOutput "   - 角色类型: $roleTypes" "White"
            
            # 按类型统计角色
            $rolesByType = $roles | Group-Object type
            Write-ColorOutput "   - 角色类型分布:" "White"
            foreach ($group in $rolesByType) {
                Write-ColorOutput "     * $($group.Name): $($group.Count)" "Gray"
            }
        }
        else {
            Write-ColorOutput "❌ 获取统计信息失败" "Red"
        }
    }
    catch {
        Write-ColorOutput "❌ 获取统计信息失败: $($_.Exception.Message)" "Red"
    }
}

# 主函数
function Main {
    Write-ColorOutput "🚀 开始添加角色管理测试数据..." "Cyan"
    Write-ColorOutput "📍 API地址: $BaseUrl" "Gray"
    Write-ColorOutput "👤 用户名: $Username" "Gray"
    Write-ColorOutput ""
    
    # 获取授权token
    $token = Get-AuthToken
    if (-not $token) {
        Write-ColorOutput "❌ 无法获取授权token，退出" "Red"
        return
    }
    
    Write-ColorOutput ""
    
    # 添加权限数据
    Add-Permissions -Token $token
    Write-ColorOutput ""
    
    # 添加角色数据
    Add-Roles -Token $token
    Write-ColorOutput ""
    
    # 获取统计信息
    Get-DataStats -Token $token
    Write-ColorOutput ""
    
    Write-ColorOutput "🎉 测试数据添加完成！" "Green"
    Write-ColorOutput ""
    Write-ColorOutput "💡 提示:" "Yellow"
    Write-ColorOutput "   - 可以访问 http://localhost:5173/admin/roles 查看角色管理页面" "White"
    Write-ColorOutput "   - 测试数据包含了不同类型、级别和状态的角色" "White"
    Write-ColorOutput "   - 可以测试搜索、筛选、编辑、删除等功能" "White"
}

# 运行主函数
Main