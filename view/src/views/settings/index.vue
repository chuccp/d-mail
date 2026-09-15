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
            <el-form-item :label="t('settings.managementPort')">
              <el-input-number v-model="settingsForm.webPort" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item :label="t('settings.apiPort')">
              <el-input-number v-model="settingsForm.apiPort" :min="1" :max="65535" />
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
        <el-button type="primary" size="large" :loading="saving" @click="handleSave">
          {{ t('settings.saveSettings') }}
        </el-button>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getSettings, updateSettings } from '@/api/settings'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)

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

onMounted(() => {
  loadData()
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
</style>
