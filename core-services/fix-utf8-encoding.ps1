# Fix UTF-8 encoding issues

Write-Host "Starting UTF-8 encoding fix..."

$goFiles = Get-ChildItem -Path "." -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*\.git\*" }

$processedCount = 0
$fixedCount = 0

foreach ($file in $goFiles) {
    $processedCount++
    Write-Progress -Activity "Processing files" -Status "Processing: $($file.Name)" -PercentComplete (($processedCount / $goFiles.Count) * 100)
    
    try {
        $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
        
        if ($content -match '[^\x00-\x7F]' -or $content -match '�') {
            Write-Host "Fixing file: $($file.FullName)"
            
            # Remove broken UTF-8 characters
            $content = $content -replace '�', ''
            
            # Save file with UTF-8 encoding without BOM
            $utf8NoBom = New-Object System.Text.UTF8Encoding $false
            [System.IO.File]::WriteAllText($file.FullName, $content, $utf8NoBom)
            
            $fixedCount++
        }
    }
    catch {
        Write-Warning "Error processing file $($file.FullName): $($_.Exception.Message)"
    }
}

Write-Host "Processing complete!"
Write-Host "Processed $processedCount files"
Write-Host "Fixed $fixedCount files"