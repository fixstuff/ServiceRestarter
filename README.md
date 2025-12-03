# Service Restarter

A Windows service management tool that automatically stops and restarts a specified service at regular intervals, with a GUI countdown timer.

## Features

- **Automatic Service Restarts**: Stops and restarts any Windows service on a configurable schedule
- **Configurable Timing**: Set custom restart intervals and hold times
- **GUI Countdown Timer**: Small, movable window displaying time until next restart
- **Visual Alerts**: Timer flashes red when restart is imminent
- **Manual Controls**: Restart immediately or close the application
- **Command-Line Interface**: Full control via CLI flags

## GUI Features

- **Yellow background** with black border
- **Countdown timer** in black text (flashes red at timeout)
- **Movable window** positioned in upper-right corner by default
- **Restart button**: Trigger immediate service restart
- **Close button**: Gracefully shut down the application

## Prerequisites

- Windows operating system
- Go 1.21 or higher (for compilation)
- Administrator privileges (required to control Windows services)

## Installation

### Step 1: Install Go

If you don't have Go installed:
1. Download from [https://golang.org/dl/](https://golang.org/dl/)
2. Install and verify with: `go version`

### Step 2: Clone or Download

Clone this repository or download the source code to your local machine.

### Step 3: Download Dependencies

Open a command prompt in the project directory and run:

```bash
go mod download
```

This will download the required GUI library (walk) and other dependencies.

### Step 4: Build the Executable

Compile the program with:

```bash
go build -o service-restarter.exe
```

For a smaller executable without debug info:

```bash
go build -ldflags="-s -w" -o service-restarter.exe
```

## Usage

### Using the Batch File (Recommended)

1. Edit `run-restarter.bat` and set your service name:
   ```batch
   set SERVICE_NAME=YourServiceName
   set RESTART_INTERVAL=60
   set HOLD_TIME=2
   ```

2. Right-click `run-restarter.bat` and select **Run as administrator**

### Direct Command-Line Usage

Run with administrator privileges:

```bash
service-restarter.exe -service <ServiceName> [-interval <minutes>] [-hold <seconds>] [-nogui]
```

### Command-Line Flags

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `-service` | string | Yes | - | Name of the Windows service to restart |
| `-interval` | int | No | 60 | Restart interval in minutes |
| `-hold` | int | No | 2 | Seconds to wait after stopping before starting |
| `-nogui` | bool | No | false | Run without GUI countdown window |

### Examples

**Restart Print Spooler every hour (default settings):**
```bash
service-restarter.exe -service Spooler
```

**Restart Windows Update every 30 minutes:**
```bash
service-restarter.exe -service wuauserv -interval 30
```

**Restart SQL Server every 2 hours with 5-second pause:**
```bash
service-restarter.exe -service MSSQLSERVER -interval 120 -hold 5
```

**Restart custom service every 15 minutes:**
```bash
service-restarter.exe -service "MyCustomService" -interval 15 -hold 3
```

**Run without GUI countdown window:**
```bash
service-restarter.exe -service Spooler -interval 60 -nogui
```

## Finding Service Names

To find the exact service name:

### Method 1: Services Management Console
1. Press `Win + R`, type `services.msc`, press Enter
2. Find your service in the list
3. Double-click to open properties
4. Copy the **Service name** field (NOT the Display name)

### Method 2: Command Line
```bash
sc query state= all | findstr "SERVICE_NAME"
```

### Common Service Names

| Display Name | Service Name |
|--------------|--------------|
| Print Spooler | Spooler |
| Windows Update | wuauserv |
| Windows Defender | WinDefend |
| DHCP Client | Dhcp |
| DNS Client | Dnscache |
| Remote Desktop Services | TermService |
| SQL Server | MSSQLSERVER |

## GUI Controls (Optional)

By default, the program shows a GUI countdown window. Use `-nogui` flag to disable it.

- **Timer Display**: Shows countdown to next restart in MM:SS or HH:MM:SS format
- **Restart Button**: Immediately stops and restarts the service, resetting the timer
- **Close Button**: Stops the restart scheduler and closes the application
- **Red Flash**: Timer text flashes red when countdown reaches zero
- **Window**: Yellow background with black border, movable, positioned in upper-right corner

## How It Works

1. The program starts and displays a startup message
2. Optionally displays the GUI countdown window (unless `-nogui` is used)
3. It runs a background loop that triggers every configured interval
4. When triggered:
   - Stops the specified service
   - Waits for the configured hold time
   - Starts the service again
5. If GUI is enabled, it updates every second showing time remaining
6. At timeout, the timer flashes red while the restart is performed
7. The cycle repeats until you stop the program

## Troubleshooting

### "Access Denied" Error
- Make sure you're running as Administrator
- Right-click the .bat file or .exe and select "Run as administrator"

### Service Not Found
- Verify the service name using `services.msc`
- Service names are case-sensitive
- Use the service name, not the display name

### GUI Doesn't Appear
- Check if the window is off-screen
- The window should appear in the upper-right corner
- Try restarting the application

### Compilation Errors
- Ensure Go 1.21+ is installed: `go version`
- Run `go mod tidy` to clean up dependencies
- Try `go clean` then rebuild

## Building for Distribution

To create a standalone executable that can be distributed:

```bash
go build -ldflags="-s -w" -o service-restarter.exe
```

The resulting `service-restarter.exe` can be copied to any Windows machine without needing Go installed.

## License

This project is provided as-is for personal and commercial use.

## Safety Notes

- Always test with non-critical services first
- Be cautious with system-critical services (can cause instability)
- Monitor the first few restart cycles to ensure proper operation
- Keep hold times reasonable (2-10 seconds is typical)

## Development

### Project Structure
```
service-restarter/
├── main.go              # Main application code
├── go.mod               # Go module dependencies
├── run-restarter.bat    # Batch file launcher with examples
├── README.md            # This file
└── .git/                # Git repository
```

### Tech Stack
- **Language**: Go 1.21+
- **GUI Library**: lxn/walk (Windows-specific GUI)
- **Service Control**: Windows `net` commands via os/exec

## Contributing

Feel free to submit issues, fork the repository, and create pull requests for any improvements.
