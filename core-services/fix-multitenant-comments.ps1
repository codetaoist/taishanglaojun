# 批量修复多租户模块中的注释
$directory = "D:\work\taishanglaojun\core-services\multi-tenant"

# 获取所有Go文件
$goFiles = Get-ChildItem -Path $directory -Filter "*.go" -Recurse

foreach ($file in $goFiles) {
    Write-Host "Processing: $($file.FullName)"
    
    # 读取文件内容
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    
    # 修复常见的注释问题
    $content = $content -replace '// \?$', '// 待完善'
    $content = $content -replace '// $', '// 待完善'
    $content = $content -replace '//\s*$', '// 待完善'
    $content = $content -replace '// ID\?', '// 租户ID检查'
    $content = $content -replace '// \?', '// 待完善'
    $content = $content -replace 'Roles:\s+\[\]string\{\},\s*//\s*$', 'Roles:       []string{}, // 角色列表'
    $content = $content -replace 'Permissions:\s+\[\]string\{\},\s*//\s*$', 'Permissions: []string{}, // 权限列表'
    
    # 写回文件
    [System.IO.File]::WriteAllText($file.FullName, $content, [System.Text.Encoding]::UTF8)
}

Write-Host "多租户模块注释修复完成"