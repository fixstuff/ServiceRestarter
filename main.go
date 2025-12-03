package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

var (
	gdi32              = syscall.NewLazyDLL("gdi32.dll")
	user32             = syscall.NewLazyDLL("user32.dll")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procFillRect       = user32.NewProc("FillRect")
	procSetWindowText  = user32.NewProc("SetWindowTextW")
	procCreateFontW    = gdi32.NewProc("CreateFontW")
)

func createSolidBrush(color uint32) win.HBRUSH {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return win.HBRUSH(ret)
}

func fillRect(hdc win.HDC, rect *win.RECT, hbrush win.HBRUSH) {
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(hbrush))
}

func setWindowText(hwnd win.HWND, text *uint16) {
	procSetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(text)))
}

func createFont(height, width, escapement, orientation, weight int32,
	italic, underline, strikeOut, charSet, outputPrecision, clipPrecision,
	quality, pitchAndFamily uint32, face *uint16) win.HFONT {
	ret, _, _ := procCreateFontW.Call(
		uintptr(height),
		uintptr(width),
		uintptr(escapement),
		uintptr(orientation),
		uintptr(weight),
		uintptr(italic),
		uintptr(underline),
		uintptr(strikeOut),
		uintptr(charSet),
		uintptr(outputPrecision),
		uintptr(clipPrecision),
		uintptr(quality),
		uintptr(pitchAndFamily),
		uintptr(unsafe.Pointer(face)))
	return win.HFONT(ret)
}

var (
	serviceName   = flag.String("service", "", "Name of the Windows service to restart")
	intervalMins  = flag.Int("interval", 60, "Restart interval in minutes")
	holdTimeSecs  = flag.Int("hold", 2, "Time to wait after stopping service (seconds)")
	noGUI         = flag.Bool("nogui", false, "Run without GUI countdown window")

	stopChan      = make(chan bool)
	restartChan   = make(chan bool)
	timeRemaining time.Duration
	isFlashing    = false
)

func main() {
	flag.Parse()

	if *serviceName == "" {
		log.Fatal("Error: Service name is required. Use -service flag to specify the service name.")
	}

	log.Printf("============================================================================")
	log.Printf("Service Restarter is RUNNING")
	log.Printf("============================================================================")
	log.Printf("Service:          %s", *serviceName)
	log.Printf("Restart Interval: %d minutes", *intervalMins)
	log.Printf("Hold Time:        %d seconds", *holdTimeSecs)
	log.Printf("Started at:       %s", time.Now().Format("2006-01-02 03:04:05 PM"))
	log.Printf("============================================================================")

	// Start the restart loop in a goroutine
	go restartLoop()

	// Start the GUI countdown window (unless -nogui flag is set)
	if !*noGUI {
		if err := createGUI(); err != nil {
			log.Printf("Failed to create GUI: %v", err)
			// Show error message box
			walk.MsgBox(nil, "Service Restarter - Error",
				fmt.Sprintf("GUI failed to start: %v\n\nContinuing without GUI...", err),
				walk.MsgBoxIconWarning)
			// Keep program running without GUI
			select {}
		}
	} else {
		// Show startup message box when running without GUI
		startupMsg := fmt.Sprintf(
			"Service Restarter is RUNNING\n\n"+
				"Service: %s\n"+
				"Restart Interval: %d minutes\n"+
				"Hold Time: %d seconds\n"+
				"Started at: %s\n\n"+
				"To stop, run stop-restarter.bat or use Task Manager",
			*serviceName,
			*intervalMins,
			*holdTimeSecs,
			time.Now().Format("2006-01-02 03:04:05 PM"))

		walk.MsgBox(nil, "Service Restarter", startupMsg, walk.MsgBoxIconInformation)

		// Keep program running without GUI
		select {}
	}
}

