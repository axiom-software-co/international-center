<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Article Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/company/news"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Browse All News
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
        <NewsBreadcrumb
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

                <NewsArticleContent :article="articleData" />
              </article>
            </div>

            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-20">
                <aside id="news-article-page-aside">
                  <NewsArticleDetails
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
                base-path="/company/news"
                default-author="International Center Team"
                :index="index"
              />
            </div>

            <!-- View All Button -->
            <div class="mt-8 text-center">
              <a
                href="/company/news"
                class="inline-block px-6 py-2 text-white bg-black hover:bg-gray-800 font-semibold rounded-sm transition-colors border border-black hover:border-gray-800"
              >
                View All News
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
import { computed } from 'vue';
import NewsBreadcrumb from './NewsBreadcrumb.vue';
import NewsArticleContent from './NewsArticleContent.vue';
import NewsArticleDetails from './NewsArticleDetails.vue';
import ContactCard from './ContactCard.vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import ArticleCard from '../ArticleCard.vue';
import { resolveAssetUrl } from '@/lib/utils/assets';
import { formatArticleDate, getDisplayDate } from '@/lib/utils/date';
import { getNewsSlugFromUrl } from '@/lib/utils/url';
import { calculateReadingTime, generateHeroImageUrl, generateImageAlt, formatAuthorName } from '@/lib/utils/content';
import { useNewsArticle, useFeaturedNews } from '@/composables/';
import type { NewsArticle } from '@/lib/clients/news/types';

interface NewsArticlePageData {
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

// Use composables for data fetching
const slug = computed(() => getNewsSlugFromUrl());
const { article, loading: isLoading, error } = useNewsArticle(slug);
const { articles: relatedArticles, loading: relatedLoading } = useFeaturedNews(3);

// Transform article data to match the expected structure
const articleData = computed((): NewsArticlePageData | null => {
  if (!article.value) return null;
  
  const displayDate = getDisplayDate(article.value.published_at, article.value.created_on);
  const content = article.value.content || article.value.summary;
  const resolvedImageUrl = resolveAssetUrl(article.value.image_url);
  
  return {
    id: article.value.news_id,
    title: article.value.title,
    slug: article.value.slug,
    description: article.value.summary,
    content,
    heroImage: {
      src: generateHeroImageUrl(resolvedImageUrl, article.value.title, 'news'),
      alt: generateImageAlt(article.value.title, 'news'),
    },
    articleDetails: {
      publishedAt: formatArticleDate(displayDate),
      author: formatAuthorName(article.value.author_name),
      readTime: calculateReadingTime(content),
    },
    category: article.value.category_id,
  };
});

// Dynamic title for related articles section
const relatedArticlesTitle = computed(() => {
  const count = relatedArticles.value?.length || 0;
  
  if (count === 0) return 'More News Articles';
  if (count === 1) return 'Related Article';
  
  return 'More News Articles';
});
</script>

<style scoped>
.article-page-container {
  overflow: visible !important;
}
</style>
