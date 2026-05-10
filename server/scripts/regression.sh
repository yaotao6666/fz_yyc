#!/bin/bash

# ============================================
# 寻梦私域管家 - 回归测试脚本
# 版本: v1.0
# 日期: 2026-05-10
# 数据库: MySQL 8.0 (Docker)
# ============================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 打印分隔线
print_line() {
    echo "════════════════════════════════════════════════════════"
}

# 打印标题
print_title() {
    echo ""
    print_line
    echo -e "${BLUE}  $1${NC}"
    print_line
}

# 打印成功消息
print_success() {
    echo -e "  ${GREEN}✓${NC} $1"
}

# 打印失败消息
print_fail() {
    echo -e "  ${RED}✗${NC} $1"
}

# 打印警告消息
print_warn() {
    echo -e "  ${YELLOW}!${NC} $1"
}

# 打印信息
print_info() {
    echo -e "  ${CYAN}→${NC} $1"
}

# 主菜单
show_menu() {
    clear
    print_title "寻梦私域管家 - 回归测试"
    echo ""
    echo "  1. 初始化数据库"
    echo "  2. 执行完整回归测试"
    echo "  3. 测试服务商接口"
    echo "  4. 查看数据库状态"
    echo "  5. 查看测试账号"
    echo "  0. 退出"
    echo ""
    print_line
    echo ""
    echo -n "  请选择操作 [0-5]: "
}

# 初始化数据库
init_database() {
    print_title "初始化数据库"
    
    print_info "创建数据库..."
    docker exec fz_yyc_mysql mysql -uroot -proot123456 -e "CREATE DATABASE IF NOT EXISTS fz_yyc_api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>&1 | grep -v "Warning"
    
    if [ $? -eq 0 ]; then
        print_success "数据库创建成功"
    else
        print_fail "数据库创建失败"
        return 1
    fi
    
    print_info "导入初始化脚本..."
    docker exec -i fz_yyc_mysql mysql -uroot -proot123456 fz_yyc_api < migrations/20240101000000_full_init.sql 2>&1 | grep -v "Warning"
    
    if [ $? -eq 0 ]; then
        print_success "初始化脚本导入成功"
    else
        print_fail "初始化脚本导入失败"
        return 1
    fi
    
    print_info "验证数据..."
    local count=$(docker exec fz_yyc_mysql mysql -uroot -proot123456 fz_yyc_api -N -e "SELECT COUNT(*) FROM service_provider_admins;" 2>&1 | grep -v "Warning")
    
    if [ "$count" = "1" ]; then
        print_success "数据验证成功"
        print_info "服务商管理员: admin / admin123"
        print_info "商家员工: merchant / merchant123"
    else
        print_warn "数据可能未正确导入"
    fi
    
    echo ""
    print_success "数据库初始化完成！"
}

