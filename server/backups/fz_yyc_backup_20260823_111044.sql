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
  `service_provider_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'æœåŠ¡å•†ID',
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
INSERT INTO `activities` VALUES (1,1,'banner','夏日餐饮节','平台联动活动，商家报名享扶持','https://example.com/images/activity_1.jpg','page','/pages/activity/summer',1,1,'2026-05-01 00:00:00','2026-06-30 23:59:59','2026-05-01 08:00:00','2026-05-15 01:57:23'),(2,1,'announcement','系统升级通知','5月中旬将进行支付能力升级','https://example.com/images/activity_2.jpg','page','/pages/notice/detail?id=2',2,1,'2026-05-05 00:00:00','2026-05-31 23:59:59','2026-05-05 09:00:00','2026-05-15 01:57:23'),(3,1,'banner','商家招募计划','邀请优质商家入驻可获得奖励','https://example.com/images/activity_3.jpg','page','/pages/invite/index',3,1,'2026-04-20 00:00:00','2026-07-31 23:59:59','2026-04-20 09:00:00','2026-05-15 01:57:23'),(4,1,'announcement','打印服务维护公告','云打印服务预计今晚 23:00 维护','https://example.com/images/activity_4.jpg','page','/pages/notice/detail?id=4',4,1,'2026-05-10 00:00:00','2026-05-20 23:59:59','2026-05-10 09:00:00','2026-05-15 01:57:23');
/*!40000 ALTER TABLE `activities` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `announcements`
--

DROP TABLE IF EXISTS `announcements`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `announcements` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `service_provider_id` bigint unsigned NOT NULL,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_announcements_sp_id` (`service_provider_id`),
  KEY `idx_announcements_status` (`status`),
  CONSTRAINT `fk_announcements_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `announcements`
--

LOCK TABLES `announcements` WRITE;
/*!40000 ALTER TABLE `announcements` DISABLE KEYS */;
INSERT INTO `announcements` VALUES (1,1,'五一活动复盘','请各商家及时查看活动复盘和经营建议。',1,'2026-05-02 10:00:00','2026-05-11 09:00:00'),(2,1,'新版打印模板上线','支持堂食和外卖分模板打印。',1,'2026-05-04 10:00:00','2026-05-11 09:00:00'),(3,1,'商家入驻资料规范','请新商家按照最新模板补充结算资料。',1,'2026-05-06 10:00:00','2026-05-11 09:00:00'),(4,1,'系统维护通知','本周日晚间将进行系统维护。',0,'2026-05-08 10:00:00','2026-05-11 09:00:00'),(5,1,'经营建议周报','系统已根据近期经营情况生成建议。',1,'2026-05-10 10:00:00','2026-05-11 09:00:00'),(6,1,'ok','ok',1,'2026-05-16 07:16:42','2026-05-16 07:16:42'),(7,1,'Web后台公告联调验证Web后台公告联调验证-已编辑','这是一条用于验证 web-admin 公告管理新增与编辑流程的测试公告。这是一条用于验证 web-admin 公告管理编辑保存流程的测试公告，内容已更新。',0,'2026-05-25 23:55:27','2026-05-25 23:58:23');
/*!40000 ALTER TABLE `announcements` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `care_plans`
--

DROP TABLE IF EXISTS `care_plans`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `care_plans` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'å±…æ°‘ç”¨æˆ·ID',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'è®¡åˆ’åç§°',
  `plan_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'ç±»åž‹:1ç”Ÿæ´»ç…§æ–™2åŸºç¡€æŠ¤ç†3åº·å¤è®­ç»ƒ4ç»¼åˆåº·å…»',
  `start_date` date DEFAULT NULL COMMENT 'å¼€å§‹æ—¥æœŸ',
  `end_date` date DEFAULT NULL COMMENT 'ç»“æŸæ—¥æœŸ',
  `frequency` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç…§æŠ¤é¢‘æ¬¡',
  `goals` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç…§æŠ¤ç›®æ ‡',
  `items` json DEFAULT NULL COMMENT 'æŠ¤ç†é¡¹é…ç½®[{name,desc}]',
  `assigned_staff_id` bigint unsigned DEFAULT NULL COMMENT 'æŒ‡æ´¾æœåŠ¡äººå‘˜ID',
  `order_id` bigint unsigned DEFAULT NULL COMMENT 'å…³è”æœåŠ¡è®¢å•ID(å¯ç©º)',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'çŠ¶æ€:0è‰ç¨¿1æ‰§è¡Œä¸­2å·²æš‚åœ3å·²å®Œæˆ',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_plans_user_id` (`user_id`),
  KEY `idx_care_plans_assigned_staff_id` (`assigned_staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ç…§æŠ¤è®¡åˆ’è¡¨';
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
  `plan_id` bigint unsigned DEFAULT NULL COMMENT 'å…³è”ç…§æŠ¤è®¡åˆ’ID(å¯ç©º)',
  `order_id` bigint unsigned DEFAULT NULL COMMENT 'å…³è”æœåŠ¡è®¢å•ID(å¯ç©º)',
  `user_id` bigint unsigned NOT NULL COMMENT 'å±…æ°‘ç”¨æˆ·ID',
  `staff_id` bigint unsigned NOT NULL COMMENT 'å½•å…¥æœåŠ¡äººå‘˜ID',
  `visit_at` datetime DEFAULT NULL COMMENT 'åˆ°è®¿æ—¶é—´',
  `nursing_items` json DEFAULT NULL COMMENT 'å®Œæˆçš„æŠ¤ç†é¡¹[{name,done,remark}]',
  `vitals` json DEFAULT NULL COMMENT 'ç”Ÿå‘½ä½“å¾{blood_pressure,blood_glucose,heart_rate,oxygen,weight}',
  `photos` json DEFAULT NULL COMMENT 'ç…§ç‰‡URLæ•°ç»„',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¤‡æ³¨',
  `follow_up_advice` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ä¸‹æ¬¡éšè®¿å»ºè®®',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_care_visits_plan_id` (`plan_id`),
  KEY `idx_care_visits_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ä¸Šé—¨ç…§æŠ¤è®°å½•è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `care_visits`
--

LOCK TABLES `care_visits` WRITE;
/*!40000 ALTER TABLE `care_visits` DISABLE KEYS */;
/*!40000 ALTER TABLE `care_visits` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单店默认1)',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '分类类型: 1=商品分类 2=服务分类',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父分类ID(空=一级)',
  `level` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '层级: 1=一级 2=二级 3=三级',
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_categories_merchant_id` (`merchant_id`),
  KEY `idx_categories_sort` (`merchant_id`,`sort`),
  KEY `idx_categories_parent` (`parent_id`),
  CONSTRAINT `fk_categories_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=743 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categories`
--

