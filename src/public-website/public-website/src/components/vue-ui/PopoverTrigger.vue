<template>
  <RadixPopoverTrigger v-bind="$attrs">
    <slot />
  </RadixPopoverTrigger>
</template>

<script setup lang="ts">
import { PopoverTrigger as RadixPopoverTrigger } from 'radix-vue';
</script>
