# AgentTeams Agent Installation Guide

This guide covers the installation and setup of the AgentTeams Agent on Windows machines.

## Prerequisites

- Windows 10/11 or Windows Server 2016+
- Administrator privileges
- Network access to the AgentTeams Server

## Installation Methods

### Method 1: MSI Installer (Recommended)

1. Download the latest MSI installer from the release page
2. Run the installer with administrator privileges
3. Follow the installation wizard
4. Configure the agent using the configuration file

```powershell
# Install silently
msiexec /i AgentTeams-Agent-1.0.0.msi /qn
```

### Method 2: Manual Installation

1. Download the agent package (ZIP)
2. Extract to `C:\Program Files\AgentTeams\`
3. Copy the configuration template:

```powershell
Copy-Item config\agent.yaml "C:\ProgramData\AgentTeams\agent.yaml"
```

4. Install as Windows Service:

```powershell
cd "C:\Program Files\AgentTeams"
.\agent.exe --install
```

## Configuration

The agent reads its configuration from `C:\ProgramData\AgentTeams\agent.yaml`.

### Minimal Configuration

```yaml
agent:
  id: ""  # Auto-generated on first run
  token: "YOUR_REGISTRATION_TOKEN"
  server_url: "wss://server.example.com:443/api/v1/agent/ws"
```

### Full Configuration

See [Configuration Reference](./configuration-reference.md) for all options.

## Registration

Before the agent can connect, it must be registered with the server:

1. Log in to the AgentTeams web interface
2. Navigate to Agents > Register New Agent
3. Enter a name for the agent
4. Copy the generated token
5. Paste the token in the configuration file

## Starting the Service

```powershell
# Start the service
net start AgentTeams

# Or via PowerShell
Start-Service AgentTeams
```

## Verifying Installation

1. Check service status:

```powershell
Get-Service AgentTeams
```

2. Check logs:

```powershell
Get-Content "C:\ProgramData\AgentTeams\logs\agent.log" -Tail 50
```

3. Verify connection in the web interface

## Firewall Configuration

The agent requires outbound access to the server:

```powershell
# Allow outbound HTTPS (if using WSS)
New-NetFirewallRule -DisplayName "AgentTeams Agent" `
    -Direction Outbound `
    -Action Allow `
    -Protocol TCP `
    -RemotePort 443
```

## Uninstallation

### MSI Installation

```powershell
msiexec /x AgentTeams-Agent-1.0.0.msi /qn
```

### Manual Installation

```powershell
cd "C:\Program Files\AgentTeams"
.\agent.exe --uninstall
```

## Troubleshooting

### Agent Won't Start

1. Check the configuration file exists
2. Verify the token is correct
3. Check network connectivity
4. Review logs for errors

### Connection Issues

1. Verify server URL is correct
2. Check if server is accessible
3. Verify TLS certificate is valid

### Authentication Failures

1. Verify the token is correct
2. Check if the agent is registered on the server
3. Verify the agent ID matches

## Log Files

Log files are located in `C:\ProgramData\AgentTeams\logs\`:
- `agent.log` - Main agent log
- `heartbeat.log` - Heartbeat worker log
- `task.log` - Task worker log

## Upgrading

The agent supports automatic updates. See the Auto-Update documentation for details.

To manually upgrade:

1. Stop the service
2. Replace executable files
3. Start the service
