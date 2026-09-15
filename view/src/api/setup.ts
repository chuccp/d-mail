import request from './request'
import { dbPayload } from './settings'

export function checkInitStatus(): Promise<ApiResponse<SystemInfo>> {
  return request.get('/set').then(res => {
    return {
      code: res.code,
      msg: res.msg,
      data: {
        initialized: res.data?.hasInit ?? false,
        dbInitialized: res.data?.hasDbInit ?? false,
        hasAdmin: res.data?.hasAdmin ?? false,
        isDocker: res.data?.isDocker ?? false,
        version: ''
      }
    }
  })
}

// Tries the supplied settings without saving them, so a typo is caught before it persists
export function testConnection(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.post('/testConnection', dbPayload(settings))
}

// Step 1: Initialize database connection
export function initDatabase(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/dbInit', {
    ...dbPayload(settings),
    manage: {
      port: settings.webPort
    },
    api: { port: settings.apiPort }
  })
}

// Step 2: Initialize admin account
export function initAdmin(username: string, password: string): Promise<ApiResponse<any>> {
  return request.put('/adminInit', {
    username,
    password
  })
}

// Step 2 alternative: Skip admin creation (admin already exists)
export function skipAdmin(): Promise<ApiResponse<any>> {
  return request.put('/adminSkip', {})
}

// Check if an admin user already exists
export function checkAdminExists(): Promise<ApiResponse<{ hasAdmin: boolean; adminName: string }>> {
  return request.get('/adminExists')
}
