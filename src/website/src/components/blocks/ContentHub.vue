<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ config.errorTitle }} Temporarily Unavailable</h3>
        <p class="text-gray-600 text-sm">
          We're unable to load {{ config.errorTitle.toLowerCase() }} information at the moment.
        </p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="space-y-8 lg:space-y-12">
      <!-- Featured Article Skeleton -->
      <div
        class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded overflow-hidden"
      >
        <div class="flex flex-col lg:flex-row">
          <div class="flex-1 p-6 md:p-8 lg:p-10 xl:p-12 flex flex-col justify-center">
            <div class="space-y-4">
              <div class="h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-24"></div>
              <div class="h-8 bg-gray-300 dark:bg-gray-600 rounded animate-pulse w-3/4"></div>
              <div v-if="config.showExcerpt" class="h-6 bg-gray-300 dark:bg-gray-600 rounded animate-pulse"></div>
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
        class="content-category bg-white dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 p-4 lg:p-6"
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

    <!-- Content -->
    <div v-if="!isLoading && !error" class="space-y-8 lg:space-y-12">
      <!-- Featured Article Section -->
      <div
        v-if="featuredArticle"
        class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded overflow-hidden"
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
                <p v-if="config.showExcerpt && featuredExcerpt" class="text-gray-600 dark:text-gray-300">
                  {{ featuredExcerpt }}
                </p>
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

      <!-- Content Categories -->
      <div
        v-for="category in articleCategories"
        :key="category.title"
        class="content-category bg-white dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 p-4 lg:p-6"
      >
        <div class="mb-4 lg:mb-6">
          <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
            {{ category.title }}
          </h2>
        </div>

        <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
          <ArticleCard
            v-for="(article, index) in category.articles.slice(0, 3)"
            :key="article.id"
            :article="article"
            :base-path="config.basePath"
            :default-author="config.defaultAuthor"
            :index="index"
          />

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

      <!-- Publications Section (hidden for events) -->
      <PublicationsSection
        v-if="config.type !== 'events'"
        :title="`All ${config.errorTitle} Publications`"
        :data-type="config.type === 'news' ? 'news' : 'research-articles'"
      />

      <!-- CTA Section -->
      <div class="pt-8 lg:pt-12">
        <UnifiedContentCTA />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import PublicationsSection from '../PublicationsSection.vue';
import ArticleCard from '../ArticleCard.vue';
import { resolveAssetUrl } from '@/lib/utils/assets';

interface ContentArticle {
  id: string | number;
  title: string;
  slug: string;
  excerpt?: string;
  description?: string;
  category?: {
    id?: string | number;
    name?: string;
    description?: string;
  } | string;
  featured_image?: string;
  published_at: string;
  author?: string;
  readingTime?: string;
  featured?: boolean;
}

interface ContentCategory {
  id?: string | number;
  title: string;
  description: string;
  articles: ContentArticle[];
}

interface ContentHubConfig {
  type: 'news' | 'research' | 'events';
  basePath: string;
  errorTitle: string;
  defaultAuthor: string;
  defaultCategory: string;
  showExcerpt: boolean;
  clientMethod: string;
  categoriesMethod: string;
}

interface Props {
  contentType: string;
}

const props = defineProps<Props>();

// Create config based on contentType
const config = computed(() => {
  if (props.contentType === 'news') {
    return {
      type: 'news' as const,
      basePath: '/company/news',
      errorTitle: 'News',
      defaultAuthor: 'International Center Team',
      defaultCategory: 'News',
      showExcerpt: false,
      clientMethod: 'getNewsArticles',
      categoriesMethod: 'getNewsCategories',
    };
  } else if (props.contentType === 'events') {
    return {
      type: 'events' as const,
      basePath: '/community/events',
      errorTitle: 'Events',
      defaultAuthor: 'International Center Team',
      defaultCategory: 'Events',
      showExcerpt: true,
      clientMethod: 'getEvents',
      categoriesMethod: 'getEventCategories',
    };
  } else {
    return {
      type: 'research' as const,
      basePath: '/community/research',
      errorTitle: 'Research',
      defaultAuthor: 'International Center Team',
      defaultCategory: 'Research',
      showExcerpt: false,
      clientMethod: 'getResearchArticles',
      categoriesMethod: 'getResearchCategories',
    };
  }
});

