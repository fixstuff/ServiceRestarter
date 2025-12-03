@echo off
REM ============================================================================
REM Stop Service Restarter
REM ============================================================================
REM
REM This batch file stops the service-restarter.exe program
REM
REM ============================================================================

echo Stopping Service Restarter...
taskkill /F /IM service-restarter.exe

if %errorLevel% equ 0 (
    echo Service Restarter has been stopped.
) else (
    echo Service Restarter is not running or could not be stopped.
)

pause
