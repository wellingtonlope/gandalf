# install.ps1 - Install gandalf on Windows without Go
# Usage (PowerShell):
#   iwr -useb https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.ps1 | iex
#   .\install.ps1 -Version v0.1.0 -Dir "$env:USERPROFILE\bin"

param(
    [string]$Version = "latest",
    [string]$Dir = "$env:USERPROFILE\bin"
)

$ErrorActionPreference = "Stop"
$Repo = "wellingtonlope/gandalf"
$Binary = "gandalf"

function Resolve-Version {
    param([string]$V)
    if ($V -eq "latest") {
        # Query GitHub API for latest tag
        $api = "https://api.github.com/repos/$Repo/releases/latest"
        try {
            $resp = Invoke-RestMethod -Uri $api -UseBasicParsing
            return $resp.tag_name
        } catch {
            Write-Error "Could not resolve latest version from $api : $_"
            exit 1
        }
    } else {
        return $V
    }
}

function Detect-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($arch -eq "AMD64") { return "amd64" }
    if ($arch -eq "ARM64") { return "arm64" }
    # Fallback check via WMI
    $wmiArch = (Get-CimInstance Win32_Processor).Architecture
    # 9 = x64, 12 = ARM64
    if ($wmiArch -eq 12) { return "arm64" }
    return "amd64"
}

$Tag = Resolve-Version -V $Version
$Arch = Detect-Arch
$Asset = "$Binary-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"

Write-Host "Installing $Binary $Tag for windows-$Arch..."
Write-Host "  URL: $Url"
Write-Host "  Dir: $Dir"

$tmp = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()) -Force
try {
    $zipPath = Join-Path $tmp.FullName $Asset
    Write-Host "Downloading..."
    Invoke-WebRequest -Uri $Url -OutFile $zipPath

    Write-Host "Extracting..."
    Expand-Archive -Path $zipPath -DestinationPath $tmp.FullName -Force
    $exe = Get-ChildItem -Path $tmp.FullName -Filter "*.exe" | Select-Object -First 1
    if (-not $exe) {
        throw "No .exe found in archive"
    }

    New-Item -ItemType Directory -Force -Path $Dir | Out-Null
    $dest = Join-Path $Dir "$Binary.exe"
    Copy-Item -Path $exe.FullName -Destination $dest -Force
    Write-Host "Installed to $dest"

    # Add Dir to user PATH if not already there
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$Dir*") {
        Write-Host "Adding $Dir to user PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$Dir", "User")
        $env:Path += ";$Dir"
        Write-Host "Added. Restart your terminal to use 'gandalf' from anywhere."
    }

    & $dest --version
    Write-Host "Try: gandalf --help"
} finally {
    Remove-Item -Recurse -Force $tmp.FullName -ErrorAction SilentlyContinue
}
