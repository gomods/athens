[CmdletBinding()]
Param(
	[Parameter(Mandatory=$false)]
	[switch]$build,

	[Parameter(Mandatory=$false)]
	[switch]$run,

	[Parameter(Mandatory=$false)]
	[switch]$docs,

	[Parameter(Mandatory=$false)]
	[Alias("setup-dev-env")]
	[switch]$setup_dev_env,

	[Parameter(Mandatory=$false)]
	[switch]$verify,

	[Parameter(Mandatory=$false)]
	[switch]$test,

	[Parameter(Mandatory=$false)]
	[Alias("test-unit")]
	[switch]$test_unit,

	[Parameter(Mandatory=$false)]
	[Alias("test-e2e")]
	[switch]$test_e2e,

	[Parameter(Mandatory=$false)]
	[switch]$docker,

	[Parameter(Mandatory=$false)]
	[Alias("proxy-docker")]
	[switch]$proxy_docker,

	[Parameter(Mandatory=$false)]
	[switch]$bench,

	[Parameter(Mandatory=$false)]
	[switch]$alldeps,

	[Parameter(Mandatory=$false)]
	[switch]$dev,

	[Parameter(Mandatory=$false)]
	[switch]$down
)
function execScript($name) {
	$scriptsDir = "$(Join-Path scripts ps)"
	& "$(Join-Path $scriptsDir $name)"
}

if ($setup_dev_env.IsPresent) {
	& docker-compose -p athensdev up -d mongo
}

if ($build.IsPresent) {
	try {
		Push-Location $(Join-Path cmd proxy)
		& go build
	}
	finally {
		Pop-Location
	}

	finally {
		Pop-Location
	}
}

if ($run.IsPresent) {
	Set-Location $(Join-Path cmd proxy)
}

if ($docs.IsPresent) {
	Set-Location docs
	& hugo
}

if ($verify.IsPresent) {
	execScript "check_deps.ps1"
}

if ($alldeps.IsPresent) {
	& docker-compose -p athensdev up -d mongo
	& docker-compose -p athensdev up -d minio
	& docker-compose -p athensdev up -d jaeger
	Write-Host "sleeping for a bit to wait for the DB to come up"
	Start-Sleep 5
}

if ($dev.IsPresent) {
	& docker-compose -p athensdev up -d mongo
}

if ($test.IsPresent) {
	try {
		Push-Location  $(Join-Path cmd proxy)
	}
	finally {
		Pop-Location
	}

	finally {
		Pop-Location
	}
}

if ($test_unit.IsPresent) {
	execScript "test_unit.ps1"
}

if ($test_e2e.IsPresent) {
	execScript "test_e2e.ps1"
}

if ($docker.IsPresent) {
	& docker build -t gomods/athens -f cmd/proxy/Dockerfile .
}


if ($proxy_docker.IsPresent) {
	& docker build -t gomods/athens -f cmd/proxy/Dockerfile .
}

if ($bench.IsPresent) {
	execScript "benchmark.ps1"
}

if ($down.IsPresent) {
	& docker-compose -p athensdev down -v
}
