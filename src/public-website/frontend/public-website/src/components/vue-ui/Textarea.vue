<template>
  <textarea
    :class="
      cn(
        'flex min-h-[60px] w-full rounded border-2 border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        props.class
      )
    "
    :value="modelValue"
    @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    v-bind="$attrs"
  />
</template>

<script setup lang="ts">
import { cn } from '@/lib/utils';

export interface TextareaProps {
  class?: string;
  modelValue?: string;
}

const props = defineProps<TextareaProps>();

defineEmits<{
  'update:modelValue': [value: string];
}>();
</script>
