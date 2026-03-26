<template>
  <div class="flex flex-wrap items-center justify-end gap-3 rounded-[22px] border border-slate-200/80 bg-white/90 p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800/80">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary min-w-[44px] px-3">
      <Icon name="refresh" size="md" :class="[loading ? 'animate-spin' : '']" />
    </button>
    <slot name="after"></slot>
    <button @click="$emit('sync')" class="btn btn-secondary">{{ t('admin.accounts.syncFromCrs') }}</button>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="btn btn-primary shadow-glow">{{ t('admin.accounts.createAccount') }}</button>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'sync', 'create'])

const { t } = useI18n()
</script>
