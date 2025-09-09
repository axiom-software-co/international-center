<template>
  <div
    :class="cn('relative', props.class)"
    role="region"
    aria-roledescription="carousel"
    @keydown="handleKeyDown"
    v-bind="$attrs"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, provide, onMounted, onUnmounted } from 'vue';
import { cn } from '@/lib/utils';

export interface CarouselProps {
  orientation?: 'horizontal' | 'vertical';
  class?: string;
}

const props = withDefaults(defineProps<CarouselProps>(), {
  orientation: 'horizontal',
});

const currentIndex = ref(0);
const itemCount = ref(0);
const carouselRef = ref<HTMLElement>();

const canScrollPrev = computed(() => currentIndex.value > 0);
const canScrollNext = computed(() => currentIndex.value < itemCount.value - 1);

const scrollPrev = () => {
  if (canScrollPrev.value) {
    currentIndex.value--;
  }
};

const scrollNext = () => {
  if (canScrollNext.value) {
    currentIndex.value++;
  }
};

const setItemCount = (count: number) => {
  itemCount.value = count;
};

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowLeft') {
    event.preventDefault();
    scrollPrev();
  } else if (event.key === 'ArrowRight') {
    event.preventDefault();
    scrollNext();
  }
};

// Provide carousel context to child components
provide('carousel', {
  orientation: props.orientation,
  currentIndex,
  scrollPrev,
  scrollNext,
  canScrollPrev,
  canScrollNext,
  setItemCount,
  carouselRef,
});
</script>
