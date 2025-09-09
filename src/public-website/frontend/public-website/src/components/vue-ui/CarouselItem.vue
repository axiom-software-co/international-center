<template>
  <div
    role="group"
    aria-roledescription="slide"
    :class="
      cn(
        'min-w-0 shrink-0 grow-0 basis-full',
        orientation === 'horizontal' ? 'pl-4' : 'pt-4',
        props.class
      )
    "
    v-bind="$attrs"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted } from 'vue';
import { cn } from '@/lib/utils';

export interface CarouselItemProps {
  class?: string;
}

const props = defineProps<CarouselItemProps>();

const carousel = inject('carousel') as any;

if (!carousel) {
  throw new Error('CarouselItem must be used within a Carousel component');
}

const { orientation, setItemCount } = carousel;

onMounted(() => {
  // Increment item count when mounted
  setItemCount(carousel.itemCount.value + 1);
});

onUnmounted(() => {
  // Decrement item count when unmounted
  setItemCount(carousel.itemCount.value - 1);
});
</script>
