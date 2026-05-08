/**
 * Pinia Store - 商家状态管理
 */

import { defineStore } from 'pinia'
import type { Category, Product, ProductListResponse } from '../types/index'

interface MerchantState {
  categories: Category[]
  products: Product[]
  currentCategory: number | null
  productLoading: boolean
  categoryLoading: boolean
}

export const useMerchantStore = defineStore('merchant', {
  state: (): MerchantState => ({
    categories: [],
    products: [],
    currentCategory: null,
    productLoading: false,
    categoryLoading: false
  }),

  getters: {
    activeCategories: (state) => state.categories.filter(c => c.status === 1),
    onSaleProducts: (state) => state.products.filter(p => p.status === 1),
    offSaleProducts: (state) => state.products.filter(p => p.status === 0)
  },

  actions: {
    // 设置当前分类
    setCurrentCategory(categoryId: number | null) {
      this.currentCategory = categoryId
    },

    // 设置分类列表
    setCategories(categories: Category[]) {
      this.categories = categories
    },

    // 添加分类
    addCategory(category: Category) {
      this.categories.push(category)
    },

    // 更新分类
    updateCategory(categoryId: number, data: Partial<Category>) {
      const index = this.categories.findIndex(c => c.id === categoryId)
      if (index !== -1) {
        this.categories[index] = { ...this.categories[index], ...data }
      }
    },

    // 删除分类
    removeCategory(categoryId: number) {
      this.categories = this.categories.filter(c => c.id !== categoryId)
    },

    // 设置商品列表
    setProducts(products: Product[]) {
      this.products = products
    },

    // 添加商品
    addProduct(product: Product) {
      this.products.unshift(product)
    },

    // 更新商品
    updateProduct(productId: number, data: Partial<Product>) {
      const index = this.products.findIndex(p => p.id === productId)
      if (index !== -1) {
        this.products[index] = { ...this.products[index], ...data }
      }
    },

    // 删除商品
    removeProduct(productId: number) {
      this.products = this.products.filter(p => p.id !== productId)
    },

    // 清空数据
    clearAll() {
      this.categories = []
      this.products = []
      this.currentCategory = null
    }
  }
})
