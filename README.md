# Claude WakeUp

Claude WakeUp is a small Windows application that sends a `wake up` prompt through the locally installed Claude Code CLI at **05:30, 10:30, 15:30, and 20:30**. It uses the computer's local date, time, and time zone. The 01:30 trigger is intentionally skipped.

## Quick start on Windows

Clone the repository and open its directory:

```cmd
git clone https://github.com/monev1905/ClaudeWakeUp.git
cd ClaudeWakeUp
```

Double-click `START-ClaudeWakeUp.bat`. A prebuilt executable is included at `dist\ClaudeWakeUp.exe`, so compilation is not required.

Keep the application window open while you want the schedule to remain active. Closing the window or pressing `Ctrl+C` stops all future wake-up runs. Starting it again resumes the schedule and waits for the next configured time.

## Requirements and authentication

- Windows 10 or newer.
- Claude Code must be installed, current, and available through an absolute directory in the Windows `PATH`.
- Claude Code must already be authenticated for the current Windows user.
- This project contains no API keys, access tokens, passwords, or other credentials.
- The application does not read or store credentials. Authentication and network communication are delegated entirely to the locally installed Claude Code CLI.

If Claude Code is unavailable or not authenticated, the wake-up command fails and a non-sensitive error is written to the log.

## Security design

- Claude Code is resolved to an absolute path before it is started.
- Relative `PATH` entries and copies of `claude.exe`, `claude.cmd`, or `claude.bat` in the application/project directory are rejected.
- Windows batch-based Claude installations are launched through the trusted command interpreter in the Windows system directory.
- Claude runs from an isolated empty directory under `%APPDATA%\ClaudeWakeUp\workspace`, not from the cloned repository.
- The request uses `--tools ""`, `--max-turns 1`, `--permission-mode plan`, `--setting-sources ""`, and `--strict-mcp-config`. This prevents the wake-up prompt from using tools or loading user/project settings and MCP servers.
- Claude's response is discarded. Only timestamps and success/failure status are logged.
- The log is capped at 1 MiB and is stored at `%APPDATA%\ClaudeWakeUp\ClaudeWakeUp.log`.
- The application does not request administrator privileges, open ports, install a service, create a scheduled task, modify the Registry, or add itself to Windows Startup.
- Only one copy can run for each signed-in Windows session.

These protections reduce accidental command hijacking and project-level configuration risks. They cannot make an already compromised Windows account or malicious system-wide Claude installation trustworthy.

## Verify the executable

The expected SHA-256 checksum is published in `dist\SHA256SUMS.txt`. Verify it from the repository root:

```cmd
certutil -hashfile dist\ClaudeWakeUp.exe SHA256
type dist\SHA256SUMS.txt
```

The two hashes must match. The prebuilt executable is currently unsigned, so Windows SmartScreen may display a warning. Users requiring stronger provenance should review the source and build it locally.

## Test immediately

Stop any running copy, open Command Prompt in the project directory, and run:

```cmd
dist\ClaudeWakeUp.exe -test
```

This sends one restricted `wake up` prompt immediately and exits.

## Normal behavior

- The computer must be powered on and awake during the scheduled minute.
- Runs missed while the computer is asleep or turned off are not executed later.
- Launching the application during a scheduled minute runs the command immediately.
- Starting a second copy does not create a second schedule.

## Build from source

Install Go 1.26.5 or newer and run `build-windows.bat`. The executable is created at `dist\ClaudeWakeUp.exe`, followed by its SHA-256 checksum.

## Important

This application only sends a Claude request at the configured times. It cannot bypass, modify, or forcibly reset limits imposed by Anthropic. Each request can count toward the user's Claude usage allowance.

See [SECURITY.md](SECURITY.md) for security reporting and distribution guidance.
