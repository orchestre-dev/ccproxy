# CCProxy Installation Script for Windows
# Downloads and installs the latest CCProxy release from GitHub

param(
    [string]$InstallDir = "$env:ProgramFiles\CCProxy",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Success { Write-ColorOutput Green $args }
function Write-Error { Write-ColorOutput Red $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Info { Write-ColorOutput Cyan $args }

# Validate JSON configuration
function Test-JsonConfig {
    param([string]$FilePath)
    
    if (!(Test-Path $FilePath)) {
        return $false
    }
    
    try {
        $null = Get-Content $FilePath | ConvertFrom-Json
        Write-Success "Configuration validated successfully"
        return $true
    } catch {
        Write-Warning "Warning: Configuration has JSON syntax issues"
        Write-Warning $_.Exception.Message
        return $false
    }
}

# Backup existing configuration
function Backup-Config {
    param([string]$ConfigFile)
    
    if (Test-Path $ConfigFile) {
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $backupFile = "$ConfigFile.backup.$timestamp"
        Copy-Item -Path $ConfigFile -Destination $backupFile
        Write-Info "Backed up existing config to: $backupFile"
    }
}

# Check for concurrent installation
function Test-ConcurrentInstall {
    $lockFile = "$env:TEMP\ccproxy_install.lock"
    $pidFile = "$env:TEMP\ccproxy_install.pid"
    
    if (Test-Path $lockFile) {
        if (Test-Path $pidFile) {
            $oldPid = Get-Content $pidFile -ErrorAction SilentlyContinue
            if ($oldPid) {
                try {
                    $process = Get-Process -Id $oldPid -ErrorAction Stop
                    if ($process) {
                        Write-Error "Another installation is already in progress (PID: $oldPid)"
                        Write-Warning "If this is incorrect, remove $lockFile and try again"
                        exit 1
                    }
                } catch {
                    # Process not found, stale lock
                }
            }
        }
        # Remove stale lock files
        Remove-Item -Path $lockFile, $pidFile -Force -ErrorAction SilentlyContinue
    }
    
    # Create lock files
    New-Item -Path $lockFile -ItemType File -Force | Out-Null
    $PID | Out-File -FilePath $pidFile -Force
}

# Cleanup function
function Remove-LockFiles {
    $lockFile = "$env:TEMP\ccproxy_install.lock"
    $pidFile = "$env:TEMP\ccproxy_install.pid"
    Remove-Item -Path $lockFile, $pidFile -Force -ErrorAction SilentlyContinue
}

# GitHub repository
$Repo = "orchestre-dev/ccproxy"
$GitHubAPI = "https://api.github.com/repos/$Repo"
$GitHubDownload = "https://github.com/$Repo/releases/download"

Write-Info "=== CCProxy Installation for Windows ==="

# Check for concurrent installation
Test-ConcurrentInstall

# Register cleanup
trap { Remove-LockFiles }

# Check if running as administrator
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
$isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Warning "Note: Running without administrator privileges."
    Write-Warning "You may need admin rights to install to Program Files or modify system PATH."
    Write-Info ""
}

# Get latest version if not specified
if (-not $Version) {
    Write-Info "Fetching latest release information..."
    try {
        $releases = Invoke-RestMethod -Uri "$GitHubAPI/releases/latest"
        $Version = $releases.tag_name
        Write-Success "Latest version: $Version"
    } catch {
        Write-Error "Failed to fetch latest release: $_"
        exit 1
    }
} else {
    Write-Info "Installing specific version: $Version"
}

# Ensure version has 'v' prefix
if (-not $Version.StartsWith('v')) {
    $Version = "v$Version"
}

# Download binary
$binaryName = "ccproxy-windows-amd64.exe"
$downloadUrl = "$GitHubDownload/$Version/$binaryName"
$checksumUrl = "$GitHubDownload/$Version/checksums.txt"
$tempFile = Join-Path $env:TEMP "ccproxy-download.exe"
$checksumFile = Join-Path $env:TEMP "ccproxy-checksums.txt"

Write-Info "Downloading CCProxy $Version for Windows (64-bit)..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    Write-Success "Download completed successfully"
} catch {
    Write-Error "Download failed: $_"
    Write-Error "The binary might not be available for this version."
    Remove-LockFiles
    exit 1
}

# Download and verify checksum
Write-Info "Downloading checksums..."
try {
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile -UseBasicParsing
    
    # Calculate SHA256 of downloaded file
    $actualHash = (Get-FileHash -Path $tempFile -Algorithm SHA256).Hash.ToLower()
    
    # Find expected hash in checksum file
    $expectedHash = $null
    Get-Content $checksumFile | ForEach-Object {
        if ($_ -match "^([a-f0-9]{64})\s+$binaryName") {
            $expectedHash = $matches[1]
        }
    }
    
    if ($expectedHash) {
        if ($actualHash -eq $expectedHash) {
            Write-Success "Checksum verification passed"
        } else {
            Write-Error "Checksum verification failed!"
            Write-Error "Expected: $expectedHash"
            Write-Error "Actual:   $actualHash"
            Remove-Item -Path $tempFile -Force
            Remove-LockFiles
            exit 1
        }
    } else {
        Write-Warning "Could not find checksum for $binaryName in checksum file"
    }
    
    # Clean up checksum file
    Remove-Item -Path $checksumFile -Force -ErrorAction SilentlyContinue
} catch {
    Write-Warning "Warning: Checksums not available for this release"
    Write-Warning "Proceeding without checksum verification"
}

