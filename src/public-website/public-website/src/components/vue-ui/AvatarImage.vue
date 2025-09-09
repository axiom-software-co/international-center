<template>
  <AvatarImage :class="cn('aspect-square h-full w-full', $props.class)" :src="src" :alt="alt" />
</template>

<script setup lang="ts">
import { AvatarImage } from 'radix-vue';
import { cn } from '@/lib/utils';

export interface AvatarImageProps {
  src?: string;
  alt?: string;
  class?: string;
}

const props = withDefaults(defineProps<AvatarImageProps>(), {});
</script>
