package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	appName        = "Claude WakeUp"
	commandTimeout = 10 * time.Minute
	pollInterval   = 5 * time.Second
)

var wakeUpTimes = []clockTime{
	{hour: 5, minute: 30},
	{hour: 10, minute: 30},
	{hour: 15, minute: 30},
	{hour: 20, minute: 30},
}

func main() {
	testOnly := flag.Bool("test", false, "run the Claude wake-up command once and exit")
	flag.Parse()

	release, alreadyRunning, err := acquireSingleInstance()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot start %s: %v\n", appName, err)
		pauseOnWindows()
		os.Exit(1)
	}
	if alreadyRunning {
		fmt.Printf("%s is already running. Close its existing window first.\n", appName)
		pauseOnWindows()
		return
	}
	defer release()

	logger, logFile, closeLog, err := newLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create the log file: %v\n", err)
		pauseOnWindows()
		os.Exit(1)
	}
	defer closeLog()

	printBanner(logFile)
	logger.Printf("Application started (local time: %s)", time.Now().Format(time.RFC3339))

	if err := checkClaudeAvailable(); err != nil {
		logger.Printf("ERROR: %v", err)
		fmt.Println("\nClaude CLI was not found. Install/login to Claude Code, make sure the")
		fmt.Println("'claude' command works in Command Prompt, and then start this app again.")
		pauseOnWindows()
		os.Exit(1)
	}

	if *testOnly {
		logger.Println("Test mode: running wake-up command now")
		if err := runWakeUp(logger); err != nil {
			logger.Printf("TEST FAILED: %v", err)
			pauseOnWindows()
			os.Exit(1)
		}
		logger.Println("TEST PASSED: Claude command completed successfully")
		pauseOnWindows()
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	next := nextWakeUp(time.Now(), wakeUpTimes)
	fmt.Printf("Next wake-up: %s\n", next.Format("02 Jan 2006 15:04 MST"))
	logger.Printf("Next wake-up: %s", next.Format(time.RFC1123))

	if err := runScheduler(ctx, logger, wakeUpTimes); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("Scheduler stopped with an error: %v", err)
		pauseOnWindows()
		os.Exit(1)
	}
	logger.Println("Application stopped")
}

func runScheduler(ctx context.Context, logger *log.Logger, schedule []clockTime) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Check immediately too, so launching during a scheduled minute still triggers it.
	ticks := make(chan time.Time, 1)
	ticks <- time.Now()
	var lastSlot string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticks:
			if isWakeUpMinute(now, schedule) {
				slot := now.Format("2006-01-02 15:04")
				if slot != lastSlot {
					lastSlot = slot
					logger.Printf("Scheduled wake-up started for %s", slot)
					if err := runWakeUp(logger); err != nil {
						logger.Printf("Wake-up FAILED: %v", err)
					} else {
						logger.Println("Wake-up completed successfully")
					}
					next := nextWakeUp(time.Now(), schedule)
					fmt.Printf("Next wake-up: %s\n", next.Format("02 Jan 2006 15:04 MST"))
					logger.Printf("Next wake-up: %s", next.Format(time.RFC1123))
				}
			}
		case now := <-ticker.C:
			// Feed the shared handler without accumulating ticks while Claude is running.
			select {
			case ticks <- now:
			default:
			}
		}
	}
}

func runWakeUp(logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cmd.exe", "/D", "/S", "/C", `claude -p "wake up"`)
	output, err := cmd.CombinedOutput()
	cleanOutput := strings.TrimSpace(string(output))
	if cleanOutput != "" {
		logger.Printf("Claude output:\n%s", cleanOutput)
	}
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("Claude command timed out after %s", commandTimeout)
	}
	if err != nil {
		return fmt.Errorf("Claude command returned an error: %w", err)
	}
	return nil
}

func checkClaudeAvailable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cmd.exe", "/D", "/C", "where claude")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("the 'claude' command is not available in PATH (%s)", strings.TrimSpace(string(output)))
	}
	return nil
}

func newLogger() (*log.Logger, string, func(), error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, "", nil, err
	}
	logDir := filepath.Join(configDir, "ClaudeWakeUp")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, "", nil, err
	}
	logPath := filepath.Join(logDir, "ClaudeWakeUp.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", nil, err
	}
	writer := io.MultiWriter(os.Stdout, file)
	return log.New(writer, "", log.Ldate|log.Ltime), logPath, func() { _ = file.Close() }, nil
}

func printBanner(logFile string) {
	fmt.Println("Claude WakeUp")
	fmt.Println("Schedule: 05:30, 10:30, 15:30, 20:30 (device local time)")
	fmt.Println("Keep this window open. Close it or press Ctrl+C to stop all wake-ups.")
	fmt.Printf("Log: %s\n\n", logFile)
}

func pauseOnWindows() {
	if runtime.GOOS != "windows" {
		return
	}
	fmt.Print("\nPress Enter to close...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
