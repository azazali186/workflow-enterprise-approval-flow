@echo off
cd /d E:\Aeroxe-Ecosystem\phase-1\workflow-enterprise\approval-flow\backend

echo Cleaning Go caches...
rmdir /s /q D:\Go\pkg\mod\cache
go clean -modcache
go clean -cache

echo Running go mod tidy...
set GONOSUMCHECK=*
set GONOSUMDB=*
set GOFLAGS=-mod=mod
go mod tidy

echo Fixing git remote in module cache...
set CACHEDIR=D:\Go\pkg\mod\cache\vcs\624b6e86962790e5fb3b4a5f6aceb6dd75f2b56a1ea35c9fa20e43a957824d59
cd %CACHEDIR%
git remote set-url origin file:///E:/Aeroxe-Ecosystem/phase-1/workflow-enterprise/approval-flow/backend
echo Git remote set to local path

echo Building server...
cd E:\Aeroxe-Ecosystem\phase-1\workflow-enterprise\approval-flow\backend
go build ./cmd/server

echo Build complete!
pause
