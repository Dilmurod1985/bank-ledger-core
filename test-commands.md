# Manual E2E Test Commands

If you prefer to run commands manually instead of using the PowerShell script, follow these steps:

## 1. Start the server
```powershell
$env:DB_DRIVER="sqlite"
go run main.go
```

## 2. Create test user account
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/accounts" -Method POST -Body '{
    "user_id": "dilmurat_01",
    "currency": "UZS",
    "balance": "0"
}' -ContentType "application/json"
```

## 3. Get system account ID
```powershell
$accounts = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/accounts" -Method GET
$systemAccount = $accounts | Where-Object { $_.user_id -eq "0" }
$systemAccount.id
```

## 4. Fund user account (replace SYSTEM_ID and USER_ID)
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/transfers/money" -Method POST -Body '{
    "from_account_id": SYSTEM_ID,
    "to_account_id": USER_ID,
    "amount": "100000"
}' -ContentType "application/json"
```

## 5. Create product
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" -Method POST -Body '{
    "name": "Oqtepa Cheeseburger",
    "description": "Delicious cheeseburger from Oqtepa",
    "price": "35000",
    "stock": 50
}' -ContentType "application/json"
```

## 6. Create order (replace PRODUCT_ID)
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/orders" -Method POST -Body '{
    "user_id": "dilmurat_01",
    "product_id": PRODUCT_ID,
    "quantity": 1
}' -ContentType "application/json"
```

## 7. Verify results
```powershell
# Check user balance
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/accounts/USER_ID" -Method GET

# Check product stock
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products/PRODUCT_ID" -Method GET

# Check order status
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/orders/ORDER_ID" -Method GET
```

## Expected Results:
- User balance: 65,000 UZS (100,000 - 35,000)
- Product stock: 49 (50 - 1)
- Order status: "paid"
