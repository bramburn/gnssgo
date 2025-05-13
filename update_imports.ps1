$files = Get-ChildItem -Path unittest -Filter *.go -Recurse | Select-String -Pattern '"gnssgo"' | Select-Object -ExpandProperty Path -Unique

foreach ($file in $files) {
    Write-Host "Updating $file"
    $content = Get-Content $file -Raw
    $updatedContent = $content -replace '"gnssgo"', '"github.com/bramburn/gnssgo"'
    Set-Content -Path $file -Value $updatedContent
}

Write-Host "All files updated!"
