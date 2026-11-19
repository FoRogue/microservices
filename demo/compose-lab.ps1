param(
    [Parameter(Mandatory)]
    [ValidateSet('part1', 'part2')]
    [string]$Lab,
    [ValidateSet('up', 'down', 'ps', 'logs', 'stop-payment', 'start-payment')]
    [string]$Action = 'up',
    [switch]$LocalGo
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
$branch = @{
    part1 = 'feature/part1-microservices'
    part2 = 'feature/part2-kafka-payments'
}[$Lab]
$project = "microservices-demo-$Lab"
$cache = Join-Path $env:TEMP $project
$pointer = Join-Path $cache 'active.txt'
$built = Join-Path $cache 'built.txt'
$override = Join-Path (Join-Path $PSScriptRoot $Lab) 'compose.ports.yml'

if ($Action -eq 'up') {
    $revision = git -C $repo rev-parse --verify --quiet "$branch^{commit}"
    if ($LASTEXITCODE -ne 0) {
        $revision = git -C $repo rev-parse --verify --quiet "origin/$branch^{commit}"
        if ($LASTEXITCODE -ne 0) { throw "Branch $branch is unavailable. Run git fetch origin." }
    }
    if ($revision -notmatch '^[a-f0-9]{40,64}$') { throw 'Invalid Git revision' }
    New-Item -ItemType Directory -Path $cache -Force | Out-Null
    $snapshot = Join-Path $cache $revision
    $ready = Join-Path $snapshot '.demo-ready'
    if (-not (Test-Path -LiteralPath $ready)) {
        $archive = Join-Path $cache "$revision.zip"
        git -C $repo archive --format=zip "--output=$archive" $revision
        if ($LASTEXITCODE -ne 0) { throw "Could not archive $branch" }
        Expand-Archive -LiteralPath $archive -DestinationPath $snapshot -Force
        New-Item -ItemType File -Path $ready -Force | Out-Null
    }
    Set-Content -LiteralPath $pointer -Value $revision -Encoding ascii
}
else {
    if (-not (Test-Path -LiteralPath $pointer)) { throw "Start $Lab first" }
    $revision = (Get-Content -LiteralPath $pointer -Raw).Trim()
    if ($revision -notmatch '^[a-f0-9]{40,64}$') { throw 'Invalid saved Git revision' }
    $snapshot = Join-Path $cache $revision
}

$base = Join-Path $snapshot 'docker-compose.yml'

function Invoke-Compose {
    docker compose -p $project -f $base -f $override @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose $args failed (exit $LASTEXITCODE)" }
}

function Build-WithLocalGo {
    $previousEnv = @{
        GOOS = $env:GOOS
        GOARCH = $env:GOARCH
        CGO_ENABLED = $env:CGO_ENABLED
    }
    $architecture = docker version --format '{{.Server.Arch}}'
    if ($LASTEXITCODE -ne 0 -or $architecture -notin 'amd64', 'arm64') {
        throw 'Could not determine Docker Linux architecture'
    }
    Invoke-Compose build auth-service catalog-service
    try {
        $env:GOOS = 'linux'
        $env:GOARCH = $architecture
        $env:CGO_ENABLED = '0'
        foreach ($service in 'order-service', 'payment-service') {
            $imageContext = Join-Path $snapshot ".demo-image-$service"
            New-Item -ItemType Directory -Path $imageContext -Force | Out-Null
            Push-Location (Join-Path $snapshot $service)
            try {
                go build -o (Join-Path $imageContext 'service') .
                if ($LASTEXITCODE -ne 0) { throw "Go build failed for $service" }
            }
            finally { Pop-Location }
            @'
FROM scratch
COPY --chmod=0755 service /service
ENTRYPOINT ["/service"]
'@ | Set-Content -LiteralPath (Join-Path $imageContext 'Dockerfile') -Encoding ascii
            docker build --tag "$project-$service" $imageContext
            if ($LASTEXITCODE -ne 0) { throw "Docker image build failed for $service" }
        }
    }
    finally {
        foreach ($name in $previousEnv.Keys) {
            [Environment]::SetEnvironmentVariable($name, $previousEnv[$name], 'Process')
        }
    }
}

switch ($Action) {
    'up' {
        $appImages = @('auth', 'catalog', 'payment', 'order') | ForEach-Object { "$project-$_-service" }
        $imageRevision = if (Test-Path -LiteralPath $built) { (Get-Content -LiteralPath $built -Raw).Trim() }
        $haveImages = $false
        if ($imageRevision -eq $revision) {
            docker image inspect @appImages 2>$null | Out-Null
            $haveImages = $LASTEXITCODE -eq 0
        }
        if (-not $haveImages) {
            if ($LocalGo) { Build-WithLocalGo }
            else { Invoke-Compose build }
            Set-Content -LiteralPath $built -Value $revision -Encoding ascii
        }
        $dependencies = @('auth-db', 'catalog-db', 'payment-db', 'order-db')
        if ($Lab -eq 'part2') { $dependencies += 'kafka' }
        Invoke-Compose up -d --wait --wait-timeout 120 @dependencies
        foreach ($service in 'auth-service', 'catalog-service', 'payment-service', 'order-service') {
            Invoke-Compose run --rm --no-deps $service migrate
        }
        Invoke-Compose up -d --no-build --wait --wait-timeout 120
        Invoke-Compose ps
    }
    'down' { Invoke-Compose down }
    'ps' { Invoke-Compose ps }
    'logs' { Invoke-Compose logs --tail 80 order-service payment-service }
    'stop-payment' { Invoke-Compose stop payment-service }
    'start-payment' { Invoke-Compose start payment-service }
}
