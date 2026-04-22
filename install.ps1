[CmdletBinding(SupportsShouldProcess = $true)]
param()

$ErrorActionPreference = "Stop"

$Repo = "gustavoz65/MonitoradorMain"
$Bin = "monimaster.exe"
$InstallRoot = Join-Path $env:LOCALAPPDATA "MoniMaster"
$InstallDir = Join-Path $InstallRoot "bin"
$ExePath = Join-Path $InstallDir $Bin

function Get-Arch {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
        "x64" { return "amd64" }
        "arm64" { return "arm64" }
        default { throw "Arquitetura Windows nao suportada: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }
}

function Get-ChecksumFromFile {
    param(
        [Parameter(Mandatory = $true)][string]$ChecksumsPath,
        [Parameter(Mandatory = $true)][string]$FileName
    )

    $line = Select-String -Path $ChecksumsPath -Pattern ([Regex]::Escape($FileName)) | Select-Object -First 1
    if (-not $line) {
        throw "Checksum nao encontrado para $FileName"
    }

    $parts = ($line.Line -split "\s+").Where({ $_ -ne "" })
    if ($parts.Count -lt 2) {
        throw "Formato invalido em checksums.txt"
    }
    return $parts[0].ToLowerInvariant()
}

function Add-UserPathEntry {
    param([Parameter(Mandatory = $true)][string]$PathEntry)

    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    $entries = @()
    if ($current) {
        $entries = $current -split ";" | Where-Object { $_.Trim() -ne "" }
    }

    if ($entries -contains $PathEntry) {
        return
    }

    $newPath = if ($current -and $current.Trim() -ne "") {
        "$current;$PathEntry"
    } else {
        $PathEntry
    }

    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
}

$Arch = Get-Arch
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Tag = $LatestRelease.tag_name
if (-not $Tag) {
    throw "Nao foi possivel identificar a ultima release."
}

$File = "monimaster_{0}_windows_{1}.zip" -f $Tag.TrimStart("v"), $Arch
$BaseUrl = "https://github.com/$Repo/releases/download/$Tag"
$ZipUrl = "$BaseUrl/$File"
$ChecksumsUrl = "$BaseUrl/checksums.txt"

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("monimaster-install-" + [guid]::NewGuid().ToString("N"))
$null = New-Item -ItemType Directory -Path $TempDir

try {
    $ZipPath = Join-Path $TempDir $File
    $ChecksumsPath = Join-Path $TempDir "checksums.txt"
    $ExtractDir = Join-Path $TempDir "extract"

    Write-Host "Baixando MoniMaster $Tag (windows/$Arch)..."
    Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipPath
    Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $ChecksumsPath

    $ExpectedChecksum = Get-ChecksumFromFile -ChecksumsPath $ChecksumsPath -FileName $File
    $ActualChecksum = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($ActualChecksum -ne $ExpectedChecksum) {
        throw "Checksum invalido para $File"
    }

    Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

    $DownloadedExe = Join-Path $ExtractDir $Bin
    if (-not (Test-Path $DownloadedExe)) {
        throw "Executavel $Bin nao encontrado no arquivo baixado."
    }

    $null = New-Item -ItemType Directory -Path $InstallDir -Force
    if ($PSCmdlet.ShouldProcess($ExePath, "Instalar MoniMaster")) {
        Copy-Item -Path $DownloadedExe -Destination $ExePath -Force
        Add-UserPathEntry -PathEntry $InstallDir
    }

    Write-Host "MoniMaster instalado em $ExePath"
    Write-Host "Se o comando ainda nao existir nesta janela, abra um novo terminal."
    & $ExePath version
}
finally {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force
    }
}
