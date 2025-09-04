<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">News Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Return Home
        </a>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="isLoading" class="space-y-8 lg:space-y-12">
      <!-- Featured Article Skeleton -->
      <div
        class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded overflow-hidden"
      >
        <div class="flex flex-col lg:flex-row">
          <div class="flex-1 p-6 md:p-8 lg:p-10 xl:p-12 flex flex-col justify-center">
            <div class="space-y-4">
              <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-24"></div>
              <div class="h-8 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-3/4"></div>
              <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-32"></div>
            </div>
          </div>
          <div class="lg:w-1/2 aspect-[16/9] lg:aspect-auto">
            <div class="w-full h-full p-6">
              <div class="w-full h-full bg-gray-300 dark:bg-gray-600 animate-pulse rounded"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Categories Skeleton -->
      <div
        v-for="index in 2"
        :key="index"
        class="news-category bg-white dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 p-4 lg:p-6"
      >
        <div class="mb-4 lg:mb-6">
          <div class="h-6 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-48"></div>
        </div>
        <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="cardIndex in 3"
            :key="cardIndex"
            class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden bg-white dark:bg-gray-800 h-full"
          >
            <div
              class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-300 dark:bg-gray-600 animate-pulse"
            ></div>
            <div class="p-4 lg:p-5">
              <div class="space-y-2 mb-4 lg:mb-6">
                <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse"></div>
                <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-3/4"></div>
              </div>
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-24"></div>
                  <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-16"></div>
                </div>
                <div class="h-3 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-20"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Content State -->
    <div v-else>
      <!-- Featured Article Section -->
      <div
        v-if="featuredArticle"
        class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded overflow-hidden mb-8 lg:mb-12"
      >
        <a :href="featuredArticleHref" class="block group featured-card">
          <div class="flex flex-col lg:flex-row">
            <!-- Content - Left Side -->
            <div class="flex-1 p-6 md:p-8 lg:p-10 xl:p-12 flex flex-col justify-center">
              <div class="space-y-4">
                <p
                  class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide"
                >
                  {{ featuredCategory }}
                </p>
                <h2
                  class="text-2xl md:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white leading-tight group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors"
                >
                  {{ featuredTitle }}
                </h2>
                <div class="flex items-center gap-3 text-gray-500 dark:text-gray-400">
                  <span class="text-sm">{{ featuredAuthor }}</span>
                  <span class="text-gray-400 dark:text-gray-500">•</span>
                  <span class="text-sm">{{ featuredDate }}</span>
                </div>
              </div>
            </div>
            <!-- Image - Right Side -->
            <div class="lg:w-1/2 aspect-[16/9] lg:aspect-auto">
              <div class="w-full h-full p-6">
                <img
                  v-if="featuredArticle.featured_image"
                  :src="resolveAssetUrl(featuredArticle.featured_image)"
                  :alt="featuredTitle"
                  class="w-full h-full object-cover rounded"
                  loading="lazy"
                />
                <div
                  v-else
                  class="w-full h-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center rounded"
                >
                  <div class="w-16 h-16 text-gray-300 dark:text-gray-600">
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
            </div>
          </div>
        </a>
      </div>

      <!-- News Categories -->
      <div
        v-if="newsCategories.length === 0"
        class="text-center py-12"
      >
        <div class="max-w-md mx-auto">
          <h3 class="text-lg font-semibold text-gray-900 mb-2">
            News Temporarily Unavailable
          </h3>
          <p class="text-gray-600 text-sm">
            We're unable to load news information at the moment.
          </p>
        </div>
      </div>

      <div v-else class="space-y-8 lg:space-y-12">
        <div
          v-for="category in newsCategories"
          :key="category.title"
          class="news-category bg-white dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 p-4 lg:p-6"
        >
          <div class="mb-4 lg:mb-6">
            <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
              {{ category.title }}
            </h2>
          </div>

          <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
            <a
              v-for="(article, index) in category.articles.slice(0, 3)"
              :key="article.id"
              :href="`/company/news/${article.slug}`"
              :class="['block group news-card', index >= 2 ? 'hidden lg:block' : '']"
            >
              <div
                class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-colors duration-200 bg-white dark:bg-gray-800 h-full hover:bg-gray-50 dark:hover:bg-gray-700/50"
              >
                <div
                  class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-800"
                >
                  <img
                    v-if="article.featured_image"
                    :src="resolveAssetUrl(article.featured_image)"
                    :alt="article.title"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
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
                <div class="p-4 lg:p-5">
                  <div class="space-y-2 mb-4 lg:mb-6">
                    <h3
                      class="text-base lg:text-lg font-semibold text-gray-900 dark:text-white group-hover:text-gray-700 dark:group-hover:text-gray-200 line-clamp-2 transition-colors"
                    >
                      {{ article.title }}
                    </h3>
                    <p
                      v-if="article.excerpt || article.description"
                      class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2"
                    >
                      {{ article.excerpt || article.description }}
                    </p>
                  </div>
                  <div class="space-y-1">
                    <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                      <span>{{ article.author || 'International Center Team' }}</span>
                      <span>{{ formatDate(article.published_at) }}</span>
                    </div>
                    <div v-if="article.readingTime" class="text-xs text-gray-400 dark:text-gray-500">
                      {{ article.readingTime }}
                    </div>
                  </div>
                </div>
              </div>
            </a>

            <!-- Placeholder cards for missing content -->
            <div
              v-if="category.articles.length === 0"
              class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-all duration-200 bg-gray-50 dark:bg-gray-800/50 h-full"
            >
              <div
                class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
              >
                <div class="w-12 h-12 text-gray-300 dark:text-gray-600 opacity-50">
                  <svg fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              </div>
              <div class="p-4 lg:p-5">
                <div class="space-y-2 mb-4 lg:mb-6">
                  <div class="text-center text-gray-400 dark:text-gray-500">
                    <p class="text-sm font-medium">More articles coming soon</p>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-if="category.articles.length === 1"
              class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-all duration-200 bg-gray-50 dark:bg-gray-800/50 h-full hidden lg:block"
            >
              <div
                class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
              >
                <div class="w-12 h-12 text-gray-300 dark:text-gray-600 opacity-50">
                  <svg fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              </div>
              <div class="p-4 lg:p-5">
                <div class="space-y-2 mb-4 lg:mb-6">
                  <div class="text-center text-gray-400 dark:text-gray-500">
                    <p class="text-sm font-medium">More articles coming soon</p>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-if="category.articles.length === 2"
              class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-all duration-200 bg-gray-50 dark:bg-gray-800/50 h-full hidden lg:block"
            >
              <div
                class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
              >
                <div class="w-12 h-12 text-gray-300 dark:text-gray-600 opacity-50">
                  <svg fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              </div>
              <div class="p-4 lg:p-5">
                <div class="space-y-2 mb-4 lg:mb-6">
                  <div class="text-center text-gray-400 dark:text-gray-500">
                    <p class="text-sm font-medium">More articles coming soon</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { resolveAssetUrl } from '@/lib/utils/assets';

