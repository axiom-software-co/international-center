<template>
  <AccordionContent
    class="overflow-hidden text-sm transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down"
    v-bind="$attrs"
  >
    <div :class="cn('pb-4 pt-0', props.class)">
      <slot />
    </div>
  </AccordionContent>
</template>

<script setup lang="ts">
import { AccordionContent } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface AccordionContentProps {
  class?: string;
}

const props = defineProps<AccordionContentProps>();
</script>
