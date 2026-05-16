# API 测试脚本
$BASE_URL = "http://localhost:8080/api/v1"
$results = @()

function Test-API {
    param([string]$Name, [scriptblock]$Test)

    Write-Host ""
    Write-Host ("=" * 60) -ForegroundColor Cyan
    Write-Host "测试: $Name" -ForegroundColor Cyan
    Write-Host ("=" * 60) -ForegroundColor Cyan

    try {
        $result = & $Test
        $script:results += $result
        return $result
    }
    catch {
        Write-Host "请求失败: $_" -ForegroundColor Red
        $script:results += @{Name=$Name; Status=$false; Detail="Error: $_"}
        return @{Name=$Name; Status=$false; Detail="Error: $_"}
    }
}

# 测试1: 服务商登录
$result = Test-API "服务商登录" {
    $body = @{username="sp"; password="tm666666"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/sp/auth/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "状态码: 200"
    Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"
    $script:spToken = $resp.data.token
    @{Name="服务商登录"; Status=($resp.code -eq 0); Detail=$resp.message}
}

# 测试2: 商家管理员登录 (验证merchant_id修复)
$result = Test-API "商家管理员登录 (验证merchant_id)" {
    $body = @{username="merchant1"; password="123456"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/merchant/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "状态码: 200"
    Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"

    $merchantId = $resp.data.merchant_id
    Write-Host ""
    Write-Host ">>> merchant_id 值: $merchantId" -ForegroundColor Yellow

    if ($merchantId -and $merchantId -gt 0) {
        Write-Host "✅ 修复成功: merchant_id 不再是 0" -ForegroundColor Green
        $passed = $true
    } else {
        Write-Host "❌ 修复失败: merchant_id 仍然是 0 或无效" -ForegroundColor Red
        $passed = $false
    }

    $script:merchantToken = $resp.data.token
    @{Name="商家登录merchant_id"; Status=$passed; Detail="merchant_id=$merchantId"}
}

# 测试3: 获取商家信息
if ($merchantToken) {
    $result = Test-API "获取商家信息" {
        $headers = @{Authorization="Bearer $merchantToken"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/profile" -Method Get -Headers $headers
        Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"
        @{Name="获取商家信息"; Status=($resp.code -eq 0); Detail=$resp.message}
    }
}

# 测试4: 创建商品带规格选项价格
if ($merchantToken) {
    $result = Test-API "创建商品带规格选项价格" {
        $body = @{
            name="测试套餐"
            price=50.0
            stock=100
            unit="份"
            description="带规格价格的测试商品"
            specs = @(
                @{
                    name="规格"
                    options = @(
                        @{name="小份"; price=48.0}
                        @{name="大份"; price=68.0}
                    )
                }
            )
        } | ConvertTo-Json -Depth 10

        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/products" -Method Post -Headers $headers -Body $body
        Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"

        $specs = $resp.data.specs
        $passed = $false
        if ($specs -and $specs[0].options[0].price) {
            Write-Host ""
            Write-Host ">>> 规格选项包含价格: $($specs[0].options[0] | ConvertTo-Json)" -ForegroundColor Yellow
            Write-Host "✅ 修复成功: 规格选项现在包含价格字段" -ForegroundColor Green
            $passed = $true
            $script:productId = $resp.data.id
        } else {
            Write-Host "❌ 修复失败: 规格选项没有价格字段" -ForegroundColor Red
        }
        @{Name="规格选项价格"; Status=$passed; Detail="specs=$($specs | ConvertTo-Json)"}
    }
}

# 测试5: 创建分类
if ($merchantToken) {
    $result = Test-API "创建分类 (验证sort字段)" {
        $body = @{name="测试分类"; sort=10} | ConvertTo-Json
        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/categories" -Method Post -Headers $headers -Body $body
        Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"

        $sortValue = $resp.data.sort
        Write-Host ""
        Write-Host ">>> sort 值为: $sortValue" -ForegroundColor Yellow

        if ($sortValue -eq 10) {
            Write-Host "✅ sort 字段设置成功" -ForegroundColor Green
            $passed = $true
            $script:categoryId = $resp.data.id
        } else {
            Write-Host "❌ sort 字段设置失败" -ForegroundColor Red
            $passed = $false
        }
        @{Name="分类sort字段"; Status=$passed; Detail="sort=$sortValue"}
    }
}

# 测试6: 更新分类sort为0
if ($merchantToken -and $categoryId) {
    $result = Test-API "更新分类sort为0 (验证sort置零)" {
        $body = @{name="测试分类"; sort=0} | ConvertTo-Json
        $headers = @{Authorization="Bearer $merchantToken"; "Content-Type"="application/json"}
        $resp = Invoke-RestMethod -Uri "$BASE_URL/merchant/categories/$categoryId" -Method Put -Headers $headers -Body $body
        Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"

        $sortValue = $resp.data.sort
        Write-Host ""
        Write-Host ">>> sort 值为: $sortValue" -ForegroundColor Yellow

        if ($sortValue -eq 0) {
            Write-Host "✅ 修复成功: sort 字段可以设置为 0" -ForegroundColor Green
            $passed = $true
        } else {
            Write-Host "❌ 修复失败: sort 值为 $sortValue，不是 0" -ForegroundColor Red
            $passed = $false
        }
        @{Name="sort置零"; Status=$passed; Detail="sort=$sortValue"}
    }
}

# 测试7: C端用户登录
$result = Test-API "C端用户登录" {
    $body = @{code="test_code_123"} | ConvertTo-Json
    $resp = Invoke-RestMethod -Uri "$BASE_URL/auth/user/login" -Method Post -Body $body -ContentType "application/json"
    Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"
    $script:userToken = $resp.data.token
    $script:userId = $resp.data.user_id
    @{Name="C端登录"; Status=($resp.code -eq 0 -and $resp.data.token -ne ""); Detail=$resp.message}
}

# 测试8: 创建订单 (库存扣减和销量更新)
if ($userToken -and $productId) {
    $result = Test-API "创建订单 (验证库存扣减和销量更新)" {
        # 先获取商品当前库存和销量
        $productResp = Invoke-RestMethod -Uri "$BASE_URL/store/1/products/$productId" -Method Get
        $beforeStock = $productResp.data.stock
        $beforeSales = $productResp.data.sales
        Write-Host ">>> 创建订单前 - stock: $beforeStock, sales: $beforeSales" -ForegroundColor Yellow

        # 创建订单
        $body = @{
            merchant_id=1
            delivery_type=1
            contact_name="测试用户"
            contact_phone="13800138000"
            delivery_address="测试地址"
            items=@(@{product_id=$productId; quantity=2})
        } | ConvertTo-Json -Depth 10

        $headers = @{Authorization="Bearer $userToken"; "Content-Type"="application/json"}
        $orderResp = Invoke-RestMethod -Uri "$BASE_URL/user/orders" -Method Post -Headers $headers -Body $body
        Write-Host "创建订单响应: $($orderResp | ConvertTo-Json -Depth 10)"

        if ($orderResp.code -eq 0) {
            # 再次获取商品信息
            Start-Sleep -Milliseconds 500
            $productResp2 = Invoke-RestMethod -Uri "$BASE_URL/store/1/products/$productId" -Method Get
            $afterStock = $productResp2.data.stock
            $afterSales = $productResp2.data.sales
            Write-Host ">>> 创建订单后 - stock: $afterStock, sales: $afterSales" -ForegroundColor Yellow

            $stockDiff = $beforeStock - $afterStock
            $salesDiff = $afterSales - $beforeSales

            $stockPassed = $false
            $salesPassed = $false

            if ($stockDiff -eq 2) {
                Write-Host "✅ 库存扣减成功: 扣减了 $stockDiff" -ForegroundColor Green
                $stockPassed = $true
            } else {
                Write-Host "❌ 库存扣减失败: 预期减少2，实际减少 $stockDiff" -ForegroundColor Red
            }

            if ($salesDiff -eq 2) {
                Write-Host "✅ 销量更新成功: 增加了 $salesDiff" -ForegroundColor Green
                $salesPassed = $true
            } else {
                Write-Host "❌ 销量更新失败: 预期增加2，实际增加 $salesDiff" -ForegroundColor Red
            }

            $script:orderId = $orderResp.data.id
            @{
                Name="库存扣减和销量更新"
                Status=($stockPassed -and $salesPassed)
                Detail="stock_diff=$stockDiff, sales_diff=$salesDiff"
            }
        } else {
            Write-Host "❌ 订单创建失败: $($orderResp.message)" -ForegroundColor Red
            @{Name="库存扣减和销量更新"; Status=$false; Detail=$orderResp.message}
        }
    }
}

# 测试9: 获取服务商配置
$result = Test-API "获取服务商配置" {
    $headers = @{Authorization="Bearer $spToken"}
    $resp = Invoke-RestMethod -Uri "$BASE_URL/sp/settings" -Method Get -Headers $headers
    Write-Host "响应: $($resp | ConvertTo-Json -Depth 10)"
    @{Name="获取服务商配置"; Status=($resp.code -eq 0); Detail=$resp.message}
}

# 汇总结果
Write-Host ""
Write-Host ("=" * 60) -ForegroundColor Cyan
Write-Host "测试结果汇总" -ForegroundColor Cyan
Write-Host ("=" * 60) -ForegroundColor Cyan

$passedCount = 0
$totalCount = $results.Count

foreach ($r in $results) {
    $statusStr = if ($r.Status) { "✅ PASS" } else { "❌ FAIL" }
    Write-Host "$statusStr - $($r.Name)"
    if ($r.Status) { $passedCount++ }
}

Write-Host ""
Write-Host "通过: $passedCount/$totalCount" -ForegroundColor $(if ($passedCount -eq $totalCount) { "Green" } else { "Yellow" })
