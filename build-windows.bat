@echo off
setlocal
if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w" -o dist\ClaudeWakeUp.exe .
if errorlevel 1 exit /b 1
echo Built dist\ClaudeWakeUp.exe
