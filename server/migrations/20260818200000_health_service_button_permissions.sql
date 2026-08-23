-- ============================================================
-- 健康服务按钮级 RBAC 权限码补齐与拆分（细粒度）
-- 表：sys_menus / sys_role_menus
-- 目标：确保 前端按钮码 == 后端接口鉴权码 == sys_menus.permission 三方一致
-- 说明：
--   - 评估量表(82)：原聚合码 assessment:config 细分为
--       assessment:create(新增量表) / assessment:update(编辑·启停) / assessment:delete(删除量表)
--   - 照护计划(85)：原聚合码 care:create 细分为
--       care:create(新增) / care:update(编辑) / care:delete(删除)
-- 幂等：UPDATE + INSERT IGNORE；绑定超级管理员(role_id=1)
-- ============================================================
SET NAMES utf8mb4;

-- 1. 评估量表(82)：821 由聚合码收敛为“新增量表” assessment:create
UPDATE sys_menus SET name = '新增量表', permission = 'assessment:create' WHERE id = 821;

-- 新增“编辑/启停量表”与“删除量表”按钮节点
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(822, 82, 2, '编辑/启停量表', NULL, NULL, 2, 1, 1, 'assessment:update'),
(823, 82, 2, '删除量表', NULL, NULL, 3, 1, 1, 'assessment:delete');

-- 2. 照护计划(85)：851 保留 care:create(新增)，补齐编辑/删除
UPDATE sys_menus SET name = '新增照护计划' WHERE id = 851;

INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(852, 85, 2, '编辑照护计划', NULL, NULL, 2, 1, 1, 'care:update'),
(853, 85, 2, '删除照护计划', NULL, NULL, 3, 1, 1, 'care:delete');

-- 3. 超级管理员(role_id=1)绑定上述拆分节点，保证既有账号权限不丢失
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (821, 822, 823, 851, 852, 853);