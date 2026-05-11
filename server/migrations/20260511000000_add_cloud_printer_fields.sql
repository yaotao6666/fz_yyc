ALTER TABLE `cloud_printers`
  ADD COLUMN `auto_print` TINYINT(1) NOT NULL DEFAULT 0 AFTER `status`,
  ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 AFTER `auto_print`,
  ADD COLUMN `print_count` INT NOT NULL DEFAULT 0 AFTER `is_default`,
  ADD COLUMN `api_url` VARCHAR(256) DEFAULT NULL AFTER `api_key`;
