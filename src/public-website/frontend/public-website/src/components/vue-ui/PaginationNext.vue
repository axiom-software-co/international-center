<template>
  <PaginationLink
    aria-label="Go to next page"
    size="default"
    :class="cn('gap-1 pr-2.5', props.class)"
    v-bind="$attrs"
  >
    <span>Next</span>
    <ChevronRight class="h-4 w-4" />
  </PaginationLink>
</template>

<script setup lang="ts">
import PaginationLink from './PaginationLink.vue';
import { ChevronRight } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface PaginationNextProps {
  class?: string;
}

const props = defineProps<PaginationNextProps>();
</script>
