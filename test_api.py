import requests
import json

BASE_URL = "http://localhost:8080/api/v1"

def test_api():
    results = []

    # 1. 测试服务商管理员登录
    print("=" * 60)
    print("测试1: 服务商管理员登录")
    print("=" * 60)
    try:
        resp = requests.post(f"{BASE_URL}/auth/admin/login", json={
            "username": "admin",
            "password": "admin123"
        })
        data = resp.json()
        print(f"状态码: {resp.status_code}")
        print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
        admin_token = data.get("data", {}).get("token", "")
        results.append(("服务商登录", resp.status_code == 200, data.get("code") == 0))
    except Exception as e:
        print(f"请求失败: {e}")
        results.append(("服务商登录", False, False))
        admin_token = ""

    if not admin_token:
        print("无法获取token，跳过后续测试")
        return results

    # 2. 测试商家管理员登录 - 验证merchant_id是否正确
    print("\n" + "=" * 60)
    print("测试2: 商家管理员登录 (验证merchant_id修复)")
    print("=" * 60)
    try:
        resp = requests.post(f"{BASE_URL}/auth/merchant/login", json={
            "username": "merchant1",
            "password": "123456"
        })
        data = resp.json()
        print(f"状态码: {resp.status_code}")
        print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")

        # 验证merchant_id
        merchant_id = data.get("data", {}).get("merchant_id")
        print(f"\n>>> merchant_id 值: {merchant_id}")
        if merchant_id and merchant_id > 0:
            print("✅ 修复成功: merchant_id 不再是 0")
        else:
            print("❌ 修复失败: merchant_id 仍然是 0 或无效")

        merchant_token = data.get("data", {}).get("token", "")
        results.append(("商家登录merchant_id", True, merchant_id > 0 if merchant_id else False))
    except Exception as e:
        print(f"请求失败: {e}")
        results.append(("商家登录merchant_id", False, False))
        merchant_token = ""

    # 3. 测试获取商家信息
    if merchant_token:
        print("\n" + "=" * 60)
        print("测试3: 获取商家信息")
        print("=" * 60)
        try:
            resp = requests.get(f"{BASE_URL}/merchant/profile", headers={
                "Authorization": f"Bearer {merchant_token}"
            })
            data = resp.json()
            print(f"状态码: {resp.status_code}")
            print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
            results.append(("获取商家信息", resp.status_code == 200, data.get("code") == 0))
        except Exception as e:
            print(f"请求失败: {e}")
            results.append(("获取商家信息", False, False))

    # 4. 测试创建商品（带规格选项价格）
    if merchant_token:
        print("\n" + "=" * 60)
        print("测试4: 创建商品带规格选项价格 (验证规格价格修复)")
        print("=" * 60)
        try:
            resp = requests.post(f"{BASE_URL}/merchant/products", headers={
                "Authorization": f"Bearer {merchant_token}"
            }, json={
                "name": "测试套餐",
                "price": 50.0,
                "stock": 100,
                "unit": "份",
                "description": "带规格价格的测试商品",
                "specs": [
                    {
                        "name": "规格",
                        "options": [
                            {"name": "小份", "price": 48.0},
                            {"name": "大份", "price": 68.0}
                        ]
                    }
                ]
            })
            data = resp.json()
            print(f"状态码: {resp.status_code}")
            print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")

            # 验证规格选项是否包含价格
            specs = data.get("data", {}).get("specs", [])
            if specs and len(specs) > 0:
                options = specs[0].get("options", [])
                if options and len(options) > 0:
                    first_option = options[0]
                    if "price" in first_option:
                        print(f"\n>>> 规格选项包含价格: {first_option}")
                        print("✅ 修复成功: 规格选项现在包含价格字段")
                        results.append(("规格选项价格", True, True))
                    else:
                        print("❌ 修复失败: 规格选项没有价格字段")
                        results.append(("规格选项价格", True, False))
                else:
                    results.append(("规格选项价格", True, False))
            else:
                results.append(("规格选项价格", True, False))

            product_id = data.get("data", {}).get("id")
        except Exception as e:
            print(f"请求失败: {e}")
            results.append(("规格选项价格", False, False))
            product_id = None

    # 5. 测试创建分类（sort字段）
    if merchant_token:
        print("\n" + "=" * 60)
        print("测试5: 创建分类 (验证sort字段)")
        print("=" * 60)
        try:
            resp = requests.post(f"{BASE_URL}/merchant/categories", headers={
                "Authorization": f"Bearer {merchant_token}"
            }, json={
                "name": "测试分类",
                "sort": 10
            })
            data = resp.json()
            print(f"状态码: {resp.status_code}")
            print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")

            sort_value = data.get("data", {}).get("sort")
            if sort_value == 10:
                print(f"\n>>> sort 值为: {sort_value}")
                print("✅ sort 字段设置成功")
                results.append(("分类sort字段", True, True))
            else:
                print("❌ sort 字段设置失败")
                results.append(("分类sort字段", True, False))

            category_id = data.get("data", {}).get("id")
        except Exception as e:
            print(f"请求失败: {e}")
            results.append(("分类sort字段", False, False))
            category_id = None

    # 6. 测试更新分类sort为0
    if merchant_token and category_id:
        print("\n" + "=" * 60)
        print("测试6: 更新分类sort为0 (验证sort置零修复)")
        print("=" * 60)
        try:
            resp = requests.put(f"{BASE_URL}/merchant/categories/{category_id}", headers={
                "Authorization": f"Bearer {merchant_token}"
            }, json={
                "name": "测试分类",
                "sort": 0
            })
            data = resp.json()
            print(f"状态码: {resp.status_code}")
            print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")

            sort_value = data.get("data", {}).get("sort")
            if sort_value == 0:
                print(f"\n>>> sort 值为: {sort_value}")
                print("✅ 修复成功: sort 字段可以设置为 0")
                results.append(("sort置零", True, True))
            else:
                print(f"❌ 修复失败: sort 值为 {sort_value}，不是 0")
                results.append(("sort置零", True, False))
        except Exception as e:
            print(f"请求失败: {e}")
            results.append(("sort置零", False, False))

    # 7. 测试C端登录
    print("\n" + "=" * 60)
    print("测试7: C端用户登录")
    print("=" * 60)
    try:
        resp = requests.post(f"{BASE_URL}/auth/user/login", json={
            "code": "test_code_123"
        })
        data = resp.json()
        print(f"状态码: {resp.status_code}")
        print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")

        user_token = data.get("data", {}).get("token", "")
        user_id = data.get("data", {}).get("user_id")
        results.append(("C端登录", resp.status_code == 200, data.get("code") == 0 and user_token != ""))
    except Exception as e:
        print(f"请求失败: {e}")
        results.append(("C端登录", False, False))
        user_token = ""

    # 8. 测试创建订单（库存扣减和销量更新）
    if user_token and product_id:
        print("\n" + "=" * 60)
        print("测试8: 创建订单 (验证库存扣减和销量更新)")
        print("=" * 60)

        # 先获取商品当前库存和销量
        try:
            resp = requests.get(f"{BASE_URL}/store/1/products/{product_id}")
            before_data = resp.json()
            before_stock = before_data.get("data", {}).get("stock", 0)
            before_sales = before_data.get("data", {}).get("sales", 0)
            print(f">>> 创建订单前 - stock: {before_stock}, sales: {before_sales}")

            # 创建订单
            resp = requests.post(f"{BASE_URL}/user/orders", headers={
                "Authorization": f"Bearer {user_token}"
            }, json={
                "merchant_id": 1,
                "delivery_type": 1,
                "contact_name": "测试用户",
                "contact_phone": "13800138000",
                "delivery_address": "测试地址",
                "items": [
                    {"product_id": product_id, "quantity": 2}
                ]
            })
            order_data = resp.json()
            print(f"创建订单响应: {json.dumps(order_data, ensure_ascii=False, indent=2)}")

            if order_data.get("code") == 0:
                # 再次获取商品信息
                resp = requests.get(f"{BASE_URL}/store/1/products/{product_id}")
                after_data = resp.json()
                after_stock = after_data.get("data", {}).get("stock", 0)
                after_sales = after_data.get("data", {}).get("sales", 0)
                print(f">>> 创建订单后 - stock: {after_stock}, sales: {after_sales}")

                stock_diff = before_stock - after_stock
                sales_diff = after_sales - before_sales

                if stock_diff == 2:
                    print(f"✅ 库存扣减成功: 扣减了 {stock_diff}")
                    results.append(("库存扣减", True, True))
                else:
                    print(f"❌ 库存扣减失败: 预期减少2，实际减少 {stock_diff}")
                    results.append(("库存扣减", True, False))

                if sales_diff == 2:
                    print(f"✅ 销量更新成功: 增加了 {sales_diff}")
                    results.append(("销量更新", True, True))
                else:
                    print(f"❌ 销量更新失败: 预期增加2，实际增加 {sales_diff}")
                    results.append(("销量更新", True, False))
            else:
                print(f"❌ 订单创建失败: {order_data.get('message')}")
                results.append(("库存扣减", False, False))
                results.append(("销量更新", False, False))

        except Exception as e:
            print(f"请求失败: {e}")
            results.append(("库存扣减", False, False))
            results.append(("销量更新", False, False))

    # 9. 获取服务商配置
    print("\n" + "=" * 60)
    print("测试9: 获取服务商配置")
    print("=" * 60)
    try:
        resp = requests.get(f"{BASE_URL}/admin/service-provider", headers={
            "Authorization": f"Bearer {admin_token}"
        })
        data = resp.json()
        print(f"状态码: {resp.status_code}")
        print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
        results.append(("获取服务商配置", resp.status_code == 200, data.get("code") == 0))
    except Exception as e:
        print(f"请求失败: {e}")
        results.append(("获取服务商配置", False, False))

    # 汇总结果
    print("\n" + "=" * 60)
    print("测试结果汇总")
    print("=" * 60)
    for name, status, success in results:
        status_str = "✅ PASS" if success else "❌ FAIL"
        print(f"{status_str} - {name}")

    passed = sum(1 for _, _, s in results if s)
    total = len(results)
    print(f"\n通过: {passed}/{total}")

    return results

if __name__ == "__main__":
    test_api()
