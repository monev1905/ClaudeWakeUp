# Security Policy

## Supported version

Only the latest commit on the `main` branch is supported. Users should pull the latest version and replace older prebuilt executables after security updates.

## Reporting a vulnerability

Do not include credentials, session data, personal paths, or sensitive log contents in a public issue. Contact the repository owner privately through their GitHub profile when possible. If no private channel is available, open a minimal public issue asking for a private contact method without disclosing exploit details.

## Trust and distribution

- Obtain the project only from `https://github.com/monev1905/ClaudeWakeUp`.
- Verify `dist\ClaudeWakeUp.exe` against `dist\SHA256SUMS.txt` before running it.
- Prefer tagged GitHub Release artifacts and verify their provenance with `gh attestation verify ClaudeWakeUp.exe --repo monev1905/ClaudeWakeUp`.
- The prebuilt executable is not currently code-signed. Windows can therefore verify its checksum, but not a publisher identity.
- For the strongest assurance, inspect the source and build the executable locally with the documented Go version.
- Never run Claude WakeUp as Administrator. It is designed to run with normal user privileges.

## Credentials and privacy

Claude WakeUp contains no credentials and does not manage Claude authentication. The installed Claude Code CLI owns authentication and network communication. Claude output is discarded and is not retained in the application log.

The child process inherits the current user's environment because Claude installations may rely on environment configuration. Users are responsible for credentials or proxy settings they intentionally place in their Windows environment; the project does not add any.

## Security boundaries

The application reduces command-search, project-configuration, tool-use, MCP, Chrome integration, session/output-retention, orphaned-process, and unbounded-log risks. It does not protect a user from:

- an already compromised Windows account;
- a malicious or compromised system-wide Claude Code installation;
- malicious changes accepted into the repository;
- compromise of the repository owner or distribution platform;
- administrator-managed Claude policies, which safe mode intentionally cannot bypass;
- risks inherent in sending the fixed prompt to the user's configured Claude service.
