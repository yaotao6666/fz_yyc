# Tasks
- [x] Task 1: 删除服务商工作台中的系统公告入口
  - [x] 从 `miniprogram/src/pages/sp/home.vue` 移除“系统公告”菜单卡片与跳转逻辑
  - [x] 确认服务商首页剩余菜单布局与样式正常

- [x] Task 2: 删除小程序中的服务商公告页面路由与页面文件
  - [x] 从 `miniprogram/src/pages.json` 移除 `pages/sp/announcements/index` 与 `pages/sp/announcements/edit` 路由
  - [x] 删除 `miniprogram/src/pages/sp/announcements/index.vue`
  - [x] 删除 `miniprogram/src/pages/sp/announcements/edit.vue`

- [x] Task 3: 清理小程序中仅供服务商公告页面使用的前端调用
  - [x] 检查 `miniprogram/src/api/index.ts` 中服务商公告相关方法是否仅被公告页面使用
  - [x] 删除仅供已移除页面使用的前端 API 封装，保留商家端公告接口
  - [x] 确认不改动后端 `/api/v1/sp/announcements` 接口注册与处理器

- [x] Task 4: 同步文档并验证
  - [x] 更新 `docs/prd/PRD-功能说明.md` 与 `PRD.md`，说明小程序端已下线服务商公告页面、接口保留供 PC 端复用
  - [x] 运行小程序诊断与 `npm run build:mp-weixin`，确认删除页面后构建正常

# Task Dependencies
- Task 2 depends on Task 1
- Task 3 depends on Task 2
- Task 4 depends on Task 1, Task 2, Task 3
