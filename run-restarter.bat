@echo off
REM ============================================================================
REM Service Restarter - Batch File Runner
REM ============================================================================
REM
REM This batch file demonstrates how to run the service-restarter program
REM with various command-line options.
REM
REM PREREQUISITES:
REM   - Run this batch file as Administrator (required for service control)
REM   - Ensure service-restarter.exe is compiled and in the same directory
REM
REM ============================================================================
REM COMMAND-LINE FLAGS:
REM ============================================================================
REM
REM -service <name>     (REQUIRED) Name of the Windows service to restart
REM                     Example: "Spooler" for Print Spooler service
REM                     Example: "wuauserv" for Windows Update service
REM                     Example: "MSSQLSERVER" for SQL Server service
REM
REM -interval <minutes> (OPTIONAL) Time between restarts in minutes
REM                     Default: 60 minutes (1 hour)
REM                     Example: -interval 30 for every 30 minutes
REM                     Example: -interval 120 for every 2 hours
REM
REM -hold <seconds>     (OPTIONAL) Time to wait after stopping before starting
REM                     Default: 2 seconds
REM                     Example: -hold 5 for 5 second pause
REM                     Example: -hold 10 for 10 second pause
REM
REM ============================================================================
REM FINDING YOUR SERVICE NAME:
REM ============================================================================
REM
REM To find the exact name of a service on your system:
REM   1. Open Services (services.msc)
REM   2. Double-click the service you want
REM   3. Copy the "Service name" field (NOT the display name)
REM
REM OR use this command in a new command prompt:
REM   sc query state= all | findstr "SERVICE_NAME"
REM
REM ============================================================================
REM USAGE EXAMPLES:
REM ============================================================================

REM Example 1: Basic usage - Restart Print Spooler every hour (default settings)
REM service-restarter.exe -service Spooler

REM Example 2: Restart Windows Update service every 30 minutes
REM service-restarter.exe -service wuauserv -interval 30

REM Example 3: Restart SQL Server every 2 hours with 5 second hold time
REM service-restarter.exe -service MSSQLSERVER -interval 120 -hold 5

REM Example 4: Restart a custom application service every 15 minutes
REM service-restarter.exe -service "MyCustomService" -interval 15 -hold 3

REM Example 5: Restart Remote Desktop service every hour with 10 second pause
REM service-restarter.exe -service TermService -interval 60 -hold 10

REM ============================================================================
REM ACTIVE CONFIGURATION:
REM ============================================================================
REM
REM Modify the values below to match your requirements:
REM

REM Set your service name here (REQUIRED - CHANGE THIS!)
set SERVICE_NAME=Spooler

REM Set restart interval in minutes (default: 60)
set RESTART_INTERVAL=60

REM Set hold time in seconds (default: 2)
set HOLD_TIME=2

REM ============================================================================
REM CHECK PREREQUISITES
REM ============================================================================

REM Check if running as administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator!
    echo Right-click this file and select "Run as administrator"
    pause
    exit /b 1
)

REM Check if the executable exists
if not exist "service-restarter.exe" (
    echo ERROR: service-restarter.exe not found in current directory!
    echo Please compile the Go program first using: go build
    pause
    exit /b 1
)

REM Check if service name is set
if "%SERVICE_NAME%"=="Spooler" (
    echo WARNING: You are using the default service name "Spooler"
    echo Make sure this is the service you want to restart!
    echo.
    echo Press Ctrl+C to cancel, or
    pause
)

REM ============================================================================
REM RUN THE SERVICE RESTARTER
REM ============================================================================

echo.
echo ============================================================================
echo Starting Service Restarter
echo ============================================================================
echo Service Name:      %SERVICE_NAME%
echo Restart Interval:  %RESTART_INTERVAL% minutes
echo Hold Time:         %HOLD_TIME% seconds
echo ============================================================================
echo.
echo A GUI window will appear in the upper-right corner showing countdown.
echo Use the Close button to stop the restarter.
echo Use the Restart button to trigger an immediate restart.
echo.

REM Run the program with configured parameters
service-restarter.exe -service %SERVICE_NAME% -interval %RESTART_INTERVAL% -hold %HOLD_TIME%

REM ============================================================================
REM CLEANUP
REM ============================================================================

echo.
echo Service Restarter has stopped.
pause
