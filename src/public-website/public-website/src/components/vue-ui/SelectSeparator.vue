<template>
  <SelectSeparator :class="cn('-mx-1 my-1 h-px bg-muted', props.class)" v-bind="$attrs" />
</template>

<script setup lang="ts">
import { SelectSeparator } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface SelectSeparatorProps {
  class?: string;
}

const props = defineProps<SelectSeparatorProps>();
</script>
