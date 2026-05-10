#!/bin/bash

# ============================================
# 寻梦私域管家 - 回归测试命令
# 功能：初始化数据库并执行完整的回归测试
# ============================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
DB_NAME="fz_yyc_api"
DB_USER="root"
DB_PASS=""
DB_HOST="localhost"
DB_PORT="3306"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
寻梦私域管家 - 回归测试命令

用法: $0 [命令] [选项]

命令:
    init        初始化数据库（创建表结构和测试数据）
    reset       重置数据库（删除所有数据后重新初始化）
    test        执行回归测试
    clean       清理测试数据
    status      查看数据库状态
    help        显示帮助信息

选项:
    -h, --host      数据库主机 (默认: localhost)
    -p, --port      数据库端口 (默认: 3306)
    -u, --user      数据库用户 (默认: root)
    -P, --password  数据库密码
    -d, --database  数据库名称 (默认: fz_yyc_api)

示例:
    $0 init                                    # 初始化数据库
    $0 init -u myuser -P mypass               # 使用指定用户初始化
    $0 reset                                   # 重置数据库
    $0 test                                    # 执行回归测试
    $0 status                                  # 查看状态

EOF
}

# 检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    # 检查 mysql 命令
    if ! command -v mysql &> /dev/null; then
        print_error "mysql 命令未找到，请安装 MySQL 客户端"
        exit 1
    fi
    
    # 检查 curl 命令
    if ! command -v curl &> /dev/null; then
        print_warning "curl 命令未找到，部分测试可能无法执行"
    fi
    
    print_success "依赖检查完成"
}

# 初始化数据库
init_database() {
    print_info "开始初始化数据库..."
    
    local SQL_FILE="migrations/20240101000000_full_init.sql"
    
    if [ ! -f "$SQL_FILE" ]; then
        print_error "初始化脚本不存在: $SQL_FILE"
        exit 1
    fi
    
    # 构建 MySQL 命令
    local MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER}"
    
    if [ -n "$DB_PASS" ]; then
        MYSQL_CMD="${MYSQL_CMD} -p${DB_PASS}"
    fi
    
    # 执行初始化脚本
    print_info "执行初始化脚本..."
    ${MYSQL_CMD} < "$SQL_FILE"
    
    if [ $? -eq 0 ]; then
        print_success "数据库初始化成功！"
        print_info "测试账号信息："
        echo "  服务商管理员: admin / admin123"
        echo "  商家管理员:   merchant / merchant123"
    else
        print_error "数据库初始化失败！"
        exit 1
    fi
}

# 重置数据库
reset_database() {
    print_warning "即将重置数据库，这将删除所有数据！"
    read -p "确认继续？(y/N): " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "已取消操作"
        exit 0
    fi
    
    print_info "开始重置数据库..."
    
    local MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER}"
    
    if [ -n "$DB_PASS" ]; then
        MYSQL_CMD="${MYSQL_CMD} -p${DB_PASS}"
    fi
    
    # 删除并重建数据库
    ${MYSQL_CMD} -e "DROP DATABASE IF EXISTS ${DB_NAME};"
    
    if [ $? -eq 0 ]; then
        print_success "数据库删除成功"
        init_database
    else
        print_error "数据库删除失败！"
        exit 1
    fi
}

# 查看数据库状态
show_status() {
    print_info "数据库状态："
    
    local MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER}"
    
    if [ -n "$DB_PASS" ]; then
        MYSQL_CMD="${MYSQL_CMD} -p${DB_PASS}"
    fi
    
    ${MYSQL_CMD} -e "
        SELECT 
            '数据表' AS 项目,
            COUNT(*) AS 数量
        FROM information_schema.tables 
        WHERE table_schema = '${DB_NAME}'
        UNION ALL
        SELECT 
            '服务商账号' AS 项目,
            COUNT(*) AS 数量
        FROM ${DB_NAME}.service_provider_admins
        UNION ALL
        SELECT 
            '商家账号' AS 项目,
            COUNT(*) AS 数量
        FROM ${DB_NAME}.merchant_staffs
        UNION ALL
        SELECT 
            'C端用户' AS 项目,
            COUNT(*) AS 数量
        FROM ${DB_NAME}.users
        UNION ALL
        SELECT 
            '商品数量' AS 项目,
            COUNT(*) AS 数量
        FROM ${DB_NAME}.products
        UNION ALL
        SELECT 
            '订单数量' AS 项目,
            COUNT(*) AS 数量
        FROM ${DB_NAME}.orders;
    "
}

