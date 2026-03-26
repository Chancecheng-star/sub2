<template>
  <button
    type="button"
    @click="toggle"
    class="relative inline-flex h-7 w-12 flex-shrink-0 cursor-pointer rounded-full border border-transparent transition-all duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500/30 focus:ring-offset-2 dark:focus:ring-offset-dark-800"
    :class="[
      modelValue
        ? 'bg-primary-600 shadow-[0_8px_22px_rgba(99,102,241,0.28)]'
        : 'bg-slate-200 shadow-inner dark:bg-dark-600'
    ]"
    role="switch"
    :aria-checked="modelValue"
  >
    <span
      class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-[0_4px_12px_rgba(15,23,42,0.18)] ring-0 transition duration-200 ease-in-out"
      :class="[modelValue ? 'translate-x-6' : 'translate-x-0.5']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>
