param(
    [string]$message = "Update: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Git Push Script - Temp-Mail-Platform" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[1/4] Checking git status..." -ForegroundColor Yellow
git status

Write-Host ""
Write-Host "[2/4] Adding all changes..." -ForegroundColor Yellow
git add -A

Write-Host ""
Write-Host "[3/4] Committing changes..." -ForegroundColor Yellow
Write-Host "Commit message: $message" -ForegroundColor Gray
git commit -m "$message"

Write-Host ""
Write-Host "[4/4] Pushing to remote repository..." -ForegroundColor Yellow
Write-Host "Target: https://github.com/santi163/Temp-Mail-Platform.git" -ForegroundColor Gray

$branch = git branch --show-current
if ([string]::IsNullOrEmpty($branch)) {
    $branch = "main"
    git branch -M $branch
}

git push -u origin $branch --force

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Push completed successfully!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "  Push failed! Please check errors above." -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
}