func restartLoop() {
	ticker := time.NewTicker(time.Duration(*intervalMins) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			log.Println("Stopping restart loop")
			return
		case <-restartChan:
			log.Println("Manual restart triggered")
			performRestart()
			ticker.Reset(time.Duration(*intervalMins) * time.Minute)
			timeRemaining = time.Duration(*intervalMins) * time.Minute
			isFlashing = false
			textIsRed = false
			// Clear status message
			if statusText != 0 {
				emptyMsg, _ := syscall.UTF16PtrFromString("")
				setWindowText(statusText, emptyMsg)
			}
		case <-ticker.C:
			performRestart()
			timeRemaining = time.Duration(*intervalMins) * time.Minute
			isFlashing = false
			textIsRed = false
			// Clear status message
			if statusText != 0 {
				emptyMsg, _ := syscall.UTF16PtrFromString("")
				setWindowText(statusText, emptyMsg)
			}
		}
	}
}

func performRestart() {
	log.Printf("Stopping service: %s", *serviceName)
	if err := stopService(*serviceName); err != nil {
		log.Printf("Error stopping service: %v", err)
		return
	}

	log.Printf("Waiting %d seconds...", *holdTimeSecs)
	time.Sleep(time.Duration(*holdTimeSecs) * time.Second)

	log.Printf("Starting service: %s", *serviceName)
	if err := startService(*serviceName); err != nil {
		log.Printf("Error starting service: %v", err)
		return
	}

	log.Println("Service restart complete")
}

func stopService(name string) error {
	cmd := exec.Command("net", "stop", name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service: %v, output: %s", err, string(output))
	}
	return nil
}

func startService(name string) error {
	cmd := exec.Command("net", "start", name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service: %v, output: %s", err, string(output))
	}
	return nil
}

var (
	hwnd          win.HWND
	timerText     win.HWND
	statusText    win.HWND
	restartButton win.HWND
	closeButton   win.HWND
	yellowBrush   win.HBRUSH
	textIsRed     bool
)

func createGUI() error {
	// Initialize time remaining
	timeRemaining = time.Duration(*intervalMins) * time.Minute

	// Create custom window using raw Windows API
	return createCustomWindow()
}

func createCustomWindow() error {
	className, _ := syscall.UTF16PtrFromString("ServiceRestarterClass")
	windowName, _ := syscall.UTF16PtrFromString("Service Restarter")

	// Register window class
	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = win.GetModuleHandle(nil)
	wc.LpszClassName = className
	wc.HbrBackground = win.COLOR_WINDOW + 1
	wc.HCursor = win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		return fmt.Errorf("RegisterClassEx failed")
	}

	// Create window
	hwnd = win.CreateWindowEx(
		0,
		className,
		windowName,
		win.WS_OVERLAPPED|win.WS_CAPTION|win.WS_SYSMENU|win.WS_VISIBLE,
		win.CW_USEDEFAULT, win.CW_USEDEFAULT,
		250, 140,
		0, 0,
		win.GetModuleHandle(nil),
		nil)

	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx failed")
	}

	// Create yellow brush for background
	yellowBrush = createSolidBrush(0x00FFFF) // BGR format: Yellow = 0x00FFFF

	// Position window in upper right
	screenWidth := win.GetSystemMetrics(win.SM_CXSCREEN)
	win.SetWindowPos(hwnd, 0, int32(screenWidth-270), 20, 0, 0,
		win.SWP_NOSIZE|win.SWP_NOZORDER)

	// Create timer text (static control)
	timerText = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("STATIC"),
		syscall.StringToUTF16Ptr(formatDuration(timeRemaining)),
		win.WS_CHILD|win.WS_VISIBLE|win.SS_CENTER,
		10, 5, 230, 40,
		hwnd, 0,
		win.GetModuleHandle(nil),
		nil)

	// Set large font for timer
	hFont := createFont(28, 0, 0, 0, 700, 0, 0, 0,
		1, 0, 0, 0, 0,
		syscall.StringToUTF16Ptr("Segoe UI"))
	win.SendMessage(timerText, win.WM_SETFONT, uintptr(hFont), 1)

	// Create status text (static control)
	statusText = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("STATIC"),
		syscall.StringToUTF16Ptr(""),
		win.WS_CHILD|win.WS_VISIBLE|win.SS_CENTER,
		10, 45, 230, 20,
		hwnd, 0,
		win.GetModuleHandle(nil),
		nil)

	// Set smaller font for status
	statusFont := createFont(12, 0, 0, 0, 400, 0, 0, 0,
		1, 0, 0, 0, 0,
		syscall.StringToUTF16Ptr("Segoe UI"))
	win.SendMessage(statusText, win.WM_SETFONT, uintptr(statusFont), 1)

	// Create Restart button
	restartButton = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("BUTTON"),
		syscall.StringToUTF16Ptr("Restart"),
		win.WS_CHILD|win.WS_VISIBLE|win.BS_PUSHBUTTON,
		10, 70, 100, 30,
		hwnd, win.HMENU(1),
		win.GetModuleHandle(nil),
		nil)

	// Create Close button
	closeButton = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("BUTTON"),
		syscall.StringToUTF16Ptr("Close"),
		win.WS_CHILD|win.WS_VISIBLE|win.BS_PUSHBUTTON,
		130, 70, 100, 30,
		hwnd, win.HMENU(2),
		win.GetModuleHandle(nil),
		nil)

	// Start timer update
	go updateTimerText()

	// Message loop
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	return nil
}

