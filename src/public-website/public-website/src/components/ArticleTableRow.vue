<template>
  <tr
    class="transition-colors duration-200 group cursor-pointer"
    @click="handleRowClick"
  >
    <td class="px-6 py-6">
      <div class="block">
        <div class="space-y-3">
          <!-- Category Label Above Title -->
          <div
            class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
          >
            {{ article.category.toUpperCase() }}
          </div>

          <!-- Title Row -->
          <div>
            <h3 class="font-semibold text-gray-900 dark:text-white line-clamp-2 transition-colors">
              {{ article.title }}
            </h3>
          </div>

          <!-- Meta Information Row -->
          <div class="flex items-center text-sm text-gray-500 dark:text-gray-400">
            <!-- Mobile: Keep original justify-between layout -->
            <div class="flex items-center justify-between w-full sm:hidden">
              <span>By {{ getArticleAuthor(article) }}</span>
              <span>{{ formatDate(getArticleDate(article)) }}</span>
            </div>

            <!-- Tablet and Desktop: Meta info immediately after each other -->
            <div class="hidden sm:flex items-center gap-3">
              <span>By {{ getArticleAuthor(article) }}</span>
              <span class="text-gray-300 dark:text-gray-600">•</span>
              <span>{{ formatDate(getArticleDate(article)) }}</span>
            </div>
          </div>
        </div>
      </div>
    </td>
  </tr>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { NewsArticle, ResearchArticle, Event } from '@/lib/clients';

type Article = NewsArticle | ResearchArticle | Event;

interface ArticleTableRowProps {
  article: Article;
  dataType: 'news' | 'research-articles' | 'events';
}

const props = defineProps<ArticleTableRowProps>();

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

const articleUrl = computed(() => {
  if (props.dataType === 'news') {
    return `/company/news/${props.article.slug}`;
  } else if (props.dataType === 'events') {
    return `/community/events/${props.article.slug}`;
  } else {
    return `/community/research/${props.article.slug}`;
  }
});

const handleRowClick = () => {
  window.location.href = articleUrl.value;
};

const getArticleAuthor = (article: Article): string => {
  const author = (article as any).author;
  if (typeof author === 'string') {
    return author;
  }
  if (typeof author === 'object' && author?.name) {
    return author.name;
  }
  return 'International Center Team';
};

const getArticleDate = (article: Article): string => {
  // For events, use event_date, for others use published_at
  if (props.dataType === 'events') {
    return (article as any).event_date || (article as any).published_at || '';
  }
  return (article as any).published_at || (article as any).publishedDate || '';
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
