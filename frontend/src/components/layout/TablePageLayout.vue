<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <div class="page-shell glass-card">
      <!-- 固定区域：操作按钮 -->
      <div v-if="$slots.actions" class="layout-section-fixed px-4 pt-4 md:px-6 md:pt-6">
        <slot name="actions" />
      </div>

      <!-- 固定区域：搜索和过滤器 -->
      <div
        v-if="$slots.filters"
        class="layout-section-fixed border-b border-white/60 px-4 py-4 md:px-6 dark:border-white/10"
      >
        <slot name="filters" />
      </div>

      <!-- 滚动区域：表格 -->
      <div class="layout-section-scrollable px-2 pb-2 md:px-3 md:pb-3">
        <div class="table-scroll-container">
          <slot name="table" />
        </div>
      </div>

      <!-- 固定区域：分页器 -->
      <div
        v-if="$slots.pagination"
        class="layout-section-fixed border-t border-white/60 px-4 py-3 md:px-6 md:py-4 dark:border-white/10"
      >
        <slot name="pagination" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  @apply flex flex-col gap-6;
  height: calc(100vh - 64px - 4rem); /* 减去 header + lg:p-8 的上下padding */
}

.page-shell {
  @apply relative flex h-full flex-col overflow-hidden rounded-[30px] border border-white/70 bg-white/80 shadow-[0_24px_80px_rgba(15,23,42,0.12)] backdrop-blur-xl dark:border-white/10 dark:bg-dark-900/80;
}

.layout-section-fixed {
  @apply flex-shrink-0;
}

.layout-section-scrollable {
  @apply flex-1 min-h-0 flex flex-col;
}

/* 表格滚动容器 - 增强版表体滚动方案 */
.table-scroll-container {
  @apply flex h-full flex-col overflow-hidden rounded-[24px] border border-slate-200/80 bg-white shadow-[0_18px_50px_rgba(148,163,184,0.18)] dark:border-dark-700 dark:bg-dark-800;
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content;
  display: table;
}

.table-scroll-container :deep(thead) {
  @apply bg-slate-50/95 backdrop-blur-sm dark:bg-dark-800/90;
}

.table-scroll-container :deep(th) {
  @apply border-b border-slate-200 px-5 py-4 text-left text-xs font-semibold uppercase tracking-[0.14em] text-slate-500 dark:border-dark-700 dark:text-dark-300;
}

.table-scroll-container :deep(td) {
  @apply border-b border-slate-100 px-5 py-4 text-sm text-slate-700 dark:border-dark-800 dark:text-gray-300;
}

.table-scroll-container :deep(tbody tr) {
  @apply transition-colors duration-150 hover:bg-slate-50/80 dark:hover:bg-dark-700/40;
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode .page-shell {
  @apply h-auto overflow-visible rounded-[24px];
}

.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none shadow-none bg-transparent;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit px-0 pb-0;
}

.table-page-layout.mobile-mode .layout-section-fixed {
  @apply px-0;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(.table-wrapper) {
  @apply overflow-visible;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
