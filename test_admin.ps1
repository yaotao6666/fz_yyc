$body = @{"username"="admin"; "password"="admin123"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/admin/login" -Method Post -Body $body -ContentType "application/json"
$resp | ConvertTo-Json -Depth 10
