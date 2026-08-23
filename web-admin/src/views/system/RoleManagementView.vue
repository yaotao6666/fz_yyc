<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoleList, createRole, updateRole, deleteRole, getMenuTree, getRoleMenus, assignRoleMenus } from '@/api/sp'
import type { SysRole, SysMenu } from '@/types/sp'

const loading = ref(false)
const list = ref<SysRole[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const statusFilter = ref<number | string>('')

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    if (statusFilter.value !== '') params.status = statusFilter.value
    const res = await getRoleList(params)
    list.value = res.list || []
    total.value = res.pagination?.total || 0
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

const statusText = (s: number) => (s === 1 ? '启用' : '禁用')
const statusType = (s: number): 'success' | 'info' => (s === 1 ? 'success' : 'info')

/* ----- 角色弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const saving = ref(false)
const form = reactive({
  id: 0,
  name: '',
  code: '',
  remark: '',
  status: 1
})

function openCreate() {
  Object.assign(form, { id: 0, name: '', code: '', remark: '', status: 1 })
  dialogMode.value = 'create'
  dialogVisible.value = true
}

function openUpdate(row: SysRole) {
  Object.assign(form, { id: row.id, name: row.name, code: row.code, remark: row.remark || '', status: row.status })
  dialogMode.value = 'update'
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name.trim() || !form.code.trim()) {
    ElMessage.warning('请填写角色名称和编码')
    return
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createRole({ name: form.name.trim(), code: form.code.trim(), remark: form.remark.trim(), status: form.status })
      ElMessage.success('创建成功')
    } else {
      await updateRole(form.id, { name: form.name.trim(), code: form.code.trim(), remark: form.remark.trim(), status: form.status })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: SysRole) {
  try {
    await ElMessageBox.confirm(`确认删除角色「${row.name}」？该角色下员工将被保留。`, '删除角色', { type: 'warning' })
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消
  }
}

/* ----- 权限分配弹窗 ----- */
const assignDialogVisible = ref(false)
const assignRoleId = ref(0)
const assignRoleName = ref('')
const menuTree = ref<SysMenu[]>([])
const checkedMenuIds = ref<number[]>([])
const assignLoading = ref(false)
const assignSaving = ref(false)

async function openAssign(row: SysRole) {
  assignRoleId.value = row.id
  assignRoleName.value = row.name
  assignLoading.value = true
  assignDialogVisible.value = true
  try {
    const [menus, roleMenus] = await Promise.all([getMenuTree(), getRoleMenus(row.id)])
    menuTree.value = menus
    checkedMenuIds.value = roleMenus.menu_ids || []
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    assignLoading.value = false
  }
}

// 收集所有菜单ID（含按钮），用于权限分配
function collectAllMenuIds(nodes: SysMenu[]): number[] {
  const ids: number[] = []
  for (const n of nodes) {
    ids.push(n.id)
    if (n.children && n.children.length) {
      ids.push(...collectAllMenuIds(n.children))
    }
  }
  return ids
}

function handleSelectAll() {
  const allIds = collectAllMenuIds(menuTree.value)
  checkedMenuIds.value = allIds
}

function handleClearAll() {
  checkedMenuIds.value = []
}

async function handleAssignSave() {
  assignSaving.value = true
  try {
    await assignRoleMenus(assignRoleId.value, { menu_ids: checkedMenuIds.value })
    ElMessage.success('分配成功')
    assignDialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    assignSaving.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <el-input v-model="keyword" placeholder="角色名称/编码" style="width: 200px" clearable @clear="() => { page = 1; loadData() }" @keyup.enter="() => { page = 1; loadData() }" />
        <el-select v-model="statusFilter" placeholder="状态" style="width: 110px" @change="() => { page = 1; loadData() }">
          <el-option label="全部" value="" />
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="() => { page = 1; loadData() }">搜索</el-button>
      </div>
      <div>
        <el-button type="primary" v-permission="'system:role:create'" @click="openCreate">新增角色</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="角色名称" width="160" />
      <el-table-column prop="code" label="角色编码" width="140" />
      <el-table-column prop="menu_count" label="绑定菜单数" width="110" />
      <el-table-column prop="remark" label="备注" min-width="160" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-permission="'system:role:assign'" @click="openAssign(row)">分配权限</el-button>
          <el-button size="small" v-permission="'system:role:create'" @click="openUpdate(row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'system:role:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      @current-change="(p: number) => { page = p; loadData() }"
    />

    <!-- 角色编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增角色' : '编辑角色'" width="480px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="角色名称" required>
          <el-input v-model="form.name" placeholder="如：运营主管" />
        </el-form-item>
        <el-form-item label="角色编码" required>
          <el-input v-model="form.code" placeholder="唯一编码，如：ops_manager" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="可选备注" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分配权限弹窗 -->
    <el-dialog v-model="assignDialogVisible" :title="`分配权限 - ${assignRoleName}`" width="520px" destroy-on-close>
      <div class="assign-toolbar">
        <el-button size="small" @click="handleSelectAll">全选</el-button>
        <el-button size="small" @click="handleClearAll">清空</el-button>
      </div>
      <el-tree
        v-loading="assignLoading"
        :data="menuTree"
        node-key="id"
        show-checkbox
        :props="{ label: 'name', children: 'children' }"
        :default-checked-keys="checkedMenuIds"
        :check-strictly="false"
        @check="(_: unknown, { checkedKeys }: { checkedKeys: number[] }) => { checkedMenuIds = checkedKeys }"
      />
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assignSaving" @click="handleAssignSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
.assign-toolbar { margin-bottom: 12px; }
</style>