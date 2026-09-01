<script setup lang="ts">
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useTabsStore, type TabItem } from '@/stores/tabs'

const route = useRoute()
const router = useRouter()
const tabsStore = useTabsStore()

// 固定页签：工作台（始终存在、不可关闭）
const AFFIX_TAB: TabItem = {
  path: '/dashboard',
  fullPath: '/dashboard',
  name: 'dashboard',
  title: '工作台',
  query: {},
  params: {},
  affix: true
}

// 初始化固定页签
if (!tabsStore.tabs.some((t) => t.affix)) {
  tabsStore.tabs.unshift(AFFIX_TAB)
}

// 路由变化时登记当前模块页签
watch(
  () => route.fullPath,
  () => tabsStore.addTab(route),
  { immediate: true }
)

/** 点击页签快速切换回该模块 */
function goTab(tab: TabItem) {
  if (tab.fullPath === route.fullPath) return
  router.push({ path: tab.path, query: tab.query })
}

/** 关闭页签：若关闭的是当前页签，则聚焦到相邻右侧页签，否则左侧 */
async function closeTab(tab: TabItem) {
  if (tab.affix) return
  const index = tabsStore.tabs.findIndex((t) => t.fullPath === tab.fullPath)
  tabsStore.removeTab(tab)
  if (tab.fullPath === route.fullPath) {
    const next = tabsStore.tabs[index - 1] || tabsStore.tabs[tabsStore.tabs.length - 1]
    if (next) router.push({ path: next.path, query: next.query })
    else router.push('/dashboard')
  }
}

/** 关闭其他 */
async function onCloseOthers() {
  const current = tabsStore.tabs.find((t) => t.fullPath === route.fullPath)
  if (!current) return
  tabsStore.closeOthers(current)
}

/** 关闭全部 */
async function onCloseAll() {
  await ElMessageBox.confirm('确认关闭全部页签？工作台将保留。', '关闭全部', { type: 'warning' })
  tabsStore.closeAll()
  if (route.fullPath !== '/dashboard') {
    router.push({ path: '/dashboard' })
  }
}
</script>

<template>
  <div class="tabs-bar">
    <div class="tabs-scroll">
      <div
        v-for="tab in tabsStore.tabs"
        :key="tab.fullPath"
        class="tab-item"
        :class="{ active: route.fullPath === tab.fullPath }"
        @click="goTab(tab)"
      >
        <span class="tab-title">{{ tab.title }}</span>
        <span v-if="!tab.affix" class="tab-close" @click.stop="closeTab(tab)">✕</span>
      </div>
    </div>
    <el-dropdown trigger="click" class="tabs-tools">
      <span class="tabs-tools-btn">操作</span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="onCloseOthers">关闭其他</el-dropdown-item>
          <el-dropdown-item divided @click="onCloseAll">关闭全部</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped>
.tabs-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 40px;
  padding: 0 16px;
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  box-sizing: border-box;
}
.tabs-scroll {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  overflow-x: auto;
  white-space: nowrap;
}
.tabs-scroll::-webkit-scrollbar { height: 4px; }
.tabs-scroll::-webkit-scrollbar-thumb { background: #d1d5db; border-radius: 4px; }
.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 13px;
  color: #4b5563;
  background: #f3f4f6;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}
.tab-item:hover { color: #111827; background: #eef2ff; }
.tab-item.active {
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  border-color: rgba(0, 122, 255, 0.35);
  font-weight: 600;
}
.tab-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  font-size: 10px;
  line-height: 1;
  border-radius: 50%;
  color: #9ca3af;
}
.tab-close:hover { background: #d1d5db; color: #111827; }
.tabs-tools-btn {
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
  white-space: nowrap;
}
.tabs-tools-btn:hover { color: #007AFF; }
</style>