# Claude WakeUp

Claude WakeUp is a small Windows application that sends a `wake up` prompt through the locally installed Claude Code CLI at **05:30, 10:30, 15:30, and 20:30**. It uses the computer's local date, time, and time zone. The 01:30 trigger is intentionally skipped.

The purpose is to start five-hour Claude usage windows ahead of the workday while keeping 10:30 as the fixed daytime anchor. The resulting fixed schedule has three five-hour gaps and one intentional nine-hour overnight gap.

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
- Claude Code must be installed, current, and available through an absolute directory in the Windows `PATH`. Claude WakeUp verifies the required security flags at startup and asks the user to update Claude Code if any are unavailable.
- Claude Code must already be authenticated for the current Windows user.
- This project contains no API keys, access tokens, passwords, or other credentials.
- The application does not read or store credentials. Authentication and network communication are delegated entirely to the locally installed Claude Code CLI.

If Claude Code is unavailable or not authenticated, the wake-up command fails and a non-sensitive error is written to the log.

## Security design

- Claude Code is resolved to an absolute path before it is started.
- Relative `PATH` entries and copies of `claude.exe`, `claude.cmd`, or `claude.bat` anywhere under the application/project tree are rejected.
- The resolved absolute Claude CLI path is displayed at startup for inspection. An absolute path reduces command-search ambiguity but does not prove publisher identity.
- Windows batch-based Claude installations are launched through the trusted command interpreter in the Windows system directory.
- Claude runs from an isolated empty directory under `%APPDATA%\ClaudeWakeUp\workspace`, not from the cloned repository.
- The request uses `--safe-mode`, `--no-session-persistence`, `--disable-slash-commands`, `--no-chrome`, `--tools ""`, `--max-turns 1`, `--permission-mode plan`, `--setting-sources ""`, and `--strict-mcp-config`. This disables customizations, tools, skills, Chrome integration, settings, and MCP servers for the scheduled prompt.
- Claude's response is discarded and the one-off session is not persisted. Only timestamps and success/failure status are logged.
- A failed scheduled request is retried after 30 seconds and then after another 60 seconds. Successful requests are never repeated.
- Each attempt has a two-minute timeout. On Windows, cancellation attempts to terminate the complete Claude process tree.
- The log is capped at 1 MiB and is stored at `%APPDATA%\ClaudeWakeUp\ClaudeWakeUp.log`.
- The application does not request administrator privileges, open ports, install a service, create a scheduled task, modify the Registry, or add itself to Windows Startup.
- Only one copy can run for each signed-in Windows session.

These protections reduce accidental command hijacking and project-level configuration risks. They cannot make an already compromised Windows account or malicious system-wide Claude installation trustworthy.

## Verify the executable

The expected SHA-256 checksum for the repository copy is published in `dist\SHA256SUMS.txt`. Verify it from the repository root:

```cmd
certutil -hashfile dist\ClaudeWakeUp.exe SHA256
type dist\SHA256SUMS.txt
```

The two hashes must match. Tagged GitHub Releases are built by GitHub Actions and include a GitHub artifact attestation. Verify a downloaded release executable with GitHub CLI:

```cmd
gh attestation verify ClaudeWakeUp.exe --repo monev1905/ClaudeWakeUp
```

The executable is currently not Authenticode-signed, so Windows SmartScreen may display a warning. Users requiring stronger assurance should verify the release attestation or review the source and build locally.

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

The five-hour reset behavior is controlled by Anthropic. Confirm after initial setup that Claude's usage page reports the expected reset times for the account.

See [SECURITY.md](SECURITY.md) for security reporting and distribution guidance.
