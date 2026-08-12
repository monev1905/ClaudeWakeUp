package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"
)

const (
	appName        = "Claude WakeUp"
	commandTimeout = 2 * time.Minute
	pollInterval   = 5 * time.Second
)

var wakeUpRetryDelays = []time.Duration{0, 30 * time.Second, 60 * time.Second}

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

	claudePath, err := resolveClaudeCLI()
	if err != nil {
		logger.Printf("ERROR: %v", err)
		fmt.Println("\nA suitable Claude CLI installation was not found. Install/login to Claude")
		fmt.Println("Code, make sure 'claude' works in Command Prompt, and start this app again.")
		pauseOnWindows()
		os.Exit(1)
	}
	workDir, err := secureWorkDirectory()
	if err != nil {
		logger.Printf("ERROR: cannot create the isolated work directory: %v", err)
		pauseOnWindows()
		os.Exit(1)
	}
	if err := validateClaudeCapabilities(claudePath, workDir); err != nil {
		logger.Printf("ERROR: %v", err)
		fmt.Println("\nClaude Code does not support the security options required by this app.")
		fmt.Println("Run 'claude update' and start Claude WakeUp again.")
		pauseOnWindows()
		os.Exit(1)
	}
	fmt.Printf("Claude CLI: %s\n", claudePath)
	logger.Println("Security checks passed; Claude CLI resolved outside the application directory")

	if *testOnly {
		logger.Println("Test mode: running wake-up command now")
		if err := runWakeUpAttempt(context.Background(), claudePath, workDir); err != nil {
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

	if err := runScheduler(ctx, logger, wakeUpTimes, claudePath, workDir); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("Scheduler stopped with an error: %v", err)
		pauseOnWindows()
		os.Exit(1)
	}
	logger.Println("Application stopped")
}

func runScheduler(ctx context.Context, logger *log.Logger, schedule []clockTime, claudePath, workDir string) error {
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
					if err := runWakeUpWithRetry(ctx, logger, claudePath, workDir); err != nil {
						if !errors.Is(err, context.Canceled) {
							logger.Printf("Wake-up FAILED after %d attempts: %v", len(wakeUpRetryDelays), err)
						}
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

func runWakeUpWithRetry(ctx context.Context, logger *log.Logger, claudePath, workDir string) error {
	return retryWithDelays(ctx, wakeUpRetryDelays, func(attempt int) error {
		if attempt > 1 {
			logger.Printf("Wake-up retry %d of %d", attempt, len(wakeUpRetryDelays))
		}
		err := runWakeUpAttempt(ctx, claudePath, workDir)
		if err != nil && attempt < len(wakeUpRetryDelays) && ctx.Err() == nil {
			logger.Printf("Wake-up attempt %d failed: %v", attempt, err)
		}
		return err
	})
}

func runWakeUpAttempt(parent context.Context, claudePath, workDir string) error {
	ctx, cancel := context.WithTimeout(parent, commandTimeout)
	defer cancel()

	cmd, err := newClaudeCommand(ctx, claudePath, claudeArguments())
	if err != nil {
		return err
	}
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("Claude command timed out after %s", commandTimeout)
	}
	if err != nil {
		return fmt.Errorf("Claude command failed (%v); run 'claude -p \"wake up\"' manually for details", err)
	}
	return nil
}

func claudeArguments() []string {
	return []string{
		"-p", "wake up",
		"--safe-mode",
		"--no-session-persistence",
		"--disable-slash-commands",
		"--no-chrome",
		"--tools", "",
		"--max-turns", "1",
		"--permission-mode", "plan",
		"--setting-sources", "",
		"--strict-mcp-config",
	}
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
