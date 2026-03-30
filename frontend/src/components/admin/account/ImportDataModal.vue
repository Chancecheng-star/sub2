<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.accounts.dataImportHint') }}
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ t('admin.accounts.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportFile') }}</label>
        <div
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200">
              {{ fileName || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              {{ fileCount > 1 ? t('admin.accounts.dataImportFileCount', { count: fileCount }) : t('admin.accounts.dataImportFileHint') }}
            </div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json"
          multiple
          @change="handleFileChange"
        />
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportGroup') }}</label>
        <select
          v-model="selectedGroupId"
          class="input-field w-full"
          :disabled="importing || loadingGroups"
        >
          <option value="">{{ t('admin.accounts.dataImportNoGroup') }}</option>
          <option
            v-for="group in groups"
            :key="group.id"
            :value="group.id"
          >
            {{ group.name }}
          </option>
        </select>
        <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.accounts.dataImportGroupHint') }}
        </div>
      </div>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.accounts.dataImportResult') }}
        </div>
        <div class="text-sm text-gray-700 dark:text-dark-300">
          {{ t('admin.accounts.dataImportResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">
            {{ t('admin.accounts.dataImportErrors') }}
          </div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-gray-50 p-3 font-mono text-xs dark:bg-dark-800"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              {{ item.kind }} {{ item.name || item.proxy_key || '-' }} — {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.accounts.dataImporting') : t('admin.accounts.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult } from '@/types'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const file = ref<File | null>(null)
const result = ref<AdminDataImportResult | null>(null)
const selectedGroupId = ref<string>('')
const groups = ref<Array<{ id: string; name: string }>>([])
const loadingGroups = ref(false)

const fileInput = ref<HTMLInputElement | null>(null)
const files = ref<File[]>([])
const fileName = computed(() => {
  if (files.value.length === 0) return ''
  if (files.value.length === 1) return files.value[0].name
  return files.value.length > 0 ? `${files.value[0].name} +${files.value.length - 1} files` : ''
})
const fileCount = computed(() => files.value.length)

const errorItems = computed(() => result.value?.errors || [])

const loadGroups = async () => {
  loadingGroups.value = true
  try {
    const res = await adminAPI.groups.list()
    groups.value = res.items || []
  } catch (error) {
    console.error('Failed to load groups:', error)
    groups.value = []
  } finally {
    loadingGroups.value = false
  }
}

watch(
  () => props.show,
  (open) => {
    if (open) {
      files.value = []
      result.value = null
      selectedGroupId.value = ''
      if (fileInput.value) {
        fileInput.value.value = ''
      }
      loadGroups()
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  files.value = Array.from(target.files || [])
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const handleImport = async () => {
  if (files.value.length === 0) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const totalResult: AdminDataImportResult = {
      account_created: 0,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      errors: []
    }

    // 逐个文件导入
    for (const sourceFile of files.value) {
      try {
        const text = await readFileAsText(sourceFile)
        const dataPayload = JSON.parse(text)

        const res: any = await adminAPI.accounts.importData({
          data: dataPayload,
          group_id: selectedGroupId.value || undefined,
          skip_default_group_bind: !selectedGroupId.value
        })

        // 累加结果
        totalResult.account_created += res.account_created || 0
        totalResult.account_failed += res.account_failed || 0
        totalResult.proxy_created += res.proxy_created || 0
        totalResult.proxy_reused += res.proxy_reused || 0
        totalResult.proxy_failed += res.proxy_failed || 0
        if (res.errors) {
          totalResult.errors = totalResult.errors.concat(res.errors)
        }
      } catch (fileError: any) {
        // 单个文件失败，记录错误继续下一个
        totalResult.account_failed++
        totalResult.errors.push({
          kind: 'account',
          name: sourceFile.name,
          message: fileError instanceof SyntaxError 
            ? t('admin.accounts.dataImportParseFailed') 
            : (fileError?.message || t('admin.accounts.dataImportFailed'))
        })
      }
    }

    result.value = totalResult

    const msgParams: Record<string, unknown> = {
      account_created: totalResult.account_created,
      account_failed: totalResult.account_failed,
      proxy_created: totalResult.proxy_created,
      proxy_reused: totalResult.proxy_reused,
      proxy_failed: totalResult.proxy_failed,
    }
    if (totalResult.account_failed > 0 || totalResult.proxy_failed > 0) {
      appStore.showError(t('admin.accounts.dataImportCompletedWithErrors', msgParams))
    } else {
      appStore.showSuccess(t('admin.accounts.dataImportSuccess', msgParams))
      emit('imported')
    }
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.dataImportFailed'))
  } finally {
    importing.value = false
  }
}
</script>
