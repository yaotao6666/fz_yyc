import { defineStore } from 'pinia'

export interface TabItem {
  path: string
  fullPath: string
  name?: string
  title: string
  query: Record<string, any>
  params: Record<string, any>
  affix?: boolean
}

interface RouteLike {
  path: string
  fullPath: string
  name?: unknown
  meta?: { title?: unknown }
  query?: Record<string, any>
  params?: Record<string, any>
}

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    tabs: [] as TabItem[]
  }),
  actions: {
    /** 路由切换时登记访问的模块页签；已存在或无标题（重定向/详情变更）则跳过 */
    addTab(route: RouteLike) {
      const title = route.meta?.title
      if (!title) return
      if (this.tabs.some((t) => t.fullPath === route.fullPath)) return
      this.tabs.push({
        path: route.path,
        fullPath: route.fullPath,
        name: typeof route.name === 'string' ? route.name : undefined,
        title: String(title),
        query: { ...(route.query || {}) },
        params: { ...(route.params || {}) }
      })
    },
    /** 关闭指定页签（固定页签不可关） */
    removeTab(tab: TabItem) {
      const i = this.tabs.findIndex((t) => t.fullPath === tab.fullPath)
      if (i === -1 || tab.affix) return
      this.tabs.splice(i, 1)
    },
    /** 关闭其他页签（保留固定页签与当前页签） */
    closeOthers(tab: TabItem) {
      this.tabs = this.tabs.filter((t) => t.affix || t.fullPath === tab.fullPath)
    },
    /** 关闭全部页签（保留固定页签） */
    closeAll() {
      this.tabs = this.tabs.filter((t) => t.affix)
    }
  }
})