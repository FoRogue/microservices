$ErrorActionPreference = 'Stop'

function Invoke-Compose {
    docker compose @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose $args failed (exit $LASTEXITCODE)" }
}

function Send-Json($Url, $Data, $Headers = @{}) {
    Invoke-RestMethod $Url -Method Post -ContentType 'application/json' `
        -Body (ConvertTo-Json -InputObject $Data -Compress -Depth 5) -Headers $Headers -TimeoutSec 15
}

function Wait-ForApi($BaseUrl) {
    for ($attempt = 0; $attempt -lt 20; $attempt++) {
        try {
            $health = Invoke-RestMethod "$BaseUrl/health" -TimeoutSec 2
            if ($health.status -eq 'ok') { return }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    throw "API unavailable at $BaseUrl"
}

function Wait-ForPublishedRequest([guid]$Id) {
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        $published = Invoke-Compose exec -T order-db psql -U orders -d orders -Atc `
            "SELECT EXISTS (SELECT 1 FROM kafka_outbox WHERE topic = 'payments.requested' AND message_key = '$Id' AND published_at IS NOT NULL)"
        if ($published -eq 't') { return }
        Start-Sleep -Seconds 1
    }
    throw "Kafka did not acknowledge payment request for order $Id"
}

function Wait-ForOrderStatus($Url, $Headers, $Expected) {
    for ($attempt = 0; $attempt -lt 45; $attempt++) {
        $order = Invoke-RestMethod $Url -Headers $Headers -TimeoutSec 5
        if ($order.status -eq $Expected) { return $order }
        if ($order.status -ne 'PENDING_PAYMENT') { throw "Unexpected order status $($order.status)" }
        Start-Sleep -Seconds 1
    }
    throw "Timed out waiting for order status $Expected"
}

Push-Location $PSScriptRoot
try {
    & "$PSScriptRoot/start.ps1"

    $auth = 'http://localhost:8081'
    $catalog = 'http://localhost:8082'
    $payment = 'http://localhost:8083'
    $orders = 'http://localhost:8084/api/orders'
    foreach ($url in $auth, $catalog, $payment, 'http://localhost:8084') { Wait-ForApi $url }

    $credentials = @{ username = "smoke-$([guid]::NewGuid().ToString('N'))"; password = 'smoke-password' }
    $user = Send-Json "$auth/api/register" $credentials
    $token = Send-Json "$auth/api/login" $credentials
    $headers = @{ Authorization = "Bearer $($token.token)" }
    $product = Send-Json "$catalog/api/products" @{ name = 'Kafka test'; price = 9.95 }
    $order = Send-Json $orders @{ items = @(@{ product_id = $product.id; quantity = 2 }) } $headers
    if ($order.user_id -ne $user.id -or $order.status -ne 'PENDING_PAYMENT' -or $order.total -ne 19.90) {
        throw 'Order did not capture the user and original purchase amount'
    }
    $orderUrl = "$orders/$($order.id)"
    $replacement = @{ name = 'Repriced Kafka test'; price = 99 } | ConvertTo-Json -Compress
    Invoke-RestMethod "$catalog/api/products/$($product.id)" -Method Put `
        -ContentType 'application/json' -Body $replacement -TimeoutSec 5 | Out-Null

    try {
        Invoke-Compose stop payment-service
        foreach ($attempt in 1..2) {
            $response = Invoke-WebRequest "$orderUrl/pay" -Method Post -Headers $headers -TimeoutSec 10 -UseBasicParsing
            if ($response.StatusCode -ne 202) { throw 'Payment was not accepted while its consumer was offline' }
        }
        Wait-ForPublishedRequest $order.id
        $pending = Invoke-RestMethod $orderUrl -Headers $headers -TimeoutSec 5
        if ($pending.status -ne 'PENDING_PAYMENT') { throw 'Order left pending status before payment was processed' }
    } finally {
        Invoke-Compose start payment-service
    }

    $paid = Wait-ForOrderStatus $orderUrl $headers PAID
    $charges = Invoke-RestMethod "$payment/api/payments?order_id=$($order.id)" -TimeoutSec 5
    if ($charges.Count -ne 1 -or $charges[0].id -ne $paid.payment_id -or $charges[0].amount -ne 19.90) {
        throw 'Payment was duplicated or did not use the original order total'
    }

    $freeProduct = Send-Json "$catalog/api/products" @{ name = 'Zero price test'; price = 0 }
    $freeOrder = Send-Json $orders @{ items = @(@{ product_id = $freeProduct.id; quantity = 1 }) } $headers
    $freeOrderUrl = "$orders/$($freeOrder.id)"
    Invoke-RestMethod "$freeOrderUrl/pay" -Method Post -Headers $headers -TimeoutSec 5 | Out-Null
    Wait-ForOrderStatus $freeOrderUrl $headers PAYMENT_FAILED | Out-Null

    foreach ($topic in 'payments.requested', 'payments.completed') {
        $description = Invoke-Compose exec -T kafka /opt/kafka/bin/kafka-topics.sh `
            --bootstrap-server kafka:29092 --describe --topic $topic
        if (-not ($description -match 'PartitionCount:\s*3\b')) { throw "Unexpected partition count for $topic" }
    }
    $logs = Invoke-Compose logs --no-color payment-service order-service
    foreach ($suffix in 'requested','completed') {
        $metadata = $logs -match ('topic=payments\.' + $suffix + ' partition=\d+ offset=\d+ key="' + $order.id + '"')
        if (-not $metadata) { throw "Missing metadata for payments.$suffix" }
        $metadata
    }
    Write-Host 'Kafka checkout smoke test passed'
} finally {
    Pop-Location
}
