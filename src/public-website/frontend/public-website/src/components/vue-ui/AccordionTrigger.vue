<template>
  <AccordionHeader class="flex">
    <AccordionTrigger
      :class="
        cn(
          'flex flex-1 items-center justify-between py-4 font-medium transition-all hover:underline [&[data-state=open]>svg]:rotate-180',
          props.class
        )
      "
      v-bind="$attrs"
    >
      <slot />
      <ChevronDown class="h-4 w-4 shrink-0 transition-transform duration-200" />
    </AccordionTrigger>
  </AccordionHeader>
</template>

<script setup lang="ts">
import { AccordionHeader, AccordionTrigger } from 'radix-vue';
import { ChevronDown } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface AccordionTriggerProps {
  class?: string;
}

const props = defineProps<AccordionTriggerProps>();
</script>
