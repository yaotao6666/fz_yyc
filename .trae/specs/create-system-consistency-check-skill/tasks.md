# Tasks

- [x] Task 1: 创建技能文件框架
  - [x] SubTask 1.1: 在 `.trae/skills/` 下创建 `system-consistency-check.md` 文件
  - [x] SubTask 1.2: 编写技能概述、前置条件、相关文件路径索引
  - [x] SubTask 1.3: 编写注意事项（并发安全、向后兼容、事务完整、数据备份、文档同步）

- [x] Task 2: 编写PRD与数据库一致性检测流程
  - [x] SubTask 2.1: 定义检测步骤（读取PRD → 提取表需求 → 对比models.go → 对比migrations → 输出差异）
  - [x] SubTask 2.2: 定义差异清单输出格式（缺失的表、缺失的字段、类型不一致、PRD未定义的扩展）
  - [x] SubTask 2.3: 列出需检测的关键文件路径（PRD.md、models.go、migrations/）

- [x] Task 3: 编写接口与模型一致性检测流程
  - [x] SubTask 3.1: 定义检测步骤（扫描handlers → 提取请求/响应结构体 → 对比models.go → 验证路由映射 → 检查中间件）
  - [x] SubTask 3.2: 定义差异清单输出格式（请求结构体不一致、响应结构体不一致、路由映射问题、参数验证问题）
  - [x] SubTask 3.3: 列出需检测的handler目录（admin/、merchant/、user/、upload/、ws/）

- [x] Task 4: 编写前端页面接口接入检测流程
  - [x] SubTask 4.1: 定义检测步骤（扫描pages → 提取API调用 → 对比后端路由 → 验证参数 → 检查认证方式）
  - [x] SubTask 4.2: 定义差异清单输出格式（接口路径错误、请求参数不一致、响应解析问题、认证方式问题）
  - [x] SubTask 4.3: 列出需检测的页面目录（merchant/、store/、auth/）和API封装文件（api/index.ts）

- [x] Task 5: 编写自测与脚本生成流程
  - [x] SubTask 5.1: 定义自测验证步骤（curl测试核心接口、验证响应结构）
  - [x] SubTask 5.2: 定义迁移脚本生成规范（命名格式、事务安全、回滚语句、初始数据）
  - [x] SubTask 5.3: 定义初始化数据清单模板（系统配置、管理员账号、基础分类、系统参数）

- [x] Task 6: 编写并行执行说明
  - [x] SubTask 6.1: 说明任务1/2/3可并行执行的方式
  - [x] SubTask 6.2: 说明任务4（自测与脚本生成）依赖前三个任务完成
  - [x] SubTask 6.3: 定义汇总输出格式（合并三个差异清单 + 修复建议 + 脚本清单）

# Task Dependencies
- [Task 2] [Task 3] [Task 4] 可并行执行，无依赖关系
- [Task 5] 依赖 [Task 2] [Task 3] [Task 4] 完成
- [Task 6] 与 [Task 2] [Task 3] [Task 4] 并行编写，但执行时作为流程控制说明
- [Task 1] 是所有任务的前置，需最先完成
