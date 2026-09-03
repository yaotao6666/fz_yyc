# Tasks

- [x] Task 1: 数据库迁移 — 创建 `store_home_recommends` 表 + 「首页推荐」菜单与按钮权限
  - [x] SubTask 1.1: 新建 `server/migrations/029_home_recommend.sql`，创建商家维度推荐表（merchant_id、product_id、target_type、title、sort、status、时间戳），与 banner 表风格一致
  - [x] SubTask 1.2: 新增「首页推荐」顶级菜单节点 + 按钮权限码 `home-recommend:view/create/update/delete/status`（复用 banner 的 sys_menus 段，需先确认下一个空闲 id，建议 96 / 961-965）
  - [x] SubTask 1.3: 绑定超管角色（INSERT IGNORE sys_role_menus role_id=1）
  - [x] SubTask 1.4: 执行迁移并验证表结构存在

- [x] Task 2: 后端模型与商家端推荐管理接口
  - [x] SubTask 2.1: 在 models.go 新增 `HomeRecommend` 模型（TableName=store_home_recommends），引用 products 表名称/类型
  - [x] SubTask 2.2: 新增 `server/internal/handlers/merchant/home_recommend.go`：List/Create/Update/Delete/Status 五接口，商家隔离（merchant_id 取自当前登录商家），product_id 需校验目标商品/服务存在且状态有效
  - [x] SubTask 2.3: 在 main.go 商户端组注册 `/home-recommends` 路由，挂 RBAC(`home-recommend:*`)
  - [x] SubTask 2.4: 编译通过（`go build ./...`）

- [x] Task 3: C端首页推荐查询接口
  - [x] SubTask 3.1: 新增 C端接口（如 `GET /store/home-recommends`），仅返回当前商家已启用且排序升序的推荐项，JOIN products 附带商品/服务快照（name、cover、image、price、product_type、sale_type、id）
  - [x] SubTask 3.2: 在 main.go 注册 C端路由（无需商户 RBAC）
  - [x] SubTask 3.3: 编译通过

- [x] Task 4: PC「首页推荐」管理页
  - [x] SubTask 4.1: 新增 `web-admin/src/views/miniprogram/HomeRecommendView.vue`（列表 + 新增/编辑弹窗：选择 target_type 后从对应产品池选择目标商品/服务、标题、排序、启停按钮）
  - [x] SubTask 4.2: 新增 PC 端 API 封装（`web-admin/src/api/`），注册路由（复用 AppLayout 侧边菜单 + 权限码）
  - [x] SubTask 4.3: `npm run build` 通过

- [x] Task 5: C端首页改造（home.vue）
  - [x] SubTask 5.1: 删除店铺头部评分星星块（rating/stars/rating-value）及相关 CSS
  - [x] SubTask 5.2: 删除下方「分类和商品」整体浏览区模板（category-sidebar + product-list + 热销推荐），替换为「推荐商品/服务」模块（推荐卡片 + 空态）
  - [x] SubTask 5.3: 新增推荐列表拉取与渲染逻辑；点击推荐项跳转——实物卡片→商品详情，服务卡片→服务详情/挑选页
  - [x] SubTask 5.4: 仅删除专属于已移除区块、确无其他引用的 JS 方法与状态；保留金刚区/领券/宣教/头部逻辑
  - [x] SubTask 5.5: `npm run build:mp-weixin` 通过

## Task Dependencies
- Task 2 依赖 Task 1（表结构）
- Task 3 依赖 Task 1
- Task 4 依赖 Task 2（接口）
- Task 5 依赖 Task 3（C端接口）