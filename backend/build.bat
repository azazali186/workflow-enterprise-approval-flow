@echo off
cd /d E:\Aeroxe-Ecosystem\phase-1\workflow-enterprise\approval-flow\backend

echo Cleaning Go caches...
rmdir /s /q D:\Go\pkg\mod\cache
go clean -modcache

echo Setting up fake git repo in module cache...
set CACHEDIR=D:\Go\pkg\mod\cache\vcs\624b6e86962790e5fb3b4a5f6aceb6dd75f2b56a1ea35c9fa20e43a957824d59
rmdir /s /q %CACHEDIR%
mkdir %CACHEDIR%
cd %CACHEDIR%
git init
git remote add origin file:///E:/Aeroxe-Ecosystem/phase-1/workflow-enterprise/approval-flow/backend
xcopy /E /I /Y "E:\Aeroxe-Ecosystem\phase-1\workflow-enterprise\approval-flow\backend\*" "%CACHEDIR%\"
git add -A
git commit -m "initial"

echo Building server...
cd E:\Aeroxe-Ecosystem\phase-1\workflow-enterprise\approval-flow\backend
set GONOSUMCHECK=*
set GONOSUMDB=*
set GOFLAGS=-mod=mod
go build ./cmd/server

echo Build complete!
pause
