<template>
  <div class="relative min-h-screen overflow-hidden dark:bg-dark-950">
    <div class="pointer-events-none fixed inset-0 bg-mesh-gradient opacity-90"></div>
    <div class="pointer-events-none fixed inset-x-0 top-0 h-56 bg-gradient-to-b from-white/70 to-transparent dark:from-dark-950/30"></div>

    <AppSidebar />

    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[96px]' : 'lg:ml-[18rem]']"
    >
      <AppHeader />

      <main class="px-4 pb-8 pt-5 md:px-6 lg:px-8 lg:pb-10 lg:pt-6">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
