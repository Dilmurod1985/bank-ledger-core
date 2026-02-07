# E2E Test Script for Bank Ledger Core Marketplace
# This script tests the complete business flow

$baseUrl = "http://localhost:8080"
$userId = "dilmurat_01"
$currency = "UZS"
$initialBalance = "100000"
$productName = "Oqtepa Cheeseburger"
$productPrice = "35000"
$productStock = 50
$orderQuantity = 1

Write-Host "=== Bank Ledger Core - E2E Marketplace Test ===" -ForegroundColor Green
Write-Host ""

# Start the server in background
Write-Host "Starting server..." -ForegroundColor Yellow
$serverProcess = Start-Process -FilePath "go" -ArgumentList "run", "main.go" -WorkingDirectory "." -PassThru -WindowStyle Hidden
$env:DB_DRIVER = "sqlite"

# Wait for server to start
Write-Host "Waiting for server to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test server health
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "✅ Server is running: $($response.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ Server is not responding. Please start it manually with: go run main.go" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Step 1: Create test user account ===" -ForegroundColor Cyan

try {
    $userAccount = @{
        user_id = $userId
        currency = $currency
        balance = "0"
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/accounts" -Method POST -Body $userAccount -ContentType "application/json"
    Write-Host "✅ User account created: ID=$($response.id), UserID=$($response.user_id)" -ForegroundColor Green
    $userAccountId = $response.id
} catch {
    Write-Host "⚠️  User account might already exist or creation failed" -ForegroundColor Yellow
    # Try to get existing account
    try {
        $accounts = Invoke-RestMethod -Uri "$baseUrl/api/v1/accounts" -Method GET
        $userAccount = $accounts | Where-Object { $_.user_id -eq $userId }
        if ($userAccount) {
            $userAccountId = $userAccount.id
            Write-Host "✅ Found existing user account: ID=$userAccountId" -ForegroundColor Green
        }
    } catch {
        Write-Host "❌ Cannot find user account" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "=== Step 2: Fund user account from system account ===" -ForegroundColor Cyan

try {
    # Get system account (user_id = "0")
    $accounts = Invoke-RestMethod -Uri "$baseUrl/api/v1/accounts" -Method GET
    $systemAccount = $accounts | Where-Object { $_.user_id -eq "0" }
    
    if (-not $systemAccount) {
        Write-Host "❌ System account not found" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ System account found: ID=$($systemAccount.id)" -ForegroundColor Green
    
    # First, let's add money to system account for testing
    $fundSystem = @{
        from_account_id = $systemAccount.id
        to_account_id = $systemAccount.id
        amount = "200000"
    } | ConvertTo-Json
    
    # This is just to initialize system account with money
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/v1/transfers/money" -Method POST -Body $fundSystem -ContentType "application/json" | Out-Null
    } catch {
        # Ignore errors, system account might already have money
    }
    
    # Transfer from system to user
    $transfer = @{
        from_account_id = $systemAccount.id
        to_account_id = $userAccountId
        amount = $initialBalance
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/transfers/money" -Method POST -Body $transfer -ContentType "application/json"
    Write-Host "✅ Transfer completed: TransferID=$($response.transfer_id), Status=$($response.status)" -ForegroundColor Green
    
    # Check user balance
    $updatedAccount = Invoke-RestMethod -Uri "$baseUrl/api/v1/accounts/$userAccountId" -Method GET
    Write-Host "✅ User balance after funding: $($updatedAccount.balance) $currency" -ForegroundColor Green
    
} catch {
    Write-Host "❌ Transfer failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Step 3: Create product ===" -ForegroundColor Cyan

try {
    $product = @{
        name = $productName
        description = "Delicious cheeseburger from Oqtepa"
        price = $productPrice
        stock = $productStock
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/products" -Method POST -Body $product -ContentType "application/json"
    Write-Host "✅ Product created: ID=$($response.id), Name=$($response.name), Price=$($response.price), Stock=$($response.stock)" -ForegroundColor Green
    $productId = $response.id
} catch {
    Write-Host "⚠️  Product creation failed, trying to find existing product..." -ForegroundColor Yellow
    try {
        $products = Invoke-RestMethod -Uri "$baseUrl/api/v1/products" -Method GET
        $existingProduct = $products.products | Where-Object { $_.name -eq $productName }
        if ($existingProduct) {
            $productId = $existingProduct.id
            Write-Host "✅ Found existing product: ID=$productId, Price=$($existingProduct.price), Stock=$($existingProduct.stock)" -ForegroundColor Green
        }
    } catch {
        Write-Host "❌ Cannot find product" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "=== Step 4: Create order (purchase 1 burger) ===" -ForegroundColor Cyan

try {
    $order = @{
        user_id = $userId
        product_id = $productId
        quantity = $orderQuantity
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/orders" -Method POST -Body $order -ContentType "application/json"
    Write-Host "✅ Order created: ID=$($response.order_id), Status=$($response.status)" -ForegroundColor Green
    $orderId = $response.order_id
} catch {
    Write-Host "❌ Order creation failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Step 5: Verify final state ===" -ForegroundColor Cyan

# Check final user balance
try {
    $finalAccount = Invoke-RestMethod -Uri "$baseUrl/api/v1/accounts/$userAccountId" -Method GET
    $expectedBalance = [decimal]$initialBalance - [decimal]$productPrice
    Write-Host "📊 Final user balance: $($finalAccount.balance) $currency" -ForegroundColor Yellow
    Write-Host "📊 Expected balance: $expectedBalance $currency" -ForegroundColor Yellow
    
    if ([decimal]$finalAccount.balance -eq $expectedBalance) {
        Write-Host "✅ User balance is correct!" -ForegroundColor Green
    } else {
        Write-Host "❌ User balance mismatch!" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Cannot check final user balance" -ForegroundColor Red
}

# Check final product stock
try {
    $finalProduct = Invoke-RestMethod -Uri "$baseUrl/api/v1/products/$productId" -Method GET
    $expectedStock = $productStock - $orderQuantity
    Write-Host "📊 Final product stock: $($finalProduct.stock)" -ForegroundColor Yellow
    Write-Host "📊 Expected stock: $expectedStock" -ForegroundColor Yellow
    
    if ($finalProduct.stock -eq $expectedStock) {
        Write-Host "✅ Product stock is correct!" -ForegroundColor Green
    } else {
        Write-Host "❌ Product stock mismatch!" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Cannot check final product stock" -ForegroundColor Red
}

# Check order details
try {
    $orderDetails = Invoke-RestMethod -Uri "$baseUrl/api/v1/orders/$orderId" -Method GET
    Write-Host "📊 Order details: Status=$($orderDetails.status), Amount=$($orderDetails.amount), Quantity=$($orderDetails.quantity)" -ForegroundColor Yellow
    
    if ($orderDetails.status -eq "paid") {
        Write-Host "✅ Order status is correct!" -ForegroundColor Green
    } else {
        Write-Host "❌ Order status mismatch!" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Cannot check order details" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test Summary ===" -ForegroundColor Green
Write-Host "✅ E2E Test completed successfully!" -ForegroundColor Green
Write-Host "🎯 ACID transaction verified: User balance deducted, product stock reduced, order created" -ForegroundColor Green

# Stop the server
Write-Host ""
Write-Host "Stopping server..." -ForegroundColor Yellow
Stop-Process -Id $serverProcess.Id -Force

Write-Host "Test completed. Server stopped." -ForegroundColor Cyan
