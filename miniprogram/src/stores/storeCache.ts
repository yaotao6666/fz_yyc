import { defineStore } from 'pinia'
import type { StoreHomeInfo, Product } from '../types'

interface StoreCacheState {
  storeInfo: StoreHomeInfo | null
  merchantId: number | null
  categoryProducts: Record<number, Product[]>
  currentCategoryIndex: number
}

export const useStoreCacheStore = defineStore('storeCache', {
  state: (): StoreCacheState => ({
    storeInfo: null,
    merchantId: null,
    categoryProducts: {},
    currentCategoryIndex: 0
  }),

  actions: {
    setStoreInfo(merchantId: number, info: StoreHomeInfo) {
      this.merchantId = merchantId
      this.storeInfo = info
    },

    setCategoryProducts(categoryId: number, products: Product[]) {
      this.categoryProducts[categoryId] = products
    },

    setCurrentCategoryIndex(index: number) {
      this.currentCategoryIndex = index
    },

    hasCache(merchantId: number): boolean {
      return this.merchantId === merchantId && this.storeInfo !== null
    },

    clearCache() {
      this.storeInfo = null
      this.merchantId = null
      this.categoryProducts = {}
      this.currentCategoryIndex = 0
    }
  }
})