// Reactive state
const articleCategories = ref<ContentCategory[]>([]);
const featuredArticle = ref<ContentArticle | null>(null);
const isLoading = ref(false);
const error = ref<string | null>(null);

// Computed properties for featured article
const featuredTitle = computed(() => featuredArticle.value?.title || '');
const featuredExcerpt = computed(() => 
  featuredArticle.value?.description || featuredArticle.value?.excerpt || ''
);
const featuredCategory = computed(() => {
  const article = featuredArticle.value;
  if (!article) return config.value.defaultCategory;
  return typeof article.category === 'object' ? article.category?.name : article.category || config.value.defaultCategory;
});
const featuredAuthor = computed(() => featuredArticle.value?.author || config.value.defaultAuthor);
const featuredDate = computed(() => {
  if (!featuredArticle.value) return '';
  return formatDate(featuredArticle.value.published_at);
});
const featuredArticleHref = computed(() => `${config.value.basePath}/${featuredArticle.value?.slug || ''}`);

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

onMounted(async () => {
  console.log(`🚀 [ContentHub] Component mounted`);
  console.log(`🔍 [ContentHub] Props received:`, props);
  console.log(`🔍 [ContentHub] Config:`, config.value);
  
  if (!props.contentType) {
    console.error(`❌ [ContentHub] No contentType prop received!`);
    error.value = 'Component configuration missing';
    return;
  }
  
  try {
    isLoading.value = true;
    error.value = null;

    console.log(`🔄 [ContentHub] Starting data load for ${config.value.type}...`);

    // Test basic API connectivity first
    if (config.value.type === 'news') {
      console.log('🔍 [ContentHub] Testing news API...');
      const { newsClient } = await import('../../lib/clients');
      console.log('📦 [ContentHub] News client imported');
      
      const pageData = await newsClient.loadNewsPageData();
      console.log('📊 [ContentHub] News page data:', pageData);
      
      articleCategories.value = pageData.articleCategories || [];
      featuredArticle.value = pageData.featuredArticle || null;
    } else if (config.value.type === 'events') {
      console.log('🔍 [ContentHub] Testing events API...');
      const { eventsClient } = await import('../../lib/clients');
      console.log('📦 [ContentHub] Events client imported');
      
      const pageData = await eventsClient.loadEventsPageData();
      console.log('📊 [ContentHub] Events page data:', pageData);
      
      articleCategories.value = pageData.articleCategories || [];
      featuredArticle.value = pageData.featuredArticle || null;
    } else {
      console.log('🔍 [ContentHub] Testing research API...');
      const { researchClient } = await import('../../lib/clients');
      console.log('📦 [ContentHub] Research client imported');
      
      const pageData = await researchClient.loadResearchPageData();
      console.log('📊 [ContentHub] Research page data:', pageData);
      
      articleCategories.value = pageData.articleCategories || [];
      featuredArticle.value = pageData.featuredArticle || null;
    }

    console.log(`✅ [ContentHub] Final - Categories: ${articleCategories.value.length}, Featured: ${featuredArticle.value?.title || 'None'}`);
  } catch (err: any) {
    console.error(`❌ [ContentHub] ERROR:`, err);
    error.value = err.message || `Failed to load ${config.value.type} content`;
    articleCategories.value = [];
    featuredArticle.value = null;
  } finally {
    isLoading.value = false;
    console.log(`🏁 [ContentHub] Done - isLoading: ${isLoading.value}`);
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

.content-card:hover img {
  transform: scale(1.05);
}

.featured-card:hover h2 {
  color: rgb(55 65 81);
}

.dark .featured-card:hover h2 {
  color: rgb(209 213 219);
}
</style>