package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 连接数据库
	dsn := "root:root@tcp(127.0.0.1:3306)/fz_yyc_api?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("无法连接到数据库:", err)
	}

	fmt.Println("✓ 数据库连接成功")

	// 验证服务商管理员密码
	var adminPwd string
	err = db.QueryRow("SELECT password FROM service_provider_admins WHERE username = ?", "admin").Scan(&adminPwd)
	if err != nil {
		log.Fatal("查询服务商管理员失败:", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(adminPwd), []byte("admin123"))
	if err != nil {
		fmt.Println("✗ 服务商管理员密码验证失败: admin123")
	} else {
		fmt.Println("✓ 服务商管理员密码验证成功: admin123")
	}

	// 验证商家员工密码
	var merchantPwd string
	err = db.QueryRow("SELECT password FROM merchant_staffs WHERE username = ?", "merchant").Scan(&merchantPwd)
	if err != nil {
		log.Fatal("查询商家员工失败:", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(merchantPwd), []byte("merchant123"))
	if err != nil {
		fmt.Println("✗ 商家员工密码验证失败: merchant123")
	} else {
		fmt.Println("✓ 商家员工密码验证成功: merchant123")
	}

	// 检查数据完整性
	fmt.Println("\n数据完整性检查:")

	// 检查服务商
	var spCount int
	db.QueryRow("SELECT COUNT(*) FROM service_providers").Scan(&spCount)
	fmt.Printf("✓ 服务商数量: %d\n", spCount)

	// 检查服务商管理员
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM service_provider_admins").Scan(&adminCount)
	fmt.Printf("✓ 服务商管理员数量: %d\n", adminCount)

	// 检查商家
	var merchantCount int
	db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&merchantCount)
	fmt.Printf("✓ 商家数量: %d\n", merchantCount)

	// 检查商家员工
	var staffCount int
	db.QueryRow("SELECT COUNT(*) FROM merchant_staffs").Scan(&staffCount)
	fmt.Printf("✓ 商家员工数量: %d\n", staffCount)

	// 检查商品分类
	var categoryCount int
	db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	fmt.Printf("✓ 商品分类数量: %d\n", categoryCount)

	// 检查商品
	var productCount int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	fmt.Printf("✓ 商品数量: %d\n", productCount)

	// 检查用户
	var userCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	fmt.Printf("✓ C端用户数量: %d\n", userCount)

	// 检查订单
	var orderCount int
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&orderCount)
	fmt.Printf("✓ 订单数量: %d\n", orderCount)

	fmt.Println("\n初始化数据验证完成！")
}
