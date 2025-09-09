<template>
  <a
    :href="`${basePath}/${article.slug}`"
    :class="['block group content-card', index !== undefined && index >= 2 ? 'hidden lg:block' : '', additionalClass]"
  >
    <div
      class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-colors duration-200 bg-white dark:bg-gray-800 h-full hover:bg-gray-50 dark:hover:bg-gray-700/50 flex flex-col"
    >
      <div
        class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-800 flex-shrink-0"
      >
        <img
          v-if="article.featured_image"
          :src="resolveAssetUrl(article.featured_image)"
          :alt="article.title"
          class="w-full h-full object-cover"
          loading="lazy"
        />
        <div
          v-else
          class="w-full h-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center"
        >
          <div class="w-12 h-12 text-gray-300 dark:text-gray-600">
            <svg fill="currentColor" viewBox="0 0 20 20">
              <path
                fill-rule="evenodd"
                d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z"
                clip-rule="evenodd"
              />
            </svg>
          </div>
        </div>
      </div>
      <div class="p-4 lg:p-5 flex flex-col flex-grow">
        <div class="flex-grow mb-4">
          <h3
            class="text-base lg:text-lg font-semibold text-gray-900 dark:text-white group-hover:text-gray-700 dark:group-hover:text-gray-200 line-clamp-2 transition-colors"
          >
            {{ article.title }}
          </h3>
        </div>
        <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mt-auto">
          <span>{{ article.author || defaultAuthor }}</span>
          <span>{{ formatDate(getArticleDate(article)) }}</span>
        </div>
      </div>
    </div>
  </a>
</template>

<script setup lang="ts">
import { resolveAssetUrl } from '@/lib/utils/assets';

interface Article {
  id: string | number;
  title: string;
  slug: string;
  featured_image?: string;
  published_at?: string;
  event_date?: string;
  author?: string;
}

interface Props {
  article: Article;
  basePath: string;
  defaultAuthor: string;
  index?: number;
  additionalClass?: string;
}

defineProps<Props>();

const getArticleDate = (article: Article): string => {
  // Try event_date first (for events), then fall back to published_at
  return article.event_date || article.published_at || '';
};

const formatDate = (dateString: string): string => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>