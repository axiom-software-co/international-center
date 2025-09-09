<template>
  <div class="max-w-6xl mx-auto">
    <!-- Opening -->
    <div class="text-center mb-20">
      <h1
        class="text-3xl lg:text-5xl font-light text-gray-900 dark:text-white leading-relaxed mb-16"
      >
        You deserve care that sees your whole story.
      </h1>

      <!-- Opening Divider -->
      <div class="flex justify-center">
        <div class="w-32 h-px bg-gray-300 dark:bg-gray-600"></div>
      </div>
    </div>

    <!-- Flowing Values -->
    <div class="space-y-24">
      <div v-for="(value, index) in values" :key="value.key">
        <!-- Value Story -->
        <div class="text-center space-y-12">
          <div class="text-9xl opacity-80 emoji-float">{{ value.icon }}</div>
          <div class="max-w-4xl mx-auto space-y-8">
            <h2 class="text-3xl lg:text-4xl font-medium text-gray-900 dark:text-white">
              We {{ value.title.toLowerCase() }}
              {{
                value.key === 'elevate'
                  ? 'every person we meet'
                  : value.key === 'educate'
                    ? 'because knowledge heals'
                    : "for tomorrow's possibilities"
              }}
            </h2>
            <div class="prose prose-xl lg:prose-2xl prose-gray dark:prose-invert mx-auto">
              <p
                class="not-prose text-xl lg:text-2xl text-gray-600 dark:text-gray-400 leading-relaxed italic"
              >
                "{{ value.story }}"
              </p>
            </div>
          </div>
        </div>

        <!-- Divider - except for last item -->
        <div v-if="index < values.length - 1" class="flex justify-center mt-20 mb-20">
          <div class="w-32 h-px bg-gray-300 dark:bg-gray-600"></div>
        </div>
      </div>
    </div>

    <!-- Final Divider -->
    <div class="flex justify-center mt-20 mb-20">
      <div class="w-32 h-px bg-gray-300 dark:bg-gray-600"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface CoreValue {
  icon: string;
  title: string;
  story: string;
}

interface ScrollRevealValuesProps {
  coreValues: {
    elevate: CoreValue;
    educate: CoreValue;
    innovate: CoreValue;
  };
}

const props = defineProps<ScrollRevealValuesProps>();

const values = computed(() => [
  { key: 'elevate', ...props.coreValues.elevate },
  { key: 'educate', ...props.coreValues.educate },
  { key: 'innovate', ...props.coreValues.innovate },
]);
</script>

<style scoped>
.emoji-float {
  animation: float 6s ease-in-out infinite;
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .emoji-float {
    animation: none;
  }
}
</style>
