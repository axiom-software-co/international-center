<template>
  <AccordionItem :class="cn('border-b', props.class)" v-bind="$attrs">
    <slot />
  </AccordionItem>
</template>

<script setup lang="ts">
import { AccordionItem } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface AccordionItemProps {
  class?: string;
}

const props = defineProps<AccordionItemProps>();
</script>
