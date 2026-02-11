# PowerShell build script for Windows
# 用于编译 Simple Exam 并注入版本信息

param(
    [string]$Output = "exam-system.exe",
    [string]$Version = "v0.1.2"
)

Write-Host "Building Simple Exam..." -ForegroundColor Green

# 获取 Git 信息
$CommitHash = "unknown"
$BuildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

try {
    $CommitHash = git rev-parse HEAD 2>$null
    if ($LASTEXITCODE -ne 0) {
        $CommitHash = "unknown"
    }
} catch {
    Write-Host "Warning: Unable to get git commit hash" -ForegroundColor Yellow
}

Write-Host "Version:     $Version" -ForegroundColor Cyan
Write-Host "Commit:      $CommitHash" -ForegroundColor Cyan
Write-Host "Build Time:  $BuildTime" -ForegroundColor Cyan

# 构建 ldflags
$ldflags = "-X 'main.Version=$Version' -X 'main.CommitHash=$CommitHash' -X 'main.BuildTime=$BuildTime' -s -w"

# 编译
Write-Host "`nCompiling..." -ForegroundColor Green
go build -ldflags $ldflags -o $Output .

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild successful! Output: $Output" -ForegroundColor Green
    
    # 显示文件信息
    $fileInfo = Get-Item $Output
    Write-Host "File size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB" -ForegroundColor Cyan
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}
