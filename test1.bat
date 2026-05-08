@echo off
chcp 65001 > nul
echo ============================================================
echo API Test Script
echo ============================================================
echo.

echo [1/9] Test Admin Login...
curl -s -X POST http://localhost:8080/api/v1/auth/admin/login -H "Content-Type: application/json" -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
echo.
echo.

echo [2/9] Test Merchant Login (verify merchant_id fix)...
curl -s -X POST http://localhost:8080/api/v1/auth/merchant/login -H "Content-Type: application/json" -d "{\"username\":\"merchant1\",\"password\":\"123456\"}"
echo.
echo.

echo Press any key to continue to next tests...
pause > nul
