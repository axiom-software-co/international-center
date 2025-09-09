<template>
  <div
    v-if="position.visible"
    class="absolute top-0 h-full overflow-hidden pointer-events-none z-0"
    :style="{
      right: `${position.rightOffset}px`,
      width: `${position.width}px`,
    }"
  >
    <div class="relative h-full flex items-center justify-center">
      <component
        :is="DnaIcon"
        class="text-gray-400/30 dark:text-gray-600/30 transform scale-125"
        :style="{
          height: '115%',
          width: 'auto',
          filter: 'grayscale(100%)',
          clipPath: 'inset(12.5% 0 12.5% 0)',
        }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { ArrowRight } from 'lucide-vue-next';


interface DynamicServiceBackgroundProps {
  targetId: string; // ID of the aside element to align with
}

interface PositionData {
  rightOffset: number;
  width: number;
  visible: boolean;
}

const props = defineProps<DynamicServiceBackgroundProps>();

const position = ref<PositionData>({
  rightOffset: 48,
  width: 256,
  visible: false,
});

let observerRef: ResizeObserver | null = null;

const calculatePosition = () => {
  const targetElement = document.getElementById(props.targetId);
  if (!targetElement) return;

  const rect = targetElement.getBoundingClientRect();
  const containerElement = document.querySelector('.service-page-container');
  if (!containerElement) return;

  const containerRect = containerElement.getBoundingClientRect();

  const viewportWidth = window.innerWidth;
  const elementLeftEdge = rect.left;
  const elementWidth = rect.width;
  const containerRightEdge = containerRect.right;

  // Account for responsive grid behavior and breakpoints
  const isLargeScreen = viewportWidth >= 1024; // lg breakpoint
  const isMediumScreen = viewportWidth >= 768; // md breakpoint

  if (!isMediumScreen) {
    // Below md breakpoint, don't show the background
    position.value = { rightOffset: 0, width: 256, visible: false };
    return;
  }

  // Calculate true visual center of the sidebar considering CSS Grid behavior
  // The sidebar uses md:col-span-5 lg:col-span-4 with md:gap-8
  const gridGap = 32; // md:gap-8 = 2rem = 32px
  const containerWidth = containerRect.width;
  const containerPadding = (viewportWidth - containerWidth) / 2; // Tailwind container padding

  // Calculate grid column widths based on breakpoint
  let sidebarRatio: number;
  if (isLargeScreen) {
    sidebarRatio = 4 / 12; // lg:col-span-4 out of 12 columns
  } else {
    sidebarRatio = 5 / 12; // md:col-span-5 out of 12 columns
  }

  // Calculate the theoretical center based on grid proportions
  const contentRatio = 1 - sidebarRatio;
  const availableWidth = containerWidth - gridGap; // Subtract gap between columns
  const contentWidth = availableWidth * contentRatio;
  const sidebarWidth = availableWidth * sidebarRatio;

  // The sidebar center is: container left + content width + gap + (sidebar width / 2)
  const theoreticalSidebarCenter = containerRect.left + contentWidth + gridGap + sidebarWidth / 2;

  // Use actual element center as primary source of truth
  const actualElementCenter = elementLeftEdge + elementWidth / 2;

  // Apply screen-size specific corrections for remaining discrepancies
  let correctionOffset = 0;

  if (viewportWidth >= 768 && viewportWidth < 1024) {
    // md breakpoint range (768px-1024px) - sidebar is 5/12 width
    // Add rightward correction to account for container padding and grid nuances
    const screenRatio = (viewportWidth - 768) / (1024 - 768); // 0 to 1 across md range
    correctionOffset = 16 - screenRatio * 8; // 16px at 768px, 8px at 1024px
  } else if (viewportWidth >= 1024 && viewportWidth < 1280) {
    // lg breakpoint range (1024px-1280px) - sidebar is 4/12 width
    correctionOffset = 8; // Smaller correction for lg screens
  } else if (viewportWidth >= 1280) {
    // xl+ screens
    correctionOffset = 4; // Minimal correction for larger screens
  }

  // Apply the correction to move icon slightly right
  const correctedCenter = actualElementCenter + correctionOffset;

  // Calculate offset from viewport right to the corrected center
  const centerOffsetFromRight = viewportWidth - correctedCenter;

  // Position icon (subtract half width to center the 256px icon)
  const iconHalfWidth = 128;
  const idealOffset = centerOffsetFromRight - iconHalfWidth;

  // Ensure minimum distance from container edge (prevent overflow)
  const minOffsetFromContainer = 16;
  const containerRightOffset = viewportWidth - containerRightEdge;
  const minAllowedOffset = containerRightOffset + minOffsetFromContainer;

  const finalOffset = Math.max(idealOffset, minAllowedOffset);

  position.value = {
    rightOffset: finalOffset,
    width: 256,
    visible: true,
  };
};

const handleResize = () => {
  calculatePosition();
};

const handleScroll = () => {
  calculatePosition();
};

onMounted(() => {
  const targetElement = document.getElementById(props.targetId);
  if (!targetElement) return;

  // Initial calculation
  calculatePosition();

  // Set up ResizeObserver to watch for size changes
  observerRef = new ResizeObserver(() => {
    calculatePosition();
  });

  observerRef.observe(targetElement);

  // Listen for window resize
  window.addEventListener('resize', handleResize);

  // Also recalculate on scroll for sticky elements
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  if (observerRef) {
    observerRef.disconnect();
  }
  window.removeEventListener('resize', handleResize);
  window.removeEventListener('scroll', handleScroll);
});
</script>