# Verify download
if (-not (Test-Path $tempFile)) {
    Write-Error "Downloaded file not found"
    Remove-LockFiles
    exit 1
}

$fileSize = (Get-Item $tempFile).Length
if ($fileSize -lt 1MB) {
    Write-Warning "Downloaded file is unusually small ($fileSize bytes)"
}

# Create installation directory
Write-Info "Installing to $InstallDir..."
try {
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Success "Created installation directory"
    }
} catch {
    Write-Error "Failed to create installation directory: $_"
    Write-Warning "Try running as administrator or choose a different location"
    Remove-LockFiles
    exit 1
}

# Install binary
$targetPath = Join-Path $InstallDir "ccproxy.exe"
try {
    Move-Item -Path $tempFile -Destination $targetPath -Force
    Write-Success "CCProxy installed successfully!"
} catch {
    Write-Error "Failed to install binary: $_"
    # Try to clean up
    Remove-Item -Path $tempFile -ErrorAction SilentlyContinue
    Remove-LockFiles
    exit 1
}

# Setup configuration
$configDir = "$env:USERPROFILE\.ccproxy"
$configFile = "$configDir\config.json"

if (-not (Test-Path $configDir)) {
    Write-Info "Creating configuration directory: $configDir"
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
}

if (-not (Test-Path $configFile)) {
    Write-Info "Creating default configuration file..."
    $defaultConfig = @'
{
  "providers": [
    {
      "name": "openai",
      "api_key": "your-openai-api-key-here",
      "api_base_url": "https://api.openai.com/v1",
      "models": ["gpt-4o", "gpt-4o-mini"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    }
  }
}
'@
    $defaultConfig | Out-File -FilePath $configFile -Encoding UTF8
    Write-Success "Created default configuration at: $configFile"
    
    # Create example config with additional providers
    $exampleFile = "$configDir\config.example.json"
    $exampleConfig = @'
{
  "_comment": "Example configuration with multiple providers",
  "providers": [
    {
      "name": "openai",
      "api_key": "your-openai-api-key-here",
      "api_base_url": "https://api.openai.com/v1",
      "models": ["gpt-4o", "gpt-4o-mini"],
      "enabled": true
    },
    {
      "_comment": "Anthropic Claude models",
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "api_base_url": "https://api.anthropic.com",
      "models": ["claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"],
      "enabled": false
    },
    {
      "_comment": "Google Gemini models",
      "name": "gemini",
      "api_key": "AI...",
      "api_base_url": "https://generativelanguage.googleapis.com/v1",
      "models": ["gemini-2.0-flash-exp", "gemini-1.5-pro"],
      "enabled": false
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    },
    "_comment_routes": "Special routes can be added here",
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022"
    }
  }
}
'@
    $exampleConfig | Out-File -FilePath $exampleFile -Encoding UTF8
    Write-Info "Example configuration saved at: $exampleFile"
    
    # Validate the generated config
    Test-JsonConfig -FilePath $configFile
} else {
    Write-Warning "Configuration already exists at: $configFile"
    # Backup existing config
    Backup-Config -ConfigFile $configFile
}

# Update PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Info "Adding $InstallDir to user PATH..."
    try {
        $newPath = $userPath + ";$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Success "Updated PATH successfully"
        Write-Warning "Note: Restart your terminal for PATH changes to take effect"
    } catch {
        Write-Warning "Failed to update PATH automatically: $_"
        Write-Info "You can manually add $InstallDir to your PATH"
    }
} else {
    Write-Success "$InstallDir is already in your PATH"
}

# Verify installation
if (Test-Path $targetPath) {
    Write-Success "CCProxy is installed at: $targetPath"
    
    # Try to get version
    try {
        $versionOutput = & $targetPath version 2>&1
        Write-Info $versionOutput
    } catch {
        # Ignore version check errors
    }
}

Write-Info ""
Write-Success "=== Installation Complete ==="
Write-Info ""
Write-Info "Next Steps:"
Write-Info ""
Write-Warning "1. Edit your configuration file:"
Write-Info "   Location: $configFile"
Write-Info "   "
Write-Info "   Run: notepad `"$configFile`""
Write-Info "   Or:  code `"$configFile`""
Write-Info "   "
Write-Info "   Replace 'your-openai-api-key-here' with your actual API key"
Write-Info ""
Write-Warning "2. Start CCProxy:"
Write-Info "   ccproxy start"
Write-Info ""
Write-Warning "3. Configure Claude Code:"
Write-Info "   ccproxy code"
Write-Info ""
Write-Info "Documentation:"
Write-Info "  Configuration Guide: https://ccproxy.orchestre.dev/guide/configuration"
Write-Info "  Provider Setup: https://ccproxy.orchestre.dev/providers/"
Write-Info "  Quick Start: https://ccproxy.orchestre.dev/guide/quick-start"
Write-Info ""
Write-Success "Tip: If 'ccproxy' command is not found, restart your terminal"

# Cleanup lock files
Remove-LockFiles