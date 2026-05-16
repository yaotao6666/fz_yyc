$body = @{"username"="sp"; "password"="tm666666"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/sp/auth/login" -Method Post -Body $body -ContentType "application/json"
$resp | ConvertTo-Json -Depth 10
