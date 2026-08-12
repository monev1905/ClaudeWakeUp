# Claude WakeUp

Claude WakeUp is a small Windows application that runs:

```cmd
claude -p "wake up"
```

at **05:30, 10:30, 15:30, and 20:30**, using the computer's local date, time, and time zone. The 01:30 trigger is intentionally skipped.

## Quick start on Windows

Clone the repository and open its directory:

```cmd
git clone https://github.com/monev1905/ClaudeWakeUp.git
cd ClaudeWakeUp
```

Double-click `START-ClaudeWakeUp.bat`. A prebuilt executable is included at `dist\ClaudeWakeUp.exe`, so no compilation is required.

Keep the application window open while you want the schedule to remain active. Closing the window or pressing `Ctrl+C` stops all future wake-up runs. Starting it again resumes the schedule and waits for the next configured time.

## Requirements and authentication

- Claude Code must be installed, and the `claude` command must be available in Command Prompt.
- Claude Code must already be authenticated on that Windows user account.
- This project contains no API keys, access tokens, passwords, or other credentials.
- The application does not read, store, or transmit credentials. It delegates authentication entirely to the locally installed Claude Code CLI.

If Claude Code is not installed, not available in `PATH`, or not authenticated, the wake-up command will fail and the error will be written to the log.

## Behavior

- The computer must be powered on and awake during the scheduled minute.
- Runs missed while the computer is asleep or turned off are not executed later.
- Launching the application during a scheduled minute runs the command immediately.
- Only one copy of the application can run at a time.
- The application does not add itself to Windows Startup.
- Logs are stored at `%APPDATA%\ClaudeWakeUp\ClaudeWakeUp.log`.

## Test immediately

Open Command Prompt in the project directory and run:

```cmd
dist\ClaudeWakeUp.exe -test
```

This sends one `wake up` prompt immediately and then exits.

## Build from source

Install Go 1.22 or newer and run `build-windows.bat`. The executable will be created at `dist\ClaudeWakeUp.exe`.

## Important

This application only sends a Claude request at the configured times. It cannot bypass, modify, or forcibly reset limits imposed by Anthropic.
