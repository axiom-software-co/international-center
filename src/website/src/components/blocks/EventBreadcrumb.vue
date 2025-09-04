<template>
  <div class="w-full bg-gray-200 dark:bg-gray-800 relative overflow-hidden">
    <div class="container mx-auto px-4 py-8 relative z-10">
      <!-- Breadcrumb Navigation -->
      <nav class="hidden md:flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 mb-6">
        <a
          href="/community/events"
          class="flex items-center gap-1 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          Events
        </a>
        <span class="text-gray-500 dark:text-gray-400">/</span>
        <span class="text-gray-900 dark:text-white font-medium">{{ eventName }}</span>
      </nav>

      <!-- Event Category Subtitle -->
      <p v-if="category" class="text-sm font-semibold text-blue-600 dark:text-blue-400 mb-2 uppercase tracking-wider">
        {{ category }}
      </p>

      <!-- Event Title -->
      <h1 class="text-3xl font-semibold text-gray-900 dark:text-white md:text-5xl mb-4">
        {{ title }}
      </h1>
    </div>

    <!-- Background Company Logo -->
    <div class="absolute right-0 top-0 w-80 h-full pointer-events-none hidden md:block">
      <div class="w-full h-full flex items-center justify-center pr-12">
        <div class="opacity-[0.08] dark:opacity-[0.12]">
          <img
            src="/rc-logo-blue.svg"
            alt="International Center Logo"
            width="240"
            height="240"
            class="w-60 h-60"
            style="filter: grayscale(100%);"
            loading="lazy"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  eventName: string;
  title: string;
  category?: string;
}

defineProps<Props>();
</script>