const newsCategories = ref<any[]>([]);
const featuredArticle = ref<any>(null);
const isLoading = ref(false);
const error = ref<string | null>(null);

// Computed properties for featured article
const featuredTitle = computed(() => featuredArticle.value?.title || '');
const featuredCategory = computed(() => 'Company News');
const featuredAuthor = computed(() => featuredArticle.value?.author || 'International Center Team');
const featuredDate = computed(() => {
  if (!featuredArticle.value) return '';
  return formatDate(featuredArticle.value.published_at);
});
const featuredArticleHref = computed(() => `/company/news/${featuredArticle.value?.slug || ''}`);

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

onMounted(async () => {
  try {
    isLoading.value = true;
    error.value = null;

    const { loadNewsPageData } = await import('../../lib/navigation-data');
    const newsData = await loadNewsPageData();
    console.log('🔍 [NewsHub] Client-side data loaded:', newsData);

    // Update reactive state with new data
    newsCategories.value = newsData.articleCategories;
    featuredArticle.value = newsData.featuredArticle;
  } catch (err: any) {
    console.error('❌ [NewsHub] Client-side data loading failed:', err);
    error.value = 'Failed to load news. Please try again later.';
    newsCategories.value = [];
    featuredArticle.value = null;
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.news-card:hover img {
  transform: scale(1.05);
}

.featured-card:hover h2 {
  color: rgb(55 65 81);
}

.dark .featured-card:hover h2 {
  color: rgb(209 213 219);
}
</style>