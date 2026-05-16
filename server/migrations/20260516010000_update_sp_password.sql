-- 将服务商默认测试账号 sp 的密码更新为 tm666666

UPDATE `service_provider_sps`
SET `password` = '$2a$10$FD5KRv2ER/jSpSCkEU/PMeMA4kRjqdv0tN6RlmRCY2ReaNFibmPre'
WHERE `username` = 'sp';
