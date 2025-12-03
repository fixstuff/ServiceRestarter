package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	serviceName   = flag.String("service", "", "Name of the Windows service to restart")
	intervalMins  = flag.Int("interval", 60, "Restart interval in minutes")
	holdTimeSecs  = flag.Int("hold", 2, "Time to wait after stopping service (seconds)")

	mainWindow    *walk.MainWindow
	timerLabel    *walk.Label
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

	log.Printf("Service Restarter started")
	log.Printf("Service: %s", *serviceName)
	log.Printf("Restart Interval: %d minutes", *intervalMins)
	log.Printf("Hold Time: %d seconds", *holdTimeSecs)

	// Start the restart loop in a goroutine
	go restartLoop()

	// Start the GUI
	if err := createGUI(); err != nil {
		log.Fatal(err)
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
		case <-ticker.C:
			performRestart()
			timeRemaining = time.Duration(*intervalMins) * time.Minute
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service: %v, output: %s", err, string(output))
	}
	return nil
}

func startService(name string) error {
	cmd := exec.Command("net", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service: %v, output: %s", err, string(output))
	}
	return nil
}

func createGUI() error {
	var closeBtn, restartBtn *walk.PushButton

	// Initialize time remaining
	timeRemaining = time.Duration(*intervalMins) * time.Minute

	err := MainWindow{
		AssignTo: &mainWindow,
		Title:    "Service Restarter",
		MinSize:  Size{Width: 250, Height: 120},
		MaxSize:  Size{Width: 250, Height: 120},
		Layout:   VBox{Margins: Margins{Top: 10, Left: 10, Right: 10, Bottom: 10}},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 0)}, // Yellow background
		Children: []Widget{
			Label{
				AssignTo: &timerLabel,
				Text:     formatDuration(timeRemaining),
				Font:     Font{PointSize: 20, Bold: true},
				TextColor: walk.RGB(0, 0, 0), // Black text
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &restartBtn,
						Text:     "Restart",
						OnClicked: func() {
							log.Println("Restart button clicked")
							restartChan <- true
						},
					},
					PushButton{
						AssignTo: &closeBtn,
						Text:     "Close",
						OnClicked: func() {
							log.Println("Close button clicked")
							stopChan <- true
							mainWindow.Close()
						},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		return err
	}

	// Position window in upper right corner
	positionWindowTopRight(mainWindow)

	// Add black border using Windows API
	addBorder(mainWindow)

	// Start timer update goroutine
	go updateTimer()

	mainWindow.Run()

	return nil
}

func positionWindowTopRight(window *walk.MainWindow) {
	screenWidth := win.GetSystemMetrics(win.SM_CXSCREEN)
	windowWidth := int32(250)

	window.SetX(int(screenWidth - windowWidth - 20))
	window.SetY(20)
}

func addBorder(window *walk.MainWindow) {
	hwnd := window.Handle()

	// Get current window style
	style := win.GetWindowLong(hwnd, win.GWL_STYLE)

	// Add border style
	style |= win.WS_BORDER

	// Set the new style
	win.SetWindowLong(hwnd, win.GWL_STYLE, style)

	// Force window to redraw
	win.SetWindowPos(hwnd, 0, 0, 0, 0, 0,
		win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED)
}

func updateTimer() {
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
			}

			mainWindow.Synchronize(func() {
				timerLabel.SetText(formatDuration(timeRemaining))
			})

		case <-flashTicker.C:
			if isFlashing && timeRemaining <= 0 {
				mainWindow.Synchronize(func() {
					// Toggle between red and black
					currentColor := timerLabel.TextColor()
					if currentColor == walk.RGB(0, 0, 0) {
						timerLabel.SetTextColor(walk.RGB(255, 0, 0)) // Red
					} else {
						timerLabel.SetTextColor(walk.RGB(0, 0, 0)) // Black
					}
				})
			} else if !isFlashing {
				mainWindow.Synchronize(func() {
					timerLabel.SetTextColor(walk.RGB(0, 0, 0)) // Ensure black when not flashing
				})
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
