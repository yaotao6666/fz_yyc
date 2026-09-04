-- ============================================================
-- 002_seed.sql  — 基础种子数据（合并最终版）
-- ------------------------------------------------------------
-- 说明：系统级先决数据，供全新环境初始化后导入。
-- 包含：单商户 merchant(1) 与管理员 merchant 账号、
--       医疗辅具分类树、全品类商品（零售+租赁）、
--       RBAC 菜单/角色/角色-菜单绑定、商家配送/活动/公告、
--       空 service_staffs 表结构。
-- 不包含业务数据（用户、订单、健康档案等），
--       此类数据后续由线上环境导出导入。
-- 执行：必须在执行 001_schema.sql 建表之后执行。
-- ============================================================
-- 源库 MySQL 8.0.46

-- MySQL dump 起始标记

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
-- Table structure for table `merchants`
--

DROP TABLE IF EXISTS `merchants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
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
  `sub_mch_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '微信支付子商户号(线下进件后回填)',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '营业状态: 1=营业中 0=休息中',
  `rating` decimal(2,1) NOT NULL DEFAULT '5.0' COMMENT '商家评分(1.0-5.0)',
  `sales_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '累计销量',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `cover_image` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '商家背景/封面图地址(七牛私有路径)',
  `payment_config_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)',
  `profit_sharing_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启自动分账',
  PRIMARY KEY (`id`),
  KEY `idx_merchants_sub_mch_id` (`sub_mch_id`),
  KEY `idx_merchants_status` (`status`),
  KEY `idx_merchants_location` (`lat`,`lng`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchants`
--

LOCK TABLES `merchants` WRITE;
/*!40000 ALTER TABLE `merchants` DISABLE KEYS */;
INSERT INTO `merchants` VALUES (1,'乐享辅具（云南财旭商贸）','https://lexiang-oss.jxxme.cn/uploads/merchant/1/1778940798537.jpg','李四','13539565631','lisi@example.com','北京市朝阳区建国路88号',39.908823,116.407470,'二类医疗器械/康复辅具 与 共享租赁','08:00-21:00','全品类二类医疗辅具，支持零售与共享租赁低价套餐，专业适老康复服务。','1112979963',1,5.0,0,'2026-05-11 00:38:11.000','2026-08-23 09:53:51.908','https://lexiang-oss.jxxme.cn/uploads/merchant/1/1778940807903.png',1,1);
/*!40000 ALTER TABLE `merchants` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `merchant_staffs`
--

DROP TABLE IF EXISTS `merchant_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `merchant_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
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
  UNIQUE KEY `uk_merchant_staffs_username` (`username`),
  KEY `idx_merchant_staffs_openid` (`openid`),
  KEY `idx_merchant_staffs_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商家员工表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `merchant_staffs`
--

LOCK TABLES `merchant_staffs` WRITE;
/*!40000 ALTER TABLE `merchant_staffs` DISABLE KEYS */;
INSERT INTO `merchant_staffs` VALUES (1,'merchant','$2a$10$SHU1i6GAOMFnglRXvuZuWu9aBuM6qPN1Scqdp1Y9MWgpgc73t2XJ6','商家管理员','13900139000','oXPw33bdkgqWh-VJaAgWPOqz6OP4','oRIl55sOtkgv6vwy7eDexHPVCRsI','2026-05-28 01:21:16','owner',1,1,1,NULL,'2026-08-23 11:04:23','2026-05-31 23:10:53','2026-05-11 00:38:11','2026-08-23 11:04:23');
/*!40000 ALTER TABLE `merchant_staffs` ENABLE KEYS */;
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
-- Table structure for table `service_staffs`
--