LOCK TABLES `categories` WRITE;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` VALUES (1,1,'轮椅',1,NULL,1,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(2,1,'助行器具',1,NULL,1,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(3,1,'无障碍扶手',1,NULL,1,3,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(4,1,'康复护理床',1,NULL,1,4,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(5,1,'康复理疗器材',1,NULL,1,5,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(6,1,'护理耗材配件',1,NULL,1,6,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(7,1,'共享租赁专区',2,NULL,1,7,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(11,1,'手动轮椅',1,1,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(12,1,'电动轮椅',1,1,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(21,1,'助行拐杖',1,2,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(22,1,'助行器/学步车',1,2,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(31,1,'马桶安全扶手',1,3,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(32,1,'走廊/浴室扶手',1,3,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(41,1,'手动护理床',1,4,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(42,1,'电动护理床',1,4,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(51,1,'四肢训练器',1,5,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(52,1,'理疗按摩仪',1,5,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(71,1,'轮椅租赁',2,7,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(72,1,'助行器租赁',2,7,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(73,1,'护理床租赁',2,7,2,3,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(74,1,'康复器械租用套餐',2,7,2,4,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(111,1,'轻便折叠款',1,11,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(112,1,'高背全躺款',1,11,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(121,1,'标准电动',1,12,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(122,1,'高续航电动',1,12,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(311,1,'可调节落地扶手',1,31,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(711,1,'周租低价套餐',2,71,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(712,1,'月租低价套餐',2,71,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(741,1,'四肢训练器租用',2,74,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(742,1,'气压按摩仪租用',2,74,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46');
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `fitting_recommendations`
--

DROP TABLE IF EXISTS `fitting_recommendations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `fitting_recommendations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'å±…æ°‘ç”¨æˆ·ID',
  `assessment_id` bigint unsigned DEFAULT NULL COMMENT 'å…³è”è¯„ä¼°è®°å½•ID',
  `symptom_desc` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç—‡çŠ¶/éœ€æ±‚æè¿°',
  `fitting_result` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'é€‚é…ç»“è®º',
  `recommended_products` json DEFAULT NULL COMMENT 'æŽ¨èå•†å“å¿«ç…§[{product_id,name,reason,sale_type}]',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT 'ç”Ÿæˆå»ºè®®çš„æœåŠ¡äººå‘˜ID',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'çŠ¶æ€:0è‰ç¨¿1å·²ç¡®è®¤2å·²ä¸‹å•',
  `order_id` bigint unsigned DEFAULT NULL COMMENT 'å…³è”è®¢å•ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_fitting_recommendations_user_id` (`user_id`),
  KEY `idx_fitting_recommendations_staff_id` (`staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='åº·å¤è¾…å…·é€‚é…å»ºè®®è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `fitting_recommendations`
--

LOCK TABLES `fitting_recommendations` WRITE;
/*!40000 ALTER TABLE `fitting_recommendations` DISABLE KEYS */;
/*!40000 ALTER TABLE `fitting_recommendations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `follow_up_tasks`
--

DROP TABLE IF EXISTS `follow_up_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `follow_up_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'å±…æ°‘ç”¨æˆ·ID',
  `task_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'ç±»åž‹:1åº·å¤éšè®¿2ç§ŸåŽå›žè®¿3æ…¢ç—…éšè®¿4è¯„ä¼°å›žè®¿',
  `source_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'æ¥æº:1æœåŠ¡å®Œæˆ2ç§Ÿèµå½’è¿˜3è¯„ä¼°å®Œæˆ4æ‰‹åŠ¨',
  `source_id` bigint unsigned DEFAULT NULL COMMENT 'æ¥æºID(è®¢å•/è¯„ä¼°ID)',
  `plan_follow_time` datetime DEFAULT NULL COMMENT 'è®¡åˆ’éšè®¿æ—¶é—´',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT 'æ‰§è¡ŒæœåŠ¡äººå‘˜ID(å¯ç©º=å¾…è®¤é¢†)',
  `contact_method` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'éšè®¿æ–¹å¼:1ç”µè¯2ä¸Šé—¨3å¾®ä¿¡',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'çŠ¶æ€:0å¾…æ‰§è¡Œ1å·²å®Œæˆ2å·²è·³è¿‡',
  `result` json DEFAULT NULL COMMENT 'éšè®¿ç»“æžœ{contact_method,content,education_article_ids,satisfaction,remark}',
  `completed_at` datetime DEFAULT NULL COMMENT 'å®Œæˆæ—¶é—´',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¤‡æ³¨',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_follow_up_tasks_user_id` (`user_id`),
  KEY `idx_follow_up_tasks_staff_id` (`staff_id`),
  KEY `idx_follow_up_tasks_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='éšè®¿ä»»åŠ¡è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `follow_up_tasks`
--

LOCK TABLES `follow_up_tasks` WRITE;
/*!40000 ALTER TABLE `follow_up_tasks` DISABLE KEYS */;
/*!40000 ALTER TABLE `follow_up_tasks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_assessment_forms`
--

DROP TABLE IF EXISTS `health_assessment_forms`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_assessment_forms` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'é‡è¡¨åç§°',
  `dimension` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'è¯„ä¼°ç»´åº¦',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'é‡è¡¨è¯´æ˜Ž',
  `questions` json DEFAULT NULL COMMENT 'é¢˜ç›®æ•°ç»„[{key,title,options:[{label,score}]}]',
  `score_rule` json DEFAULT NULL COMMENT 'è¯„åˆ†è§„åˆ™[{min,max,level,conclusion}]',
  `version` int unsigned NOT NULL DEFAULT '1' COMMENT 'ç‰ˆæœ¬å·',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '0è‰ç¨¿1å¯ç”¨',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='å¥åº·è¯„ä¼°é‡è¡¨è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_assessment_forms`
--

LOCK TABLES `health_assessment_forms` WRITE;
/*!40000 ALTER TABLE `health_assessment_forms` DISABLE KEYS */;
INSERT INTO `health_assessment_forms` VALUES (1,'AD L','adl','量表描述','[{\"key\": \"q1787367018821_1\", \"title\": \"题目内容1\", \"options\": [{\"label\": \"选项1-10\", \"score\": 10}, {\"label\": \"选项2-20\", \"score\": 20}]}, {\"key\": \"q1787367041633_2\", \"title\": \"题目内容2\", \"options\": [{\"label\": \"选项1-30\", \"score\": 30}, {\"label\": \"选项2-40\", \"score\": 40}]}]','[{\"max\": 50, \"min\": 40, \"level\": \"高风险\", \"conclusion\": \"1\"}, {\"max\": 60, \"min\": 50, \"level\": \"中高风险\", \"conclusion\": \"1\"}, {\"max\": 100, \"min\": 60, \"level\": \"低风险\", \"conclusion\": \"1\"}]',1,0,'2026-08-22 10:52:24','2026-08-22 11:27:32');
/*!40000 ALTER TABLE `health_assessment_forms` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_assessments`
--

DROP TABLE IF EXISTS `health_assessments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_assessments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'ç”¨æˆ·ID',
  `form_id` bigint unsigned NOT NULL COMMENT 'é‡è¡¨ID',
  `form_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'é‡è¡¨åç§°å¿«ç…§',
  `assessor_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '1è‡ªåŠ©2æœåŠ¡äººå‘˜',
  `staff_id` bigint unsigned DEFAULT NULL COMMENT 'è¯„ä¼°æœåŠ¡äººå‘˜ID',
  `answers` json DEFAULT NULL COMMENT 'ç­”æ¡ˆ{key:é€‰ä¸­label}',
  `total_score` decimal(6,1) DEFAULT NULL COMMENT 'æ€»åˆ†',
  `level` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'è¯„ä¼°ç­‰çº§',
  `conclusion` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'è¯„ä¼°ç»“è®º',
  `suggestions` json DEFAULT NULL COMMENT 'å»ºè®®æ•°ç»„',
  `symptom_desc` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç—‡çŠ¶æè¿°',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_health_assessments_user_id` (`user_id`),
  KEY `idx_health_assessments_form_id` (`form_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='å¥åº·è¯„ä¼°è®°å½•è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_assessments`
--

LOCK TABLES `health_assessments` WRITE;
/*!40000 ALTER TABLE `health_assessments` DISABLE KEYS */;
/*!40000 ALTER TABLE `health_assessments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_education_articles`
--

DROP TABLE IF EXISTS `health_education_articles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_education_articles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'æ ‡é¢˜',
  `category` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'åˆ†ç±»',
  `cover` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å°é¢å›¾URL',
  `content` text COLLATE utf8mb4_unicode_ci COMMENT 'æ­£æ–‡',
  `tags` json DEFAULT NULL COMMENT 'å®šå‘æ…¢ç—…æ ‡ç­¾æ•°ç»„',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'çŠ¶æ€:0è‰ç¨¿1å‘å¸ƒ',
  `publish_at` datetime DEFAULT NULL COMMENT 'å‘å¸ƒæ—¶é—´',
  `views` int unsigned NOT NULL DEFAULT '0' COMMENT 'æµè§ˆé‡',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='å¥åº·å®£æ•™å†…å®¹è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_education_articles`
--

LOCK TABLES `health_education_articles` WRITE;
/*!40000 ALTER TABLE `health_education_articles` DISABLE KEYS */;
/*!40000 ALTER TABLE `health_education_articles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_monitoring`
--

DROP TABLE IF EXISTS `health_monitoring`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_monitoring` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'ç”¨æˆ·ID',
  `record_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'ç±»åž‹:1è¡€åŽ‹2è¡€ç³–3å¿ƒçŽ‡4è¡€æ°§5ä½“é‡',
  `value` decimal(8,2) NOT NULL DEFAULT '0.00' COMMENT 'æµ‹é‡å€¼',
  `unit` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å•ä½',
  `extra` json DEFAULT NULL COMMENT 'æ‰©å±•(å¦‚è¡€åŽ‹é«˜ä½ŽåŽ‹)',
  `recorded_by` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'å½•å…¥äºº(0=ç”¨æˆ·æœ¬äºº)',
  `recorded_at` datetime DEFAULT NULL COMMENT 'æµ‹é‡æ—¶é—´',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¤‡æ³¨',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_health_monitoring_user_id` (`user_id`),
  KEY `idx_health_monitoring_record_type` (`record_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ç”Ÿå‘½ä½“å¾ç›‘æµ‹è¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_monitoring`
--

LOCK TABLES `health_monitoring` WRITE;
/*!40000 ALTER TABLE `health_monitoring` DISABLE KEYS */;
/*!40000 ALTER TABLE `health_monitoring` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_records`
--

DROP TABLE IF EXISTS `health_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `health_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT 'ç”¨æˆ·ID',
  `real_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'çœŸå®žå§“å',
  `gender` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'æ€§åˆ«:1ç”·2å¥³',
  `birth_date` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å‡ºç”Ÿæ—¥æœŸ',
  `id_card` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'èº«ä»½è¯å·',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'è”ç³»ç”µè¯',
  `emergency_contact` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç´§æ€¥è”ç³»äºº',
  `emergency_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ç´§æ€¥è”ç³»ç”µè¯',
  `address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¸¸ä½åœ°å€',
  `height_cm` decimal(5,1) DEFAULT NULL COMMENT 'èº«é«˜(cm)',
  `weight_kg` decimal(5,1) DEFAULT NULL COMMENT 'ä½“é‡(kg)',
  `blood_type` varchar(8) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'è¡€åž‹',
  `past_history` json DEFAULT NULL COMMENT 'æ—¢å¾€ç—…å²æ•°ç»„',
  `allergy_history` json DEFAULT NULL COMMENT 'è¿‡æ•å²æ•°ç»„',
  `family_history` json DEFAULT NULL COMMENT 'å®¶æ—ç—…å²æ•°ç»„',
  `surgery_history` json DEFAULT NULL COMMENT 'æ‰‹æœ¯å²æ•°ç»„',
  `medication_list` json DEFAULT NULL COMMENT 'é•¿æœŸç”¨è¯æ•°ç»„',
  `chronic_tags` json DEFAULT NULL COMMENT 'æ…¢ç—…æ ‡ç­¾æ•°ç»„',
  `smoking` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¸çƒŸæƒ…å†µ',
  `drinking` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'é¥®é…’æƒ…å†µ',
  `assessment_level` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'æœ€è¿‘ä¸€æ¬¡è¯„ä¼°ç­‰çº§',
  `remark` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'å¤‡æ³¨',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '0æœªå»ºæ¡£1æ­£å¸¸2å·²å½’æ¡£',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_health_records_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='å±…æ°‘å¥åº·æ¡£æ¡ˆè¡¨';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_records`
--

LOCK TABLES `health_records` WRITE;
/*!40000 ALTER TABLE `health_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `health_records` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_delivery_settings`
--

DROP TABLE IF EXISTS `merchant_delivery_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_delivery_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `base_fee` decimal(10,2) NOT NULL DEFAULT '0.00',
  `free_delivery_amount` decimal(10,2) NOT NULL DEFAULT '0.00',
  `max_distance` int unsigned NOT NULL DEFAULT '10',
  `distance_rules` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_delivery_settings_merchant_id` (`merchant_id`),
  CONSTRAINT `fk_merchant_delivery_settings_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家配送设置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_delivery_settings`
--

LOCK TABLES `merchant_delivery_settings` WRITE;
/*!40000 ALTER TABLE `merchant_delivery_settings` DISABLE KEYS */;
INSERT INTO `merchant_delivery_settings` VALUES (1,1,1,3.00,20.00,10,'[{\"fee\": 3, \"max_distance\": 3, \"min_distance\": 0}, {\"fee\": 5, \"max_distance\": 5, \"min_distance\": 3}]','2026-05-11 00:38:11','2026-05-19 00:29:00');
/*!40000 ALTER TABLE `merchant_delivery_settings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_fees`
--

DROP TABLE IF EXISTS `merchant_fees`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_fees` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL,
  `year` int unsigned NOT NULL,
  `amount` decimal(10,2) NOT NULL DEFAULT '0.00',
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pay_time` datetime DEFAULT NULL,
  `free_reason` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_merchant_fees_merchant_id` (`merchant_id`),
  CONSTRAINT `fk_merchant_fees_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家年费表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_fees`
--

LOCK TABLES `merchant_fees` WRITE;
/*!40000 ALTER TABLE `merchant_fees` DISABLE KEYS */;
INSERT INTO `merchant_fees` VALUES (1,1,2025,2000.00,'paid','2025-01-10 10:00:00',NULL,'2025-01-01 00:00:00','2026-05-11 09:00:00'),(2,1,2026,2000.00,'paid','2026-01-12 10:00:00',NULL,'2026-01-01 00:00:00','2026-05-11 09:00:00');
/*!40000 ALTER TABLE `merchant_fees` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_rates`
--

DROP TABLE IF EXISTS `merchant_rates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_rates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL,
  `rate_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `rate` decimal(5,4) NOT NULL,
  `effective_time` datetime NOT NULL,
  `expire_time` datetime DEFAULT NULL,
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_merchant_rates_merchant_id` (`merchant_id`),
  KEY `idx_merchant_rates_status` (`status`),
  CONSTRAINT `fk_merchant_rates_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家手续费率表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_rates`
--

LOCK TABLES `merchant_rates` WRITE;
/*!40000 ALTER TABLE `merchant_rates` DISABLE KEYS */;
INSERT INTO `merchant_rates` VALUES (1,1,'wechat_pay',0.0060,'2025-01-01 00:00:00','2025-12-31 23:59:59','上一年度标准费率',0,'2025-01-01 00:00:00','2026-05-11 09:00:00'),(2,1,'wechat_pay',0.0050,'2026-01-01 00:00:00',NULL,'当前生效费率',1,'2026-01-01 00:00:00','2026-05-11 09:00:00');
/*!40000 ALTER TABLE `merchant_rates` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_staff_roles`
--

DROP TABLE IF EXISTS `merchant_staff_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_staff_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `staff_id` bigint unsigned NOT NULL COMMENT '员工ID(merchant_staffs.id)',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID(sys_roles.id)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_staff_roles` (`staff_id`,`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_staff_roles`
--

LOCK TABLES `merchant_staff_roles` WRITE;
/*!40000 ALTER TABLE `merchant_staff_roles` DISABLE KEYS */;
INSERT INTO `merchant_staff_roles` VALUES (1,20,1,'2026-08-18 12:32:01'),(2,3,1,'2026-08-18 12:32:01'),(3,5,1,'2026-08-18 12:32:01'),(4,7,1,'2026-08-18 12:32:01'),(5,9,1,'2026-08-18 12:32:01'),(6,11,1,'2026-08-18 12:32:01'),(7,13,1,'2026-08-18 12:32:01'),(8,15,1,'2026-08-18 12:32:01'),(9,17,1,'2026-08-18 12:32:01'),(10,19,1,'2026-08-18 12:32:01'),(11,2,1,'2026-08-18 12:32:01'),(12,4,1,'2026-08-18 12:32:01'),(13,6,1,'2026-08-18 12:32:01'),(14,8,1,'2026-08-18 12:32:01'),(15,10,1,'2026-08-18 12:32:01'),(16,12,1,'2026-08-18 12:32:01'),(17,14,1,'2026-08-18 12:32:01'),(18,16,1,'2026-08-18 12:32:01'),(19,18,1,'2026-08-18 12:32:01'),(20,1,1,'2026-08-18 12:32:01');
/*!40000 ALTER TABLE `merchant_staff_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_staffs`
--

DROP TABLE IF EXISTS `merchant_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL,
  `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `unionid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信UnionID',
  `wechat_bound_at` datetime DEFAULT NULL,
  `role` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'staff',
  `notify_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `browse_notify_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `department_id` bigint unsigned DEFAULT NULL COMMENT '部门ID(sys_departments.id)',
  `last_login_at` datetime DEFAULT NULL,
  `last_wechat_login_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_staffs_merchant_username` (`merchant_id`,`username`),
  KEY `idx_merchant_staffs_openid` (`openid`),
  KEY `idx_merchant_staffs_phone` (`phone`),
  CONSTRAINT `fk_merchant_staffs_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家员工表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_staffs`
--

LOCK TABLES `merchant_staffs` WRITE;
/*!40000 ALTER TABLE `merchant_staffs` DISABLE KEYS */;
INSERT INTO `merchant_staffs` VALUES (1,1,'merchant','$2a$10$SHU1i6GAOMFnglRXvuZuWu9aBuM6qPN1Scqdp1Y9MWgpgc73t2XJ6','商家管理员','13900139000','oXPw33bdkgqWh-VJaAgWPOqz6OP4','oRIl55sOtkgv6vwy7eDexHPVCRsI','2026-05-28 01:21:16','owner',1,1,1,NULL,'2026-08-23 11:04:23','2026-05-31 23:10:53','2026-05-11 00:38:11','2026-08-23 11:04:23');
/*!40000 ALTER TABLE `merchant_staffs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchants`
--

DROP TABLE IF EXISTS `merchants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `service_provider_id` bigint unsigned NOT NULL COMMENT '所属服务商ID',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '商家名称',
  `logo` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家Logo图片地址(七牛私有路径)',
  `contact_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人姓名',
  `contact_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话(用于用户端拨打退款)',
  `contact_email` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系邮箱',
  `address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家地址',
  `lat` decimal(10,6) DEFAULT NULL COMMENT '纬度',
  `lng` decimal(10,6) DEFAULT NULL COMMENT '经度',
  `business_category` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '经营类目',
  `business_hours` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '营业时间描述',
  `announcement` text COLLATE utf8mb4_unicode_ci COMMENT '商家公告',
  `min_order_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '最低起送金额',
  `takeout_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否开启配送: true=开启 false=关闭',
  `dine_in_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否开启堂食: true=开启 false=关闭',
  `sub_mch_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付子商户号(线下进件后回填)',
  `sub_mch_status` tinyint unsigned NOT NULL DEFAULT '0',
  `applyment_status` tinyint unsigned NOT NULL DEFAULT '0',
  `audit_status` tinyint unsigned NOT NULL DEFAULT '0',
  `audit_remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '营业状态: 1=营业中 0=休息中',
  `rating` decimal(2,1) NOT NULL DEFAULT '5.0' COMMENT '商家评分(1.0-5.0)',
  `sales_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '累计销量',
  `qrcode_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `cover_image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家背景/封面图地址(七牛私有路径)',
  `payment_config_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)',
  `profit_sharing_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启自动分账',
  `qr_code_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家小程序码图片地址',
  `pickup_enabled_2` tinyint(1) NOT NULL DEFAULT '1',
  `pickup_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否开启自提: true=开启 false=关闭',
  PRIMARY KEY (`id`),
  KEY `idx_merchants_sp_id` (`service_provider_id`),
  KEY `idx_merchants_sub_mch_id` (`sub_mch_id`),
  KEY `idx_merchants_audit_status` (`audit_status`),
  KEY `idx_merchants_status` (`status`),
  KEY `idx_merchants_location` (`lat`,`lng`),
  CONSTRAINT `fk_merchants_sp` FOREIGN KEY (`service_provider_id`) REFERENCES `service_providers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchants`
--

LOCK TABLES `merchants` WRITE;
/*!40000 ALTER TABLE `merchants` DISABLE KEYS */;
INSERT INTO `merchants` VALUES (1,1,'乐享辅具（云南财旭商贸）','https://lexiang-oss.jxxme.cn/uploads/merchant/1/1778940798537.jpg','李四','13539565631','lisi@example.com','北京市朝阳区建国路88号',39.908823,116.407470,'二类医疗器械/康复辅具 与 共享租赁','08:00-21:00','全品类二类医疗辅具，支持零售与共享租赁低价套餐，专业适老康复服务。',0.00,1,1,'1112979963',2,2,1,'资质审核通过',1,5.0,0,'https://example.com/qrcode/merchant_1.png','2026-05-11 00:38:11.000','2026-08-23 09:53:51.908','https://lexiang-oss.jxxme.cn/uploads/merchant/1/1778940807903.png',1,1,NULL,1,1);
/*!40000 ALTER TABLE `merchants` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `order_items`
--

DROP TABLE IF EXISTS `order_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `merchant_id` bigint unsigned NOT NULL,
  `product_id` bigint unsigned NOT NULL,
  `product_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `price` decimal(10,2) NOT NULL,
  `quantity` int unsigned NOT NULL DEFAULT '1',
  `spec_info` json DEFAULT NULL,
  `subtotal` decimal(10,2) NOT NULL,
  `sale_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '销售类型快照: 1=一口价 2=租赁',
  `rental_unit` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '租赁计费周期快照',
  `rental_duration` int unsigned NOT NULL DEFAULT '0' COMMENT '租赁时长(下单时选择)',
  `unit_rental_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单位租金快照',
  `rental_subtotal` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '租金小计=单价×时长',
  `deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单商品押金快照',
  `deposit_deduct` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金扣除金额(损坏赔偿)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_items_order_id` (`order_id`),
  KEY `idx_order_items_merchant_id` (`merchant_id`),
  KEY `idx_order_items_product_id` (`product_id`),
  CONSTRAINT `fk_order_items_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_order_items_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `order_items`
--

LOCK TABLES `order_items` WRITE;
/*!40000 ALTER TABLE `order_items` DISABLE KEYS */;
INSERT INTO `order_items` VALUES (1,1,1,10001,'铝合金折叠轮椅',NULL,1280.00,1,NULL,1280.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 09:00:00'),(2,2,1,10006,'高续航电动轮椅',NULL,5500.00,1,NULL,5500.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 10:08:00'),(3,3,1,10003,'高背全躺轮椅',NULL,1680.00,1,NULL,1680.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 15:30:00'),(4,4,1,10007,'四脚助行拐杖',NULL,89.00,2,NULL,178.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 09:30:00'),(5,5,1,10011,'老人三脚助行架',NULL,249.00,1,NULL,249.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 12:08:00'),(6,6,1,10012,'宝宝学步车',NULL,199.00,1,NULL,199.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-14 18:00:00'),(7,7,1,10014,'U型马桶安全扶手',NULL,399.00,1,NULL,399.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 09:10:00'),(8,8,1,10015,'浴室L型扶手',NULL,129.00,1,NULL,129.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 13:30:00'),(9,8,1,10016,'走廊连续扶手',NULL,59.00,1,NULL,59.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 13:30:00'),(10,9,1,10013,'可调节马桶助力架',NULL,359.00,1,NULL,359.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 17:05:00'),(11,10,1,10017,'手动三折护理床',NULL,2680.00,1,NULL,2680.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-14 09:30:00'),(12,11,1,10019,'翻身防压疮电动床',NULL,8990.00,1,NULL,8990.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 11:10:00'),(13,12,1,10021,'肩关节滑轮训练器',NULL,299.00,1,NULL,299.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-14 08:00:00'),(14,13,1,10022,'弹力带康复套装',NULL,49.00,1,NULL,49.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 08:40:00'),(15,14,1,10023,'四肢气压按摩仪',NULL,1290.00,1,NULL,1290.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 16:20:00'),(16,15,1,10024,'防褥疮充气床垫',NULL,499.00,1,NULL,499.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-14 07:00:00'),(17,15,1,10025,'轮椅防压疮坐垫',NULL,129.00,1,NULL,129.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-14 07:00:00'),(18,16,1,10026,'透气护理垫',NULL,39.00,3,NULL,117.00,1,0,0,0.00,0.00,0.00,0.00,'2026-08-15 14:00:00'),(19,17,1,20001,'手动轮椅·周租套餐',NULL,50.00,1,NULL,50.00,2,2,1,50.00,50.00,200.00,0.00,'2026-08-12 09:00:00'),(20,18,1,20004,'手动护理床·月租套餐',NULL,300.00,1,NULL,300.00,2,3,1,300.00,300.00,800.00,0.00,'2026-07-15 09:00:00'),(21,19,1,20003,'四轮助行器·月租套餐',NULL,60.00,1,NULL,60.00,2,3,1,60.00,60.00,100.00,0.00,'2026-08-15 18:20:00'),(22,20,1,20006,'气压按摩仪·周租套餐',NULL,60.00,1,NULL,60.00,2,2,1,60.00,60.00,300.00,0.00,'2026-08-15 10:20:00');
/*!40000 ALTER TABLE `order_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `orders`
--

DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号',
  `user_id` bigint unsigned NOT NULL COMMENT '下单用户ID',
  `merchant_id` bigint unsigned NOT NULL COMMENT '商家ID',
  `order_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '订单类型: 1=零售 2=租赁 3=康养上门 4=陪诊 5=科普体验 6=长护险服务',
  `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '商品总金额(元)',
  `delivery_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '配送费(元)',
  `discount_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额(元)',
  `pay_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '实付金额(元)=总金额+配送费-优惠',
  `total_deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '总押金(元)',
  `deposit_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '押金状态: 0=无押金 1=已收 2=已退 3=部分扣除',
  `deposit_refund_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金退还金额',
  `deposit_deduct_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金扣除总额(损坏赔偿)',
  `deposit_refunded_at` datetime DEFAULT NULL COMMENT '押金退还时间',
  `rental_returned_at` datetime DEFAULT NULL COMMENT '租赁归还时间',
  `rental_return_remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '归还备注(验机情况)',
  `delivery_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '取餐方式: 1=配送 2=堂食 3=自提',
  `delivery_distance` decimal(5,2) DEFAULT NULL COMMENT '配送距离(km,配送时用户选择)',
  `delivery_address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货地址(配送时填写)',
  `contact_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人姓名(配送时填写)',
  `contact_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话(配送时填写)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款',
  `biz_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '业务子状态(按order_type语义不同)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户备注',
  `verify_code` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '核销码(已支付订单出示给商家)',
  `transaction_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付交易单号',
  `paid_at` datetime(3) DEFAULT NULL COMMENT '支付完成时间',
  `pay_notify_payload` json DEFAULT NULL COMMENT '微信支付回调原始报文JSON',
  `profit_sharing_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '分账状态:0未分账1分账中2分账成功3分账失败4已跳过',
  `profit_sharing_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '分账总额(元)',
  `profit_sharing_order_no` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账单号',
  `profit_sharing_at` datetime DEFAULT NULL COMMENT '分账完成时间',
  `profit_sharing_error` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账失败原因',
  `completed_at` datetime(3) DEFAULT NULL COMMENT '核销完成时间',
  `completed_by_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '核销操作人姓名',
  `cancelled_at` datetime(3) DEFAULT NULL COMMENT '取消时间',
  `refunded_at` datetime(3) DEFAULT NULL COMMENT '退款完成时间(status=6时写入)',
  `scheduled_at` datetime(3) DEFAULT NULL COMMENT '预约服务开始时间(康养/陪诊/科普)',
  `assigned_staff_id` bigint unsigned DEFAULT NULL COMMENT '指派的服务人员ID(service_staffs.id)',
  `actual_started_at` datetime(3) DEFAULT NULL COMMENT '实际服务开始时间(签到时间)',
  `actual_ended_at` datetime(3) DEFAULT NULL COMMENT '实际服务结束时间(签退时间)',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `pickup_point_id` bigint unsigned DEFAULT NULL COMMENT '自提点ID(自提订单)',
  `pickup_point_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '自提点名称快照',
  `pickup_point_address` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '自提点地址快照',
  `pickup_point_lat` decimal(10,6) DEFAULT NULL COMMENT '自提点纬度快照',
  `pickup_point_lng` decimal(10,6) DEFAULT NULL COMMENT '自提点经度快照',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orders_order_no` (`order_no`),
  UNIQUE KEY `idx_orders_order_no` (`order_no`),
  UNIQUE KEY `order_no` (`order_no`),
  KEY `idx_orders_user_id` (`user_id`),
  KEY `idx_orders_merchant_id` (`merchant_id`),
  KEY `idx_orders_status` (`status`),
  KEY `idx_orders_created_at` (`created_at`),
  KEY `idx_orders_paid_at` (`paid_at`),
  KEY `idx_orders_verify_code` (`verify_code`),
  KEY `idx_orders_pickup_point_id` (`pickup_point_id`),
  KEY `idx_orders_deposit_status` (`deposit_status`),
  KEY `idx_orders_order_type` (`merchant_id`,`order_type`,`status`),
  KEY `idx_orders_scheduled_at` (`merchant_id`,`scheduled_at`),
  KEY `idx_orders_assigned_staff` (`assigned_staff_id`,`status`),
  CONSTRAINT `fk_orders_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `orders`
--

LOCK TABLES `orders` WRITE;
/*!40000 ALTER TABLE `orders` DISABLE KEYS */;
INSERT INTO `orders` VALUES (1,'20260815090000112301',1,1,1,1280.00,0.00,0.00,1280.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,'北京市朝阳区建国路88号','张三','13500000001',3,0,NULL,'884201','4200001234202608151106','2026-08-15 09:05:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-15 11:20:00.000','张师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 09:00:00.000','2026-08-15 09:00:00.000',NULL,NULL,NULL,NULL,NULL),(2,'20260815100000112302',2,1,1,5500.00,0.00,50.00,5450.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,'北京市朝阳区望京SOHO T1','李四','13500000002',2,0,NULL,'386512','4200001206202608151144','2026-08-15 10:10:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 10:08:00.000','2026-08-15 10:08:00.000',NULL,NULL,NULL,NULL,NULL),(3,'20260815150000112303',3,1,1,1680.00,10.00,0.00,1690.00,0.00,0,0.00,0.00,NULL,NULL,NULL,2,NULL,NULL,NULL,NULL,1,0,NULL,NULL,NULL,NULL,NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 15:30:00.000','2026-08-15 15:30:00.000',NULL,NULL,NULL,NULL,NULL),(4,'20260815093000112304',4,1,1,178.00,5.00,0.00,183.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,3,0,NULL,'771209','4200001198202608150936','2026-08-15 09:32:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-15 10:05:00.000','王师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 09:30:00.000','2026-08-15 09:30:00.000',NULL,NULL,NULL,NULL,NULL),(5,'20260815120000112305',5,1,1,249.00,0.00,0.00,249.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,2,0,NULL,'550781','4200001216202608151210','2026-08-15 12:10:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 12:08:00.000','2026-08-15 12:08:00.000',NULL,NULL,NULL,NULL,NULL),(6,'20260814180000112306',6,1,1,199.00,0.00,0.00,199.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,4,0,NULL,NULL,NULL,NULL,NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,'2026-08-14 18:20:00.000',NULL,NULL,NULL,NULL,NULL,'2026-08-14 18:00:00.000','2026-08-14 18:20:00.000',NULL,NULL,NULL,NULL,NULL),(7,'20260815091000112307',7,1,1,399.00,0.00,0.00,399.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,3,0,NULL,'118045','4200001205202608150915','2026-08-15 09:15:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-15 10:30:00.000','王师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 09:10:00.000','2026-08-15 09:10:00.000',NULL,NULL,NULL,NULL,NULL),(8,'20260815133000112308',8,1,1,188.00,8.00,0.00,196.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,2,0,NULL,'602411','4200001218202608151335','2026-08-15 13:35:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 13:30:00.000','2026-08-15 13:30:00.000',NULL,NULL,NULL,NULL,NULL),(9,'20260815170000112309',9,1,1,359.00,0.00,0.00,359.00,0.00,0,0.00,0.00,NULL,NULL,NULL,2,NULL,NULL,NULL,NULL,1,0,NULL,NULL,NULL,NULL,NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 17:05:00.000','2026-08-15 17:05:00.000',NULL,NULL,NULL,NULL,NULL),(10,'20260814093000112310',10,1,1,2680.00,100.00,0.00,2780.00,0.00,0,0.00,0.00,NULL,NULL,NULL,3,NULL,NULL,NULL,NULL,3,0,NULL,'993045','4200001195202608140935','2026-08-14 09:35:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-14 15:40:00.000','安装师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-14 09:30:00.000','2026-08-14 09:30:00.000',NULL,NULL,NULL,NULL,NULL),(11,'20260815110000112311',11,1,1,8990.00,0.00,200.00,8790.00,0.00,0,0.00,0.00,NULL,NULL,NULL,3,NULL,NULL,NULL,NULL,2,0,NULL,'445702','4200001217202608151115','2026-08-15 11:15:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 11:10:00.000','2026-08-15 11:10:00.000',NULL,NULL,NULL,NULL,NULL),(12,'20260814080000112312',12,1,1,299.00,5.00,0.00,304.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,3,0,NULL,'223671','4200001193202608140805','2026-08-14 08:05:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-14 09:50:00.000','康复师',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-14 08:00:00.000','2026-08-14 08:00:00.000',NULL,NULL,NULL,NULL,NULL),(13,'20260815084000112313',1,1,1,49.00,0.00,0.00,49.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,5,0,NULL,NULL,'4200001202202608150845','2026-08-15 08:45:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 08:40:00.000','2026-08-15 08:50:00.000',NULL,NULL,NULL,NULL,NULL),(14,'20260815160000112314',2,1,1,1290.00,0.00,0.00,1290.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,1,0,NULL,NULL,NULL,NULL,NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 16:20:00.000','2026-08-15 16:20:00.000',NULL,NULL,NULL,NULL,NULL),(15,'20260814070000112315',3,1,1,628.00,0.00,0.00,628.00,0.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,3,0,NULL,'660139','4200001190202608140708','2026-08-14 07:08:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-14 08:30:00.000','李师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-14 07:00:00.000','2026-08-14 07:00:00.000',NULL,NULL,NULL,NULL,NULL),(16,'20260815140000112316',4,1,1,117.00,0.00,0.00,117.00,0.00,0,0.00,0.00,NULL,NULL,NULL,2,NULL,NULL,NULL,NULL,2,0,NULL,'302156','4200001213202608151410','2026-08-15 14:10:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 14:00:00.000','2026-08-15 14:00:00.000',NULL,NULL,NULL,NULL,NULL),(17,'20260812090000112317',5,1,2,50.00,0.00,0.00,250.00,200.00,1,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,2,0,NULL,'881204','4200001200202608120910','2026-08-12 09:10:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-12 09:00:00.000','2026-08-15 09:00:00.000',NULL,NULL,NULL,NULL,NULL),(18,'20260715090000112318',6,1,2,300.00,300.00,0.00,1400.00,800.00,2,800.00,0.00,'2026-08-15 09:40:00','2026-08-15 09:30:00','设备完好，已归还',3,NULL,NULL,NULL,NULL,3,0,NULL,'992071','4200001195202607150915','2026-07-15 09:15:00.000',NULL,0,0.00,NULL,NULL,NULL,'2026-08-15 09:30:00.000','张师傅',NULL,NULL,NULL,NULL,NULL,NULL,'2026-07-15 09:00:00.000','2026-08-15 09:30:00.000',NULL,NULL,NULL,NULL,NULL),(19,'20260815182000112319',7,1,2,60.00,0.00,0.00,160.00,100.00,0,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,1,0,NULL,NULL,NULL,NULL,NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 18:20:00.000','2026-08-15 18:20:00.000',NULL,NULL,NULL,NULL,NULL),(20,'20260815100000112320',8,1,2,60.00,0.00,0.00,360.00,300.00,1,0.00,0.00,NULL,NULL,NULL,1,NULL,NULL,NULL,NULL,2,0,NULL,'511648','4200001208202608151030','2026-08-15 10:30:00.000',NULL,0,0.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2026-08-15 10:20:00.000','2026-08-15 10:20:00.000',NULL,NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `orders` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `product_specs`
--

DROP TABLE IF EXISTS `product_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `product_specs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `product_id` bigint unsigned NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `options` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_specs_product_id` (`product_id`),
  CONSTRAINT `fk_product_specs_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3088 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品规格表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `product_specs`
--

LOCK TABLES `product_specs` WRITE;
/*!40000 ALTER TABLE `product_specs` DISABLE KEYS */;
/*!40000 ALTER TABLE `product_specs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单店默认1)',
  `category_id` bigint unsigned DEFAULT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `images` json DEFAULT NULL,
  `price` decimal(10,2) NOT NULL,
  `original_price` decimal(10,2) DEFAULT NULL,
  `stock` int unsigned NOT NULL DEFAULT '0',
  `unit` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '份',
  `product_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普活动 6=长护险服务',
  `service_content` json DEFAULT NULL COMMENT '服务型商品的内容描述JSON',
  `sale_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '销售类型: 1=一口价 2=租赁',
  `rental_unit` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月',
  `rental_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单位租金(元)',
  `deposit` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '押金(元)',
  `max_rental_duration` int unsigned NOT NULL DEFAULT '0' COMMENT '最大租赁时长(0=不限)',
  `sales` int unsigned NOT NULL DEFAULT '0',
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `deleted_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_products_merchant_id` (`merchant_id`),
  KEY `idx_products_category_id` (`category_id`),
  KEY `idx_products_status` (`status`),
  KEY `idx_products_sales` (`merchant_id`,`sales`),
  KEY `idx_products_sort` (`merchant_id`,`sort`),
  KEY `idx_products_sale_type` (`merchant_id`,`sale_type`),
  CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_products_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=20007 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `products`
--

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES (10001,1,111,'铝合金折叠轮椅','轻量铝合金车架，一键折叠，坐宽45cm，承重100kg，适合居家出行。','[]',1280.00,1580.00,20,'台',1,NULL,1,0,0.00,0.00,0,12,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10002,1,111,'轻便便携旅行轮椅','超轻6kg，可放入后备箱，配旅行袋，出行便携首选。','[]',1580.00,1880.00,15,'台',1,NULL,1,0,0.00,0.00,0,8,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10003,1,112,'高背全躺轮椅','大轮高背，靠背可调至全躺，适合长时间坐卧者。','[]',1680.00,2080.00,12,'台',1,NULL,1,0,0.00,0.00,0,6,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10004,1,112,'全躺式看护轮椅','座便两用，可全躺可拆卸腿托，方便护理。','[]',1880.00,2280.00,10,'台',1,NULL,1,0,0.00,0.00,0,5,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10005,1,121,'基础款电动轮椅','锂电池 12km，可折叠，遥控/手推双模式。','[]',3990.00,4590.00,8,'台',1,NULL,1,0,0.00,0.00,0,9,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10006,1,122,'高续航电动轮椅','锂电池续航20km，防后倾，上下肢驱动助力。','[]',5500.00,6290.00,6,'台',1,NULL,1,0,0.00,0.00,0,4,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10007,1,21,'四脚助行拐杖','铝合金四脚，10档高度可调，承重125kg。','[]',89.00,109.00,80,'支',1,NULL,1,0,0.00,0.00,0,20,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10008,1,21,'可调高度单拐','轻便单拐，磨砂握把防滑，出街轻巧。','[]',59.00,79.00,100,'支',1,NULL,1,0,0.00,0.00,0,25,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10009,1,21,'铝合金肘拐','肘托承重，适合单侧下肢支撑，康复期常用。','[]',129.00,159.00,60,'支',1,NULL,1,0,0.00,0.00,0,15,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10010,1,22,'四轮助行器带刹车','一键刹车，带坐垫可歇脚，适老助行。','[]',299.00,359.00,30,'个',1,NULL,1,0,0.00,0.00,0,18,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10011,1,22,'老人三脚助行架','三角稳定结构，防滑底脚，轻便稳固。','[]',249.00,299.00,25,'个',1,NULL,1,0,0.00,0.00,0,10,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10012,1,22,'宝宝学步车','护栏可拆，静音万向轮，安全学步。','[]',199.00,239.00,20,'辆',1,NULL,1,0,0.00,0.00,0,14,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10013,1,311,'可调节马桶助力架','免打孔可调宽，起身支撑更省力。','[]',359.00,419.00,20,'个',1,NULL,1,0,0.00,0.00,0,11,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10014,1,31,'U型马桶安全扶手','带防滑盖板，不锈钢承重，适老化改造。','[]',399.00,469.00,18,'个',1,NULL,1,0,0.00,0.00,0,9,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10015,1,32,'浴室L型扶手','304不锈钢，承重200kg，浴缸/淋浴墙面安装。','[]',129.00,169.00,50,'个',1,NULL,1,0,0.00,0.00,0,16,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10016,1,32,'走廊连续扶手','墙面长扶手，适老化走廊/过道安装。','[]',59.00,79.00,100,'米',1,NULL,1,0,0.00,0.00,0,13,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10017,1,41,'手动三折护理床','背/腿/脚多段升降，配护栏，家庭护理。','[]',2680.00,3180.00,10,'台',1,NULL,1,0,0.00,0.00,0,7,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10018,1,42,'五功能电动护理床','背腿升降+翻身+便孔，配防褥疮床垫。','[]',6800.00,7980.00,6,'台',1,NULL,1,0,0.00,0.00,0,5,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10019,1,42,'翻身防压疮电动床','定时翻身，减轻护理负担，适合长期卧床。','[]',8990.00,10900.00,4,'台',1,NULL,1,0,0.00,0.00,0,3,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10020,1,51,'手指握力康复训练器','五档阻力，手部精细动作康复。','[]',79.00,99.00,90,'个',1,NULL,1,0,0.00,0.00,0,22,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10021,1,51,'肩关节滑轮训练器','家用滑轮，爬墙训练，肩部术后康复。','[]',299.00,359.00,30,'个',1,NULL,1,0,0.00,0.00,0,12,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10022,1,51,'弹力带康复套装','五色拉力带，全身肌力训练。','[]',49.00,69.00,120,'套',1,NULL,1,0,0.00,0.00,0,26,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10023,1,52,'四肢气压按摩仪','多档气压，促进血液循环，缓解浮肿。','[]',1290.00,1590.00,15,'台',1,NULL,1,0,0.00,0.00,0,8,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10024,1,6,'防褥疮充气床垫','交替充气，分散压力，预防压疮。','[]',499.00,599.00,30,'个',1,NULL,1,0,0.00,0.00,0,17,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10025,1,6,'轮椅防压疮坐垫','凝胶减压，透气久坐不闷。','[]',129.00,159.00,60,'个',1,NULL,1,0,0.00,0.00,0,19,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10026,1,6,'透气护理垫','加厚防水，一次性医疗级护理垫。','[]',39.00,49.00,200,'包',1,NULL,1,0,0.00,0.00,0,30,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20001,1,711,'手动轮椅·周租套餐','共享租赁低价体验，按周计费，含押金，到期归还。','[]',50.00,0.00,5,'台/周',2,NULL,2,2,50.00,200.00,12,6,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20002,1,712,'手动轮椅·月租套餐','共享租赁低价月租，长期使用更划算。','[]',80.00,0.00,8,'台/月',2,NULL,2,3,80.00,200.00,6,9,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20003,1,72,'四轮助行器·月租套餐','带刹车助行器按月租，康复期安心用。','[]',60.00,0.00,6,'个/月',2,NULL,2,3,60.00,100.00,6,7,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20004,1,73,'手动护理床·月租套餐','护理床按月租，含安装指导，居家照护。','[]',300.00,0.00,4,'台/月',2,NULL,2,3,300.00,800.00,6,4,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20005,1,741,'四肢联动训练器·月租套餐','共享康复器械，按月租用，配合康复指导。','[]',200.00,0.00,5,'台/月',2,NULL,2,3,200.00,500.00,6,3,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20006,1,742,'气压按摩仪·周租套餐','康复理疗器械低价租用，促血液循环。','[]',60.00,0.00,6,'台/周',2,NULL,2,2,60.00,300.00,12,5,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `profit_sharing_receivers`
--

DROP TABLE IF EXISTS `profit_sharing_receivers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_receivers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单商户恒为1)',
  `receiver_type` tinyint unsigned NOT NULL COMMENT '接收方类型: 1=商户号 2=个人微信openid',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '显示名称',
  `account` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方账号(商户号 或 个人openid)',
  `personal_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人真实姓名(个人类型时, 微信实名校验, 可空)',
  `relation_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'SERVICE_PROVIDER' COMMENT '与特约商户关系, 默认SERVICE_PROVIDER',
  `default_ratio` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '自动分账默认比例(%)',
  `wechat_bound` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '微信接收方关系是否已建立: 0=未建立 1=已建立',
  `wechat_error` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信建立接收方关系失败原因',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=停用',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值(越小越靠前)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_receivers_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_receivers_account` (`account`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账接收方表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `profit_sharing_receivers`
--

LOCK TABLES `profit_sharing_receivers` WRITE;
/*!40000 ALTER TABLE `profit_sharing_receivers` DISABLE KEYS */;
INSERT INTO `profit_sharing_receivers` VALUES (1,1,2,'尧涛','o4mtI3RXnsd-YwxOUO2F-wvwuLJc','尧涛','SERVICE_PROVIDER',0.02,1,'',1,0,'test','2026-08-23 10:14:23','2026-08-23 10:42:28');
/*!40000 ALTER TABLE `profit_sharing_receivers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `profit_sharing_record_receivers`
--

DROP TABLE IF EXISTS `profit_sharing_record_receivers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_record_receivers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `record_id` bigint unsigned NOT NULL COMMENT '分账单ID',
  `receiver_id` bigint unsigned DEFAULT NULL COMMENT '接收方ID',
  `receiver_type` tinyint unsigned NOT NULL COMMENT '接收方类型快照: 1=商户号 2=个人微信openid',
  `receiver_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方名称快照',
  `account` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方账号快照',
  `amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '分账金额(元)',
  `result_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账结果: PROCESSING/SUCCESS/CLOSED/FAILED/FINISHED',
  `detail_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信分账明细单号',
  `fail_reason` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账失败原因',
  `finish_time` datetime DEFAULT NULL COMMENT '分账完成时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_record_receivers_record_id` (`record_id`),
  KEY `idx_profit_sharing_record_receivers_receiver_id` (`receiver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `profit_sharing_record_receivers`
--

LOCK TABLES `profit_sharing_record_receivers` WRITE;
/*!40000 ALTER TABLE `profit_sharing_record_receivers` DISABLE KEYS */;
/*!40000 ALTER TABLE `profit_sharing_record_receivers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `profit_sharing_records`
--

DROP TABLE IF EXISTS `profit_sharing_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `profit_sharing_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT '1' COMMENT '商家ID(单商户恒为1)',
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `order_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号',
  `sp_mchid` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务商商户号',
  `sub_mchid` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '特约商户号(分账出资方)',
  `appid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分账请求使用的appid(特约商户主体小程序)',
  `transaction_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付交易单号',
  `out_order_no` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '微信分账单号',
  `total_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '订单实付金额(元)',
  `total_share_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '本次分账总额(元)',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态: 0=待分账 1=分账中 2=分账成功 3=分账失败 4=已跳过',
  `share_time` datetime DEFAULT NULL COMMENT '分账完成时间',
  `error_message` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '失败原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_profit_sharing_records_merchant_id` (`merchant_id`),
  KEY `idx_profit_sharing_records_order_id` (`order_id`),
  KEY `idx_profit_sharing_records_order_no` (`order_no`),
  KEY `idx_profit_sharing_records_out_order_no` (`out_order_no`),
  KEY `idx_profit_sharing_records_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `profit_sharing_records`
--

LOCK TABLES `profit_sharing_records` WRITE;
/*!40000 ALTER TABLE `profit_sharing_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `profit_sharing_records` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `refunds`
--

DROP TABLE IF EXISTS `refunds`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `refunds` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `refund_no` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `refund_amount` decimal(10,2) NOT NULL,
  `refund_reason` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned NOT NULL DEFAULT '0',
  `refund_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `refunded_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refunds_refund_no` (`refund_no`),
  KEY `idx_refunds_order_id` (`order_id`),
  KEY `idx_refunds_status` (`status`),
  CONSTRAINT `fk_refunds_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `refunds`
--

LOCK TABLES `refunds` WRITE;
/*!40000 ALTER TABLE `refunds` DISABLE KEYS */;
/*!40000 ALTER TABLE `refunds` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_staffs`
--

DROP TABLE IF EXISTS `service_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `service_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '服务人员ID',
  `merchant_id` bigint unsigned NOT NULL COMMENT '所属商家ID',
  `username` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录用户名',
  `password` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加密密码(不返回)',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID(用于快捷登录)',
  `avatar` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像URL',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态: 0=待审核 1=启用 2=禁用',
  `last_login_at` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_service_staffs_merchant_id` (`merchant_id`),
  KEY `idx_service_staffs_open_id` (`openid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_staffs`
--

LOCK TABLES `service_staffs` WRITE;
/*!40000 ALTER TABLE `service_staffs` DISABLE KEYS */;
INSERT INTO `service_staffs` VALUES (1,1,'yaotao','$2a$10$lKkm0izHgfEO1r9lx9P/qea7AJxb5CeLkL7yjiysdQTWl.vpVcsaq','yaotao','13539565631','','',1,'2026-08-23 10:53:17.455','2026-08-15 22:41:30.819','2026-08-23 10:53:17.456');
/*!40000 ALTER TABLE `service_staffs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_departments`
--

DROP TABLE IF EXISTS `sys_departments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_departments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父部门ID(0=顶级)',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `leader` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sys_departments_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统部门表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_departments`
--

LOCK TABLES `sys_departments` WRITE;
/*!40000 ALTER TABLE `sys_departments` DISABLE KEYS */;
/*!40000 ALTER TABLE `sys_departments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_menus`
--

DROP TABLE IF EXISTS `sys_menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父菜单ID(0=顶级)',
  `menu_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '类型: 1=菜单/目录 2=按钮',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单名称',
  `path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '前端路由路径(菜单时)',
  `icon` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '菜单图标',
  `sort` int unsigned NOT NULL DEFAULT '0' COMMENT '排序值(越小越靠前)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `visible` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '是否显示: 1=显示 0=隐藏',
  `permission` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限标识(如 order:view)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sys_menus_parent_id` (`parent_id`),
  KEY `idx_sys_menus_permission` (`permission`)
) ENGINE=InnoDB AUTO_INCREMENT=923 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统菜单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_menus`
--

LOCK TABLES `sys_menus` WRITE;
/*!40000 ALTER TABLE `sys_menus` DISABLE KEYS */;
INSERT INTO `sys_menus` VALUES (1,0,1,'工作台','/dashboard','HomeFilled',1,1,1,'dashboard:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(2,0,1,'订单管理','/orders','Tickets',3,1,1,'orders:view','2026-08-18 12:25:47','2026-08-18 16:33:19'),(3,0,1,'商品管理','/products','Goods',2,1,1,'','2026-08-18 12:25:47','2026-08-18 16:33:15'),(4,0,1,'服务人员','/staff','User',4,1,1,'staff:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(5,0,1,'数据分析','/analytics','DataAnalysis',5,1,1,'analytics:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(6,0,1,'商家管理','/profile','Setting',6,1,1,'profile:view','2026-08-18 12:25:47','2026-08-18 15:38:41'),(7,0,1,'系统管理','/system','Setting',10,1,1,'','2026-08-18 12:25:47','2026-08-19 08:05:20'),(8,0,1,'健康服务','/health','FirstAidKit',8,1,1,NULL,'2026-08-18 22:27:42','2026-08-19 07:57:15'),(9,0,1,'分账管理','/settlement','Money',9,1,1,NULL,'2026-08-22 17:16:16','2026-08-22 17:16:16'),(21,745,2,'核销/完成','','',1,1,1,'orders:complete','2026-08-18 12:25:47','2026-08-18 15:33:20'),(22,745,2,'退款','','',2,1,1,'orders:refund','2026-08-18 12:25:47','2026-08-18 15:33:26'),(23,745,2,'归还押金','','',3,1,1,'orders:return','2026-08-18 12:25:47','2026-08-18 15:33:34'),(31,3,1,'商品管理','/products','',2,1,1,'products:view','2026-08-18 12:25:47','2026-08-18 16:33:51'),(32,3,1,'分类管理','/categories','',1,1,1,'categories:view','2026-08-18 12:25:47','2026-08-18 16:33:54'),(41,747,2,'审核/启停','','',1,1,1,'staff:update','2026-08-18 12:25:47','2026-08-18 15:40:55'),(42,747,2,'重置密码','','',2,1,1,'staff:reset-password','2026-08-18 12:25:47','2026-08-18 15:41:03'),(43,747,2,'删除','','',3,1,1,'staff:delete','2026-08-18 12:25:47','2026-08-18 15:41:08'),(44,747,2,'添加服务人员',NULL,NULL,1,1,1,'staff:create','2026-08-23 11:08:16','2026-08-23 11:08:16'),(61,746,2,'保存资料','','',1,1,1,'profile:update','2026-08-18 12:25:47','2026-08-18 15:38:57'),(62,746,2,'支付配置','','',2,1,1,'profile:payment','2026-08-18 12:25:47','2026-08-18 15:39:03'),(63,746,2,'修改密码','','',3,1,1,'profile:password','2026-08-18 12:25:47','2026-08-18 15:39:09'),(71,7,1,'菜单管理','/system/menus',NULL,1,1,1,'system:menu:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(72,7,1,'角色管理','/system/roles',NULL,2,1,1,'system:role:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(73,7,1,'部门管理','/system/departments',NULL,3,1,1,'system:dept:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(74,7,1,'员工管理','/system/staff',NULL,4,1,1,'system:staff:view','2026-08-18 12:25:47','2026-08-18 12:25:47'),(81,8,1,'健康档案','/health/records',NULL,1,1,1,'health:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(82,8,1,'评估量表','/health/assessment-forms',NULL,2,1,1,'assessment:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(83,8,1,'评估记录','/health/assessments',NULL,3,1,1,'assessment:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(84,8,1,'适配建议','/health/fitting',NULL,4,1,1,'fitting:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(85,8,1,'照护计划','/health/care-plans',NULL,5,1,1,'care:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(86,8,1,'随访任务','/health/follow-ups',NULL,6,1,1,'followup:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(87,8,1,'生命体征','/health/monitoring',NULL,7,1,1,'monitor:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(88,8,1,'健康宣教','/health/education',NULL,8,1,1,'education:view','2026-08-18 22:27:42','2026-08-19 07:57:15'),(91,9,1,'分账接收方','/settlement/receivers',NULL,1,1,1,'profit:view','2026-08-22 17:16:16','2026-08-22 17:16:16'),(92,9,1,'分账记录','/settlement/records',NULL,2,1,1,'profit:view','2026-08-22 17:16:16','2026-08-22 17:16:16'),(311,31,2,'新增/编辑',NULL,NULL,1,1,1,'products:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(312,31,2,'删除',NULL,NULL,2,1,1,'products:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(313,31,2,'上/下架/批量',NULL,NULL,3,1,1,'products:status','2026-08-18 12:25:47','2026-08-18 12:25:47'),(314,31,2,'库存',NULL,NULL,4,1,1,'products:stock','2026-08-18 12:25:47','2026-08-18 12:25:47'),(315,31,2,'规格',NULL,NULL,5,1,1,'products:specs','2026-08-18 12:25:47','2026-08-18 12:25:47'),(321,32,2,'新增/编辑',NULL,NULL,1,1,1,'categories:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(322,32,2,'删除',NULL,NULL,2,1,1,'categories:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(323,32,2,'排序',NULL,NULL,3,1,1,'categories:sort','2026-08-18 12:25:47','2026-08-18 12:25:47'),(711,71,2,'新增/编辑',NULL,NULL,1,1,1,'system:menu:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(712,71,2,'删除',NULL,NULL,2,1,1,'system:menu:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(721,72,2,'新增/编辑',NULL,NULL,1,1,1,'system:role:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(722,72,2,'删除',NULL,NULL,2,1,1,'system:role:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(723,72,2,'分配权限',NULL,NULL,3,1,1,'system:role:assign','2026-08-18 12:25:47','2026-08-18 12:25:47'),(731,73,2,'新增/编辑',NULL,NULL,1,1,1,'system:dept:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(732,73,2,'删除',NULL,NULL,2,1,1,'system:dept:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(741,74,2,'新增',NULL,NULL,1,1,1,'system:staff:create','2026-08-18 12:25:47','2026-08-18 12:25:47'),(742,74,2,'编辑',NULL,NULL,2,1,1,'system:staff:update','2026-08-18 12:25:47','2026-08-18 12:25:47'),(743,74,2,'删除',NULL,NULL,3,1,1,'system:staff:delete','2026-08-18 12:25:47','2026-08-18 12:25:47'),(744,74,2,'重置密码',NULL,NULL,4,1,1,'system:staff:reset-password','2026-08-18 12:25:47','2026-08-18 12:25:47'),(745,2,1,'订单列表','/orders','Tickets',0,1,1,'orders:view','2026-08-18 15:33:04','2026-08-18 15:44:02'),(746,6,1,'商家资料','/profile','Setting',0,1,1,'profile:view','2026-08-18 15:38:32','2026-08-18 15:39:16'),(747,4,1,'服务人员','/staff','',0,1,1,'staff:view','2026-08-18 15:40:31','2026-08-18 15:40:48'),(811,81,2,'编辑档案',NULL,NULL,1,1,1,'health:update','2026-08-18 22:27:42','2026-08-19 07:57:15'),(821,82,2,'新增量表',NULL,NULL,1,1,1,'assessment:create','2026-08-18 22:27:42','2026-08-22 11:38:54'),(822,82,2,'编辑/启停量表',NULL,NULL,2,1,1,'assessment:update','2026-08-22 11:38:54','2026-08-22 11:38:54'),(823,82,2,'删除量表',NULL,NULL,3,1,1,'assessment:delete','2026-08-22 11:38:54','2026-08-22 11:38:54'),(831,83,2,'手动登记',NULL,NULL,1,1,1,'assessment:create','2026-08-18 22:27:42','2026-08-19 07:57:15'),(841,84,2,'编辑/确认',NULL,NULL,1,1,1,'fitting:update','2026-08-18 22:27:42','2026-08-19 07:57:15'),(851,85,2,'新增照护计划',NULL,NULL,1,1,1,'care:create','2026-08-18 22:27:42','2026-08-22 11:38:54'),(852,85,2,'编辑照护计划',NULL,NULL,2,1,1,'care:update','2026-08-22 11:38:54','2026-08-22 11:38:54'),(853,85,2,'删除照护计划',NULL,NULL,3,1,1,'care:delete','2026-08-22 11:38:54','2026-08-22 11:38:54'),(861,86,2,'执行/登记',NULL,NULL,1,1,1,'followup:update','2026-08-18 22:27:42','2026-08-19 07:57:15'),(871,87,2,'录入',NULL,NULL,1,1,1,'monitor:create','2026-08-18 22:27:42','2026-08-19 07:57:15'),(881,88,2,'新增/编辑',NULL,NULL,1,1,1,'education:create','2026-08-18 22:27:42','2026-08-19 07:57:15'),(911,91,2,'新增/编辑/删除/同步接收方',NULL,NULL,1,1,1,'profit:receiver:manage','2026-08-22 17:16:16','2026-08-22 17:16:16'),(921,92,2,'分账/重试',NULL,NULL,1,1,1,'profit:share','2026-08-22 17:16:16','2026-08-22 17:16:16'),(922,92,2,'分账配置',NULL,NULL,2,1,1,'profit:config','2026-08-22 17:16:16','2026-08-22 17:16:16');
/*!40000 ALTER TABLE `sys_menus` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_role_menus`
--

DROP TABLE IF EXISTS `sys_role_menus`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role_menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_menus` (`role_id`,`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=112 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_role_menus`
--

LOCK TABLES `sys_role_menus` WRITE;
/*!40000 ALTER TABLE `sys_role_menus` DISABLE KEYS */;
INSERT INTO `sys_role_menus` VALUES (1,1,1,'2026-08-18 12:25:47'),(2,1,2,'2026-08-18 12:25:47'),(3,1,3,'2026-08-18 12:25:47'),(4,1,4,'2026-08-18 12:25:47'),(5,1,5,'2026-08-18 12:25:47'),(6,1,6,'2026-08-18 12:25:47'),(7,1,7,'2026-08-18 12:25:47'),(8,1,21,'2026-08-18 12:25:47'),(9,1,22,'2026-08-18 12:25:47'),(10,1,23,'2026-08-18 12:25:47'),(11,1,31,'2026-08-18 12:25:47'),(12,1,32,'2026-08-18 12:25:47'),(13,1,41,'2026-08-18 12:25:47'),(14,1,42,'2026-08-18 12:25:47'),(15,1,43,'2026-08-18 12:25:47'),(16,1,61,'2026-08-18 12:25:47'),(17,1,62,'2026-08-18 12:25:47'),(18,1,63,'2026-08-18 12:25:47'),(19,1,71,'2026-08-18 12:25:47'),(20,1,72,'2026-08-18 12:25:47'),(21,1,73,'2026-08-18 12:25:47'),(22,1,74,'2026-08-18 12:25:47'),(23,1,311,'2026-08-18 12:25:47'),(24,1,312,'2026-08-18 12:25:47'),(25,1,313,'2026-08-18 12:25:47'),(26,1,314,'2026-08-18 12:25:47'),(27,1,315,'2026-08-18 12:25:47'),(28,1,321,'2026-08-18 12:25:47'),(29,1,322,'2026-08-18 12:25:47'),(30,1,323,'2026-08-18 12:25:47'),(31,1,711,'2026-08-18 12:25:47'),(32,1,712,'2026-08-18 12:25:47'),(33,1,721,'2026-08-18 12:25:47'),(34,1,722,'2026-08-18 12:25:47'),(35,1,723,'2026-08-18 12:25:47'),(36,1,731,'2026-08-18 12:25:47'),(37,1,732,'2026-08-18 12:25:47'),(38,1,741,'2026-08-18 12:25:47'),(39,1,742,'2026-08-18 12:25:47'),(40,1,743,'2026-08-18 12:25:47'),(41,1,744,'2026-08-18 12:25:47'),(66,1,8,'2026-08-18 22:27:42'),(67,1,745,'2026-08-18 22:27:42'),(68,1,747,'2026-08-18 22:27:42'),(69,1,746,'2026-08-18 22:27:42'),(70,1,81,'2026-08-18 22:27:42'),(71,1,82,'2026-08-18 22:27:42'),(72,1,83,'2026-08-18 22:27:42'),(73,1,811,'2026-08-18 22:27:42'),(74,1,821,'2026-08-18 22:27:42'),(75,1,831,'2026-08-18 22:27:42'),(81,1,84,'2026-08-18 22:27:42'),(82,1,841,'2026-08-18 22:27:42'),(84,1,85,'2026-08-18 22:27:42'),(85,1,851,'2026-08-18 22:27:42'),(87,1,86,'2026-08-18 22:27:42'),(88,1,87,'2026-08-18 22:27:42'),(89,1,88,'2026-08-18 22:27:42'),(90,1,861,'2026-08-18 22:27:42'),(91,1,871,'2026-08-18 22:27:42'),(92,1,881,'2026-08-18 22:27:42'),(94,1,822,'2026-08-22 11:38:54'),(95,1,823,'2026-08-22 11:38:54'),(96,1,852,'2026-08-22 11:38:54'),(97,1,853,'2026-08-22 11:38:54'),(104,1,9,'2026-08-22 17:16:16'),(105,1,91,'2026-08-22 17:16:16'),(106,1,92,'2026-08-22 17:16:16'),(107,1,911,'2026-08-22 17:16:16'),(108,1,921,'2026-08-22 17:16:16'),(109,1,922,'2026-08-22 17:16:16'),(111,1,44,'2026-08-23 11:08:16');
/*!40000 ALTER TABLE `sys_role_menus` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sys_roles`
--

DROP TABLE IF EXISTS `sys_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `code` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色编码(唯一)',
  `remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 0=禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_roles_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sys_roles`
--

LOCK TABLES `sys_roles` WRITE;
/*!40000 ALTER TABLE `sys_roles` DISABLE KEYS */;
INSERT INTO `sys_roles` VALUES (1,'超级管理员','admin','拥有全部菜单权限(系统预置)',1,'2026-08-18 12:25:47','2026-08-18 12:25:47');
/*!40000 ALTER TABLE `sys_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_addresses`
--

DROP TABLE IF EXISTS `user_addresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_addresses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `province` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `district` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `lat` decimal(10,6) DEFAULT NULL,
  `lng` decimal(10,6) DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_addresses_user_id` (`user_id`),
  CONSTRAINT `fk_user_addresses_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=111 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户收货地址表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_addresses`
--

LOCK TABLES `user_addresses` WRITE;
/*!40000 ALTER TABLE `user_addresses` DISABLE KEYS */;
INSERT INTO `user_addresses` VALUES (1,1,'收货人1','13500000001','北京市','北京市','朝阳区','模拟地址1号',39.860000,116.310000,1,'2026-05-07 12:00:00','2026-05-11 09:00:00'),(2,2,'收货人2','13500000002','北京市','北京市','朝阳区','模拟地址2号',39.870000,116.320000,1,'2026-05-06 12:00:00','2026-05-11 09:00:00'),(3,3,'收货人3','13500000003','北京市','北京市','朝阳区','模拟地址3号',39.880000,116.330000,1,'2026-05-05 12:00:00','2026-05-11 09:00:00'),(4,4,'收货人4','13500000004','北京市','北京市','朝阳区','模拟地址4号',39.890000,116.340000,1,'2026-05-04 12:00:00','2026-05-11 09:00:00'),(5,5,'收货人5','13500000005','北京市','北京市','朝阳区','模拟地址5号',39.900000,116.350000,1,'2026-05-03 12:00:00','2026-05-11 09:00:00'),(6,6,'收货人6','13500000006','北京市','北京市','朝阳区','模拟地址6号',39.910000,116.360000,1,'2026-05-02 12:00:00','2026-05-11 09:00:00'),(7,7,'收货人7','13500000007','北京市','北京市','朝阳区','模拟地址7号',39.920000,116.370000,1,'2026-05-01 12:00:00','2026-05-11 09:00:00'),(8,8,'收货人8','13500000008','北京市','北京市','朝阳区','模拟地址8号',39.930000,116.380000,1,'2026-04-30 12:00:00','2026-05-11 09:00:00'),(9,9,'收货人9','13500000009','北京市','北京市','朝阳区','模拟地址9号',39.940000,116.390000,1,'2026-04-29 12:00:00','2026-05-11 09:00:00'),(10,10,'收货人10','13500000010','北京市','北京市','朝阳区','模拟地址10号',39.950000,116.400000,1,'2026-04-28 12:00:00','2026-05-11 09:00:00'),(11,11,'收货人11','13500000011','北京市','北京市','朝阳区','模拟地址11号',39.960000,116.410000,1,'2026-04-27 12:00:00','2026-05-11 09:00:00'),(12,12,'收货人12','13500000012','北京市','北京市','朝阳区','模拟地址12号',39.970000,116.420000,1,'2026-04-26 12:00:00','2026-05-11 09:00:00'),(13,13,'收货人13','13500000013','北京市','北京市','朝阳区','模拟地址13号',39.980000,116.430000,1,'2026-04-25 12:00:00','2026-05-11 09:00:00'),(14,14,'收货人14','13500000014','北京市','北京市','朝阳区','模拟地址14号',39.990000,116.440000,1,'2026-04-24 12:00:00','2026-05-11 09:00:00'),(15,15,'收货人15','13500000015','北京市','北京市','朝阳区','模拟地址15号',40.000000,116.450000,1,'2026-04-23 12:00:00','2026-05-11 09:00:00'),(16,16,'收货人16','13500000016','北京市','北京市','朝阳区','模拟地址16号',40.010000,116.460000,1,'2026-04-22 12:00:00','2026-05-11 09:00:00'),(17,17,'收货人17','13500000017','北京市','北京市','朝阳区','模拟地址17号',40.020000,116.470000,1,'2026-04-21 12:00:00','2026-05-11 09:00:00'),(18,18,'收货人18','13500000018','北京市','北京市','朝阳区','模拟地址18号',40.030000,116.480000,1,'2026-05-08 12:00:00','2026-05-11 09:00:00'),(19,19,'收货人19','13500000019','北京市','北京市','朝阳区','模拟地址19号',40.040000,116.490000,1,'2026-05-07 12:00:00','2026-05-11 09:00:00'),(20,20,'收货人20','13500000020','北京市','北京市','朝阳区','模拟地址20号',40.050000,116.500000,1,'2026-05-06 12:00:00','2026-05-11 09:00:00'),(21,21,'收货人21','13500000021','北京市','北京市','朝阳区','模拟地址21号',40.060000,116.510000,1,'2026-05-05 12:00:00','2026-05-11 09:00:00'),(22,22,'收货人22','13500000022','北京市','北京市','朝阳区','模拟地址22号',40.070000,116.520000,1,'2026-05-04 12:00:00','2026-05-11 09:00:00'),(23,23,'收货人23','13500000023','北京市','北京市','朝阳区','模拟地址23号',40.080000,116.530000,1,'2026-05-03 12:00:00','2026-05-11 09:00:00'),(24,24,'收货人24','13500000024','北京市','北京市','朝阳区','模拟地址24号',40.090000,116.540000,1,'2026-05-02 12:00:00','2026-05-11 09:00:00'),(25,25,'收货人25','13500000025','北京市','北京市','朝阳区','模拟地址25号',40.100000,116.550000,1,'2026-05-01 12:00:00','2026-05-11 09:00:00'),(26,26,'收货人26','13500000026','北京市','北京市','朝阳区','模拟地址26号',40.110000,116.560000,1,'2026-04-30 12:00:00','2026-05-11 09:00:00'),(27,27,'收货人27','13500000027','北京市','北京市','朝阳区','模拟地址27号',40.120000,116.570000,1,'2026-04-29 12:00:00','2026-05-11 09:00:00'),(28,28,'收货人28','13500000028','北京市','北京市','朝阳区','模拟地址28号',40.130000,116.580000,1,'2026-04-28 12:00:00','2026-05-11 09:00:00'),(29,29,'收货人29','13500000029','北京市','北京市','朝阳区','模拟地址29号',40.140000,116.590000,1,'2026-04-27 12:00:00','2026-05-11 09:00:00'),(30,30,'收货人30','13500000030','北京市','北京市','朝阳区','模拟地址30号',40.150000,116.600000,1,'2026-04-26 12:00:00','2026-05-11 09:00:00'),(101,1,'备用联系人1','13599990001','北京市','北京市','海淀区','备用地址1号',39.905000,116.355000,0,'2026-05-07 12:00:00','2026-05-11 09:00:00'),(102,2,'备用联系人2','13599990002','北京市','北京市','海淀区','备用地址2号',39.910000,116.360000,0,'2026-05-06 12:00:00','2026-05-11 09:00:00'),(103,3,'备用联系人3','13599990003','北京市','北京市','海淀区','备用地址3号',39.915000,116.365000,0,'2026-05-05 12:00:00','2026-05-11 09:00:00'),(104,4,'备用联系人4','13599990004','北京市','北京市','海淀区','备用地址4号',39.920000,116.370000,0,'2026-05-04 12:00:00','2026-05-11 09:00:00'),(105,5,'备用联系人5','13599990005','北京市','北京市','海淀区','备用地址5号',39.925000,116.375000,0,'2026-05-03 12:00:00','2026-05-11 09:00:00'),(106,6,'备用联系人6','13599990006','北京市','北京市','海淀区','备用地址6号',39.930000,116.380000,0,'2026-05-02 12:00:00','2026-05-11 09:00:00'),(107,7,'备用联系人7','13599990007','北京市','北京市','海淀区','备用地址7号',39.935000,116.385000,0,'2026-05-01 12:00:00','2026-05-11 09:00:00'),(108,8,'备用联系人8','13599990008','北京市','北京市','海淀区','备用地址8号',39.940000,116.390000,0,'2026-04-30 12:00:00','2026-05-11 09:00:00'),(109,9,'备用联系人9','13599990009','北京市','北京市','海淀区','备用地址9号',39.945000,116.395000,0,'2026-04-29 12:00:00','2026-05-11 09:00:00'),(110,10,'备用联系人10','13599990010','北京市','北京市','海淀区','备用地址10号',39.950000,116.400000,0,'2026-04-28 12:00:00','2026-05-11 09:00:00');
/*!40000 ALTER TABLE `user_addresses` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_behavior_events`
--

DROP TABLE IF EXISTS `user_behavior_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_behavior_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL COMMENT '商家ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID',
  `event_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型: page_view=页面浏览 product_view=商品查看 submit_order=提交订单 pay_success=支付成功',
  `page` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '页面标识(如store_home/store_product)',
  `product_id` bigint unsigned DEFAULT NULL COMMENT '关联商品ID(商品查看事件)',
  `order_id` bigint unsigned DEFAULT NULL COMMENT '关联订单ID(下单/支付事件)',
  `source` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '事件来源: scan=扫码 direct=直接进入',
  `payload` json DEFAULT NULL COMMENT '事件附加数据JSON',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `open_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_behavior_events_merchant_id` (`merchant_id`),
  KEY `idx_user_behavior_events_user_id` (`user_id`),
  KEY `idx_user_behavior_events_openid` (`openid`),
  KEY `idx_user_behavior_events_event_type` (`event_type`),
  KEY `idx_user_behavior_events_product_id` (`product_id`),
  KEY `idx_user_behavior_events_order_id` (`order_id`),
  KEY `idx_user_behavior_events_created_at` (`created_at`),
  KEY `idx_user_behavior_events_open_id` (`open_id`)
) ENGINE=InnoDB AUTO_INCREMENT=671 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户行为事件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_behavior_events`
--

LOCK TABLES `user_behavior_events` WRITE;
/*!40000 ALTER TABLE `user_behavior_events` DISABLE KEYS */;
INSERT INTO `user_behavior_events` VALUES (1,1,1,'mock_user_openid_1','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-11 11:00:00.000',NULL),(2,1,2,'mock_user_openid_2','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-10 09:57:00.000',NULL),(3,1,2,'mock_user_openid_2','submit_order','pages/order/create',NULL,2,'miniapp','{\"merchant_id\": 1}','2026-05-10 10:02:00.000',NULL),(4,1,2,'mock_user_openid_2','pay_success','pages/order/pay',NULL,2,'miniapp','{\"merchant_id\": 1}','2026-05-10 10:07:00.000',NULL),(5,1,3,'mock_user_openid_3','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-09 08:54:00.000',NULL),(6,1,3,'mock_user_openid_3','submit_order','pages/order/create',NULL,3,'miniapp','{\"merchant_id\": 1}','2026-05-09 08:59:00.000',NULL),(7,1,3,'mock_user_openid_3','pay_success','pages/order/pay',NULL,3,'miniapp','{\"merchant_id\": 1}','2026-05-09 09:04:00.000',NULL),(8,1,4,'mock_user_openid_4','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-08 07:51:00.000',NULL),(9,1,4,'mock_user_openid_4','submit_order','pages/order/create',NULL,4,'miniapp','{\"merchant_id\": 1}','2026-05-08 07:56:00.000',NULL),(10,1,4,'mock_user_openid_4','pay_success','pages/order/pay',NULL,4,'miniapp','{\"merchant_id\": 1}','2026-05-08 08:01:00.000',NULL),(11,1,5,'mock_user_openid_5','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-07 06:48:00.000',NULL),(12,1,5,'mock_user_openid_5','submit_order','pages/order/create',NULL,5,'miniapp','{\"merchant_id\": 1}','2026-05-07 06:53:00.000',NULL),(13,1,5,'mock_user_openid_5','pay_success','pages/order/pay',NULL,5,'miniapp','{\"merchant_id\": 1}','2026-05-07 06:58:00.000',NULL),(14,1,6,'mock_user_openid_6','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-06 05:45:00.000',NULL),(15,1,6,'mock_user_openid_6','submit_order','pages/order/create',NULL,6,'miniapp','{\"merchant_id\": 1}','2026-05-06 05:50:00.000',NULL),(16,1,6,'mock_user_openid_6','pay_success','pages/order/pay',NULL,6,'miniapp','{\"merchant_id\": 1}','2026-05-06 05:55:00.000',NULL),(17,1,7,'mock_user_openid_7','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-05 04:42:00.000',NULL),(18,1,8,'mock_user_openid_8','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-04 11:39:00.000',NULL),(19,1,8,'mock_user_openid_8','submit_order','pages/order/create',NULL,8,'miniapp','{\"merchant_id\": 1}','2026-05-04 11:44:00.000',NULL),(20,1,8,'mock_user_openid_8','pay_success','pages/order/pay',NULL,8,'miniapp','{\"merchant_id\": 1}','2026-05-04 11:49:00.000',NULL),(21,1,9,'mock_user_openid_9','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-03 10:36:00.000',NULL),(22,1,9,'mock_user_openid_9','submit_order','pages/order/create',NULL,9,'miniapp','{\"merchant_id\": 1}','2026-05-03 10:41:00.000',NULL),(23,1,9,'mock_user_openid_9','pay_success','pages/order/pay',NULL,9,'miniapp','{\"merchant_id\": 1}','2026-05-03 10:46:00.000',NULL),(24,1,10,'mock_user_openid_10','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-11 09:33:00.000',NULL),(25,1,10,'mock_user_openid_10','submit_order','pages/order/create',NULL,10,'miniapp','{\"merchant_id\": 1}','2026-05-11 09:38:00.000',NULL),(26,1,10,'mock_user_openid_10','pay_success','pages/order/pay',NULL,10,'miniapp','{\"merchant_id\": 1}','2026-05-11 09:43:00.000',NULL),(27,1,11,'mock_user_openid_11','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-10 08:30:00.000',NULL),(28,1,11,'mock_user_openid_11','submit_order','pages/order/create',NULL,11,'miniapp','{\"merchant_id\": 1}','2026-05-10 08:35:00.000',NULL),(29,1,11,'mock_user_openid_11','pay_success','pages/order/pay',NULL,11,'miniapp','{\"merchant_id\": 1}','2026-05-10 08:40:00.000',NULL),(30,1,12,'mock_user_openid_12','store_visit','pages/store/home',NULL,NULL,'miniapp','{\"merchant_id\": 1}','2026-05-09 07:27:00.000',NULL),(31,1,12,'mock_user_openid_12','submit_order','pages/order/create',NULL,12,'miniapp','{\"merchant_id\": 1}','2026-05-09 07:32:00.000',NULL),(32,1,12,'mock_user_openid_12','pay_success','pages/order/pay',NULL,12,'miniapp','{\"merchant_id\": 1}','2026-05-09 07:37:00.000',NULL),(321,1,1,'','submit_order','store_confirm',NULL,123,'store','{\"pay_amount\": 26.8, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-15 02:27:30.948',NULL),(322,1,1,'','submit_order','store_confirm',NULL,124,'store','{\"pay_amount\": 26.8, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-15 02:29:12.398',NULL),(323,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 08:19:28.716',NULL),(324,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 08:19:28.740',NULL),(325,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-15 08:19:34.467',NULL),(326,1,62,'','submit_order','store_confirm',NULL,125,'store','{\"pay_amount\": 31.9, \"delivery_type\": 2, \"delivery_distance\": 0}','2026-05-15 08:23:03.966',NULL),(327,1,62,'','submit_order','store_confirm',NULL,126,'store','{\"pay_amount\": 31.9, \"delivery_type\": 2, \"delivery_distance\": 0}','2026-05-15 08:24:00.090',NULL),(328,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 14:51:36.355',NULL),(329,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 14:51:36.387',NULL),(330,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-15 14:51:46.264',NULL),(331,1,62,'','submit_order','store_confirm',NULL,127,'store','{\"pay_amount\": 0.01, \"delivery_type\": 2, \"delivery_distance\": 0}','2026-05-15 14:54:22.970',NULL),(332,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','pay_success','store_payment_result',NULL,127,'store','{\"amount\": 0.01, \"order_id\": 127}','2026-05-15 14:54:47.491',NULL),(333,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 20:40:41.173',NULL),(334,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 20:40:41.198',NULL),(335,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 21:58:21.008',NULL),(336,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 21:58:21.444',NULL),(337,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 22:08:45.529',NULL),(338,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 22:08:45.552',NULL),(339,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 22:09:00.253',NULL),(340,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 22:09:00.292',NULL),(341,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 22:12:21.705',NULL),(342,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 22:12:21.726',NULL),(343,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 22:12:43.130',NULL),(344,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 22:12:43.178',NULL),(345,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 22:41:38.308',NULL),(346,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 22:41:38.327',NULL),(347,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 23:17:42.434',NULL),(348,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 23:17:48.421',NULL),(349,1,62,'','submit_order','store_confirm',NULL,128,'store','{\"pay_amount\": 0.02, \"delivery_type\": 2, \"delivery_distance\": 0}','2026-05-15 23:18:00.255',NULL),(350,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 23:25:33.616',NULL),(351,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 23:25:41.003',NULL),(352,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-15 23:27:10.612',NULL),(353,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-15 23:27:10.635',NULL),(354,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:52:08.474',NULL),(355,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:52:08.522',NULL),(356,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:53:38.162',NULL),(357,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:53:38.185',NULL),(358,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:54:35.277',NULL),(359,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:54:35.416',NULL),(360,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:56:43.026',NULL),(361,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:56:43.046',NULL),(362,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2008,NULL,'store','{\"product_id\": 2008}','2026-05-16 00:56:53.798',NULL),(363,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-16 00:56:57.089',NULL),(364,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2008,NULL,'store','{\"product_id\": 2008}','2026-05-16 00:57:06.966',NULL),(365,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:57:08.477',NULL),(366,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:57:08.505',NULL),(367,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:57:13.513',NULL),(368,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:57:13.565',NULL),(369,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:57:33.151',NULL),(370,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:57:33.170',NULL),(371,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 00:57:40.991',NULL),(372,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 00:57:41.025',NULL),(373,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 01:16:17.236',NULL),(374,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 01:16:17.304',NULL),(375,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 01:22:44.248',NULL),(376,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 01:22:44.272',NULL),(377,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:13:48.563',NULL),(378,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:13:48.638',NULL),(379,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:39:30.516',NULL),(380,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:39:30.560',NULL),(381,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:39:51.678',NULL),(382,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:39:51.703',NULL),(383,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:49:22.345',NULL),(384,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:49:22.390',NULL),(385,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:51:23.425',NULL),(386,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:51:23.489',NULL),(387,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:51:49.451',NULL),(388,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:51:49.466',NULL),(389,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:58:57.977',NULL),(390,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:58:57.991',NULL),(391,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 22:59:01.230',NULL),(392,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 22:59:02.960',NULL),(393,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 22:59:03.095',NULL),(394,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:01:27.239',NULL),(395,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:01:27.259',NULL),(396,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:01:33.557',NULL),(397,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:01:33.568',NULL),(398,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:00.622',NULL),(399,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:00.638',NULL),(400,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 23:11:03.363',NULL),(401,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:04.936',NULL),(402,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:04.945',NULL),(403,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 23:11:13.467',NULL),(404,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:15.315',NULL),(405,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:15.326',NULL),(406,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 23:11:17.393',NULL),(407,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:18.687',NULL),(408,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:18.705',NULL),(409,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2008,NULL,'store','{\"product_id\": 2008}','2026-05-16 23:11:20.292',NULL),(410,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:21.792',NULL),(411,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:21.805',NULL),(412,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 23:11:22.787',NULL),(413,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:24.483',NULL),(414,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:24.493',NULL),(415,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2004,NULL,'store','{\"product_id\": 2004}','2026-05-16 23:11:26.108',NULL),(416,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:27.565',NULL),(417,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:27.574',NULL),(418,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2008,NULL,'store','{\"product_id\": 2008}','2026-05-16 23:11:34.974',NULL),(419,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:39.810',NULL),(420,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:39.824',NULL),(421,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-16 23:11:42.163',NULL),(422,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:11:43.284',NULL),(423,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:11:43.295',NULL),(424,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-16 23:13:12.821',NULL),(425,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-16 23:13:22.783',NULL),(426,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-16 23:13:22.795',NULL),(427,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:43:10.924',NULL),(428,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:43:10.946',NULL),(429,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:46:47.836',NULL),(430,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:46:47.847',NULL),(431,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 00:46:57.703',NULL),(432,1,62,'','submit_order','store_confirm',NULL,129,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-17 00:47:15.509',NULL),(433,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','pay_success','store_payment_result',NULL,129,'store','{\"amount\": 0.01, \"order_id\": 129}','2026-05-17 00:47:27.924',NULL),(434,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 00:47:28.010',NULL),(435,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:47:41.983',NULL),(436,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:47:41.991',NULL),(437,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:47:47.176',NULL),(438,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:47:48.313',NULL),(439,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:47:48.326',NULL),(440,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:47:55.134',NULL),(441,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:47:56.364',NULL),(442,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:47:56.375',NULL),(443,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:48:10.205',NULL),(444,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:48:12.948',NULL),(445,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:48:12.957',NULL),(446,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2004,NULL,'store','{\"product_id\": 2004}','2026-05-17 00:48:55.103',NULL),(447,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:48:59.861',NULL),(448,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:48:59.873',NULL),(449,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:49:03.226',NULL),(450,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:49:11.358',NULL),(451,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:49:11.367',NULL),(452,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:49:17.570',NULL),(453,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:49:19.525',NULL),(454,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:49:19.533',NULL),(455,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-17 00:50:31.408',NULL),(456,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:50:32.813',NULL),(457,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:50:32.829',NULL),(458,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:54:54.716',NULL),(459,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:54:54.731',NULL),(460,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:58:52.722',NULL),(461,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:58:52.733',NULL),(462,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:59:17.174',NULL),(463,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:59:17.187',NULL),(464,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:59:22.049',NULL),(465,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:59:22.059',NULL),(466,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 00:59:33.423',NULL),(467,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 00:59:33.433',NULL),(468,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 01:00:00.419',NULL),(469,1,62,'','submit_order','store_confirm',NULL,130,'store','{\"pay_amount\": 5.01, \"delivery_type\": 1, \"delivery_distance\": 3}','2026-05-17 01:00:14.117',NULL),(470,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','pay_success','store_payment_result',NULL,130,'store','{\"amount\": 5.01, \"order_id\": 130}','2026-05-17 01:00:26.010',NULL),(471,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 01:00:26.120',NULL),(472,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 01:00:38.875',NULL),(473,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 01:00:38.885',NULL),(474,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 01:00:42.280',NULL),(475,1,62,'','submit_order','store_confirm',NULL,131,'store','{\"pay_amount\": 3.01, \"delivery_type\": 1, \"delivery_distance\": 5}','2026-05-17 01:01:06.478',NULL),(476,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'store','{\"page\": \"store_confirm\"}','2026-05-17 01:01:13.781',NULL),(477,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 01:02:36.225',NULL),(478,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 01:02:36.237',NULL),(479,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 01:05:20.139',NULL),(480,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 01:05:20.154',NULL),(481,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 14:27:27.519',NULL),(482,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 14:27:27.560',NULL),(483,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:38:39.881',NULL),(484,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:38:39.913',NULL),(485,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:39:39.406',NULL),(486,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:39:39.432',NULL),(487,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:41:04.382',NULL),(488,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:41:04.398',NULL),(489,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-17 17:41:08.188',NULL),(490,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:41:10.455',NULL),(491,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:41:10.454',NULL),(492,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:44:55.940',NULL),(493,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:44:55.962',NULL),(494,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:45:06.857',NULL),(495,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:45:06.863',NULL),(496,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:46:12.884',NULL),(497,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:46:12.905',NULL),(498,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 17:55:20.499',NULL),(499,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 17:55:20.517',NULL),(500,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-17 18:21:56.777',NULL),(501,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-17 18:21:56.791',NULL),(502,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:36:51.925',NULL),(503,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:36:51.986',NULL),(504,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:36:58.997',NULL),(505,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:37:02.524',NULL),(506,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:37:02.537',NULL),(507,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:39:41.626',NULL),(508,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:39:41.643',NULL),(509,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:39:43.575',NULL),(510,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:39:51.413',NULL),(511,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:39:51.438',NULL),(512,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:40:06.614',NULL),(513,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:40:06.624',NULL),(514,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:40:41.001',NULL),(515,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:40:45.472',NULL),(516,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:40:45.499',NULL),(517,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:40:46.995',NULL),(518,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:47:00.728',NULL),(519,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:47:00.741',NULL),(520,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:47:02.679',NULL),(521,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:47:25.761',NULL),(522,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:47:25.780',NULL),(523,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:47:27.402',NULL),(524,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:47:44.054',NULL),(525,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:47:44.059',NULL),(526,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:47:47.519',NULL),(527,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-18 23:48:09.702',NULL),(528,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-18 23:48:09.716',NULL),(529,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-18 23:48:12.209',NULL),(530,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:27:43.724',NULL),(531,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:27:43.746',NULL),(532,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-19 00:27:46.255',NULL),(533,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:28:09.910',NULL),(534,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:28:09.921',NULL),(535,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-19 00:28:11.529',NULL),(536,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:28:43.943',NULL),(537,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:28:43.958',NULL),(538,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-19 00:28:45.683',NULL),(539,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:29:09.341',NULL),(540,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:29:09.344',NULL),(541,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-19 00:29:10.492',NULL),(542,1,62,'','submit_order','store_confirm',NULL,132,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-19 00:29:22.276',NULL),(543,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:29:26.616',NULL),(544,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:29:26.632',NULL),(545,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:32:46.971',NULL),(546,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:32:46.983',NULL),(547,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:35:11.867',NULL),(548,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:35:11.880',NULL),(549,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-19 00:35:13.890',NULL),(550,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-19 00:35:16.063',NULL),(551,1,62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-19 00:35:16.072',NULL),(552,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:23:27.981',NULL),(553,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:23:28.001',NULL),(554,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:23:40.668',NULL),(555,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:23:40.686',NULL),(556,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:23:49.715',NULL),(557,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:24:11.182',NULL),(558,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:24:11.196',NULL),(559,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:24:55.071',NULL),(560,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:24:55.101',NULL),(561,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:25:06.489',NULL),(562,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:25:19.956',NULL),(563,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:25:19.968',NULL),(564,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:25:25.497',NULL),(565,1,63,'','submit_order','store_confirm',NULL,133,'store','{\"pay_amount\": 92.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-28 01:25:37.619',NULL),(566,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:25:57.600',NULL),(567,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:25:57.620',NULL),(568,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:26:06.680',NULL),(569,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:26:06.698',NULL),(570,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:26:16.771',NULL),(571,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:26:16.800',NULL),(572,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:26:31.851',NULL),(573,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:26:31.869',NULL),(574,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:26:52.270',NULL),(575,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:26:52.286',NULL),(576,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:30:34.367',NULL),(577,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:30:34.382',NULL),(578,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:33:08.093',NULL),(579,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:33:08.103',NULL),(580,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:33:10.857',NULL),(581,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:33:29.987',NULL),(582,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:33:30.003',NULL),(583,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:33:34.039',NULL),(584,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:33:35.749',NULL),(585,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:33:35.760',NULL),(586,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:33:37.920',NULL),(587,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:33:37.929',NULL),(588,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:33:51.173',NULL),(589,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:33:51.187',NULL),(590,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','product_view','store_product',2002,NULL,'store','{\"product_id\": 2002}','2026-05-28 01:33:58.740',NULL),(591,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:34:00.575',NULL),(592,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:34:00.587',NULL),(593,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:34:14.828',NULL),(594,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:34:29.259',NULL),(595,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:34:29.274',NULL),(596,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:34:44.572',NULL),(597,1,63,'','submit_order','store_confirm',NULL,134,'store','{\"pay_amount\": 49.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-28 01:34:57.582',NULL),(598,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:47:49.029',NULL),(599,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:47:49.050',NULL),(600,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:47:54.457',NULL),(601,1,63,'','submit_order','store_confirm',NULL,135,'store','{\"pay_amount\": 49.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-28 01:47:58.534',NULL),(602,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:48:25.554',NULL),(603,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:48:25.564',NULL),(604,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:48:36.236',NULL),(605,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:48:36.246',NULL),(606,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:48:50.834',NULL),(607,1,63,'','submit_order','store_confirm',NULL,136,'store','{\"pay_amount\": 48.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-28 01:48:58.116',NULL),(608,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-28 01:52:16.578',NULL),(609,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-28 01:52:16.594',NULL),(610,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-28 01:52:30.848',NULL),(611,1,63,'','submit_order','store_confirm',NULL,137,'store','{\"pay_amount\": 46.02, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-28 01:52:35.697',NULL),(612,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:08:06.501',NULL),(613,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:08:06.511',NULL),(614,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-31 23:08:14.128',NULL),(615,1,63,'','submit_order','store_confirm',NULL,138,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:08:19.395',NULL),(616,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:09:27.583',NULL),(617,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:09:27.594',NULL),(618,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-31 23:10:19.856',NULL),(619,1,63,'','submit_order','store_confirm',NULL,139,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:10:22.367',NULL),(620,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','pay_success','store_payment_result',NULL,139,'store','{\"amount\": 0.01, \"order_id\": 139}','2026-05-31 23:10:39.585',NULL),(621,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:10:46.926',NULL),(622,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:10:46.934',NULL),(623,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:21:46.818',NULL),(624,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:21:46.828',NULL),(625,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-31 23:21:59.618',NULL),(626,1,64,'','submit_order','store_confirm',NULL,140,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:22:02.084',NULL),(627,1,64,'','submit_order','store_confirm',NULL,141,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:22:25.596',NULL),(628,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:26:19.469',NULL),(629,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:26:19.482',NULL),(630,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:26:26.620',NULL),(631,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:26:26.635',NULL),(632,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-31 23:26:31.482',NULL),(633,1,64,'','submit_order','store_confirm',NULL,143,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:26:33.841',NULL),(634,1,64,'','submit_order','store_confirm',NULL,144,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:27:35.440',NULL),(635,1,64,'','submit_order','store_confirm',NULL,145,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:28:01.468',NULL),(636,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:30:39.461',NULL),(637,1,64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:30:39.477',NULL),(649,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:34:56.488',NULL),(650,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:34:56.499',NULL),(651,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_confirm',NULL,NULL,'scan','{\"page\": \"store_confirm\"}','2026-05-31 23:34:59.408',NULL),(652,1,63,'','submit_order','store_confirm',NULL,147,'store','{\"pay_amount\": 0.01, \"delivery_type\": 3, \"delivery_distance\": 0}','2026-05-31 23:35:01.686',NULL),(653,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:35:20.948',NULL),(654,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:35:20.959',NULL),(655,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-05-31 23:37:43.186',NULL),(656,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-05-31 23:37:43.200',NULL),(657,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-14 14:14:32.627',NULL),(658,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-14 14:14:32.650',NULL),(659,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-14 23:15:42.733',NULL),(660,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-14 23:15:42.757',NULL),(661,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-14 23:17:27.655',NULL),(662,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-14 23:17:27.672',NULL),(663,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-15 22:46:51.606',NULL),(664,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-15 22:46:51.639',NULL),(665,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-18 09:31:40.030',NULL),(666,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-18 09:31:40.051',NULL),(667,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-18 11:06:20.154',NULL),(668,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-18 11:06:20.183',NULL),(669,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','page_view','store_home',NULL,NULL,'scan','{\"page\": \"store_home\"}','2026-08-18 11:06:31.093',NULL),(670,1,63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','store_visit','store_home',NULL,NULL,'scan','{\"source\": \"scan\"}','2026-08-18 11:06:31.114',NULL);
/*!40000 ALTER TABLE `user_behavior_events` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_visits`
--

DROP TABLE IF EXISTS `user_visits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_visits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `merchant_id` bigint unsigned NOT NULL COMMENT '商家ID',
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID',
  `visit_time` datetime(3) DEFAULT NULL COMMENT '访问时间',
  `source` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问来源: scan=扫码 direct=直接进入',
  PRIMARY KEY (`id`),
  KEY `idx_user_visits_user_id` (`user_id`),
  KEY `idx_user_visits_merchant_id` (`merchant_id`),
  KEY `idx_user_visits_openid` (`openid`),
  KEY `idx_user_visits_open_id` (`openid`),
  CONSTRAINT `fk_user_visits_merchant` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_visits_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=252 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户访问记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_visits`
--

LOCK TABLES `user_visits` WRITE;
/*!40000 ALTER TABLE `user_visits` DISABLE KEYS */;
INSERT INTO `user_visits` VALUES (1,1,1,'mock_user_openid_1','2026-05-11 11:00:00.000','share'),(2,2,1,'mock_user_openid_2','2026-05-10 09:57:00.000','miniapp'),(3,3,1,'mock_user_openid_3','2026-05-09 08:54:00.000','miniapp'),(4,4,1,'mock_user_openid_4','2026-05-08 07:51:00.000','share'),(5,5,1,'mock_user_openid_5','2026-05-07 06:48:00.000','miniapp'),(6,6,1,'mock_user_openid_6','2026-05-06 05:45:00.000','miniapp'),(7,7,1,'mock_user_openid_7','2026-05-05 04:42:00.000','share'),(8,8,1,'mock_user_openid_8','2026-05-04 11:39:00.000','miniapp'),(9,9,1,'mock_user_openid_9','2026-05-03 10:36:00.000','miniapp'),(10,10,1,'mock_user_openid_10','2026-05-11 09:33:00.000','share'),(11,11,1,'mock_user_openid_11','2026-05-10 08:30:00.000','miniapp'),(12,12,1,'mock_user_openid_12','2026-05-09 07:27:00.000','miniapp'),(121,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 08:19:28.684','scan'),(122,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 14:51:36.314','scan'),(123,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 20:40:41.133','scan'),(124,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 21:58:20.984','scan'),(125,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 22:08:45.502','scan'),(126,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 22:09:00.231','scan'),(127,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 22:12:21.682','scan'),(128,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 22:12:43.105','scan'),(129,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 22:41:38.287','scan'),(130,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 23:17:42.419','scan'),(131,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 23:25:33.608','scan'),(132,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-15 23:27:10.590','scan'),(133,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:52:08.454','scan'),(134,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:53:38.150','scan'),(135,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:54:35.257','scan'),(136,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:56:43.009','scan'),(137,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:57:08.430','scan'),(138,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:57:13.483','scan'),(139,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:57:33.139','scan'),(140,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 00:57:40.969','scan'),(141,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 01:16:17.207','scan'),(142,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 01:22:44.235','scan'),(143,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:13:48.543','scan'),(144,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:39:30.493','scan'),(145,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:39:51.659','scan'),(146,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:49:22.315','scan'),(147,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:51:23.394','scan'),(148,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:51:49.441','scan'),(149,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:58:57.971','scan'),(150,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 22:59:02.954','scan'),(151,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:01:27.230','scan'),(152,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:01:33.552','scan'),(153,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:00.612','scan'),(154,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:04.935','scan'),(155,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:15.313','scan'),(156,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:18.685','scan'),(157,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:21.790','scan'),(158,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:24.476','scan'),(159,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:27.564','scan'),(160,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:39.803','scan'),(161,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:11:43.280','scan'),(162,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-16 23:13:22.778','scan'),(163,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:43:10.920','scan'),(164,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:46:47.828','scan'),(165,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:47:41.979','scan'),(166,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:47:48.310','scan'),(167,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:47:56.362','scan'),(168,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:48:12.944','scan'),(169,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:48:59.857','scan'),(170,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:49:11.356','scan'),(171,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:49:19.514','scan'),(172,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:50:32.811','scan'),(173,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:54:54.706','scan'),(174,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:58:52.720','scan'),(175,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:59:17.171','scan'),(176,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:59:22.044','scan'),(177,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 00:59:33.421','scan'),(178,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 01:00:38.871','scan'),(179,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 01:02:36.220','scan'),(180,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 01:05:20.135','scan'),(181,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 14:27:27.458','scan'),(182,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:38:39.868','scan'),(183,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:39:39.404','scan'),(184,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:41:04.381','scan'),(185,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:41:10.440','scan'),(186,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:44:55.938','scan'),(187,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:45:06.849','scan'),(188,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:46:12.881','scan'),(189,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 17:55:20.496','scan'),(190,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-17 18:21:56.773','scan'),(191,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:36:51.921','scan'),(192,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:37:02.523','scan'),(193,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:39:41.621','scan'),(194,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:39:51.405','scan'),(195,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:40:06.611','scan'),(196,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:40:45.468','scan'),(197,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:47:00.726','scan'),(198,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:47:25.759','scan'),(199,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:47:44.046','scan'),(200,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-18 23:48:09.695','scan'),(201,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:27:43.723','scan'),(202,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:28:09.904','scan'),(203,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:28:43.937','scan'),(204,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:29:09.322','scan'),(205,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:29:26.614','scan'),(206,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:32:46.969','scan'),(207,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:35:11.864','scan'),(208,62,1,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','2026-05-19 00:35:16.059','scan'),(209,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:23:27.980','scan'),(210,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:23:40.666','scan'),(211,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:24:11.177','scan'),(212,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:24:55.068','scan'),(213,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:25:19.946','scan'),(214,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:25:57.594','scan'),(215,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:26:06.673','scan'),(216,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:26:16.757','scan'),(217,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:26:31.818','scan'),(218,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:26:52.267','scan'),(219,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:30:34.366','scan'),(220,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:33:08.092','scan'),(221,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:33:29.984','scan'),(222,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:33:35.746','scan'),(223,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:33:37.917','scan'),(224,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:33:51.167','scan'),(225,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:34:00.573','scan'),(226,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:34:29.255','scan'),(227,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:47:49.026','scan'),(228,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:48:25.550','scan'),(229,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:48:36.231','scan'),(230,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-28 01:52:16.575','scan'),(231,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:08:06.487','scan'),(232,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:09:27.581','scan'),(233,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:10:46.923','scan'),(234,64,1,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','2026-05-31 23:21:46.816','scan'),(235,64,1,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','2026-05-31 23:26:19.466','scan'),(236,64,1,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','2026-05-31 23:26:26.617','scan'),(237,64,1,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','2026-05-31 23:30:39.443','scan'),(242,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:34:56.486','scan'),(243,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:35:20.945','scan'),(244,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-05-31 23:37:43.182','scan'),(245,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-14 14:14:32.619','scan'),(246,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-14 23:15:42.727','scan'),(247,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-14 23:17:27.651','scan'),(248,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-15 22:46:51.603','scan'),(249,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-18 09:31:40.002','scan'),(250,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-18 11:06:20.145','scan'),(251,63,1,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','2026-08-18 11:06:31.085','scan');
/*!40000 ALTER TABLE `user_visits` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `openid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信OpenID(用户唯一标识)',
  `union_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信UnionID(跨小程序唯一)',
  `nickname` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户昵称(默认微信用户)',
  `avatar` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户头像URL',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户手机号',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态: 1=正常 0=禁用',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `first_visit_at` datetime(3) DEFAULT NULL COMMENT '首次访问时间',
  `last_visit_at` datetime(3) DEFAULT NULL COMMENT '最后访问时间',
  `visit_count` bigint unsigned NOT NULL DEFAULT '1' COMMENT '累计访问次数',
  `has_ordered` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否下过单: true=是 false=否',
  `total_orders` bigint unsigned NOT NULL DEFAULT '0' COMMENT '累计订单数',
  `total_spent` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '累计消费金额(元)',
  `has_paid` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否完成过支付: true=是 false=否',
  `first_paid_at` datetime(3) DEFAULT NULL COMMENT '首次支付时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_openid` (`openid`),
  UNIQUE KEY `idx_users_open_id` (`openid`),
  UNIQUE KEY `openid` (`openid`),
  KEY `idx_users_union_id` (`union_id`),
  KEY `idx_users_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='C端用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'mock_user_openid_1','mock_union_1','测试用户1','https://example.com/avatar/1.png','13500000001',1,'2026-05-07 12:00:00.000','2026-05-15 02:29:12.400',NULL,NULL,1,1,2,53.60,1,'2026-05-15 02:27:30.952'),(2,'mock_user_openid_2','mock_union_2','测试用户2','https://example.com/avatar/2.png','13500000002',1,'2026-05-06 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(3,'mock_user_openid_3','mock_union_3','测试用户3','https://example.com/avatar/3.png','13500000003',1,'2026-05-05 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(4,'mock_user_openid_4','mock_union_4','测试用户4','https://example.com/avatar/4.png','13500000004',1,'2026-05-04 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(5,'mock_user_openid_5','mock_union_5','测试用户5','https://example.com/avatar/5.png','13500000005',1,'2026-05-03 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(6,'mock_user_openid_6','mock_union_6','测试用户6','https://example.com/avatar/6.png','13500000006',1,'2026-05-02 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(7,'mock_user_openid_7','mock_union_7','测试用户7','https://example.com/avatar/7.png','13500000007',1,'2026-05-01 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(8,'mock_user_openid_8','mock_union_8','测试用户8','https://example.com/avatar/8.png','13500000008',1,'2026-04-30 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(9,'mock_user_openid_9','mock_union_9','测试用户9','https://example.com/avatar/9.png','13500000009',1,'2026-04-29 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(10,'mock_user_openid_10','mock_union_10','测试用户10','https://example.com/avatar/10.png','13500000010',1,'2026-04-28 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(11,'mock_user_openid_11','mock_union_11','测试用户11','https://example.com/avatar/11.png','13500000011',1,'2026-04-27 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(12,'mock_user_openid_12','mock_union_12','测试用户12','https://example.com/avatar/12.png','13500000012',1,'2026-04-26 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(13,'mock_user_openid_13','mock_union_13','测试用户13','https://example.com/avatar/13.png','13500000013',1,'2026-04-25 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(14,'mock_user_openid_14','mock_union_14','测试用户14','https://example.com/avatar/14.png','13500000014',1,'2026-04-24 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(15,'mock_user_openid_15','mock_union_15','测试用户15','https://example.com/avatar/15.png','13500000015',1,'2026-04-23 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(16,'mock_user_openid_16','mock_union_16','测试用户16','https://example.com/avatar/16.png','13500000016',1,'2026-04-22 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(17,'mock_user_openid_17','mock_union_17','测试用户17','https://example.com/avatar/17.png','13500000017',1,'2026-04-21 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(18,'mock_user_openid_18','mock_union_18','测试用户18','https://example.com/avatar/18.png','13500000018',1,'2026-05-08 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(19,'mock_user_openid_19','mock_union_19','测试用户19','https://example.com/avatar/19.png','13500000019',1,'2026-05-07 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(20,'mock_user_openid_20','mock_union_20','测试用户20','https://example.com/avatar/20.png','13500000020',1,'2026-05-06 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(21,'mock_user_openid_21','mock_union_21','测试用户21','https://example.com/avatar/21.png','13500000021',1,'2026-05-05 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(22,'mock_user_openid_22','mock_union_22','测试用户22','https://example.com/avatar/22.png','13500000022',1,'2026-05-04 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(23,'mock_user_openid_23','mock_union_23','测试用户23','https://example.com/avatar/23.png','13500000023',1,'2026-05-03 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(24,'mock_user_openid_24','mock_union_24','测试用户24','https://example.com/avatar/24.png','13500000024',1,'2026-05-02 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(25,'mock_user_openid_25','mock_union_25','测试用户25','https://example.com/avatar/25.png','13500000025',1,'2026-05-01 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(26,'mock_user_openid_26','mock_union_26','测试用户26','https://example.com/avatar/26.png','13500000026',1,'2026-04-30 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(27,'mock_user_openid_27','mock_union_27','测试用户27','https://example.com/avatar/27.png','13500000027',1,'2026-04-29 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(28,'mock_user_openid_28','mock_union_28','测试用户28','https://example.com/avatar/28.png','13500000028',1,'2026-04-28 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(29,'mock_user_openid_29','mock_union_29','测试用户29','https://example.com/avatar/29.png','13500000029',1,'2026-04-27 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(30,'mock_user_openid_30','mock_union_30','测试用户30','https://example.com/avatar/30.png','13500000030',1,'2026-04-26 12:00:00.000','2026-05-11 09:00:00.000',NULL,NULL,1,0,0,0.00,0,NULL),(31,'wx_dev_login_fix2','','微信用户','','',1,'2026-05-12 12:30:13.767','2026-05-12 12:30:13.767',NULL,NULL,1,0,0,0.00,0,NULL),(32,'wx_0b1gJ7Ga1MIHGL032MFa1Sv3Fo4gJ7GI','','微信用户','','',1,'2026-05-12 12:32:38.211','2026-05-12 12:32:38.211',NULL,NULL,1,0,0,0.00,0,NULL),(33,'wx_0e1gGQFa1QRwHL040RHa10bIIq1gGQFR','','微信用户','','',1,'2026-05-12 12:32:38.562','2026-05-12 12:32:38.562',NULL,NULL,1,0,0,0.00,0,NULL),(34,'wx_0f1lMoGa1NLYGL0lleJa19bUpr0lMoG1','','微信用户','','',1,'2026-05-12 12:32:38.881','2026-05-12 12:32:38.881',NULL,NULL,1,0,0,0.00,0,NULL),(35,'wx_0e1kYoGa17DYGL09OoFa16uuqS2kYoGp','','微信用户','','',1,'2026-05-12 12:35:49.245','2026-05-12 12:35:49.245',NULL,NULL,1,0,0,0.00,0,NULL),(36,'wx_0f1M7Tkl2JzbHh4tllnl2i1iJA3M7TkL','','微信用户','','',1,'2026-05-12 12:35:49.605','2026-05-12 12:35:49.605',NULL,NULL,1,0,0,0.00,0,NULL),(37,'wx_0b1S3c200o7onW1UN0000euFB81S3c2j','','微信用户','','',1,'2026-05-12 12:38:58.510','2026-05-12 12:38:58.510',NULL,NULL,1,0,0,0.00,0,NULL),(38,'wx_0f1YSA100K5PnW14VD1000SHD52YSA1o','','微信用户','','',1,'2026-05-12 21:07:43.944','2026-05-12 21:07:43.944',NULL,NULL,1,0,0,0.00,0,NULL),(39,'wx_0c1Th7ll2FulHh4VGGml2blcnk2Th7lQ','','微信用户','','',1,'2026-05-12 21:07:44.331','2026-05-12 21:07:44.331',NULL,NULL,1,0,0,0.00,0,NULL),(40,'wx_0a1phQkl2PyCHh4cFKnl2rqKDB0phQk5','','微信用户','','',1,'2026-05-12 21:08:24.679','2026-05-12 21:08:24.679',NULL,NULL,1,0,0,0.00,0,NULL),(41,'wx_0c1Xnoll2Vu4Hh4mJ0ml2rFviM1Xnols','','微信用户','','',1,'2026-05-12 21:08:32.431','2026-05-12 21:08:32.431',NULL,NULL,1,0,0,0.00,0,NULL),(42,'wx_0f1zeUGa1zQvIL0EqJHa11PdHe2zeUGy','','微信用户','','',1,'2026-05-12 21:08:32.787','2026-05-12 21:08:32.787',NULL,NULL,1,0,0,0.00,0,NULL),(43,'wx_0f1bBuml2QGaIh4GDsnl2Tb0vb0bBumr','','微信用户','','',1,'2026-05-12 21:08:49.755','2026-05-12 21:08:49.755',NULL,NULL,1,0,0,0.00,0,NULL),(44,'wx_0c1c39200M2hnW1Tgo100GqVCf4c392t','','微信用户','','',1,'2026-05-12 21:08:50.116','2026-05-12 21:08:50.116',NULL,NULL,1,0,0,0.00,0,NULL),(45,'wx_0a1ckQkl2JmCHh43U2ll2E2II91ckQk-','','微信用户','','',1,'2026-05-12 21:09:09.008','2026-05-12 21:09:09.008',NULL,NULL,1,0,0,0.00,0,NULL),(46,'wx_0d1Xlqll2522Hh4yVknl2HzPqy3Xlql1','','微信用户','','',1,'2026-05-12 21:40:48.675','2026-05-12 21:40:48.675',NULL,NULL,1,0,0,0.00,0,NULL),(47,'wx_0d1u6oGa1x5PGL08U4Ga1wjwZT3u6oGO','','微信用户','','',1,'2026-05-12 21:40:49.019','2026-05-12 21:40:49.019',NULL,NULL,1,0,0,0.00,0,NULL),(48,'wx_0f1qGEkl2OjpHh4hJGll2FXzXv0qGEk2','','微信用户','','',1,'2026-05-12 22:37:50.262','2026-05-12 22:37:50.262',NULL,NULL,1,0,0,0.00,0,NULL),(49,'wx_0e148N0w3aSg173Hbe2w37BJLk348N03','','微信用户','','',1,'2026-05-12 22:37:50.678','2026-05-12 22:37:50.678',NULL,NULL,1,0,0,0.00,0,NULL),(50,'wx_0e1ohA000ji4nW1OwU300eUtxW1ohA0K','','微信用户','','',1,'2026-05-15 01:55:13.804','2026-05-15 01:55:13.804',NULL,NULL,1,0,0,0.00,0,NULL),(51,'wx_0b1mb2000qoCnW1eTQ000iRXwb0mb20S','','微信用户','','',1,'2026-05-15 01:55:14.205','2026-05-15 01:55:14.205',NULL,NULL,1,0,0,0.00,0,NULL),(52,'wx_0f1pvlFa1TbjIL0k4yHa11f4Xb1pvlF6','','微信用户','','',1,'2026-05-15 01:55:38.520','2026-05-15 01:55:38.520',NULL,NULL,1,0,0,0.00,0,NULL),(53,'wx_0a1WNEkl2g5fIh4Gsdol2C9caV3WNEkk','','微信用户','','',1,'2026-05-15 01:55:38.860','2026-05-15 01:55:38.860',NULL,NULL,1,0,0,0.00,0,NULL),(54,'wx_0f1hSEkl2iafIh4ORGnl29a7tP3hSEke','','微信用户','','',1,'2026-05-15 01:56:47.772','2026-05-15 01:56:47.772',NULL,NULL,1,0,0,0.00,0,NULL),(55,'wx_0d1mYcll2U3HHh4c2Kkl2aQWWg2mYclH','','微信用户','','',1,'2026-05-15 01:56:48.167','2026-05-15 01:56:48.167',NULL,NULL,1,0,0,0.00,0,NULL),(56,'wx_0f1Zh2000GxCnW1LMd2003Wsqc2Zh201','','微信用户','','',1,'2026-05-15 01:56:59.539','2026-05-15 01:56:59.539',NULL,NULL,1,0,0,0.00,0,NULL),(57,'wx_0c1ywGkl2bPeIh4x3qll245Dib2ywGkR','','微信用户','','',1,'2026-05-15 02:23:51.231','2026-05-15 02:23:51.231',NULL,NULL,1,0,0,0.00,0,NULL),(58,'wx_0b1DzXkl2aSvIh4n3Unl2RQfNn2DzXkt','','微信用户','','',1,'2026-05-15 02:23:51.719','2026-05-15 02:23:51.719',NULL,NULL,1,0,0,0.00,0,NULL),(59,'wx_0d1NzXkl2LLvIh4yHLll2f2gIA3NzXk2','','微信用户','','',1,'2026-05-15 02:23:54.324','2026-05-15 02:23:54.324',NULL,NULL,1,0,0,0.00,0,NULL),(60,'wx_0a1lBGkl2edfIh4BN0ll2jgbWq4lBGkr','','微信用户','','',1,'2026-05-15 02:25:07.071','2026-05-15 02:25:07.071',NULL,NULL,1,0,0,0.00,0,NULL),(61,'wx_0c1mBGkl2ldfIh4WPFml2S3CYg2mBGkz','','微信用户','','',1,'2026-05-15 02:25:07.443','2026-05-15 02:25:07.443',NULL,NULL,1,0,0,0.00,0,NULL),(62,'oXPw33VXT3-fzY6PH_R-QYvqAj8o','oRIl55tdTYZ9pFVFmB9g4VvVmeBk','微信用户','','',1,'2026-05-15 06:03:56.311','2026-05-19 00:35:16.061','2026-05-15 06:03:56.322','2026-05-19 00:35:16.059',132,1,8,71.87,1,'2026-05-15 08:23:03.969'),(63,'oXPw33bdkgqWh-VJaAgWPOqz6OP4','oRIl55sOtkgv6vwy7eDexHPVCRsI','微信用户','','',1,'2026-05-28 01:23:27.958','2026-08-18 11:06:31.086','2026-05-28 01:23:27.964','2026-08-18 11:06:31.085',41,1,8,284.09,1,'2026-05-28 01:25:37.622'),(64,'o4mtI3Tai1BJTC-MNKOCcBJHVqv4','','微信用户','','',1,'2026-05-31 23:21:46.786','2026-05-31 23:32:35.517','2026-05-31 23:21:46.791','2026-05-31 23:32:35.514',10,1,6,0.10,1,'2026-05-31 23:22:02.087');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-08-23 11:10:45
