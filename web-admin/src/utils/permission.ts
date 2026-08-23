import type { Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

/**
 * 按钮级权限指令：v-permission="'products:create'"
 * 无权限时直接移除 DOM 节点。
 */
export const vPermission: Directive<HTMLElement, string | string[] | undefined> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string | string[] | undefined>) {
    const authStore = useAuthStore()
    const required = binding.value
    if (!required) return

    const codes = Array.isArray(required) ? required : [required]
    const allowed = codes.some((code) => authStore.hasPermission(code))
    if (!allowed && el.parentNode) {
      el.parentNode.removeChild(el)
    }
  },
}

/**
 * 模板内判断权限是否可用（用于 disabled 等场景）。
 */
export function hasPermission(code?: string): boolean {
  if (!code) return true
  return useAuthStore().hasPermission(code)
}