# 执行完整回归测试
run_full_test() {
    print_title "执行完整回归测试"
    
    print_info "步骤 1/3: 初始化数据库..."
    init_database
    
    if [ $? -ne 0 ]; then
        print_fail "数据库初始化失败，终止测试"
        return 1
    fi
    
    echo ""
    print_info "步骤 2/3: 测试服务商登录..."
    local response=$(curl -s -X POST http://localhost:8080/api/v1/auth/admin/login \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    if echo "$response" | grep -q '"code":0'; then
        print_success "服务商登录测试通过"
        print_info "Token: $(echo $response | grep -o '"token":"[^"]*"' | cut -d'"' -f4 | head -c 50)..."
    else
        print_fail "服务商登录测试失败"
        print_info "响应: $response"
    fi
    
    echo ""
    print_info "步骤 3/3: 检查数据库状态..."
    docker exec fz_yyc_mysql mysql -uroot -proot123456 fz_yyc_api -e \
        "SELECT 'Service Providers' as type, COUNT(*) as count FROM service_provider_admins
         UNION ALL SELECT 'Merchants', COUNT(*) FROM merchant_staffs
         UNION ALL SELECT 'Products', COUNT(*) FROM products
         UNION ALL SELECT 'Users', COUNT(*) FROM users
         UNION ALL SELECT 'Orders', COUNT(*) FROM orders;" 2>&1 | grep -v "Warning"
    
    echo ""
    print_success "回归测试完成！"
}

# 测试服务商接口
test_sp_api() {
    print_title "测试服务商接口"
    
    print_info "测试服务商登录..."
    echo ""
    
    local response=$(curl -s -X POST http://localhost:8080/api/v1/auth/admin/login \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    if echo "$response" | grep -q '"code":0'; then
        print_success "登录成功"
        echo ""
        print_info "完整响应:"
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        print_fail "登录失败"
        echo ""
        print_info "响应:"
        echo "$response"
    fi
}

# 查看数据库状态
show_db_status() {
    print_title "数据库状态"
    
    print_info "容器状态:"
    docker ps --filter "name=fz_yyc" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    echo ""
    print_info "数据库统计:"
    docker exec fz_yyc_mysql mysql -uroot -proot123456 fz_yyc_api -e \
        "SELECT 'Tables' as item, COUNT(*) as count FROM information_schema.tables WHERE table_schema = 'fz_yyc_api'
         UNION ALL SELECT 'Service Providers', COUNT(*) FROM service_provider_admins
         UNION ALL SELECT 'Merchants', COUNT(*) FROM merchant_staffs
         UNION ALL SELECT 'Products', COUNT(*) FROM products
         UNION ALL SELECT 'Users', COUNT(*) FROM users
         UNION ALL SELECT 'Orders', COUNT(*) FROM orders;" 2>&1 | grep -v "Warning"
}

# 显示测试账号
show_accounts() {
    print_title "测试账号信息"
    
    echo ""
    echo -e "  ${CYAN}服务商管理员${NC}"
    echo "  ───────────────────────────────────"
    echo "  用户名: admin"
    echo "  密码:   admin123"
    echo "  角色:   超级管理员"
    echo "  路径:   POST /api/v1/auth/admin/login"
    
    echo ""
    echo -e "  ${CYAN}商家员工${NC}"
    echo "  ───────────────────────────────────"
    echo "  用户名: merchant"
    echo "  密码:   merchant123"
    echo "  角色:   商家管理员"
    echo "  商家:   美味餐厅 (ID=1)"
    
    echo ""
    echo -e "  ${CYAN}初始化数据说明${NC}"
    echo "  ───────────────────────────────────"
    echo "  当前单文件初始化脚本仅保留服务商与商家登录账号"
    echo "  C端用户、商品、订单与行为数据由联调或测试过程生成"
    echo ""
}

# 主函数
main() {
    # 检查 Docker 是否运行
    if ! docker ps &> /dev/null; then
        print_fail "Docker 未运行，请先启动 Docker"
        exit 1
    fi
    
    # 检查必要的容器
    if ! docker ps --format '{{.Names}}' | grep -q "fz_yyc_mysql"; then
        print_fail "MySQL 容器未运行"
        print_info "请运行: docker-compose up -d"
        exit 1
    fi
    
    # 显示菜单并处理选择
    while true; do
        show_menu
        read choice
        echo ""
        
        case $choice in
            1)
                init_database
                echo ""
                echo -n "按 Enter 键继续..."
                read
                ;;
            2)
                run_full_test
                echo ""
                echo -n "按 Enter 键继续..."
                read
                ;;
            3)
                test_sp_api
                echo ""
                echo -n "按 Enter 键继续..."
                read
                ;;
            4)
                show_db_status
                echo ""
                echo -n "按 Enter 键继续..."
                read
                ;;
            5)
                show_accounts
                echo ""
                echo -n "按 Enter 键继续..."
                read
                ;;
            0)
                print_info "感谢使用！"
                exit 0
                ;;
            *)
                print_fail "无效选择，请重新输入"
                sleep 1
                ;;
        esac
    done
}

# 执行主函数
main
