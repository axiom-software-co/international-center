<template>
  <PaginationLink
    aria-label="Go to previous page"
    size="default"
    :class="cn('gap-1 pl-2.5', props.class)"
    v-bind="$attrs"
  >
    <ChevronLeft class="h-4 w-4" />
    <span>Previous</span>
  </PaginationLink>
</template>

<script setup lang="ts">
import PaginationLink from './PaginationLink.vue';
import { ChevronLeft } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface PaginationPreviousProps {
  class?: string;
}

const props = defineProps<PaginationPreviousProps>();
</script>
