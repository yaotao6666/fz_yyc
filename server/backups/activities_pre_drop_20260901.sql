-- MySQL dump 10.13  Distrib 8.0.46, for Linux (aarch64)
--
-- Host: localhost    Database: fz_yyc_api
-- ------------------------------------------------------
-- Server version	8.0.46

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `activities`
--

DROP TABLE IF EXISTS `activities`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `activities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `link_type` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `link_value` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_activities_type` (`type`),
  KEY `idx_activities_status` (`status`),
  KEY `idx_activities_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台活动表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `activities`
--

LOCK TABLES `activities` WRITE;
/*!40000 ALTER TABLE `activities` DISABLE KEYS */;
INSERT INTO `activities` VALUES (1,'banner','夏日餐饮节','平台联动活动，商家报名享扶持','https://example.com/images/activity_1.jpg','page','/pages/activity/summer',1,1,'2026-05-01 00:00:00','2026-06-30 23:59:59','2026-05-01 08:00:00','2026-05-15 01:57:23'),(2,'announcement','系统升级通知','5月中旬将进行支付能力升级','https://example.com/images/activity_2.jpg','page','/pages/notice/detail?id=2',2,1,'2026-05-05 00:00:00','2026-05-31 23:59:59','2026-05-05 09:00:00','2026-05-15 01:57:23'),(3,'banner','商家招募计划','邀请优质商家入驻可获得奖励','https://example.com/images/activity_3.jpg','page','/pages/invite/index',3,1,'2026-04-20 00:00:00','2026-07-31 23:59:59','2026-04-20 09:00:00','2026-05-15 01:57:23'),(4,'announcement','打印服务维护公告','云打印服务预计今晚 23:00 维护','https://example.com/images/activity_4.jpg','page','/pages/notice/detail?id=4',4,1,'2026-05-10 00:00:00','2026-05-20 23:59:59','2026-05-10 09:00:00','2026-05-15 01:57:23');
/*!40000 ALTER TABLE `activities` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-09-02  9:42:38
