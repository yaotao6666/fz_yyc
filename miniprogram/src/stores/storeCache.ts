import { defineStore } from 'pinia'
import type { StoreHomeInfo, Product } from '../types'

interface StoreCacheState {
  storeInfo: StoreHomeInfo | null
  categoryProducts: Record<number, Product[]>
  currentCategoryIndex: number
}

export const useStoreCacheStore = defineStore('storeCache', {
  state: (): StoreCacheState => ({
    storeInfo: null,
    categoryProducts: {},
    currentCategoryIndex: 0
  }),

  actions: {
    setStoreInfo(info: StoreHomeInfo) {
      this.storeInfo = info
    },

    setCategoryProducts(categoryId: number, products: Product[]) {
      this.categoryProducts[categoryId] = products
    },

    setCurrentCategoryIndex(index: number) {
      this.currentCategoryIndex = index
    },

    hasCache(): boolean {
      return this.storeInfo !== null
    },

    clearCache() {
      this.storeInfo = null
      this.categoryProducts = {}
      this.currentCategoryIndex = 0
    }
  }
})
