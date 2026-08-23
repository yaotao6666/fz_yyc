$BASE_URL = "http://localhost:8080/api/v1"

Write-Host "============================================================"
Write-Host "API Test Script"
Write-Host "============================================================"

# Test 1: Service Provider Login
Write-Host ""
Write-Host "[Test 1] Service Provider Login..."
$body = @{"username"="sp"; "password"="tm666666"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "$BASE_URL/sp/auth/login" -Method Post -Body $body -ContentType "application/json"
$resp | ConvertTo-Json -Depth 10

# Test 2: Merchant Login
Write-Host ""
Write-Host "[Test 2] Merchant Login (verify RBAC login state)..."
$body = @{"username"="merchant1"; "password"="123456"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "$BASE_URL/auth/merchant/login" -Method Post -Body $body -ContentType "application/json"
$resp | ConvertTo-Json -Depth 10

$menus = $resp.data.menus
$permissions = $resp.data.permissions
Write-Host ""
if ($resp.data.token -and $menus -is [array] -and $permissions -is [array]) {
    Write-Host ">>> token/menus/permissions present (menus=$($menus.Count), perms=$($permissions.Count)) - PASS" -ForegroundColor Green
} else {
    Write-Host ">>> missing token/menus/permissions - FAIL" -ForegroundColor Red
}
