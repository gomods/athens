$repoDir = Join-Path $PSScriptRoot ".." | Join-Path -ChildPath ".."
if (-not (Join-Path $repoDir config.toml | Test-Path)) {
    $example = Join-Path $repoDir config-example.toml
    $target = Join-Path $repoDir config.toml
    Copy-Item $example $target
}
