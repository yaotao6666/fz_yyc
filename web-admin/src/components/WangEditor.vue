<script setup lang="ts">
import { onBeforeUnmount, shallowRef, watch } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import type { IDomEditor } from '@wangeditor/editor'
import { uploadSpImage } from '@/utils/qiniu'
import { signMediaUrl } from '@/api/sp'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorRef = shallowRef<IDomEditor | null>(null)
const valueHtml = shallowRef<string>(props.modelValue || '')

const toolbarConfig = {}
const editorConfig: any = {
  placeholder: props.placeholder || '请输入正文内容…',
  MENU_CONF: {
    // 工具栏插入与直接粘贴图片统一走「上传 → 签名 → 插入」，保证七牛私有图可即时预览
    uploadImage: {
      customUpload: async (file: File, insertFn: (url: string, alt: string, href: string) => void) => {
        try {
          const { url } = await uploadSpImage(file)
          const signed = await signMediaUrl(url)
          insertFn(signed, '', '')
        } catch (error: any) {
          // 上传失败由拦截器提示
        }
      }
    }
  }
}

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor
}

// 外部重置内容时同步编辑器
watch(
  () => props.modelValue,
  (val) => {
    // 避免内部输入被外部覆盖导致光标跳动：仅当两者不一致时回写
    if (editorRef.value && val !== valueHtml.value) {
      editorRef.value.setHtml(val || '')
      valueHtml.value = val || ''
    }
  }
)

function handleChange(editor: IDomEditor) {
  valueHtml.value = editor.getHtml()
  emit('update:modelValue', editor.getHtml())
}

onBeforeUnmount(() => {
  editorRef.value?.destroy()
  editorRef.value = null
})
</script>

<template>
  <div class="wang-editor">
    <Toolbar
      class="wang-editor-toolbar"
      :editor="editorRef"
      :default-config="toolbarConfig"
      mode="default"
    />
    <Editor
      class="wang-editor-body"
      v-model="valueHtml"
      :default-config="editorConfig"
      mode="default"
      @on-created="handleCreated"
      @on-change="handleChange"
    />
  </div>
</template>

<style scoped>
.wang-editor {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}
.wang-editor-toolbar {
  border-bottom: 1px solid #e5e7eb;
}
.wang-editor-body {
  height: 360px;
  overflow-y: hidden;
}
</style>