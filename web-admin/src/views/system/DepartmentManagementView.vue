<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDepartmentTree, createDepartment, updateDepartment, deleteDepartment } from '@/api/sp'
import type { SysDepartment } from '@/types/sp'

const loading = ref(false)
const tree = ref<SysDepartment[]>([])

async function loadTree() {
  loading.value = true
  try {
    tree.value = await getDepartmentTree()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

const statusText = (s: number) => (s === 1 ? '启用' : '禁用')
const statusType = (s: number): 'success' | 'info' => (s === 1 ? 'success' : 'info')

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const saving = ref(false)
const form = reactive({
  id: 0,
  parent_id: 0,
  name: '',
  leader: '',
  phone: '',
  sort: 0,
  status: 1
})

function openCreate(parent?: SysDepartment) {
  Object.assign(form, { id: 0, parent_id: parent ? parent.id : 0, name: '', leader: '', phone: '', sort: 0, status: 1 })
  dialogMode.value = 'create'
  dialogVisible.value = true
}

function openUpdate(row: SysDepartment) {
  Object.assign(form, {
    id: row.id,
    parent_id: row.parent_id,
    name: row.name,
    leader: row.leader || '',
    phone: row.phone || '',
    sort: row.sort,
    status: row.status
  })
  dialogMode.value = 'update'
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入部门名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      parent_id: form.parent_id,
      name: form.name.trim(),
      leader: form.leader.trim() || undefined,
      phone: form.phone.trim() || undefined,
      sort: form.sort || 0,
      status: form.status
    }
    if (dialogMode.value === 'create') {
      await createDepartment(payload)
      ElMessage.success('创建成功')
    } else {
      await updateDepartment(form.id, payload)
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

async function handleDelete(row: SysDepartment) {
  try {
    await ElMessageBox.confirm(`确认删除部门「${row.name}」？存在子部门或关联员工时将被拒绝。`, '删除部门', {
      type: 'warning'
    })
    await deleteDepartment(row.id)
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
        <el-button type="primary" v-permission="'system:dept:create'" @click="openCreate()">新增部门</el-button>
      </div>
      <el-button :loading="loading" @click="loadTree">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="tree" row-key="id" stripe :tree-props="{ children: 'children' }" default-expand-all>
      <el-table-column prop="name" label="部门名称" min-width="180" />
      <el-table-column prop="leader" label="负责人" width="140" />
      <el-table-column prop="phone" label="联系电话" width="150" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-permission="'system:dept:create'" @click="openCreate(row)">新增子部门</el-button>
          <el-button size="small" v-permission="'system:dept:create'" @click="openUpdate(row)">编辑</el-button>
          <el-button size="small" type="danger" v-permission="'system:dept:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增部门' : '编辑部门'" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="form.parent_id"
            :data="tree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            value-key="id"
            check-strictly
            clearable
            placeholder="不选则为顶级部门"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="部门名称" required>
          <el-input v-model="form.name" placeholder="如：运营部" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.leader" placeholder="负责人姓名" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="联系电话" />
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
</style>