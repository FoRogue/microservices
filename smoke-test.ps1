$ErrorActionPreference = 'Stop'

function Invoke-Compose {
    docker compose @args
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose $args failed (exit $LASTEXITCODE)"
    }
}

function Wait-ForService($BaseUrl) {
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        try {
            $health = Invoke-RestMethod "$BaseUrl/health" -TimeoutSec 2
            if ($health.status -eq 'ok') { return }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    throw "Service at $BaseUrl did not become ready"
}

function Send-Json($Url, $Data, $Headers = @{}) {
    Invoke-RestMethod $Url -Method Post -ContentType 'application/json' `
        -Body (ConvertTo-Json -InputObject $Data -Compress) -Headers $Headers -TimeoutSec 10
}

Push-Location $PSScriptRoot
try {
    Invoke-Compose build
    Invoke-Compose up --wait --wait-timeout 90 auth-db catalog-db payment-db order-db
    foreach ($service in 'auth-service', 'catalog-service', 'payment-service', 'order-service') {
        Invoke-Compose run --rm $service migrate
    }
    Invoke-Compose up --wait --wait-timeout 90
    $auth = 'http://localhost:8081'
    $catalog = 'http://localhost:8082'
    $payment = 'http://localhost:8083'
    $orders = 'http://localhost:8084/api/orders'
    foreach ($url in $auth, $catalog, $payment, 'http://localhost:8084') { Wait-ForService $url }

    $tag = [guid]::NewGuid().ToString('N')
    $credentials = @{ username = "smoke-$tag"; password = 'smoke-password' }
    $user = Send-Json "$auth/api/register" $credentials
    $token = Send-Json "$auth/api/login" $credentials
    $headers = @{
        Authorization = "Bearer $($token.token)"
    }
    $profile = Invoke-RestMethod "$auth/api/me" -TimeoutSec 5 -Headers $headers
    if ($profile.id -ne $user.id -or $profile.username -ne $credentials.username) {
        throw 'Auth profile does not match the registered user'
    }

    $product = Send-Json "$catalog/api/products" @{ name = "smoke-$tag"; price = 99999999.99 }
    if ($product.price -ne 99999999.99) { throw 'Product price lost precision' }
    $cartUrl = "$catalog/api/cart?user_id=$tag"
    $item = Send-Json "$catalog/api/cart/items" @{
        user_id = $tag; product_id = $product.id; quantity = 2
    }
    $charge = Send-Json "$payment/api/payments" @{ order_id = $tag; amount = 1.005 }
    if ($charge.status -ne 'succeeded' -or $charge.amount -ne 1.01) {
        throw 'Payment status or decimal rounding is incorrect'
    }
    $cart = Invoke-RestMethod $cartUrl -TimeoutSec 5
    if ($cart.Count -ne 1 -or $cart[0].id -ne $item.id -or $cart[0].quantity -ne 2) {
        throw 'Payment changed the cart or the cart item was not saved'
    }

    try {
        Invoke-Compose stop payment-service
        $otherProduct = Send-Json "$catalog/api/products" @{ name = "offline-$tag"; price = 10 }
        $savedProduct = Invoke-RestMethod "$catalog/api/products/$($otherProduct.id)" -TimeoutSec 5
        if ($savedProduct.name -ne "offline-$tag") {
            throw 'Catalog could not save a product while payment was stopped'
        }
        $order = Send-Json $orders @{ items = @(@{ product_id = $otherProduct.id; quantity = 2 }) } $headers
        if ($order.status -ne 'PENDING_PAYMENT' -or $order.total -ne 20 -or $order.user_id -ne $user.id) {
            throw 'Order did not capture the expected user, price and pending status'
        }
        $orderUrl = "$orders/$($order.id)"
        try {
            Invoke-RestMethod "$orderUrl/pay" -Method Post -Headers $headers -TimeoutSec 5 | Out-Null
            throw 'Order payment unexpectedly worked with payment-service stopped'
        } catch {
            if ($null -eq $_.Exception.Response -or [int]$_.Exception.Response.StatusCode -ne 503) { throw }
        }
        $pending = Invoke-RestMethod $orderUrl -Headers $headers -TimeoutSec 5
        if ($pending.status -ne 'PENDING_PAYMENT') { throw 'The order is no longer pending' }
        $replacement = @{ name = "repriced-$tag"; price = 99 } | ConvertTo-Json -Compress
        Invoke-RestMethod "$catalog/api/products/$($otherProduct.id)" -Method Put `
            -ContentType 'application/json' -Body $replacement -TimeoutSec 5 | Out-Null
    } finally {
        Invoke-Compose start payment-service
    }

    Wait-ForService $payment
    $paid = Invoke-RestMethod "$orderUrl/pay" -Method Post -Headers $headers -TimeoutSec 5
    $replayed = Invoke-RestMethod "$orderUrl/pay" -Method Post -Headers $headers -TimeoutSec 5
    $paymentKey = @{ 'Idempotency-Key' = "order:$($order.id)" }
    $replayedCharge = Send-Json "$payment/api/payments" @{ order_id = $order.id; amount = 20 } $paymentKey
    $charges = Invoke-RestMethod "$payment/api/payments?order_id=$($order.id)" -TimeoutSec 5
    if ($paid.status -ne 'PAID' -or $paid.total -ne 20 -or $paid.items[0].unit_price -ne 10 -or
        $paid.items[0].product_name -ne "offline-$tag" -or $paid.payment_id -ne $replayed.payment_id -or
        $charges.Count -ne 1 -or $charges[0].id -ne $paid.payment_id -or
        $replayedCharge.id -ne $paid.payment_id -or $charges[0].amount -ne 20) {
        throw 'Order payment lost its price snapshot or was duplicated'
    }
    $savedCharge = Invoke-RestMethod "$payment/api/payments/$($charge.id)" -TimeoutSec 5
    if ($savedCharge.order_id -ne $tag -or $savedCharge.amount -ne 1.01 -or
        $savedCharge.status -ne 'succeeded') {
        throw 'Payment was not readable after restart'
    }
    Write-Host 'Smoke test passed'
} finally {
    Pop-Location
}
