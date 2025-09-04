<template>
  <label :class="cn(labelVariants(), props.class)" v-bind="$attrs">
    <slot />
  </label>
</template>

<script setup lang="ts">
import { cva } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const labelVariants = cva(
  'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
);

export interface LabelProps {
  class?: string;
}

const props = defineProps<LabelProps>();
</script>
