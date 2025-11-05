# Build script for QOTP Wireshark Plugin
# This script builds both the Go DLL and the C Lua module

Write-Host "Building QOTP Wireshark Decryption Plugin..." -ForegroundColor Green
Write-Host ""

# Step 1: Build Go shared library
Write-Host "Step 1: Building Go shared library (qotp_crypto.dll)..." -ForegroundColor Cyan
$env:CGO_ENABLED = "1"
go build -buildmode=c-shared -o qotp_crypto.dll qotp_export.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "qotp_crypto.dll created" -ForegroundColor Green
Write-Host ""

# Step 2: Build C wrapper as Lua module
Write-Host "Step 2: Building C Lua module (qotp_decrypt.dll)..." -ForegroundColor Cyan
$luaInclude = "C:\Users\gian\sa\wireshark\wireshark-libs\lua-5.4.6-unicode-win64-vc14\include"
$luaLib = "C:\Users\gian\sa\wireshark\wireshark-libs\lua-5.4.6-unicode-win64-vc14\lua54.lib"
$currentDir = (Get-Location).Path

$buildCmd = "vcvars64.bat & cd /d `"$currentDir`" & cl /LD /O2 /TP qotp_decrypt.c /I`"`"$luaInclude`"`" /link `"`"$luaLib`"`" User32.lib /OUT:qotp_decrypt.dll"

cmd /c $buildCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: C compilation failed!" -ForegroundColor Red
    Write-Host "Make sure vcvars64.bat is in your PATH or run from VS Developer Command Prompt" -ForegroundColor Yellow
    exit 1
}

Write-Host "qotp_decrypt.dll created" -ForegroundColor Green
Write-Host ""

# Step 3: Show deployment instructions
Write-Host "Build complete! Deployment instructions:" -ForegroundColor Green
Write-Host ""
Write-Host "1. Copy these files to C:\Program Files\Wireshark\:" -ForegroundColor Yellow
Write-Host "   - qotp_decrypt.dll"
Write-Host "   - qotp_crypto.dll"
Write-Host ""
Write-Host "2. Copy qotp_dissector.lua to:" -ForegroundColor Yellow  
Write-Host "   C:\Users\$env:USERNAME\AppData\Roaming\Wireshark\plugins\4.6\"
Write-Host ""
Write-Host "3. Restart Wireshark" -ForegroundColor Yellow
Write-Host ""

# Offer to copy files
$response = Read-Host "Copy files now? (y/n)"
if ($response -eq 'y') {
    Write-Host "Copying DLLs..." -ForegroundColor Cyan
    Copy-Item qotp_decrypt.dll "C:\Program Files\Wireshark\" -Force
    Copy-Item qotp_crypto.dll "C:\Program Files\Wireshark\" -Force
    
    $pluginDir = "C:\Users\$env:USERNAME\AppData\Roaming\Wireshark\plugins\4.6"
    if (!(Test-Path $pluginDir)) {
        New-Item -ItemType Directory -Path $pluginDir -Force | Out-Null
    }
    Copy-Item qotp_dissector.lua $pluginDir -Force
    
    Write-Host "Files copied successfully!" -ForegroundColor Green
}
