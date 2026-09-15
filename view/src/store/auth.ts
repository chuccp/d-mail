import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout, getSystemInfo } from '@/api/auth'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const username = ref<string>(localStorage.getItem('http2smtp-username') || '')
  const isAdmin = ref<boolean>(localStorage.getItem('http2smtp-isAdmin') === 'true')
  const rememberUsername = ref<boolean>(localStorage.getItem('http2smtp-remember') === 'true')

  // The session credential is an HttpOnly cookie the server sets, so it is deliberately
  // not stored here: page script cannot read it and it is not a routing credential.
  // This flag drives navigation only — every request is authorised server-side.
  const loggedIn = ref<boolean>(localStorage.getItem('http2smtp-logged-in') === 'true')

  const isLoggedIn = computed(() => loggedIn.value)

  const getUsername = computed(() => username.value)

  const getIsAdmin = computed(() => isAdmin.value)

  function setSession(name: string, admin: boolean) {
    username.value = name
    isAdmin.value = admin
    loggedIn.value = true
    localStorage.setItem('http2smtp-username', name)
    localStorage.setItem('http2smtp-isAdmin', String(admin))
    localStorage.setItem('http2smtp-logged-in', 'true')
  }

  async function loginAction(user: string, pass: string, remember: boolean): Promise<boolean> {
    try {
      const res = await login(user, pass)
      if (res.code === 0 || res.code === 200) {
        // The token is delivered as a Set-Cookie header; only display state is kept here
        setSession(user, res.data?.isAdmin ?? false)
        localStorage.setItem('http2smtp-remember', String(remember))
        rememberUsername.value = remember
        return true
      }
      return false
    } catch (e) {
      console.error('Login error', e)
      return false
    }
  }

  async function logoutAction() {
    try {
      // Clears the session cookie so the server stops recognising this browser
      await logout()
    } catch {
      // A failed call must still clear the local state
    } finally {
      username.value = ''
      isAdmin.value = false
      loggedIn.value = false
      localStorage.removeItem('http2smtp-username')
      localStorage.removeItem('http2smtp-isAdmin')
      localStorage.removeItem('http2smtp-logged-in')
      router.push('/login')
    }
  }

  async function checkSystemInit(): Promise<boolean> {
    try {
      const res = await getSystemInfo()
      if (res.code === 0 || res.code === 200) {
        return res.data.initialized
      }
      return false
    } catch (e) {
      console.error('Check init error', e)
      return false
    }
  }

  return {
    username,
    isAdmin,
    rememberUsername,
    isLoggedIn,
    getUsername,
    getIsAdmin,
    setSession,
    loginAction,
    logoutAction,
    checkSystemInit
  }
})
