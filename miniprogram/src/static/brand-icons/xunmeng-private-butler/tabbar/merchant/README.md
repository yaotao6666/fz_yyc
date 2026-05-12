# 商家端 TabBar 图标（SVG 源文件）

对应 `pages.json` 的商家端底部四个 Tab：工作台 / 数据分析 / 订单 / 设置。

文件与状态：

- `home.svg`、`home-active.svg`
- `analytics.svg`、`analytics-active.svg`
- `order.svg`、`order-active.svg`
- `settings.svg`、`settings-active.svg`

配色建议：

- 未选中：`#9AA7B4`
- 选中：`#FFB020`（点缀橙：`#FF8A00`）

导出 PNG 建议：

- 画布源尺寸：`1024x1024`（本目录已统一）
- 导出 TabBar PNG 常用：`81x81`（或 `64x64`，按你现有资源口径）
- 建议保留 14%~18% 内边距，避免小图标贴边发糊。

一键转换脚本：

- `bash miniprogram/scripts/convert-merchant-tabbar-svg-to-png.sh`（默认导出 `81x81`）
- `bash miniprogram/scripts/convert-merchant-tabbar-svg-to-png.sh 64`（导出 `64x64`）
