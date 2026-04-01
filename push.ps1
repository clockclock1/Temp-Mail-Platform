param(
    [string]$message = "Update: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')",
    [switch]$useProxy,
    [string]$proxy = "http://127.0.0.1:10700"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Git Push Script - Temp-Mail-Platform" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($useProxy) {
    Write-Host "[0/5] Setting up proxy..." -ForegroundColor Yellow
    Write-Host "Proxy: $proxy" -ForegroundColor Gray
    git config --global http.proxy $proxy
    git config --global https.proxy $proxy
    Write-Host "Proxy configured successfully!" -ForegroundColor Green
    Write-Host ""
}

Write-Host "[1/5] Checking git status..." -ForegroundColor Yellow
git status

Write-Host ""
Write-Host "[2/5] Adding all changes..." -ForegroundColor Yellow
git add -A

Write-Host ""
Write-Host "[3/5] Committing changes..." -ForegroundColor Yellow
Write-Host "Commit message: $message" -ForegroundColor Gray
git commit -m "$message"

Write-Host ""
Write-Host "[4/5] Pushing to remote repository..." -ForegroundColor Yellow
Write-Host "Target: https://github.com/clockclock1/Temp-Mail-Platform.git" -ForegroundColor Gray

$branch = git branch --show-current
if ([string]::IsNullOrEmpty($branch)) {
    $branch = "main"
    git branch -M $branch
}

git push -u origin $branch --force

if ($useProxy) {
    Write-Host ""
    Write-Host "[5/5] Cleaning up proxy..." -ForegroundColor Yellow
    git config --global --unset http.proxy
    git config --global --unset https.proxy
    Write-Host "Proxy settings cleared!" -ForegroundColor Green
}

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
