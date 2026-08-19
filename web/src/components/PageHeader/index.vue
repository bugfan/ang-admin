<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import BackIcon from "~icons/ep/back";

defineProps<{
  title: string;
  description?: string;
  backTitle?: string;
}>();

const emit = defineEmits<{
  (e: "back"): void;
}>();

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === "Escape" || e.code === "Escape") {
    // If an Element Plus overlay (e.g. dialog, popconfirm, select dropdown) is active, prioritize closing it
    const activeOverlay = document.querySelector(
      ".el-overlay:not([style*='display: none'])"
    );
    if (activeOverlay) {
      return;
    }
    emit("back");
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeyDown);
});
</script>

<template>
  <div
    class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 mb-4 border-b border-(--el-border-color-lighter)"
  >
    <div class="flex items-center">
      <el-button
        circle
        class="mr-3! shrink-0 shadow-2xs hover:scale-105 transition-transform"
        :icon="useRenderIcon(BackIcon)"
        :title="backTitle || '返回 (Esc)'"
        @click="emit('back')"
      />
      <div>
        <h2
          class="text-base sm:text-lg font-bold text-(--el-text-color-primary)"
        >
          {{ title }}
        </h2>
        <div
          v-if="description"
          class="text-xs text-(--el-text-color-secondary) mt-0.5"
        >
          {{ description }}
        </div>
      </div>
    </div>

    <!-- Right Side Actions Slot -->
    <div
      v-if="$slots.actions"
      class="flex items-center space-x-2 sm:space-x-3 shrink-0 self-end sm:self-auto"
    >
      <slot name="actions" />
    </div>
  </div>
</template>
