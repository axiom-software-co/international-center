<template>
  <AvatarFallback
    :class="
      cn('flex h-full w-full items-center justify-center rounded-full bg-muted', $props.class)
    "
  >
    <slot />
  </AvatarFallback>
</template>

<script setup lang="ts">
import { AvatarFallback } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface AvatarFallbackProps {
  class?: string;
}

const props = withDefaults(defineProps<AvatarFallbackProps>(), {});
</script>
