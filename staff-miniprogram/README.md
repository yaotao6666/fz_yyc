# 财旭商贸服务人员端小程序 (staff-miniprogram)

第 0 期：空工程搭建，对齐技术栈，搭建 4 个一级 Tab 页面脚手架。

## 技术栈
- **框架**：uni-app 3.0 (Vue 3.5 + Vite 5)
- **语言**：TypeScript 5.4
- **状态管理**：Pinia 2.1
- **样式**：SASS，主题色 `#517528`（财旭绿）

## 4 个核心 Tab
1. **待办 (pages/todo)** — 今日数据概览 + 高优待办清单（签到/验机/紧急）
2. **工单 (pages/workorder)** — 全部工单：待出发 / 服务中 / 已完成 / Tab 切换
3. **排班 (pages/schedule)** — 月视图排班日历（早班/晚班/休息/待命）
4. **我的 (pages/profile)** — 个人信息 + 签到历史 + 服务记录 + 设置 + 退出登录

## 目录结构
```
staff-miniprogram/
├── src/
│   ├── api/index.ts            # 第2期接口占位（工单/签到/排班/验机/个人）
│   ├── stores/auth.ts          # 登录态 Pinia Store
│   ├── types/index.ts          # OrderType / BizStatus / StaffUser / Workorder
│   ├── utils/format.ts         # 日期工具
│   ├── config/env.ts           # 读取 env
│   ├── pages/
│   │   ├── todo/index.vue
│   │   ├── workorder/index.vue
│   │   ├── schedule/index.vue
│   │   └── profile/index.vue
│   ├── App.vue                 # 财旭绿主题色 CSS 变量
│   ├── pages.json              # TabBar + 4 个页面
│   ├── manifest.json           # 微信小程序配置（appid 待填）
│   └── ...
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 启动
```bash
cd staff-miniprogram
npm install
npm run dev:mp-weixin      # 开发
npm run build:mp-weixin    # 构建（微信小程序 dist/build/mp-weixin）
```

## 第 2 期实现要点（占位未开发）
1. 登录：微信授权 code → 后端换 token + 服务人员资料
2. 工作台：`assigned_staff_id = 当前服务人员ID` 查询订单
3. 签到/签退：`wx.getLocation` 上传坐标 + 更新 `actual_started_at/actual_ended_at`
4. 租赁验机归还：调用后端 ReturnRentalOrder 接口
5. 服务记录附件：七牛云上传图片 + 服务记录
6. 排班：按月调度 + 换班/请假申请

## TabBar 图标
页面配置已引用 `static/tabbar/*.png`，需要第 1 期将 8 张图标放入：
- todo.png / todo-active.png
- workorder.png / workorder-active.png
- schedule.png / schedule-active.png
- profile.png / profile-active.png

建议尺寸 81×81 px，不超过 40 KB。
