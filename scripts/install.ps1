$ErrorActionPreference = "Stop"

$Owner = "programmersd21"
$Repo  = "kairo"
$App   = "kairo"

function Die($msg) { Write-Error $msg; exit 1 }

$os = "windows"
$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -eq "AMD64") { $arch = "x86_64" }
elseif ($arch -eq "ARM64") { $arch = "arm64" }
else { Die "Unsupported architecture: $arch" }

$asset = "${App}_${os}_${arch}.zip"
$base  = "https://github.com/$Owner/$Repo/releases/latest/download"
$archiveUrl   = "$base/$asset"
$checksumsUrl = "$base/checksums.txt"

$installDir = Join-Path $env:USERPROFILE "AppData\Local\Programs\$App"
$exePath    = Join-Path $installDir "$App.exe"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$tmp = New-Item -ItemType Directory -Force -Path ([IO.Path]::Combine([IO.Path]::GetTempPath(), "$App-install-$([Guid]::NewGuid().ToString('N'))"))
try {
  $checksumsPath = Join-Path $tmp "checksums.txt"
  $archivePath   = Join-Path $tmp $asset

  Write-Host "Downloading $App ($os/$arch)..."
  Invoke-WebRequest -UseBasicParsing -Uri $checksumsUrl -OutFile $checksumsPath
  Invoke-WebRequest -UseBasicParsing -Uri $archiveUrl -OutFile $archivePath

  $want = (Get-Content $checksumsPath | ForEach-Object {
    $parts = ($_ -split '\s+')
    if ($parts.Length -ge 2) {
      $name = $parts[$parts.Length - 1].TrimStart('*')
      if ($name -eq $asset) { return $parts[0] }
    }
  } | Select-Object -First 1)

  if (-not $want) { Die "Checksum for $asset not found in checksums.txt" }

  $got = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
  if ($got -ne $want.ToLowerInvariant()) { Die "Checksum mismatch for $asset" }

  Expand-Archive -Force -Path $archivePath -DestinationPath $tmp
  $srcExe = Join-Path $tmp "$App.exe"
  if (-not (Test-Path $srcExe)) { Die "Archive did not contain $App.exe" }

  Copy-Item -Force -Path $srcExe -Destination $exePath
  Write-Host "Installed to $exePath"
}
finally {
  Remove-Item -Recurse -Force -Path $tmp | Out-Null
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath) { $userPath = "" }
if ($userPath -notmatch [Regex]::Escape($installDir)) {
  $newPath = if ($userPath -eq "") { $installDir } else { "$userPath;$installDir" }
  [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
  $env:Path = "$env:Path;$installDir"
  Write-Host "Added $installDir to your user PATH (new terminals will pick it up)."
}

Write-Host ""
Write-Host "Verify:"
Write-Host "  $App version"

