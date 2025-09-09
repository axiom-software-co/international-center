<template>
  <SelectTrigger
    :class="
      cn(
        'flex h-9 w-full items-center justify-between rounded border-2 border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1 [&_svg]:pointer-events-none group',
        props.class
      )
    "
    v-bind="$attrs"
  >
    <slot />
    <SelectIcon>
      <div class="flex items-center justify-center h-full">
        <ChevronDown
          class="h-4 w-4 opacity-50 shrink-0 transition-transform duration-200 data-[state=open]:rotate-180 group-data-[state=open]:rotate-180"
        />
      </div>
    </SelectIcon>
  </SelectTrigger>
</template>

<script setup lang="ts">
import { SelectTrigger, SelectIcon } from 'radix-vue';
import { ChevronDown } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface SelectTriggerProps {
  class?: string;
}

const props = defineProps<SelectTriggerProps>();
</script>
