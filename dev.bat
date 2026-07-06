@echo off
cd /d "%~dp0"
echo Installing frontend dependencies...
cd frontend
call npm install
cd ..
echo Starting dev mode...
wails dev
