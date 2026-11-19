$ErrorActionPreference = 'Stop'

function Invoke-Compose {
    docker compose @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose $args failed (exit $LASTEXITCODE)" }
}

Push-Location $PSScriptRoot
try {
    Invoke-Compose build
    Invoke-Compose up -d --wait --wait-timeout 120 auth-db catalog-db payment-db order-db kafka
    foreach ($service in 'auth-service', 'catalog-service', 'payment-service', 'order-service') {
        Invoke-Compose run --rm --no-deps $service migrate
    }
    Invoke-Compose up -d --wait --wait-timeout 120
} finally {
    Pop-Location
}
