# check_govet.ps1
# Run the linter on everything

$out = & go vet ./...
if ($LastExitCode -ne 0) {
    Write-Error $out
}
