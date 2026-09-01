-- MySQL dump 10.13  Distrib 8.1.0, for macos13.3 (arm64)
--
-- Host: 127.0.0.1    Database: fz_yyc_api
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
-- Table structure for table `care_plans`
--

DROP TABLE IF EXISTS `care_plans`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `care_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '计划名称',
  `plan_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1生活照料2基础护理3康复训练4综合康养',
  `start_date` date DEFAULT NULL COMMENT '开始日期',
  `end_date` date DEFAULT NULL COMMENT '结束日期',
  `frequency` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '照护频次',
  `goals` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '照护目标',
  `items` json DEFAULT NULL COMMENT '护理项配置[{name,desc}]',
  `assigned_staff_id` bigint unsigned DEFAULT NULL COMMENT '指派服务人员ID',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联服务订单ID(可空)',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0草稿1执行中2已暂停3已完成',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_plans_user_id` (`user_id`),
  KEY `idx_care_plans_assigned_staff_id` (`assigned_staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='照护计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `care_plans`
--

LOCK TABLES `care_plans` WRITE;
/*!40000 ALTER TABLE `care_plans` DISABLE KEYS */;
/*!40000 ALTER TABLE `care_plans` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `care_visits`
--

DROP TABLE IF EXISTS `care_visits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `care_visits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `plan_id` bigint unsigned DEFAULT NULL COMMENT '关联照护计划ID(可空)',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联服务订单ID(可空)',
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `staff_id` bigint unsigned NOT NULL COMMENT '录入服务人员ID',
  `visit_at` datetime DEFAULT NULL COMMENT '到访时间',
  `nursing_items` json DEFAULT NULL COMMENT '完成的护理项[{name,done,remark}]',
  `vitals` json DEFAULT NULL COMMENT '生命体征{blood_pressure,blood_glucose,heart_rate,oxygen,weight}',
  `photos` json DEFAULT NULL COMMENT '照片URL数组',
  `remark` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `follow_up_advice` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '下次随访建议',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_visits_plan_id` (`plan_id`),
  KEY `idx_care_visits_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上门照护记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `care_visits`
--

LOCK TABLES `care_visits` WRITE;
/*!40000 ALTER TABLE `care_visits` DISABLE KEYS */;
/*!40000 ALTER TABLE `care_visits` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `follow_up_tasks`
--

DROP TABLE IF EXISTS `follow_up_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `follow_up_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '居民用户ID',
  `task_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1康复随访2租后回访3慢病随访4评估回访',
  `source_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '来源:1服务完成2租赁归还3评估完成4手动',
  `source_id` bigint unsigned DEFAULT NULL COMMENT '来源ID(订单/评估ID)',
  `plan_follow_time` datetime DEFAULT NULL COMMENT '计划随访时间',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT '执行服务人员ID(可空=待认领)',
  `contact_method` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '随访方式:1电话2上门3微信',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态:0待执行1已完成2已跳过',
  `result` json DEFAULT NULL COMMENT '随访结果{contact_method,content,education_article_ids,satisfaction,remark}',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `remark` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_follow_up_tasks_user_id` (`user_id`),
  KEY `idx_follow_up_tasks_staff_id` (`staff_id`),
  KEY `idx_follow_up_tasks_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='随访任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `follow_up_tasks`
--

LOCK TABLES `follow_up_tasks` WRITE;
/*!40000 ALTER TABLE `follow_up_tasks` DISABLE KEYS */;
INSERT INTO `follow_up_tasks` VALUES (1,8,4,3,1,'2026-08-27 00:33:42',1,0,0,NULL,NULL,'','2026-08-24 00:33:42'),(2,8,4,3,2,'2026-08-27 00:34:02',1,1,1,'{\"remark\": \"\", \"content\": \"随访内容啊啊啊\", \"satisfaction\": 5, \"contact_method\": 1, \"education_article_ids\": []}','2026-08-24 00:45:37','','2026-08-24 00:34:02'),(3,8,4,3,3,'2026-08-27 00:34:09',1,0,0,NULL,NULL,'','2026-08-24 00:34:09'),(4,8,2,2,20,'2026-08-27 00:35:51',1,0,0,NULL,NULL,'','2026-08-24 00:35:51'),(5,8,4,3,4,'2026-08-27 00:47:52',1,0,0,NULL,NULL,'','2026-08-24 00:47:52'),(6,8,4,3,5,'2026-08-27 00:50:06',1,1,1,'{\"remark\": \"11\", \"content\": \"111\", \"satisfaction\": 5, \"contact_method\": 1, \"education_article_ids\": []}','2026-08-24 00:50:27','','2026-08-24 00:50:06'),(7,8,1,4,NULL,'2026-09-02 16:24:46',1,0,0,NULL,NULL,'','2026-08-30 16:24:46'),(8,8,1,4,NULL,'2026-09-02 16:24:50',1,0,0,NULL,NULL,'','2026-08-30 16:24:50'),(9,8,1,4,NULL,'2026-09-02 16:24:56',1,0,0,NULL,NULL,'','2026-08-30 16:24:56');
/*!40000 ALTER TABLE `follow_up_tasks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_monitoring`
--

DROP TABLE IF EXISTS `health_monitoring`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_monitoring` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `record_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型:1血压2血糖3心率4血氧5体重',
  `value` decimal(8,2) NOT NULL DEFAULT '0.00' COMMENT '测量值',
  `unit` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '单位',
  `extra` json DEFAULT NULL COMMENT '扩展(如血压高低压)',
  `recorded_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT '录入人(0=用户本人)',
  `recorded_at` datetime DEFAULT NULL COMMENT '测量时间',
  `remark` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_health_monitoring_user_id` (`user_id`),
  KEY `idx_health_monitoring_record_type` (`record_type`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生命体征监测表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_monitoring`
--

LOCK TABLES `health_monitoring` WRITE;
/*!40000 ALTER TABLE `health_monitoring` DISABLE KEYS */;
INSERT INTO `health_monitoring` VALUES (1,64,1,123.00,'mmHg','{}',0,'2026-08-24 23:50:00','','2026-08-24 23:50:31'),(2,64,2,12.00,'mmol/L','{}',0,'2026-08-24 23:50:00','','2026-08-24 23:50:40');
/*!40000 ALTER TABLE `health_monitoring` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-08-31 21:56:04
