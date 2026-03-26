<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="space-y-4">
      <div class="rounded-[20px] border border-slate-200/80 bg-slate-50/80 px-4 py-4 dark:border-dark-700 dark:bg-dark-800/70">
        <p class="text-sm leading-6 text-slate-600 dark:text-gray-300">{{ message }}</p>
      </div>
      <slot></slot>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleCancel"
          type="button"
          class="rounded-2xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 transition-all hover:-translate-y-0.5 hover:bg-slate-50 hover:shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:ring-offset-2 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 dark:focus:ring-offset-dark-800"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="[
            'rounded-2xl px-4 py-2.5 text-sm font-medium text-white transition-all hover:-translate-y-0.5 hover:shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-dark-800',
            danger
              ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500/30'
              : 'bg-primary-600 hover:bg-primary-700 focus:ring-primary-500/30 shadow-glow'
          ]"
        >
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>