DROP TABLE IF EXISTS `service_staffs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `service_staffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '服务人员ID',
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
  KEY `idx_service_staffs_open_id` (`openid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_staffs`
--

LOCK TABLES `service_staffs` WRITE;
/*!40000 ALTER TABLE `service_staffs` DISABLE KEYS */;
INSERT INTO `service_staffs` VALUES (1,'yaotao','$2a$10$lKkm0izHgfEO1r9lx9P/qea7AJxb5CeLkL7yjiysdQTWl.vpVcsaq','yaotao','13539565631','','',1,'2026-08-23 10:53:17.455','2026-08-15 22:41:30.819','2026-08-23 10:53:17.456');
/*!40000 ALTER TABLE `service_staffs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '分类类型: 1=商品分类 2=服务分类',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父分类ID(空=一级)',
  `level` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '层级: 1=一级 2=二级 3=三级',
  `sort` int unsigned NOT NULL DEFAULT '0',
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_categories_parent` (`parent_id`),
  KEY `idx_categories_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=743 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categories`
--

LOCK TABLES `categories` WRITE;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` VALUES (1,'轮椅',1,NULL,1,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(2,'助行器具',1,NULL,1,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(3,'无障碍扶手',1,NULL,1,3,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(4,'康复护理床',1,NULL,1,4,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(5,'康复理疗器材',1,NULL,1,5,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(6,'护理耗材配件',1,NULL,1,6,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(7,'共享租赁专区',2,NULL,1,7,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(11,'手动轮椅',1,1,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(12,'电动轮椅',1,1,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(21,'助行拐杖',1,2,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(22,'助行器/学步车',1,2,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(31,'马桶安全扶手',1,3,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(32,'走廊/浴室扶手',1,3,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(41,'手动护理床',1,4,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(42,'电动护理床',1,4,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(51,'四肢训练器',1,5,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(52,'理疗按摩仪',1,5,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(71,'轮椅租赁',2,7,2,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(72,'助行器租赁',2,7,2,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(73,'护理床租赁',2,7,2,3,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(74,'康复器械租用套餐',2,7,2,4,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(111,'轻便折叠款',1,11,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(112,'高背全躺款',1,11,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(121,'标准电动',1,12,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(122,'高续航电动',1,12,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(311,'可调节落地扶手',1,31,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(711,'周租低价套餐',2,71,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(712,'月租低价套餐',2,71,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(741,'四肢训练器租用',2,74,3,1,1,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(742,'气压按摩仪租用',2,74,3,2,1,'2026-08-18 16:29:46','2026-08-18 16:29:46');
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
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
  KEY `idx_products_category_id` (`category_id`),
  KEY `idx_products_status` (`status`),
  KEY `idx_products_sale_type` (`sale_type`),
  KEY `idx_products_sales` (`sales`),
  KEY `idx_products_sort` (`sort`),
  KEY `idx_products_product_type` (`product_type`,`status`),
  CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=20007 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `products`
--

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES (10001,111,'铝合金折叠轮椅','轻量铝合金车架，一键折叠，坐宽45cm，承重100kg，适合居家出行。','[]',1280.00,1580.00,20,'台',1,NULL,1,0,0.00,0.00,0,12,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10002,111,'轻便便携旅行轮椅','超轻6kg，可放入后备箱，配旅行袋，出行便携首选。','[]',1580.00,1880.00,15,'台',1,NULL,1,0,0.00,0.00,0,8,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10003,112,'高背全躺轮椅','大轮高背，靠背可调至全躺，适合长时间坐卧者。','[]',1680.00,2080.00,12,'台',1,NULL,1,0,0.00,0.00,0,6,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10004,112,'全躺式看护轮椅','座便两用，可全躺可拆卸腿托，方便护理。','[]',1880.00,2280.00,10,'台',1,NULL,1,0,0.00,0.00,0,5,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10005,121,'基础款电动轮椅','锂电池 12km，可折叠，遥控/手推双模式。','[]',3990.00,4590.00,8,'台',1,NULL,1,0,0.00,0.00,0,9,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10006,122,'高续航电动轮椅','锂电池续航20km，防后倾，上下肢驱动助力。','[]',5500.00,6290.00,6,'台',1,NULL,1,0,0.00,0.00,0,4,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10007,21,'四脚助行拐杖','铝合金四脚，10档高度可调，承重125kg。','[]',89.00,109.00,80,'支',1,NULL,1,0,0.00,0.00,0,20,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10008,21,'可调高度单拐','轻便单拐，磨砂握把防滑，出街轻巧。','[]',59.00,79.00,100,'支',1,NULL,1,0,0.00,0.00,0,25,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10009,21,'铝合金肘拐','肘托承重，适合单侧下肢支撑，康复期常用。','[]',129.00,159.00,60,'支',1,NULL,1,0,0.00,0.00,0,15,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10010,22,'四轮助行器带刹车','一键刹车，带坐垫可歇脚，适老助行。','[]',299.00,359.00,30,'个',1,NULL,1,0,0.00,0.00,0,18,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10011,22,'老人三脚助行架','三角稳定结构，防滑底脚，轻便稳固。','[]',249.00,299.00,25,'个',1,NULL,1,0,0.00,0.00,0,10,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10012,22,'宝宝学步车','护栏可拆，静音万向轮，安全学步。','[]',199.00,239.00,20,'辆',1,NULL,1,0,0.00,0.00,0,14,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10013,311,'可调节马桶助力架','免打孔可调宽，起身支撑更省力。','[]',359.00,419.00,20,'个',1,NULL,1,0,0.00,0.00,0,11,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10014,31,'U型马桶安全扶手','带防滑盖板，不锈钢承重，适老化改造。','[]',399.00,469.00,18,'个',1,NULL,1,0,0.00,0.00,0,9,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10015,32,'浴室L型扶手','304不锈钢，承重200kg，浴缸/淋浴墙面安装。','[]',129.00,169.00,50,'个',1,NULL,1,0,0.00,0.00,0,16,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10016,32,'走廊连续扶手','墙面长扶手，适老化走廊/过道安装。','[]',59.00,79.00,100,'米',1,NULL,1,0,0.00,0.00,0,13,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10017,41,'手动三折护理床','背/腿/脚多段升降，配护栏，家庭护理。','[]',2680.00,3180.00,10,'台',1,NULL,1,0,0.00,0.00,0,7,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10018,42,'五功能电动护理床','背腿升降+翻身+便孔，配防褥疮床垫。','[]',6800.00,7980.00,6,'台',1,NULL,1,0,0.00,0.00,0,5,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10019,42,'翻身防压疮电动床','定时翻身，减轻护理负担，适合长期卧床。','[]',8990.00,10900.00,4,'台',1,NULL,1,0,0.00,0.00,0,3,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10020,51,'手指握力康复训练器','五档阻力，手部精细动作康复。','[]',79.00,99.00,90,'个',1,NULL,1,0,0.00,0.00,0,22,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10021,51,'肩关节滑轮训练器','家用滑轮，爬墙训练，肩部术后康复。','[]',299.00,359.00,30,'个',1,NULL,1,0,0.00,0.00,0,12,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10022,51,'弹力带康复套装','五色拉力带，全身肌力训练。','[]',49.00,69.00,120,'套',1,NULL,1,0,0.00,0.00,0,26,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10023,52,'四肢气压按摩仪','多档气压，促进血液循环，缓解浮肿。','[]',1290.00,1590.00,15,'台',1,NULL,1,0,0.00,0.00,0,8,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10024,6,'防褥疮充气床垫','交替充气，分散压力，预防压疮。','[]',499.00,599.00,30,'个',1,NULL,1,0,0.00,0.00,0,17,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10025,6,'轮椅防压疮坐垫','凝胶减压，透气久坐不闷。','[]',129.00,159.00,60,'个',1,NULL,1,0,0.00,0.00,0,19,2,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(10026,6,'透气护理垫','加厚防水，一次性医疗级护理垫。','[]',39.00,49.00,200,'包',1,NULL,1,0,0.00,0.00,0,30,3,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20001,711,'手动轮椅·周租套餐','共享租赁低价体验，按周计费，含押金，到期归还。','[]',50.00,0.00,5,'台/周',2,NULL,2,2,50.00,200.00,12,6,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20002,712,'手动轮椅·月租套餐','共享租赁低价月租，长期使用更划算。','[]',80.00,0.00,8,'台/月',2,NULL,2,3,80.00,200.00,6,9,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20003,72,'四轮助行器·月租套餐','带刹车助行器按月租，康复期安心用。','[]',60.00,0.00,6,'个/月',2,NULL,2,3,60.00,100.00,6,7,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20004,73,'手动护理床·月租套餐','护理床按月租，含安装指导，居家照护。','[]',300.00,0.00,4,'台/月',2,NULL,2,3,300.00,800.00,6,4,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20005,741,'四肢联动训练器·月租套餐','共享康复器械，按月租用，配合康复指导。','[]',200.00,0.00,5,'台/月',2,NULL,2,3,200.00,500.00,6,3,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46'),(20006,742,'气压按摩仪·周租套餐','康复理疗器械低价租用，促血液循环。','[]',60.00,0.00,6,'台/周',2,NULL,2,2,60.00,300.00,12,5,1,1,NULL,'2026-08-18 16:29:46','2026-08-18 16:29:46');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
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
-- Table structure for table `announcements`
--

DROP TABLE IF EXISTS `announcements`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `announcements` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `status` tinyint unsigned NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_announcements_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统公告表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `announcements`
--

LOCK TABLES `announcements` WRITE;
/*!40000 ALTER TABLE `announcements` DISABLE KEYS */;
INSERT INTO `announcements` VALUES (1,'五一活动复盘','请各商家及时查看活动复盘和经营建议。',1,'2026-05-02 10:00:00','2026-05-11 09:00:00'),(2,'新版打印模板上线','支持堂食和外卖分模板打印。',1,'2026-05-04 10:00:00','2026-05-11 09:00:00'),(3,'商家入驻资料规范','请新商家按照最新模板补充结算资料。',1,'2026-05-06 10:00:00','2026-05-11 09:00:00'),(4,'系统维护通知','本周日晚间将进行系统维护。',0,'2026-05-08 10:00:00','2026-05-11 09:00:00'),(5,'经营建议周报','系统已根据近期经营情况生成建议。',1,'2026-05-10 10:00:00','2026-05-11 09:00:00'),(6,'ok','ok',1,'2026-05-16 07:16:42','2026-05-16 07:16:42'),(7,'Web后台公告联调验证Web后台公告联调验证-已编辑','这是一条用于验证 web-admin 公告管理新增与编辑流程的测试公告。这是一条用于验证 web-admin 公告管理编辑保存流程的测试公告，内容已更新。',0,'2026-05-25 23:55:27','2026-05-25 23:58:23');
/*!40000 ALTER TABLE `announcements` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-08-23 11:23:14
