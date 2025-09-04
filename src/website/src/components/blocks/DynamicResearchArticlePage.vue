<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Research Article Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/community/research"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Browse All Research
        </a>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="isLoading">
      <section class="pb-8">
        <!-- Breadcrumb Loading -->
        <div class="bg-gray-50 py-6">
          <div class="container">
            <div class="animate-pulse">
              <div class="h-4 bg-gray-300 rounded w-64 mb-4"></div>
              <div class="h-8 bg-gray-300 rounded w-96 mb-2"></div>
              <div class="h-4 bg-gray-300 rounded w-80"></div>
            </div>
          </div>
        </div>

        <div class="container article-page-container">
          <div class="mt-8 grid gap-12 md:grid-cols-12 md:gap-8">
            <!-- Main Content Loading -->
            <div class="order-2 md:order-none md:col-span-7 md:col-start-1 lg:col-span-8">
              <article class="prose dark:prose-invert mx-auto">
                <div class="animate-pulse">
                  <div class="mb-8 mt-0 aspect-video w-full rounded bg-gray-300"></div>
                  <div class="space-y-4">
                    <div class="h-8 bg-gray-300 rounded w-3/4"></div>
                    <div class="h-4 bg-gray-300 rounded w-full"></div>
                    <div class="h-4 bg-gray-300 rounded w-5/6"></div>
                    <div class="h-4 bg-gray-300 rounded w-4/5"></div>
                  </div>
                </div>
              </article>
            </div>

            <!-- Sidebar Loading -->
            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-20">
                <aside>
                  <div class="animate-pulse space-y-8">
                    <div class="bg-gray-200 rounded-lg p-6">
                      <div class="h-6 bg-gray-300 rounded w-3/4 mb-4"></div>
                      <div class="h-10 bg-gray-300 rounded w-full"></div>
                    </div>
                    <div class="bg-gray-200 rounded-lg p-6">
                      <div class="space-y-3">
                        <div class="h-4 bg-gray-300 rounded w-1/2"></div>
                        <div class="h-4 bg-gray-300 rounded w-3/4"></div>
                        <div class="h-4 bg-gray-300 rounded w-2/3"></div>
                      </div>
                    </div>
                  </div>
                </aside>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Content State -->
    <div v-else-if="articleData">
      <section class="pb-0">
        <ResearchBreadcrumb
          :articleName="articleData.title"
          :title="articleData.title"
          :category="articleData.category"
        />

        <div class="container article-page-container">
          <div class="mt-8 grid gap-12 md:grid-cols-12 md:gap-8">
            <div class="order-2 md:order-none md:col-span-7 md:col-start-1 lg:col-span-8">
              <article class="prose dark:prose-invert mx-auto">
                <div>
                  <img
                    :src="articleData.heroImage.src"
                    :alt="articleData.heroImage.alt"
                    class="mb-8 mt-0 aspect-video w-full rounded object-cover"
                  />
                </div>

                <ResearchArticleContent :article="articleData" />
              </article>
            </div>

            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-20">
                <aside id="research-article-page-aside">
                  <ResearchArticleDetails
                    :publishedAt="articleData.articleDetails.publishedAt"
                    :author="articleData.articleDetails.author"
                    :readTime="articleData.articleDetails.readTime"
                  />

                  <ContactCard class="mt-8" />
                </aside>
              </div>
            </div>
          </div>
        </div>

        <!-- Related Articles Section -->
        <div v-if="!isLoading && !error && relatedArticles.length > 0" class="pt-16 lg:pt-20 pb-8 lg:pb-12">
          <div class="container">
            <div class="mb-4 lg:mb-6">
              <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
                {{ relatedArticlesTitle }}
              </h2>
            </div>

            <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
              <ArticleCard
                v-for="(article, index) in relatedArticles"
                :key="article.id"
                :article="article"
                base-path="/community/research"
                default-author="International Center Research Team"
                :index="index"
              />
            </div>

            <!-- View All Button -->
            <div class="mt-8 text-center">
              <a
                href="/community/research"
                class="inline-block px-6 py-2 text-white bg-black hover:bg-gray-800 font-semibold rounded-sm transition-colors border border-black hover:border-gray-800"
              >
                View All Research
              </a>
            </div>
          </div>
        </div>

        <!-- CTA Section -->
        <div class="pt-0 pb-0">
          <UnifiedContentCTA />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import ResearchBreadcrumb from './ResearchBreadcrumb.vue';
import ResearchArticleContent from './ResearchArticleContent.vue';
import ResearchArticleDetails from './ResearchArticleDetails.vue';
import ContactCard from './ContactCard.vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import ArticleCard from '../ArticleCard.vue';
import { resolveAssetUrl } from '@/lib/utils/assets';

