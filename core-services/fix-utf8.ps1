# Simple UTF-8 fix script
$files = Get-ChildItem -Path "." -Recurse -Include "*.go" | Where-Object { $_.FullName -notlike "*\vendor\*" }

foreach ($file in $files) {
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $originalContent = $content
    
    # Remove problematic UTF-8 characters
    $content = $content -replace '[^\x00-\x7F\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]', ''
    
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -Encoding UTF8 -NoNewline
        Write-Host "Fixed: $($file.FullName)"
    }
}

Write-Host "UTF-8 encoding fix completed!"