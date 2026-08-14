import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { StaffUser } from '@/types'

export const useStaffAuthStore = defineStore('staff-auth', () => {
  const token = ref<string>(uni.getStorageSync('staff_token') || '')
  const user = ref<StaffUser | null>(null)

  function setToken(t: string) {
    token.value = t
    if (t) uni.setStorageSync('staff_token', t)
    else uni.removeStorageSync('staff_token')
  }

  function setUser(u: StaffUser | null) {
    user.value = u
  }

  function logout() {
    token.value = ''
    user.value = null
    uni.removeStorageSync('staff_token')
  }

  return { token, user, setToken, setUser, logout }
})
