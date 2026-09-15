import request from './request'

// The stored MySQL password is masked by the backend; a blank value on save means
// "keep the existing one", so an untouched field never wipes the credential.
export function getSettings(): Promise<ApiResponse<SetInfo>> {
  return request.get('/readSet').then(res => {
    const d = res.data
    return {
      code: res.code,
      msg: res.msg,
      data: {
        webPort: d?.manage?.port ?? 12566,
        apiPort: d?.api?.port ?? 12567,
        dbType: d?.core?.dbType ?? 'sqlite',
        dbHost: d?.mysql?.host ?? '127.0.0.1',
        dbPort: d?.mysql?.port ?? 3306,
        dbName: d?.mysql?.dbname ?? '',
        dbUser: d?.mysql?.username ?? '',
        dbPass: '',
        dbCharset: d?.mysql?.charset ?? 'utf8mb4',
        dbFile: d?.sqlite?.filename ?? 'data.db',
        adminUser: '',
        adminPass: ''
      }
    }
  })
}

// Database settings must be nested under `core`, matching model.Config on the backend
export function dbPayload(settings: SetInfo) {
  return {
    core: { dbType: settings.dbType },
    sqlite: { filename: settings.dbFile },
    mysql: {
      host: settings.dbHost,
      port: settings.dbPort,
      dbname: settings.dbName,
      username: settings.dbUser,
      password: settings.dbPass,
      charset: settings.dbCharset
    }
  }
}

export function updateSettings(settings: SetInfo): Promise<ApiResponse<any>> {
  return request.put('/reSet', {
    ...dbPayload(settings),
    manage: { port: settings.webPort },
    api: { port: settings.apiPort }
  })
}

// Restarts the process so startup-only settings (the ports) take effect. The response
// carries the ports now in effect — the management port may no longer be the one this
// page was loaded from.
export function restartSystem(): Promise<ApiResponse<{ managePort: number; apiPort: number }>> {
  return request.post('/restart', {})
}