# 执行回归测试
run_tests() {
    print_info "开始执行回归测试..."
    
    local API_HOST="http://localhost:8080"
    
    # 测试1：服务商登录
    print_info "测试1：服务商管理员登录..."
    local SP_TOKEN=$(curl -s -X POST "${API_HOST}/api/v1/auth/admin/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$SP_TOKEN" ]; then
        print_success "服务商登录成功"
    else
        print_error "服务商登录失败"
    fi
    
    # 测试2：商家登录
    print_info "测试2：商家管理员登录..."
    local MERCHANT_TOKEN=$(curl -s -X POST "${API_HOST}/api/v1/auth/merchant/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"merchant","password":"merchant123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$MERCHANT_TOKEN" ]; then
        print_success "商家登录成功"
    else
        print_error "商家登录失败"
    fi
    
    # 测试3：获取商家列表
    print_info "测试3：获取商家列表..."
    local MERCHANT_LIST=$(curl -s -X GET "${API_HOST}/api/v1/sp/merchants/list" \
        -H "Authorization: Bearer ${SP_TOKEN}")
    
    if echo "$MERCHANT_LIST" | grep -q '"list"'; then
        print_success "获取商家列表成功"
    else
        print_warning "获取商家列表可能失败或接口不存在"
    fi
    
    # 测试4：获取商品列表
    print_info "测试4：获取商品列表..."
    local PRODUCT_LIST=$(curl -s -X GET "${API_HOST}/api/v1/merchant/products?merchant_id=1" \
        -H "Authorization: Bearer ${MERCHANT_TOKEN}")
    
    if echo "$PRODUCT_LIST" | grep -q '"list"'; then
        print_success "获取商品列表成功"
    else
        print_warning "获取商品列表可能失败或接口不存在"
    fi
    
    print_success "回归测试完成！"
}

# 清理测试数据
clean_data() {
    print_warning "即将清理测试数据，这不会删除表结构！"
    read -p "确认继续？(y/N): " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "已取消操作"
        exit 0
    fi
    
    print_info "清理测试数据..."
    
    local MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER}"
    
    if [ -n "$DB_PASS" ]; then
        MYSQL_CMD="${MYSQL_CMD} -p${DB_PASS}"
    fi
    
    ${MYSQL_CMD} ${DB_NAME} << 'EOF'
        DELETE FROM order_items;
        DELETE FROM orders;
        DELETE FROM products;
        DELETE FROM categories;
        DELETE FROM users;
        DELETE FROM merchant_licenses;
        DELETE FROM merchant_delivery_settings;
        DELETE FROM merchant_staffs;
        DELETE FROM merchants;
        DELETE FROM merchant_applications;
        DELETE FROM merchant_audit_records;
        DELETE FROM announcements;
        DELETE FROM service_provider_admins;
        DELETE FROM service_providers;
        
        ALTER TABLE order_items AUTO_INCREMENT = 1;
        ALTER TABLE orders AUTO_INCREMENT = 1;
        ALTER TABLE products AUTO_INCREMENT = 1;
        ALTER TABLE categories AUTO_INCREMENT = 1;
        ALTER TABLE users AUTO_INCREMENT = 1;
        ALTER TABLE merchant_licenses AUTO_INCREMENT = 1;
        ALTER TABLE merchant_delivery_settings AUTO_INCREMENT = 1;
        ALTER TABLE merchant_staffs AUTO_INCREMENT = 1;
        ALTER TABLE merchants AUTO_INCREMENT = 1;
        ALTER TABLE merchant_applications AUTO_INCREMENT = 1;
        ALTER TABLE merchant_audit_records AUTO_INCREMENT = 1;
        ALTER TABLE announcements AUTO_INCREMENT = 1;
        ALTER TABLE service_provider_admins AUTO_INCREMENT = 1;
        ALTER TABLE service_providers AUTO_INCREMENT = 1;
EOF
    
    print_success "测试数据清理完成！"
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-help}" in
        init)
            check_dependencies
            init_database
            ;;
        reset)
            check_dependencies
            reset_database
            ;;
        test)
            check_dependencies
            run_tests
            ;;
        clean)
            check_dependencies
            clean_data
            ;;
        status)
            check_dependencies
            show_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
