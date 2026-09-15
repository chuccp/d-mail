<template>
  <div class="settings-container app-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card v-loading="loading">
          <template #header>
            <h3>{{ t('settings.generalSettings') }}</h3>
          </template>
          <el-form
            ref="settingsFormRef"
            :model="settingsForm"
            label-width="140px"
          >
            <!-- Docker instances are started with their ports, and the container's port
                 mapping depends on them, so here they are only reported -->
            <el-form-item :label="t('settings.managementPort')">
              <el-input-number v-if="!isDocker" v-model="settingsForm.webPort" :min="1" :max="65535" />
              <span v-else class="port-value">{{ settingsForm.webPort }}</span>
            </el-form-item>
            <el-form-item :label="t('settings.apiPort')">
              <el-input-number v-if="!isDocker" v-model="settingsForm.apiPort" :min="1" :max="65535" />
              <span v-else class="port-value">{{ settingsForm.apiPort }}</span>
            </el-form-item>
            <el-form-item v-if="isDocker">
              <span class="port-hint">{{ t('settings.portLockedByDocker') }}</span>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card v-loading="loading">
          <template #header>
            <h3>{{ t('settings.databaseConfig') }}</h3>
          </template>
          <el-form
            :model="settingsForm"
            label-width="140px"
          >
            <el-form-item :label="t('settings.databaseType')">
              <el-radio-group v-model="settingsForm.dbType">
                <el-radio value="sqlite">{{ t('settings.sqlite') }}</el-radio>
                <el-radio value="mysql">{{ t('settings.mysql') }}</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="settingsForm.dbType === 'sqlite'" :label="t('settings.sqliteFile')">
              <el-input v-model="settingsForm.dbFile" />
            </el-form-item>
            <template v-if="settingsForm.dbType === 'mysql'">
              <el-form-item :label="t('settings.mysqlHost')">
                <el-input v-model="settingsForm.dbHost" />
              </el-form-item>
              <el-form-item :label="t('settings.mysqlPort')">
                <el-input-number v-model="settingsForm.dbPort" :min="1" :max="65535" />
              </el-form-item>
              <el-form-item :label="t('settings.mysqlDatabase')">
                <el-input v-model="settingsForm.dbName" />
              </el-form-item>
              <el-form-item :label="t('settings.mysqlUser')">
                <el-input v-model="settingsForm.dbUser" />
              </el-form-item>
              <el-form-item :label="t('settings.mysqlPassword')">
                <el-input
                  v-model="settingsForm.dbPass"
                  type="password"
                  show-password
                  :placeholder="t('settings.passwordKeepHint')"
                />
              </el-form-item>
              <el-form-item :label="t('settings.mysqlCharset')">
                <el-input v-model="settingsForm.dbCharset" />
              </el-form-item>
            </template>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <el-row style="margin-top: 20px;">
      <el-col :span="24">
        <el-button type="primary" size="large" :loading="saving" :disabled="restarting" @click="handleSave">
          {{ t('settings.saveSettings') }}
        </el-button>
        <el-button type="warning" size="large" :loading="restarting" :disabled="saving" @click="handleSaveAndRestart">
          {{ t('settings.saveAndRestart') }}
        </el-button>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getSettings, updateSettings, restartSystem } from '@/api/settings'
import { checkInitStatus } from '@/api/setup'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const restarting = ref(false)
const isDocker = ref(false)

const settingsForm = reactive<SetInfo>({
  webPort: 12566,
  apiPort: 12567,
  dbType: 'sqlite',
  dbFile: 'data.db',
  dbHost: '127.0.0.1',
  dbPort: 3306,
  dbName: '',
  dbUser: '',
  dbPass: '',
  dbCharset: 'utf8mb4',
  adminUser: '',
  adminPass: ''
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await getSettings()
    Object.assign(settingsForm, res.data)
  } catch {
    // interceptor already surfaced the error
  } finally {
    loading.value = false
  }
}

// A Docker instance is started with its ports (flag or environment), and a saved port
// would not move the listener, so the fields are read-only there.
const loadDeployMode = async () => {
  try {
    const res = await checkInitStatus()
    isDocker.value = res.data?.isDocker ?? false
  } catch {
    // interceptor already surfaced the error
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    await updateSettings(settingsForm)
    ElMessage.success(t('settings.settingsSaved'))
    // Reload so the masked password field does not keep a stale value
    await loadData()
  } catch {
    // interceptor already surfaced the error
  } finally {
    saving.value = false
  }
}

// The ports only bind at startup, so saving them is pointless without the restart.
const handleSaveAndRestart = async () => {
  try {
    await ElMessageBox.confirm(t('settings.restartConfirm'), t('settings.restartSystem'), {
      type: 'warning',
      confirmButtonText: t('settings.saveAndRestart'),
      cancelButtonText: t('common.cancel')
    })
  } catch {
    return // dismissed
  }

  restarting.value = true
  try {
    await updateSettings(settingsForm)
    const res = await restartSystem()
    const managePort = res.data?.managePort
    const currentPort = Number(window.location.port || (window.location.protocol === 'https:' ? 443 : 80))
    if (managePort && managePort !== currentPort) {
      // Reloading would only hit a dead port — point at the new origin instead.
      ElMessage.warning(
        t('settings.restartPortChanged', { manage: managePort, api: res.data.apiPort })
      )
      restarting.value = false
    } else {
      // Stay busy until the reload: a second click would restart an instance mid-startup.
      ElMessage.success(t('settings.systemRestarted'))
      setTimeout(() => window.location.reload(), 3000)
    }
  } catch {
    // interceptor already surfaced the error
    restarting.value = false
  }
}

onMounted(() => {
  loadData()
  loadDeployMode()
})
</script>

<style scoped lang="scss">
.settings-container {
  background: #f0f2f5;
}

h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.port-value {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.port-hint {
  font-size: 12px;
  line-height: 1.6;
  color: #909399;
}
</style>
