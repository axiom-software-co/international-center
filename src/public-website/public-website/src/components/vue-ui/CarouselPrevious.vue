<template>
  <Button
    :variant="variant"
    :size="size"
    :class="
      cn(
        'absolute h-8 w-8 rounded-full',
        orientation === 'horizontal'
          ? '-left-12 top-1/2 -translate-y-1/2'
          : '-top-12 left-1/2 -translate-x-1/2 rotate-90',
        props.class
      )
    "
    :disabled="!canScrollPrev"
    @click="scrollPrev"
    v-bind="$attrs"
  >
    <ArrowLeft class="h-4 w-4" />
    <span class="sr-only">Previous slide</span>
  </Button>
</template>

<script setup lang="ts">
import { inject } from 'vue';
import Button from './Button.vue';
import { ArrowLeft } from 'lucide-vue-next';
import { cn } from '@/lib/utils';
import type { VariantProps } from 'class-variance-authority';

export interface CarouselPreviousProps {
  variant?: VariantProps<any>['variant'];
  size?: VariantProps<any>['size'];
  class?: string;
}

const props = withDefaults(defineProps<CarouselPreviousProps>(), {
  variant: 'outline',
  size: 'icon',
});

const carousel = inject('carousel') as any;

if (!carousel) {
  throw new Error('CarouselPrevious must be used within a Carousel component');
}

const { orientation, scrollPrev, canScrollPrev } = carousel;
</script>
