param(
  [string]$Version = "latest"
)

$repo = "EmmyCodes234/kode"
$releases = "https://api.github.com/repos/$repo/releases"

if ($Version -eq "latest") {
  $tag = (Invoke-RestMethod "$releases/latest").tag_name
} else {
  $tag = "v$Version"
}

Write-Host "Installing Kode $tag..."

$binary = "kode-windows-amd64.exe"
$download = "https://github.com/$repo/releases/download/$tag/$binary"

$installDir = if ($env:KODE_INSTALL_DIR) { $env:KODE_INSTALL_DIR } else { "$env:LOCALAPPDATA\kode" }
$installPath = "$installDir\kode.exe"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

Write-Host "Downloading $binary..."
Invoke-WebRequest -Uri $download -OutFile $installPath

# Add to PATH if not already
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
  $env:Path = "$env:Path;$installDir"
}

Write-Host "Kode $tag installed to $installPath"
Write-Host ""
Write-Host "Run: kode init"
Write-Host "Run: kode --help"
