@echo off
title CAPA Downloader

REM Add Git's bin directory to PATH for dependencies
set PATH=%PATH%;%LOCALAPPDATA%\Programs\Git\mingw64\bin

echo Options:
echo   1. Download all open RCH CAPAs
echo   2. Download all closed RCH CAPAs
echo   3. Download a single CAPA
echo.
set /p CHOICE=Choose (1/2/3):

set /p OUTPUT_DIR=Enter output folder:

if "%CHOICE%"=="1" (
    "%~dp0capa-downloader.exe" --output "%OUTPUT_DIR%"
) else if "%CHOICE%"=="2" (
    "%~dp0capa-downloader.exe" --output "%OUTPUT_DIR%" --closed
) else if "%CHOICE%"=="3" (
    set /p CAPA_NUM=Enter CAPA number (e.g., CAPA-2025-000054):
    "%~dp0capa-downloader.exe" --output "%OUTPUT_DIR%" --capa %CAPA_NUM%
) else (
    echo Invalid choice.
)

echo.
pause
