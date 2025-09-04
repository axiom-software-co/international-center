<template>
  <div class="space-y-8">
    <!-- Article Header -->
    <div class="space-y-4">
      <div class="flex flex-wrap gap-2">
        <Badge
          variant="outline"
          class="px-3 py-1 text-sm font-medium border border-gray-300 bg-white text-gray-900 rounded hover:bg-white hover:text-gray-900 hover:border-gray-300"
        >
          {{ article.category }}
        </Badge>
        <Badge
          v-if="article.featured"
          variant="outline"
          class="px-3 py-1 text-sm font-medium border border-blue-300 bg-blue-50 text-blue-700 rounded"
        >
          Featured
        </Badge>
      </div>

      <div class="flex flex-wrap gap-2">
        <span
          v-for="tag in article.tags"
          :key="tag"
          class="inline-block px-2 py-1 text-xs bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-full"
        >
          #{{ tag.toLowerCase().replace(/\s+/g, '') }}
        </span>
      </div>
    </div>

    <!-- Article Excerpt -->
    <div class="prose dark:prose-invert max-w-none">
      <p class="text-lg text-gray-600 dark:text-gray-300 leading-relaxed font-medium">
        {{ article.excerpt }}
      </p>
    </div>

    <!-- Article Content -->
    <div class="prose dark:prose-invert max-w-none">
      <div v-html="article.content"></div>
    </div>

    <!-- Author Information -->
    <div v-if="article.author" class="bg-gray-50 dark:bg-gray-800 rounded p-4">
      <p class="text-sm text-gray-600 dark:text-gray-300">
        <strong>Author:</strong> {{ article.author }}
      </p>
      <p v-if="article.published_at" class="text-sm text-gray-500 dark:text-gray-400">
        Published: {{ formatDate(article.published_at) }}
      </p>
    </div>

    <!-- Call to Action -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded p-6 text-center">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        Interested in Learning More?
      </h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">
        Discover how our regenerative medicine treatments can help you achieve optimal health and
        wellness.
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <a
          href="/services"
          class="inline-flex items-center justify-center px-6 py-3 bg-blue-600 text-white font-medium rounded transition-colors"
        >
          Explore Our Services
        </a>
        <a
          href="/company/contact"
          class="inline-flex items-center justify-center px-6 py-3 border border-gray-300 hover:border-gray-400 text-gray-700 dark:text-gray-300 font-medium rounded transition-colors"
        >
          Contact Us
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Badge from '@/components/vue-ui/Badge.vue';
import type { NewsArticle } from '@/lib/clients/news/types';

interface NewsContentProps {
  article: NewsArticle;
}

defineProps<NewsContentProps>();

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString();
};
</script>
