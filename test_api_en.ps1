$BASE_URL = "http://localhost:8080/api/v1"
$results = @()
$adminToken = ""
$merchantToken = ""
$userToken = ""
$productId = $null
$categoryId = $null
$orderId = $null

function Test-API {
    param([string]$Name, [scriptblock]$Test)
    Write-Host ""
    Write-Host ("=" * 60)
    Write-Host "TEST: $Name"
    Write-Host ("=" * 60)
    try {
        $result = & $Test
        $script:results += $result
        return $result
    }
    catch {
        Write-Host "Error: $_"
        $script:results += @{Name=$Name; Status=$false; Detail="Error: $_"}
        return @{Name=$Name; Status=$false; Detail="Error: $_"}
    }
}

# Test 1: Admin Login
$result = Test-API "Admin Login" {
    $body = @{username="admin"; password="admin123"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/admin/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
    $script:adminToken = $resp.data.token
    @{Name="Admin Login"; Status=($resp.code -eq 0); Detail=$resp.message}
}

# Test 2: Merchant Login (verify merchant_id fix)
$result = Test-API "Merchant Login (merchant_id fix)" {
    $body = @{username="merchant1"; password="123456"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/merchant/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
    $merchantId = $resp.data.merchant_id
    Write-Host ""
    Write-Host ">>> merchant_id value: $merchantId"
    if ($merchantId -and $merchantId -gt 0) {
        Write-Host "PASS: merchant_id is not 0 anymore"
        $passed = $true
    } else {
        Write-Host "FAIL: merchant_id is still 0 or invalid"
        $passed = $false
    }
    $script:merchantToken = $resp.data.token
    @{Name="Merchant Login merchant_id"; Status=$passed; Detail="merchant_id=$merchantId"}
}

# Test 3: Get Merchant Profile
if ($merchantToken) {
    $result = Test-API "Get Merchant Profile" {
        $headers = @{Authorization="Bearer $merchantToken"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/profile" -Method Get -Headers $headers
        Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
        @{Name="Get Merchant Profile"; Status=($resp.code -eq 0); Detail=$resp.message}
    }
}

# Test 4: Create Product with Spec Options Price
if ($merchantToken) {
    $result = Test-API "Create Product with Spec Options Price" {
        $body = @{
            name="Test Product"
            price=50.0
            stock=100
            unit="份"
            description="Product with spec options price"
            specs = @(
                @{
                    name="Size"
                    options = @(
                        @{name="Small"; price=48.0}
                        @{name="Large"; price=68.0}
                    )
                }
            )
        } | ConvertTo-Json -Depth 10
        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/products" -Method Post -Headers $headers -Body $body
        Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
        $specs = $resp.data.specs
        $passed = $false
        if ($specs -and $specs[0].options[0].price) {
            Write-Host ""
            Write-Host ">>> Spec options include price: $($specs[0].options[0] | ConvertTo-Json)"
            Write-Host "PASS: Spec options now include price field"
            $passed = $true
            $script:productId = $resp.data.id
        } else {
            Write-Host "FAIL: Spec options do not include price field"
        }
        @{Name="Spec Options Price"; Status=$passed; Detail="specs=$($specs | ConvertTo-Json)"}
    }
}

# Test 5: Create Category
if ($merchantToken) {
    $result = Test-API "Create Category (verify sort field)" {
        $body = @{name="Test Category"; sort=10} | ConvertTo-Json
        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/categories" -Method Post -Headers $headers -Body $body
        Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
        $sortValue = $resp.data.sort
        Write-Host ""
        Write-Host ">>> sort value: $sortValue"
        if ($sortValue -eq 10) {
            Write-Host "PASS: sort field set successfully"
            $passed = $true
            $script:categoryId = $resp.data.id
        } else {
            Write-Host "FAIL: sort field not set correctly"
            $passed = $false
        }
        @{Name="Category sort field"; Status=$passed; Detail="sort=$sortValue"}
    }
}

# Test 6: Update Category sort to 0
if ($merchantToken -and $categoryId) {
    $result = Test-API "Update Category sort to 0" {
        $body = @{name="Test Category"; sort=0} | ConvertTo-Json
        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/categories/$categoryId" -Method Put -Headers $headers -Body $body
        Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
        $sortValue = $resp.data.sort
        Write-Host ""
        Write-Host ">>> sort value: $sortValue"
        if ($sortValue -eq 0) {
            Write-Host "PASS: sort field can be set to 0"
            $passed = $true
        } else {
            Write-Host "FAIL: sort value is $sortValue, not 0"
            $passed = $false
        }
        @{Name="sort set to 0"; Status=$passed; Detail="sort=$sortValue"}
    }
}

# Test 7: C端 User Login
$result = Test-API "C-end User Login" {
    $body = @{code="test_code_123"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/user/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
    $script:userToken = $resp.data.token
    $script:userId = $resp.data.user_id
    @{Name="C-end Login"; Status=($resp.code -eq 0 -and $resp.data.token -ne ""); Detail=$resp.message}
}

# Test 8: Create Order (stock deduction and sales update)
if ($userToken -and $productId) {
    $result = Test-API "Create Order (stock & sales)" {
        $productResp = Invoke-RestMethod -Uri "$BASE_URL/store/1/products/$productId" -Method Get
        $beforeStock = $productResp.data.stock
        $beforeSales = $productResp.data.sales
        Write-Host ">>> Before order - stock: $beforeStock, sales: $beforeSales"

        $body = @{
            merchant_id=1
            delivery_type=1
            contact_name="Test User"
            contact_phone="13800138000"
            delivery_address="Test Address"
            items=@(@{product_id=$productId; quantity=2})
        } | ConvertTo-Json -Depth 10
        $headers = @{Authorization="Bearer $userToken"; "Content-Type"="application/json"}
        $orderResp = Invoke-RestMethod -Uri "$BASE_URL/user/orders" -Method Post -Headers $headers -Body $body
        Write-Host "Order Response: $($orderResp | ConvertTo-Json -Depth 10)"

        if ($orderResp.code -eq 0) {
            Start-Sleep -Milliseconds 500
            $productResp2 = Invoke-RestMethod -Uri "$BASE_URL/store/1/products/$productId" -Method Get
            $afterStock = $productResp2.data.stock
            $afterSales = $productResp2.data.sales
            Write-Host ">>> After order - stock: $afterStock, sales: $afterSales"

            $stockDiff = $beforeStock - $afterStock
            $salesDiff = $afterSales - $beforeSales

            $stockPassed = $stockDiff -eq 2
            $salesPassed = $salesDiff -eq 2

            if ($stockPassed) {
                Write-Host "PASS: Stock deducted $stockDiff"
            } else {
                Write-Host "FAIL: Expected stock -2, actual -$stockDiff"
            }

            if ($salesPassed) {
                Write-Host "PASS: Sales increased $salesDiff"
            } else {
                Write-Host "FAIL: Expected sales +2, actual +$salesDiff"
            }

            $script:orderId = $orderResp.data.id
            @{
                Name="Stock & Sales Update"
                Status=($stockPassed -and $salesPassed)
                Detail="stock_diff=$stockDiff, sales_diff=$salesDiff"
            }
        } else {
            Write-Host "FAIL: Order creation failed: $($orderResp.message)"
            @{Name="Stock & Sales Update"; Status=$false; Detail=$orderResp.message}
        }
    }
}

# Test 9: Get Service Provider Config
$result = Test-API "Get Service Provider Config" {
    $headers = @{Authorization="Bearer $adminToken"}
    $resp = Invoke-RestMethod -Uri "$BASE_URL/admin/service-provider" -Method Get -Headers $headers
    Write-Host "Response: $($resp | ConvertTo-Json -Depth 10)"
    @{Name="Get Service Provider"; Status=($resp.code -eq 0); Detail=$resp.message}
}

# Summary
Write-Host ""
Write-Host ("=" * 60)
Write-Host "TEST RESULTS SUMMARY"
Write-Host ("=" * 60)

$passedCount = 0
$totalCount = $results.Count

foreach ($r in $results) {
    $statusStr = if ($r.Status) { "PASS" } else { "FAIL" }
    Write-Host "$statusStr - $($r.Name)"
    if ($r.Status) { $passedCount++ }
}

Write-Host ""
Write-Host "Passed: $passedCount/$totalCount"
