<template>
  <AccordionRoot v-bind="$attrs" :class="props.class">
    <slot />
  </AccordionRoot>
</template>

<script setup lang="ts">
import { AccordionRoot } from 'radix-vue';

export interface AccordionProps {
  class?: string;
}

const props = defineProps<AccordionProps>();
</script>
