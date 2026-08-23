-- ============================================================
-- 阶段一：通用 RBAC 数据模型 + 种子数据
-- 表：sys_menus / sys_roles / sys_role_menus / sys_departments / merchant_staff_roles
-- 约束：纯 SQL 单条执行，禁止存储过程 / BEGIN...END / WHILE / IF...THEN
-- ============================================================

-- ============================================================
-- 1. 系统菜单表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_menus (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    parent_id   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID(0=顶级)',
    menu_type   TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '类型: 1=菜单/目录 2=按钮',
    name        VARCHAR(64) NOT NULL COMMENT '菜单名称',
    path        VARCHAR(128) DEFAULT NULL COMMENT '前端路由路径(菜单时)',
    icon        VARCHAR(64) DEFAULT NULL COMMENT '菜单图标',
    sort        INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值(越小越靠前)',
    status      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
    visible     TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否显示: 1=显示 0=隐藏',
    permission  VARCHAR(128) DEFAULT NULL COMMENT '权限标识(如 order:view)',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_sys_menus_parent_id (parent_id),
    KEY idx_sys_menus_permission (permission)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统菜单表';

-- ============================================================
-- 2. 系统角色表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_roles (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(64) NOT NULL COMMENT '角色名称',
    code        VARCHAR(64) NOT NULL COMMENT '角色编码(唯一)',
    remark      VARCHAR(256) DEFAULT NULL COMMENT '备注',
    status      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';

-- ============================================================
-- 3. 角色-菜单关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role_menus (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_id     BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    menu_id     BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_role_menus (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

-- ============================================================
-- 4. 系统部门表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_departments (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    parent_id   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父部门ID(0=顶级)',
    name        VARCHAR(64) NOT NULL COMMENT '部门名称',
    leader      VARCHAR(64) DEFAULT NULL COMMENT '负责人',
    phone       VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    sort        INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值',
    status      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_sys_departments_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统部门表';

-- ============================================================
-- 5. 员工-角色关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS merchant_staff_roles (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    staff_id    BIGINT UNSIGNED NOT NULL COMMENT '员工ID(merchant_staffs.id)',
    role_id     BIGINT UNSIGNED NOT NULL COMMENT '角色ID(sys_roles.id)',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_merchant_staff_roles (staff_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工角色关联表';

-- ============================================================
-- 6. merchant_staffs 新增部门ID列
-- ============================================================
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_staffs' AND COLUMN_NAME='department_id');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE merchant_staffs ADD COLUMN department_id BIGINT UNSIGNED DEFAULT NULL COMMENT '部门ID(sys_departments.id)' AFTER status");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 7. 种子菜单
-- ============================================================
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(1, 0, 1, '工作台', '/dashboard', 'HomeFilled', 1, 1, 1, 'dashboard:view'),
(2, 0, 1, '订单管理', '/orders', 'Tickets', 2, 1, 1, 'orders:view'),
(21, 2, 2, '核销/完成', NULL, NULL, 1, 1, 1, 'orders:complete'),
(22, 2, 2, '退款', NULL, NULL, 2, 1, 1, 'orders:refund'),
(23, 2, 2, '归还押金', NULL, NULL, 3, 1, 1, 'orders:return'),
(3, 0, 1, '商品管理', '/products', 'Goods', 3, 1, 1, NULL),
(31, 3, 1, '商品管理', '/products', NULL, 1, 1, 1, 'products:view'),
(311, 31, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'products:create'),
(312, 31, 2, '删除', NULL, NULL, 2, 1, 1, 'products:delete'),
(313, 31, 2, '上/下架/批量', NULL, NULL, 3, 1, 1, 'products:status'),
(314, 31, 2, '库存', NULL, NULL, 4, 1, 1, 'products:stock'),
(315, 31, 2, '规格', NULL, NULL, 5, 1, 1, 'products:specs'),
(32, 3, 1, '分类管理', '/categories', NULL, 2, 1, 1, 'categories:view'),
(321, 32, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'categories:create'),
(322, 32, 2, '删除', NULL, NULL, 2, 1, 1, 'categories:delete'),
(323, 32, 2, '排序', NULL, NULL, 3, 1, 1, 'categories:sort'),
(4, 0, 1, '服务人员', '/staff', 'User', 4, 1, 1, 'staff:view'),
(41, 4, 2, '审核/启停', NULL, NULL, 1, 1, 1, 'staff:update'),
(42, 4, 2, '重置密码', NULL, NULL, 2, 1, 1, 'staff:reset-password'),
(43, 4, 2, '删除', NULL, NULL, 3, 1, 1, 'staff:delete'),
(5, 0, 1, '数据分析', '/analytics', 'DataAnalysis', 5, 1, 1, 'analytics:view'),
(6, 0, 1, '商家资料', '/profile', 'Setting', 6, 1, 1, 'profile:view'),
(61, 6, 2, '保存资料', NULL, NULL, 1, 1, 1, 'profile:update'),
(62, 6, 2, '支付配置', NULL, NULL, 2, 1, 1, 'profile:payment'),
(63, 6, 2, '修改密码', NULL, NULL, 3, 1, 1, 'profile:password'),
(7, 0, 1, '系统管理', '/system', 'Setting', 7, 1, 1, NULL),
(71, 7, 1, '菜单管理', '/system/menus', NULL, 1, 1, 1, 'system:menu:view'),
(711, 71, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'system:menu:create'),
(712, 71, 2, '删除', NULL, NULL, 2, 1, 1, 'system:menu:delete'),
(72, 7, 1, '角色管理', '/system/roles', NULL, 2, 1, 1, 'system:role:view'),
(721, 72, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'system:role:create'),
(722, 72, 2, '删除', NULL, NULL, 2, 1, 1, 'system:role:delete'),
(723, 72, 2, '分配权限', NULL, NULL, 3, 1, 1, 'system:role:assign'),
(73, 7, 1, '部门管理', '/system/departments', NULL, 3, 1, 1, 'system:dept:view'),
(731, 73, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'system:dept:create'),
(732, 73, 2, '删除', NULL, NULL, 2, 1, 1, 'system:dept:delete'),
(74, 7, 1, '员工管理', '/system/staff', NULL, 4, 1, 1, 'system:staff:view'),
(741, 74, 2, '新增', NULL, NULL, 1, 1, 1, 'system:staff:create'),
(742, 74, 2, '编辑', NULL, NULL, 2, 1, 1, 'system:staff:update'),
(743, 74, 2, '删除', NULL, NULL, 3, 1, 1, 'system:staff:delete'),
(744, 74, 2, '重置密码', NULL, NULL, 4, 1, 1, 'system:staff:reset-password');

-- ============================================================
-- 8. 种子角色：超级管理员（绑定全部菜单）
-- ============================================================
INSERT INTO sys_roles (id, name, code, remark, status) VALUES
(1, '超级管理员', 'admin', '拥有全部菜单权限(系统预置)', 1);

INSERT INTO sys_role_menus (role_id, menu_id)
SELECT 1, id FROM sys_menus;
