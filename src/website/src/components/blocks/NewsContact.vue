<template>
  <div :class="cn('rounded border bg-card p-6', props.class)">
    <h3 class="mb-4 text-lg font-semibold">More News</h3>
    <div class="space-y-3">
      <Button class="w-full bg-black hover:bg-gray-800 text-white rounded-sm" asChild>
        <a href="/company/news" class="flex items-center justify-center">
          Browse All News
        </a>
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { cn } from '@/lib/vue-utils';
import Button from '@/components/vue-ui/Button.vue';

interface Props {
  class?: string;
}

const props = defineProps<Props>();
</script>