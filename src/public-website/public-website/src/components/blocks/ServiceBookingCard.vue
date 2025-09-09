<template>
  <div class="rounded border bg-card p-6">
    <h3 v-if="isComingSoon" class="text-lg font-semibold mb-3">Coming Soon</h3>
    <h3 v-else class="text-lg font-semibold mb-3">Ready to Get Started?</h3>

    <p v-if="isComingSoon" class="text-sm text-muted-foreground">
      {{ serviceName }} will be available at our future clinic location. We're working to bring you
      the most advanced treatments.
    </p>

  </div>
</template>

<script setup lang="ts">
interface Props {
  serviceName?: string;
  isComingSoon?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  serviceName: 'this treatment',
  isComingSoon: false,
});
</script>