/**
 * Pinia Store - 购物车状态管理
 */

import { defineStore } from 'pinia'
import type { ProductType } from '../types'

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
  product_type?: ProductType | number
  sale_type?: number
  rental_unit?: number
  rental_price?: number
  deposit?: number
  rental_duration?: number
}

interface CartState {
  items: CartItem[]
  merchantId: number | null
  merchantName: string
  buyNowItem: CartItem | null
}

export const useCartStore = defineStore('cart', {
  state: (): CartState => ({
    items: uni.getStorageSync('cartItems') || [],
    merchantId: uni.getStorageSync('cartMerchantId') || null,
    merchantName: uni.getStorageSync('cartMerchantName') || '',
    buyNowItem: uni.getStorageSync('cartBuyNowItem') || null
  }),

  getters: {
    totalAmount(): number {
      return this.items.reduce((sum, item) => sum + getItemPayAmount(item), 0)
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
        i => i.product_id === item.product_id
          && i.specs === item.specs
          && i.rental_duration === item.rental_duration
      )

      if (existIndex !== -1) {
        this.items[existIndex].quantity += item.quantity
      } else {
        this.items.push({ ...item })
      }

      this.saveToStorage()
    },

    setBuyNowItem(item: CartItem | null) {
      this.buyNowItem = item
      if (item) {
        uni.setStorageSync('cartBuyNowItem', item)
      } else {
        uni.removeStorageSync('cartBuyNowItem')
      }
    },

    clearBuyNowItem() {
      this.buyNowItem = null
      uni.removeStorageSync('cartBuyNowItem')
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
      this.buyNowItem = uni.getStorageSync('cartBuyNowItem') || null
    }
  }
})

// 单个购物车项的应付金额（一口价=单价×数量；租赁=租金×时长×数量 + 押金×数量）
export function getItemPayAmount(item: CartItem): number {
  if (Number(item.sale_type) === 2) {
    const duration = Number(item.rental_duration || 0)
    const rentalPrice = Number(item.rental_price || 0)
    const deposit = Number(item.deposit || 0)
    const quantity = Number(item.quantity || 1)
    return (rentalPrice * duration + deposit) * quantity
  }
  return Number(item.price || 0) * Number(item.quantity || 1)
}

// 单个购物车项的租金小计（不含押金）
export function getItemRentalSubtotal(item: CartItem): number {
  if (Number(item.sale_type) !== 2) return 0
  const duration = Number(item.rental_duration || 0)
  const rentalPrice = Number(item.rental_price || 0)
  const quantity = Number(item.quantity || 1)
  return rentalPrice * duration * quantity
}

// 单个购物车项的押金小计
export function getItemDeposit(item: CartItem): number {
  if (Number(item.sale_type) !== 2) return 0
  return Number(item.deposit || 0) * Number(item.quantity || 1)
}

export function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}
