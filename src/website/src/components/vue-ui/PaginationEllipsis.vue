<template>
  <span
    aria-hidden
    :class="cn('flex h-9 w-9 items-center justify-center', props.class)"
    v-bind="$attrs"
  >
    <MoreHorizontal class="h-4 w-4" />
    <span class="sr-only">More pages</span>
  </span>
</template>

<script setup lang="ts">
import { MoreHorizontal } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface PaginationEllipsisProps {
  class?: string;
}

const props = defineProps<PaginationEllipsisProps>();
</script>
