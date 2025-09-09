<template>
  <RadixPopoverRoot v-bind="$attrs">
    <slot />
  </RadixPopoverRoot>
</template>

<script setup lang="ts">
import { PopoverRoot as RadixPopoverRoot } from 'radix-vue';
</script>
