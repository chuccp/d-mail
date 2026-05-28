<template>
  <div class="mail-selector">
    <div class="selector-trigger" @click="dialogVisible = true">
      <template v-if="selectedMail">
        <el-tag size="small" closable @close.stop="handleClear" type="success" class="selector-tag">
          {{ selectedMail.name }} &lt;{{ selectedMail.mail }}&gt;
        </el-tag>
      </template>
      <span v-else class="selector-placeholder">{{ placeholder }}</span>
      <el-button size="small" class="add-btn">
        <el-icon><Plus /></el-icon>
        {{ selectedMail ? '更换' : '选择' }}
      </el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="600px"
    >
      <el-table
        :data="paginatedList"
        border
        stripe
        v-loading="loading"
        highlight-current-row
        :row-key="(row: MailConfig) => row.id"
        max-height="400"
      >
        <el-table-column label="" width="50" align="center">
          <template #default="{ row }">
            <el-radio v-model="tempSelectedMail" :value="row.mail">
              &nbsp;
            </el-radio>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('mail.recipientName')" />
        <el-table-column prop="mail" :label="t('mail.emailAddress')" />
        <el-table-column v-if="authStore.getIsAdmin" prop="userName" :label="t('common.creator')" width="100" />
      </el-table>

      <div class="pagination-wrapper">
        <div class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}-{{ Math.min(currentPage * pageSize, mailList.length) }} / 共 {{ mailList.length }} 条
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="mailList.length"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="handleConfirm" :disabled="!tempSelectedMail">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getMails } from '@/api/mail'
import { useAuthStore } from '@/store/auth'

const { t } = useI18n()
const authStore = useAuthStore()

interface Props {
  modelValue: string
  title?: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  placeholder: ''
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const dialogVisible = ref(false)
const loading = ref(false)
const mailList = ref<MailConfig[]>([])
const selectedMailAddress = ref('')
const tempSelectedMail = ref('')
const currentPage = ref(1)
const pageSize = 10

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return mailList.value.slice(start, start + pageSize)
})

const selectedMail = computed(() => {
  return mailList.value.find(m => m.mail === selectedMailAddress.value) || null
})

const loadMailList = async () => {
  loading.value = true
  try {
    const res = await getMails(1, 1000)
    if (res.code === 0 || res.code === 200) {
      mailList.value = res.data.list
    }
  } finally {
    loading.value = false
  }
}

const handleClear = () => {
  emit('update:modelValue', '')
}

const handleConfirm = () => {
  if (tempSelectedMail.value) {
    emit('update:modelValue', tempSelectedMail.value)
    dialogVisible.value = false
  }
}

const handlePageChange = () => {}

watch(() => props.modelValue, (val) => {
  selectedMailAddress.value = val
  tempSelectedMail.value = val
})

watch(dialogVisible, (val) => {
  if (val) {
    currentPage.value = 1
    loadMailList()
  }
})

onMounted(() => {
  selectedMailAddress.value = props.modelValue
  tempSelectedMail.value = props.modelValue
})
</script>

<style scoped>
.mail-selector {
  width: 100%;
}
.selector-trigger {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-height: 32px;
  padding: 4px 8px;
  cursor: pointer;
  background-color: var(--el-bg-color);
}
.selector-placeholder {
  color: var(--el-text-color-placeholder);
  font-size: 14px;
}
.selector-tag {
  margin: 0;
}
.add-btn {
  flex-shrink: 0;
}
.pagination-wrapper {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.page-info {
  font-size: 13px;
  color: var(--el-text-color-regular);
}
</style>
