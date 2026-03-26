<template>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-[24px] border border-white/70 bg-white/85 p-5 shadow-[0_18px_50px_rgba(148,163,184,0.16)] dark:border-white/10 dark:bg-dark-900/80">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-400">{{ t('usage.totalRequests') }}</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">{{ stats?.total_requests?.toLocaleString() || '0' }}</p>
          <p class="mt-2 text-sm text-slate-500 dark:text-dark-300">{{ t('usage.inSelectedRange') }}</p>
        </div>
        <div class="rounded-2xl bg-blue-100 p-3 text-blue-600 dark:bg-blue-900/30">
          <Icon name="document" size="md" />
        </div>
      </div>
    </div>

    <div class="rounded-[24px] border border-white/70 bg-white/85 p-5 shadow-[0_18px_50px_rgba(148,163,184,0.16)] dark:border-white/10 dark:bg-dark-900/80">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-400">{{ t('usage.totalTokens') }}</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="mt-2 text-sm text-slate-500 dark:text-dark-300">
            {{ t('usage.in') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
            {{ t('usage.out') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
        <div class="rounded-2xl bg-amber-100 p-3 text-amber-600 dark:bg-amber-900/30">
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
          </svg>
        </div>
      </div>
    </div>

    <div class="rounded-[24px] border border-white/70 bg-white/85 p-5 shadow-[0_18px_50px_rgba(148,163,184,0.16)] dark:border-white/10 dark:bg-dark-900/80">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-400">{{ t('usage.totalCost') }}</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-emerald-600 dark:text-emerald-400">
            ${{ ((stats?.total_account_cost ?? stats?.total_actual_cost) || 0).toFixed(4) }}
          </p>
          <p v-if="stats?.total_account_cost != null" class="mt-2 text-sm text-slate-500 dark:text-dark-300">
            {{ t('usage.userBilled') }}:
            <span class="text-slate-700 dark:text-dark-100">${{ (stats?.total_actual_cost || 0).toFixed(4) }}</span>
            ¡¤ {{ t('usage.standardCost') }}:
            <span class="text-slate-700 dark:text-dark-100">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
          </p>
          <p v-else class="mt-2 text-sm text-slate-500 dark:text-dark-300">
            {{ t('usage.standardCost') }}:
            <span class="line-through">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
          </p>
        </div>
        <div class="rounded-2xl bg-emerald-100 p-3 text-emerald-600 dark:bg-emerald-900/30">
          <Icon name="dollar" size="md" />
        </div>
      </div>
    </div>

    <div class="rounded-[24px] border border-white/70 bg-white/85 p-5 shadow-[0_18px_50px_rgba(148,163,184,0.16)] dark:border-white/10 dark:bg-dark-900/80">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400 dark:text-dark-400">{{ t('usage.avgDuration') }}</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
          <p class="mt-2 text-sm text-slate-500 dark:text-dark-300">Observed across the selected log window</p>
        </div>
        <div class="rounded-2xl bg-violet-100 p-3 text-violet-600 dark:bg-violet-900/30">
          <Icon name="clock" size="md" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ stats: AdminUsageStatsResponse | null }>()

const { t } = useI18n()

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}
</script>
