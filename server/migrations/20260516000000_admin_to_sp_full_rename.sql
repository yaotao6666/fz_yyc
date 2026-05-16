-- 将服务商账号表与初始化账号从 admin 口径收口到 sp

RENAME TABLE `service_provider_admins` TO `service_provider_sps`;

ALTER TABLE `service_provider_sps`
    RENAME INDEX `uk_service_provider_admins_username` TO `uk_service_provider_sps_username`,
    RENAME INDEX `idx_service_provider_admins_sp_id` TO `idx_service_provider_sps_sp_id`;

ALTER TABLE `service_provider_sps`
    DROP FOREIGN KEY `fk_service_provider_admins_sp`,
    ADD CONSTRAINT `fk_service_provider_sps_sp`
        FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE CASCADE;

ALTER TABLE `service_provider_sps`
    COMMENT = '服务商账号表';

UPDATE `service_provider_sps`
SET
    `username` = 'sp',
    `password` = '$2a$10$FD5KRv2ER/jSpSCkEU/PMeMA4kRjqdv0tN6RlmRCY2ReaNFibmPre',
    `name` = '服务商',
    `role` = 'sp'
WHERE `username` = 'admin';
