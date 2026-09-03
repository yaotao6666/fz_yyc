-- MySQL dump 10.13  Distrib 5.7.39, for osx10.17 (x86_64)
--
-- Host: 127.0.0.1    Database: fz_yyc_api
-- ------------------------------------------------------
-- Server version	8.0.46

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `alert_settings`
--

DROP TABLE IF EXISTS `alert_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `alert_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '预警总开关',
  `goods_unverified_hours` int NOT NULL DEFAULT '24' COMMENT '实物超时未核销(小时)',
  `service_unassigned_hours` int NOT NULL DEFAULT '2' COMMENT '服务超时未指派(小时)',
  `escort_unfinished_minutes` int NOT NULL DEFAULT '120' COMMENT '陪诊超时未完成(分钟)',
  `service_unstarted_minutes` int NOT NULL DEFAULT '30' COMMENT '指派超时未签到(分钟)',
  `rental_overdue_hours` int NOT NULL DEFAULT '24' COMMENT '租赁逾期未归还(小时)',
  `refund_stuck_hours` int NOT NULL DEFAULT '24' COMMENT '退款卡在处理中(小时)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='预警配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `alert_settings`
--

LOCK TABLES `alert_settings` WRITE;
/*!40000 ALTER TABLE `alert_settings` DISABLE KEYS */;
INSERT INTO `alert_settings` VALUES (1,1,24,2,120,30,24,24,'2026-09-02 00:15:30','2026-09-02 00:31:35');
/*!40000 ALTER TABLE `alert_settings` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-09-02 11:16:58