func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		cmdID := win.LOWORD(uint32(wParam))
		if cmdID == 1 { // Restart button
			log.Println("Restart button clicked")
			// Reset timer
			timeRemaining = time.Duration(*intervalMins) * time.Minute
			isFlashing = false
			textIsRed = false
			// Clear status message
			emptyMsg, _ := syscall.UTF16PtrFromString("")
			setWindowText(statusText, emptyMsg)
			// Update display immediately
			text, _ := syscall.UTF16PtrFromString(formatDuration(timeRemaining))
			setWindowText(timerText, text)
			win.InvalidateRect(timerText, nil, true)
			// Trigger service restart
			restartChan <- true
			return 0
		} else if cmdID == 2 { // Close button
			log.Println("Close button clicked")
			win.PostQuitMessage(0)
			stopChan <- true
			return 0
		}

	case win.WM_CTLCOLORSTATIC:
		// Set yellow background for static text
		hdc := win.HDC(wParam)
		control := win.HWND(lParam)
		win.SetBkColor(hdc, win.COLORREF(0x00FFFF)) // Yellow

		// Check if this is the timer text and if it should be red
		if control == timerText && textIsRed {
			win.SetTextColor(hdc, win.COLORREF(0x0000FF)) // Red (BGR format)
		} else {
			win.SetTextColor(hdc, win.COLORREF(0x000000)) // Black
		}
		return uintptr(yellowBrush)

	case win.WM_ERASEBKGND:
		// Paint yellow background
		hdc := win.HDC(wParam)
		var rect win.RECT
		win.GetClientRect(hwnd, &rect)
		fillRect(hdc, &rect, yellowBrush)
		return 1

	case win.WM_CLOSE:
		win.PostQuitMessage(0)
		stopChan <- true
		return 0

	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}

	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

func updateTimerText() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	flashTicker := time.NewTicker(500 * time.Millisecond)
	defer flashTicker.Stop()

	for {
		select {
		case <-ticker.C:
			timeRemaining -= time.Second
			if timeRemaining <= 0 {
				timeRemaining = 0
				isFlashing = true
				// Show restarting message
				statusMsg, _ := syscall.UTF16PtrFromString("Restarting service...")
				setWindowText(statusText, statusMsg)
			} else if timeRemaining == time.Duration(*intervalMins)*time.Minute {
				// Clear status message when timer resets
				emptyMsg, _ := syscall.UTF16PtrFromString("")
				setWindowText(statusText, emptyMsg)
			}

			// Update text
			text, _ := syscall.UTF16PtrFromString(formatDuration(timeRemaining))
			setWindowText(timerText, text)

		case <-flashTicker.C:
			if isFlashing && timeRemaining <= 0 {
				// Flash between red and black
				textIsRed = !textIsRed
				win.InvalidateRect(timerText, nil, true)
			} else {
				textIsRed = false
				isFlashing = false
			}
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
