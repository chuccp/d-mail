<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import zhTw from 'element-plus/es/locale/lang/zh-tw'
import ja from 'element-plus/es/locale/lang/ja'

const { locale } = useI18n()

// Keep Element Plus built-in strings (pagination, date pickers, …) in step with the app language
const elementLocales: Record<string, typeof en> = {
  'zh-cn': zhCn,
  'zh-tw': zhTw,
  'en': en,
  'ja': ja
}

const elementLocale = computed(() => elementLocales[locale.value] ?? en)
</script>
