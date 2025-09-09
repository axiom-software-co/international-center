<template>
  <div class="w-full bg-gray-200 dark:bg-gray-800 relative overflow-hidden">
    <div class="container mx-auto px-4 py-8 relative z-10">
      <!-- Breadcrumb Navigation -->
      <nav class="hidden md:flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 mb-6">
        <a
          :href="parentPath"
          class="flex items-center gap-1 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          {{ parentLabel }}
        </a>
        <span class="text-gray-500 dark:text-gray-400">/</span>
        <span class="text-gray-900 dark:text-white font-medium">{{ itemName }}</span>
      </nav>

      <!-- Category Subtitle -->
      <p v-if="category" class="text-sm font-semibold text-blue-600 dark:text-blue-400 mb-2 uppercase tracking-wider">
        {{ category }}
      </p>

      <!-- Title -->
      <h1 class="text-3xl font-semibold text-gray-900 dark:text-white md:text-5xl mb-4">
        {{ title }}
      </h1>
    </div>

    <!-- Background Company Logo -->
    <div class="absolute right-0 top-0 w-80 h-full pointer-events-none hidden md:block">
      <div class="w-full h-full flex items-center justify-center pr-12">
        <div class="opacity-[0.08] dark:opacity-[0.12]">
          <!-- Company logo SVG -->
          <svg
            class="w-40 h-40"
            viewBox="0 0 200 200"
            fill="currentColor"
            xmlns="http://www.w3.org/2000/svg"
          >
            <!-- Medical cross symbol -->
            <rect x="85" y="40" width="30" height="120" rx="4" />
            <rect x="40" y="85" width="120" height="30" rx="4" />
            
            <!-- Medical symbol decorative elements -->
            <circle cx="100" cy="100" r="8" fill="none" stroke="currentColor" stroke-width="2" />
            <circle cx="100" cy="100" r="60" fill="none" stroke="currentColor" stroke-width="1" opacity="0.5" />
            
            <!-- Text placeholder -->
            <text x="100" y="185" text-anchor="middle" font-size="12" font-family="sans-serif" opacity="0.7">
              INTERNATIONAL CENTER
            </text>
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  parentPath: string;
  parentLabel: string;
  itemName: string;
  title: string;
  category?: string;
}

defineProps<Props>();
</script>