interface ResearchArticle {
  id: string;
  title: string;
  slug: string;
  description: string;
  content?: string;
  excerpt?: string;
  featured_image?: string;
  published_at: string;
  author?: string;
  category?: string;
  reading_time?: string;
}

interface ResearchArticlePageData {
  id: string;
  title: string;
  slug: string;
  description: string;
  content?: string;
  heroImage: {
    src: string;
    alt: string;
  };
  articleDetails: {
    publishedAt: string;
    author: string;
    readTime: string;
  };
  category?: string;
}

// Reactive state
const article = ref<ResearchArticle | null>(null);
const relatedArticles = ref<ResearchArticle[]>([]);
const isLoading = ref(false);
const error = ref<string | null>(null);
const relatedLoading = ref(false);

// Utility function to format dates
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

// Transform article data to match the expected structure
const articleData = computed((): ResearchArticlePageData | null => {
  if (!article.value) return null;
  
  return {
    id: article.value.id,
    title: article.value.title,
    slug: article.value.slug,
    description: article.value.description,
    content: article.value.content || article.value.description || article.value.excerpt,
    heroImage: {
      src: resolveAssetUrl(article.value.featured_image) || `https://placehold.co/1200x600/2563eb/ffffff/png?text=${encodeURIComponent(article.value.title)}`,
      alt: `${article.value.title} - International Center Research`,
    },
    articleDetails: {
      publishedAt: formatDate(article.value.published_at),
      author: article.value.author || 'International Center Research Team',
      readTime: article.value.reading_time || '8 min read',
    },
    category: article.value.category,
  };
});

// Dynamic title for related articles section
const relatedArticlesTitle = computed(() => {
  if (articleData.value?.category) {
    return `More from ${articleData.value.category}`;
  }
  return 'More Research Articles';
});

// Get slug from current URL
const getSlugFromUrl = (): string => {
  if (typeof window === 'undefined') return '';
  const pathParts = window.location.pathname.split('/');
  return pathParts[pathParts.length - 1] || '';
};

// Client-side data loading following services pattern
onMounted(async () => {
  try {
    isLoading.value = true;
    error.value = null;

    const slug = getSlugFromUrl();
    if (!slug) {
      throw new Error('Research article slug not found');
    }

    console.log(`🔍 [DynamicResearchArticlePage] Loading article: ${slug}`);

    // Dynamic import for client-side code splitting
    const { researchClient } = await import('../../lib/clients');
    
    const articleResult = await researchClient.getResearchArticleBySlug(slug);
    
    if (articleResult) {
      article.value = articleResult;
      console.log('✅ [DynamicResearchArticlePage] Article loaded:', articleResult);

      // Fetch related articles from same category (excluding current article)
      if (articleResult.category) {
        try {
          relatedLoading.value = true;
          console.log(`🔍 [DynamicResearchArticlePage] Loading related articles for category: ${articleResult.category}`);
          
          // Get all articles and filter client-side for more reliable results  
          const relatedResult = await researchClient.getResearchArticles({
            pageSize: 50, // Get more articles to filter from
            sortBy: 'date-desc'
          });
          
          console.log(`🔍 [DynamicResearchArticlePage] Related articles API response:`, relatedResult);
          
          if (relatedResult.data) {
            // Filter by category and exclude current article
            const filteredRelated = relatedResult.data
              .filter(relatedArticle => 
                relatedArticle.id !== articleResult.id && 
                relatedArticle.category === articleResult.category
              )
              .slice(0, 3);
            
            // If no articles in same category, show recent articles instead
            if (filteredRelated.length === 0) {
              const recentArticles = relatedResult.data
                .filter(relatedArticle => relatedArticle.id !== articleResult.id)
                .slice(0, 3);
              relatedArticles.value = recentArticles;
              console.log(`✅ [DynamicResearchArticlePage] No same-category articles, showing recent: ${recentArticles.length}`);
            } else {
              relatedArticles.value = filteredRelated;
              console.log(`✅ [DynamicResearchArticlePage] Related articles loaded: ${filteredRelated.length}`, filteredRelated);
            }
          } else {
            console.warn('⚠️ [DynamicResearchArticlePage] No data in related articles response');
          }
        } catch (relatedErr) {
          console.warn('⚠️ [DynamicResearchArticlePage] Failed to load related articles:', relatedErr);
          // Don't set error, just continue without related articles
        } finally {
          relatedLoading.value = false;
        }
      } else {
        console.log('ℹ️ [DynamicResearchArticlePage] No category found for current article, skipping related articles');
      }
    } else {
      throw new Error(`Research article not found: ${slug}`);
    }
  } catch (err: any) {
    console.error('❌ [DynamicResearchArticlePage] Failed to load article:', err.message);
    error.value = err.message || 'Failed to load research article data';
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
.article-page-container {
  overflow: visible !important;
}

.prose {
  max-width: none;
}
</style>