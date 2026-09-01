# Checklist

- [ ] 迁移脚本 `016_alert_center_extend.sql` 存在且已执行：staff\_id 可空、summary 列、orders.assigned\_at、alert\_settings 表+默认行、sys\_menus 972 节点

- [ ] models.go 中 ServiceAlertEvent（StaffID \*uint64、Summary）、Order（AssignedAt）、AlertSettings 与迁移一致

- [ ] 预警类型 3\~8 常量与默认阈值在 constants.go 定义

- [ ] GET/PUT `/merchant/alert-settings` 接口已实现并带 RBAC 权限

- [ ] DispatchOrder / AcceptOrder 写入 assigned\_at

- [ ] 6 个扫描函数按配置阈值生成预警且幂等（同 order\_id+alert\_type 仅一条未闭环）

- [ ] `StartOrderTimeoutTasks` 已注册（启动即扫 + 5 分钟 ticker），enabled=false 跳过

- [ ] AlertEventsView 支持新类型筛选、空 staff 显示"-"、summary 列、"预警设置"弹窗（v-permission）

- [ ] `go build ./...` 与 web-admin `npm run build` 通过

- [ ] 接口冒烟验证通过（GET/PUT 配置、按类型筛选、staff 空值）

