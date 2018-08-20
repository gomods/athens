# Execute end-to-end (e2e) tests to verify that everything is working right
# from the end user perpsective
Get-Process -Name *buffalo*
$REPO_DIR = Join-Path $PSScriptRoot ".." ".."
if (-not (Test-Path env:GO_BINARY_PATH)) { $env:GO_BINARY_PATH = "go" }
$globalTmpDir = [System.IO.Path]::GetTempPath()
$tmpDirName = [GUID]::NewGuid()
$tmpDirPath = Join-Path $globalTmpDir $tmpDirName
New-Item $tmpDirPath -ItemType Directory | Out-Null
$goPath = $env:GOPATH
$GOMOD_CACHE = Join-Path $tmpDirPath "pkg" "mod"
$env:Path += ";" + "${$(Join-Path $REPO_DIR "bin")}"

function clearGoModCache () {
  if ($IsLinux) {
    # this is required because deps are read-only
    chmod -R 0770 $GOMOD_CACHE
    Get-ChildItem -Path $GOMOD_CACHE -Recurse | Remove-Item -Recurse -Force -Confirm:$false -ErrorAction SilentlyContinue
  }
  if ($IsWindows) {
    # on Win -Force passed to Remove-Item should do the trick
    Get-ChildItem -Path $GOMOD_CACHE -Recurse | Remove-Item -Recurse -Force -Confirm:$false -ErrorAction SilentlyContinue
  }
}

function stopProcesses () {
  Get-Process -Name buffalo -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  Get-Process -Name athens-build -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}

function teardown () {
  # Cleanup after our tests
  $env:GOPATH = $goPath
  stopProcesses
  Pop-Location 
  Pop-Location
}

try {
  ## Start the proxy in the background and wait for it to be ready
  Push-Location $(Join-Path $REPO_DIR cmd proxy)
  ## just in case something is still running
  stopProcesses
  Start-Process -NoNewWindow buffalo dev

  $retryNum = 0
  $maxRetryNum = 3
  $proxyUp = $false
  do {
    try {
      if ($retryNum -gt $maxRetryNum) { ThrowError -ExceptionMessage "could not start proxy" }
      $proxyUp = (Invoke-WebRequest  -Method GET -Uri http://localhost:3000).StatusCode -eq "200"
      $retryNum++
    }
    catch {
      Start-Sleep -Seconds 5
    }
  } while(-not $proxyUp)

  ## Clone our test repo
  $TEST_SOURCE = Join-Path $tmpDirPath "happy-path"
  git clone https://github.com/athens-artifacts/happy-path.git ${TEST_SOURCE}
  Push-Location ${TEST_SOURCE}

  ## set modules on after running buffalo dev, not sure why
  ## issue https://github.com/gomods/athens/issues/412
  $env:GO111MODULE = "on"

  $env:GOPATH = $tmpDirPath
  ## Make sure that our test repo works without the GOPROXY first
  if (Test-Path env:GOPROXY) { Remove-Item env:GOPROXY }
  
  & $env:GO_BINARY_PATH run .
  clearGoModCache

  ## Verify that the test works against the proxy
  $env:GOPROXY = "http://localhost:3000"
  & $env:GO_BINARY_PATH run .
  #clearGoModCache
}
finally {
  teardown
}
