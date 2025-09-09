<template>
  <div ref="carouselRef" class="overflow-hidden">
    <div
      :class="
        cn(
          'flex transition-transform duration-300 ease-in-out',
          orientation === 'horizontal' ? '-ml-4' : '-mt-4 flex-col',
          props.class
        )
      "
      :style="transformStyle"
      v-bind="$attrs"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, computed, onMounted, onUnmounted } from 'vue';
import { cn } from '@/lib/utils';

export interface CarouselContentProps {
  class?: string;
}

const props = defineProps<CarouselContentProps>();

const carousel = inject('carousel') as any;

if (!carousel) {
  throw new Error('CarouselContent must be used within a Carousel component');
}

const { orientation, currentIndex, carouselRef } = carousel;

const transformStyle = computed(() => {
  if (orientation === 'horizontal') {
    return {
      transform: `translateX(-${currentIndex.value * 100}%)`,
    };
  } else {
    return {
      transform: `translateY(-${currentIndex.value * 100}%)`,
    };
  }
});
</script>
