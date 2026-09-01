<script setup lang="ts">
import { computed, ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { getMenuTree, createMenu, updateMenu, deleteMenu } from '@/api/sp'
import type { SysMenu } from '@/types/sp'

const loading = ref(false)
const tree = ref<SysMenu[]>([])

// 可选图标库：Element Plus 全部图标名（与侧边栏动态渲染组件名一致）
const iconNames = computed<string[]>(() =>
  Object.keys(ElementPlusIconsVue).filter((key) => !key.startsWith('_'))
)
const iconOf = (name?: string) => {
  if (!name) return null
  return (ElementPlusIconsVue as Record<string, unknown>)[name] || null
}

async function loadTree() {
  loading.value = true
  try {
    tree.value = await getMenuTree()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

const menuTypeText = (t: number) => (t === 2 ? '按钮' : '菜单/目录')
const statusText = (s: number) => (s === 1 ? '启用' : '禁用')
const visibleText = (v: number) => (v === 1 ? '显示' : '隐藏')

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const saving = ref(false)

const emptyForm = () => ({
  id: 0,
  parent_id: 0,
  menu_type: 1,
  name: '',
  path: '',
  icon: '',
  sort: 0,
  status: 1,
  visible: 1,
  permission: ''
})
const form = reactive<SysMenu>(emptyForm())

function openCreate(parent?: SysMenu) {
  Object.assign(form, emptyForm())
  if (parent) {
    form.parent_id = parent.menu_type === 2 ? parent.parent_id : parent.id
    form.menu_type = parent.menu_type === 2 ? 2 : 1
  }
  dialogMode.value = 'create'
  dialogVisible.value = true
}

function openCreateButton(parent?: SysMenu) {
  Object.assign(form, emptyForm())
  form.parent_id = parent ? parent.id : 0
  form.menu_type = 2
  dialogMode.value = 'create'
  dialogVisible.value = true
}

function openUpdate(row: SysMenu) {
  Object.assign(form, {
    id: row.id,
    parent_id: row.parent_id,
    menu_type: row.menu_type,
    name: row.name,
    path: row.path || '',
    icon: row.icon || '',
    sort: row.sort,
    status: row.status,
    visible: row.visible,
    permission: row.permission || ''
  })
  dialogMode.value = 'update'
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      parent_id: form.parent_id,
      menu_type: form.menu_type,
      name: form.name.trim(),
      path: (form.path || '').trim() || undefined,
      icon: (form.icon || '').trim() || undefined,
      sort: form.sort || 0,
      status: form.status,
      visible: form.visible,
      permission: (form.permission || '').trim() || undefined
    }
    if (dialogMode.value === 'create') {
      await createMenu(payload)
      ElMessage.success('创建成功')
    } else {
      await updateMenu(form.id, payload)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadTree()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: SysMenu) {
  try {
    await ElMessageBox.confirm(`确认删除菜单「${row.name}」？存在子菜单或被角色引用时将被拒绝。`, '删除菜单', {
      type: 'warning'
    })
    await deleteMenu(row.id)
    ElMessage.success('删除成功')
    loadTree()
  } catch (_e) {
    // 取消
  }
}

onMounted(loadTree)
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div>
        <el-button type="primary" v-permission="'system:menu:create'" @click="openCreate()">新增菜单</el-button>
        <el-button v-permission="'system:menu:create'" @click="openCreateButton()">新增按钮</el-button>
      </div>
      <el-button :loading="loading" @click="loadTree">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="tree" row-key="id" stripe :tree-props="{ children: 'children' }" default-expand-all>
      <el-table-column prop="name" label="菜单名称" min-width="160" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="row.menu_type === 2 ? 'warning' : 'primary'">{{ menuTypeText(row.menu_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="图标" width="120">
        <template #default="{ row }">
          <span v-if="row.icon" class="icon-cell">
            <el-icon v-if="iconOf(row.icon)"><component :is="iconOf(row.icon)" /></el-icon>
            <span>{{ row.icon }}</span>
          </span>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="路径" min-width="120" />
      <el-table-column prop="permission" label="权限标识" min-width="150" />
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="显示" width="80">
        <template #default="{ row }">
          <span>{{ visibleText(row.visible) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-permission="'system:menu:create'" @click="openCreate(row)">新增子级</el-button>
          <el-button size="small" v-permission="'system:menu:create'" @click="openUpdate(row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'system:menu:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增菜单' : '编辑菜单'" width="520px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select
            v-model="form.parent_id"
            :data="tree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            value-key="id"
            check-strictly
            clearable
            placeholder="不选则为顶级菜单"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.menu_type">
            <el-radio :value="1">菜单/目录</el-radio>
            <el-radio :value="2">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如：商品管理" />
        </el-form-item>
        <el-form-item label="路由路径">
          <el-input v-model="form.path" placeholder="菜单必填，如 /products；按钮留空" />
        </el-form-item>
        <el-form-item label="图标">
          <el-select
            v-model="form.icon"
            filterable
            clearable
            placeholder="选择图标（可输入搜索）"
            style="width: 100%"
          >
            <template #prefix>
              <el-icon v-if="iconOf(form.icon)" class="icon-prefix">
                <component :is="iconOf(form.icon)" />
              </el-icon>
            </template>
            <el-option v-for="name in iconNames" :key="name" :label="name" :value="name">
              <span class="icon-option">
                <el-icon v-if="iconOf(name)"><component :is="iconOf(name)" /></el-icon>
                <span>{{ name }}</span>
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="权限标识">
          <el-input v-model="form.permission" placeholder="如 products:view" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否显示">
          <el-radio-group v-model="form.visible">
            <el-radio :value="1">显示</el-radio>
            <el-radio :value="0">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.icon-prefix { margin-right: 4px; font-size: 15px; }
.icon-option { display: inline-flex; align-items: center; gap: 8px; }
.icon-cell { display: inline-flex; align-items: center; gap: 6px; }
</style>
