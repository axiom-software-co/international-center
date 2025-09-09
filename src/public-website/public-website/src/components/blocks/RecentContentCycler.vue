<template>
  <div>
    <!-- Latest Research Section -->
    <section class="pt-16 pb-20">
      <div class="container">
        <div class="text-center mb-16">
          <h2 class="text-3xl font-semibold lg:text-5xl mb-4">Latest Research</h2>
        </div>

        <!-- Error State -->
        <div v-if="error" class="text-center py-12">
          <div class="max-w-md mx-auto bg-red-50 border border-red-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-red-900 mb-2">Unable to Load Research</h3>
            <p class="text-red-700 text-sm">{{ error }}</p>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="index in 3"
            :key="index"
            class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden bg-white dark:bg-gray-800 h-full"
          >
            <div
              class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-300 dark:bg-gray-600 animate-pulse"
            ></div>
            <div class="p-4 lg:p-5">
              <div class="space-y-2 mb-4 lg:mb-6">
                <div class="h-4 lg:h-5 bg-gray-300 dark:bg-gray-600 rounded animate-pulse"></div>
                <div
                  class="h-4 lg:h-5 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-3/4"
                ></div>
              </div>
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <div
                    class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-24"
                  ></div>
                  <div
                    class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-16"
                  ></div>
                </div>
                <div
                  class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-20"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Content Grid -->
        <div
          v-if="!isLoading && !error && researchArticles.length > 0"
          :class="`grid gap-4 md:gap-6 lg:gap-8 ${
            researchArticles.length === 1
              ? 'md:grid-cols-1 lg:grid-cols-1 max-w-md mx-auto'
              : researchArticles.length === 2
                ? 'md:grid-cols-2 lg:grid-cols-2 max-w-2xl mx-auto'
                : 'md:grid-cols-2 lg:grid-cols-3'
          }`"
        >
          <a
            v-for="researchArticle in researchArticles.slice(0, 3)"
            :key="researchArticle.id"
            :href="`/community/research/${researchArticle.slug}`"
            class="block group research-article-card"
          >
            <div
              class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-colors duration-200 bg-white dark:bg-gray-800 h-full hover:bg-gray-50 dark:hover:bg-gray-700/50"
            >
              <div
                class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-200 dark:bg-gray-600"
              >
                <img
                  v-if="researchArticle.featured_image"
                  :src="resolveAssetUrl(researchArticle.featured_image)"
                  :alt="researchArticle.title"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
              </div>
              <div class="p-4 lg:p-5">
                <div class="mb-4 lg:mb-6">
                  <h3
                    class="text-sm lg:text-base font-semibold text-gray-900 dark:text-white line-clamp-2"
                  >
                    {{ researchArticle.title }}
                  </h3>
                </div>
                <div class="space-y-1">
                  <div
                    class="flex items-center justify-between text-xs lg:text-sm text-gray-500 dark:text-gray-400"
                  >
                    <span>By Research Team</span>
                    <span>{{ researchArticle.readingTime || '8 min read' }}</span>
                  </div>
                  <div class="text-left">
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ formatDate(researchArticle.published_at) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </a>
        </div>

        <!-- No Content State -->
        <div v-if="!isLoading && !error && researchArticles.length === 0" class="text-center py-12">
          <div class="max-w-md mx-auto">
            <h3 class="text-lg font-semibold text-gray-900 mb-2">No Research Articles Available</h3>
            <p class="text-gray-600 text-sm">Check back soon for the latest research findings.</p>
          </div>
        </div>

        <!-- View All Button -->
        <div class="mt-12 text-center">
          <a
            href="/community/research"
            class="inline-block px-6 py-2 text-white bg-black hover:bg-gray-800 font-semibold rounded-sm transition-colors border border-black hover:border-gray-800"
          >
            View All Research
          </a>
        </div>
      </div>
    </section>

    <!-- Latest News Section -->
    <section class="pt-16 pb-20 bg-gray-50">
      <div class="container">
        <div class="text-center mb-16">
          <h2 class="text-3xl font-semibold lg:text-5xl mb-4">Latest News</h2>
        </div>

        <!-- Error State -->
        <div v-if="error" class="text-center py-12">
          <div class="max-w-md mx-auto bg-red-50 border border-red-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-red-900 mb-2">Unable to Load News Articles</h3>
            <p class="text-red-700 text-sm">{{ error }}</p>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="index in 3"
            :key="index"
            class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden bg-white dark:bg-gray-800 h-full"
          >
            <div
              class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-300 dark:bg-gray-600 animate-pulse"
            ></div>
            <div class="p-4 lg:p-5">
              <div class="space-y-2 mb-4 lg:mb-6">
                <div class="h-4 lg:h-5 bg-gray-300 dark:bg-gray-600 rounded animate-pulse"></div>
                <div
                  class="h-4 lg:h-5 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-3/4"
                ></div>
              </div>
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <div
                    class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-24"
                  ></div>
                  <div
                    class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-16"
                  ></div>
                </div>
                <div
                  class="h-3 lg:h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-20"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Content Grid -->
        <div
          v-if="!isLoading && !error && newsArticles.length > 0"
          :class="`grid gap-4 md:gap-6 lg:gap-8 ${
            newsArticles.length === 1
              ? 'md:grid-cols-1 lg:grid-cols-1 max-w-md mx-auto'
              : newsArticles.length === 2
                ? 'md:grid-cols-2 lg:grid-cols-2 max-w-2xl mx-auto'
                : 'md:grid-cols-2 lg:grid-cols-3'
          }`"
        >
          <a
            v-for="article in newsArticles.slice(0, 3)"
            :key="article.id"
            :href="`/company/news/${article.slug}`"
            class="block group news-card"
          >
            <div
              class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-colors duration-200 bg-white dark:bg-gray-800 h-full hover:bg-gray-50 dark:hover:bg-gray-700/50"
            >
              <div
                class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-200 dark:bg-gray-600"
              >
                <img
                  v-if="article.featured_image"
                  :src="resolveAssetUrl(article.featured_image)"
                  :alt="article.title"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
              </div>
              <div class="p-4 lg:p-5">
                <div class="mb-4 lg:mb-6">
                  <h3
                    class="text-sm lg:text-base font-semibold text-gray-900 dark:text-white line-clamp-2"
                  >
                    {{ article.title }}
                  </h3>
                </div>
                <div class="space-y-1">
                  <div
                    class="flex items-center justify-between text-xs lg:text-sm text-gray-500 dark:text-gray-400"
                  >
                    <span>By {{ article.author || 'International Center Team' }}</span>
                    <span>{{ article.readingTime || '5 min read' }}</span>
                  </div>
                  <div class="text-left">
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ formatDate(article.published_at) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </a>
        </div>

        <!-- No Content State -->
        <div v-if="!isLoading && !error && newsArticles.length === 0" class="text-center py-12">
          <div class="max-w-md mx-auto">
            <h3 class="text-lg font-semibold text-gray-900 mb-2">No News Articles Available</h3>
            <p class="text-gray-600 text-sm">
              Check back soon for the latest updates and insights.
            </p>
          </div>
        </div>

        <!-- View All Button -->
        <div class="mt-12 text-center">
          <a
            href="/company/news"
            class="inline-block px-6 py-2 text-white bg-black hover:bg-gray-800 font-semibold rounded-sm transition-colors border border-black hover:border-gray-800"
          >
            View All News
          </a>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { resolveAssetUrl } from '@/lib/utils/assets';

interface Article {
  id: string | number;
  title: string;
  slug: string;
  featured_image?: string;
  published_at: string;
  readingTime?: string;
  author?: string;
}

interface Props {
  researchArticles?: Article[];
  newsArticles?: Article[];
}

interface RecentContentData {
  researchArticles: Article[];
  newsArticles: Article[];
}

const props = withDefaults(defineProps<Props>(), {
  researchArticles: () => [],
  newsArticles: () => [],
});

// Reactive state for content data that can be updated on client side
const contentData = ref<RecentContentData>({
  researchArticles: props.researchArticles || [],
  newsArticles: props.newsArticles || [],
});

// Loading and error states
const isLoading = ref(false);
const error = ref<string | null>(null);

// Use reactive data for computed properties
const researchArticles = computed(() => contentData.value.researchArticles);
const newsArticles = computed(() => contentData.value.newsArticles);

onMounted(async () => {
  // If no meaningful data provided (client:only or empty data), load data on client side
  if (contentData.value.researchArticles.length === 0 && contentData.value.newsArticles.length === 0) {
    try {
      isLoading.value = true;
      error.value = null;

      const { loadRecentContentData } = await import('../../lib/navigation-data');
      const clientData = await loadRecentContentData();
      console.log('🔍 [RecentContentCycler] Client-side data loaded:', clientData);

      // Update reactive state with new data
      contentData.value = clientData;
    } catch (err) {
      console.error('❌ [RecentContentCycler] Client-side data loading failed:', err);
      error.value = 'Failed to load content. Please try again later.';
    } finally {
      isLoading.value = false;
    }
  }
});

// Utility function to format dates
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};
</script>
