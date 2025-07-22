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

# GitHub repository
$Repo = "orchestre-dev/ccproxy"
$GitHubAPI = "https://api.github.com/repos/$Repo"
$GitHubDownload = "https://github.com/$Repo/releases/download"

Write-Info "=== CCProxy Installation for Windows ==="

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
$tempFile = Join-Path $env:TEMP "ccproxy-download.exe"

Write-Info "Downloading CCProxy $Version for Windows (64-bit)..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    Write-Success "Download completed successfully"
} catch {
    Write-Error "Download failed: $_"
    Write-Error "The binary might not be available for this version."
    exit 1
}

# Verify download
if (-not (Test-Path $tempFile)) {
    Write-Error "Downloaded file not found"
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
    // Add more providers below (uncomment and configure as needed):
    // {
    //   "name": "anthropic",
    //   "api_key": "sk-ant-...",
    //   "api_base_url": "https://api.anthropic.com",
    //   "models": ["claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"],
    //   "enabled": true
    // },
    // {
    //   "name": "gemini",
    //   "api_key": "AI...",
    //   "api_base_url": "https://generativelanguage.googleapis.com/v1",
    //   "models": ["gemini-2.0-flash-exp", "gemini-1.5-pro"],
    //   "enabled": true
    // }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    }
    // Special routes (uncomment to enable):
    // "longContext": {
    //   "provider": "anthropic",
    //   "model": "claude-3-5-sonnet-20241022"
    // }
  }
}
'@
    $defaultConfig | Out-File -FilePath $configFile -Encoding UTF8
    Write-Success "Created default configuration at: $configFile"
} else {
    Write-Warning "Configuration already exists at: $configFile"
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