/**
 * Pinia Store - 购物车状态管理
 */

import { defineStore } from 'pinia'

export interface CartItem {
  merchant_id?: number
  merchant_name?: string
  product_id: number
  product_name: string
  image: string
  price: number
  quantity: number
  specs?: string
  max_stock?: number
}

interface CartState {
  items: CartItem[]
  merchantId: number | null
  merchantName: string
}

export const useCartStore = defineStore('cart', {
  state: (): CartState => ({
    items: uni.getStorageSync('cartItems') || [],
    merchantId: uni.getStorageSync('cartMerchantId') || null,
    merchantName: uni.getStorageSync('cartMerchantName') || ''
  }),

  getters: {
    totalAmount(): number {
      return this.items.reduce((sum, item) => sum + item.price * item.quantity, 0)
    },
    
    totalCount(): number {
      return this.items.reduce((sum, item) => sum + item.quantity, 0)
    },

    isEmpty(): boolean {
      return this.items.length === 0
    }
  },

  actions: {
    // 添加商品到购物车
    addItem(item: CartItem) {
      const incomingMerchantId = item.merchant_id ?? null
      const incomingMerchantName = item.merchant_name ?? ''

      if (incomingMerchantId && this.merchantId && this.merchantId !== incomingMerchantId) {
        this.clearCart()
      }

      if (incomingMerchantId) {
        this.merchantId = incomingMerchantId
      }
      if (incomingMerchantName) {
        this.merchantName = incomingMerchantName
      }

      const existIndex = this.items.findIndex(
        i => i.product_id === item.product_id && i.specs === item.specs
      )

      if (existIndex !== -1) {
        this.items[existIndex].quantity += item.quantity
      } else {
        this.items.push({ ...item })
      }

      this.saveToStorage()
    },

    // 更新商品数量
    updateQuantity(productId: number, specs: string | undefined, quantity: number) {
      const index = this.items.findIndex(
        i => i.product_id === productId && i.specs === specs
      )

      if (index !== -1) {
        if (quantity <= 0) {
          this.items.splice(index, 1)
        } else {
          this.items[index].quantity = quantity
        }
      }

      if (this.items.length === 0) {
        this.clearCart()
      } else {
        this.saveToStorage()
      }
    },

    // 删除商品
    removeItem(productId: number, specs?: string) {
      this.items = this.items.filter(
        i => !(i.product_id === productId && i.specs === specs)
      )

      if (this.items.length === 0) {
        this.clearCart()
      } else {
        this.saveToStorage()
      }
    },

    // 清空购物车
    clearCart() {
      this.items = []
      this.merchantId = null
      this.merchantName = ''
      
      uni.removeStorageSync('cartItems')
      uni.removeStorageSync('cartMerchantId')
      uni.removeStorageSync('cartMerchantName')
    },

    // 保存到本地存储
    saveToStorage() {
      uni.setStorageSync('cartItems', this.items)
      uni.setStorageSync('cartMerchantId', this.merchantId)
      uni.setStorageSync('cartMerchantName', this.merchantName)
    },

    // 从存储恢复
    restoreFromStorage() {
      this.items = uni.getStorageSync('cartItems') || []
      this.merchantId = uni.getStorageSync('cartMerchantId') || null
      this.merchantName = uni.getStorageSync('cartMerchantName') || ''
    }
  }
})
