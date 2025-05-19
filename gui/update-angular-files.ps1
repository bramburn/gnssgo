# Update Angular files script

# Update app.component.html
Get-Content -Path "frontend/src/app/app.component.html.part1" | Set-Content -Path "frontend/src/app/app.component.html"
Write-Host "Updated app.component.html"

# Update app.config.ts
Get-Content -Path "frontend/src/app/app.config.ts.part1" | Set-Content -Path "frontend/src/app/app.config.ts"
Write-Host "Updated app.config.ts"

# Update index.html
Get-Content -Path "frontend/src/index.html.part1" | Set-Content -Path "frontend/src/index.html"
Write-Host "Updated index.html"

Write-Host "All files updated successfully!"
