<template>
  <SelectLabel :class="cn('py-1.5 pl-8 pr-2 text-sm font-semibold', props.class)" v-bind="$attrs">
    <slot />
  </SelectLabel>
</template>

<script setup lang="ts">
import { SelectLabel } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface SelectLabelProps {
  class?: string;
}

const props = defineProps<SelectLabelProps>();
</script>
