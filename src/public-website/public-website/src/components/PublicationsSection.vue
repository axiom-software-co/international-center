<template>
  <section class="pt-8 lg:pt-12 pb-0">
    <div class="container mx-auto px-4">
      <!-- Section Header -->
      <div class="mb-8 lg:mb-6">
        <div class="text-center max-w-3xl mx-auto mb-6 lg:mb-6">
          <h2 class="text-3xl lg:text-4xl font-bold">{{ title }}</h2>
        </div>

        <!-- Filters and Search -->
        <div class="flex flex-col lg:flex-row gap-4 lg:gap-6 lg:items-center lg:justify-between">
          <!-- Search Section -->
          <div class="flex flex-row items-center gap-3 order-1 sm:order-2 lg:order-2">
            <div class="relative flex-1 lg:w-80 lg:flex-none">
              <SearchIcon
                :class="`absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 ${
                  isSearching ? 'text-blue-500 animate-pulse' : 'text-muted-foreground'
                }`"
              />
              <Input
                type="text"
                :placeholder="`Search ${dataType === 'news' ? 'articles' : dataType === 'events' ? 'events' : 'research articles'}...`"
                v-model="searchQuery"
                :class="`pl-10 pr-10 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700 ${isSearching ? 'border-blue-300' : ''}`"
              />

              <!-- Right side icons -->
              <div class="absolute right-3 top-1/2 transform -translate-y-1/2">
                <div
                  v-if="isSearching"
                  class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"
                ></div>
                <button
                  v-else-if="searchQuery"
                  @click="searchQuery = ''"
                  class="flex items-center justify-center h-4 w-4 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
                  aria-label="Clear search"
                >
                  <XMarkIcon class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>

          <!-- Filters Section -->
          <div
            class="flex flex-col sm:flex-row gap-4 sm:items-center order-2 sm:order-1 lg:order-1"
          >
            <!-- Category Filter -->
            <Select v-model="activeCategory">
              <SelectTrigger
                class="w-full sm:w-auto sm:min-w-[120px] sm:max-w-[250px] touch-manipulation bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
              >
                <SelectValue placeholder="Select category" />
              </SelectTrigger>
              <SelectContent position="popper" side="bottom" align="start" :side-offset="5">
                <SelectItem v-for="category in categoryOptions" :key="category" :value="category">
                  {{ category }}
                </SelectItem>
              </SelectContent>
            </Select>

            <!-- Sort Filter -->
            <Select v-model="sortBy">
              <SelectTrigger
                class="w-full sm:w-auto sm:min-w-[110px] sm:max-w-[180px] touch-manipulation bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
              >
                <SelectValue placeholder="Sort by" />
              </SelectTrigger>
              <SelectContent position="popper" side="bottom" align="start" :side-offset="5">
                <SelectItem v-for="option in sortOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>

            <!-- Reset Filters -->
            <Button
              v-if="hasActiveFilters"
              variant="outline"
              @click="resetFilters"
              class="flex items-center gap-2 text-sm bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700 md:hover:bg-gray-50 md:dark:hover:bg-gray-700 transition-all duration-200 w-full sm:w-auto whitespace-nowrap touch-manipulation"
            >
              <XMarkIcon class="h-4 w-4" />
              <span>Reset Filters</span>
            </Button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="flex justify-center items-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span class="ml-3 text-muted-foreground">Loading articles...</span>
      </div>

      <!-- Error State -->
      <div
        v-else-if="error"
        class="bg-white dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm"
      >
        <div class="text-center py-12">
          <div class="max-w-md mx-auto">
            <h3 class="text-lg font-semibold text-gray-900 mb-2">
              Content Temporarily Unavailable
            </h3>
            <p class="text-gray-600 text-sm">We're unable to load content at the moment.</p>
          </div>
        </div>
      </div>

      <!-- Articles Table -->
      <div
        v-else
        ref="tableRef"
        class="bg-white dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm mt-6 lg:mt-3"
      >
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead
              class="bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700"
            >
              <tr>
                <th class="px-6 py-4 text-left">
                  <div
                    class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                  >
                    Articles
                  </div>
                </th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <ArticleTableRow
                v-for="article in currentArticles"
                :key="article.id"
                :article="article"
                :data-type="dataType"
              />
            </tbody>
          </table>
        </div>
      </div>

      <!-- Pagination - Hide when in search mode -->
      <div
        v-if="!isSearchMode && currentArticles.length > 0"
        class="mt-8 flex flex-col sm:flex-row items-center justify-between gap-4"
      >
        <!-- Page Size Selector -->
        <div class="flex items-center gap-2 whitespace-nowrap flex-shrink-0">
          <span class="text-sm text-gray-600 dark:text-gray-400">Show:</span>
          <Select v-model="pageSize">
            <SelectTrigger class="w-20 h-9 flex-shrink-0 touch-manipulation bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
              <SelectValue />
            </SelectTrigger>
            <SelectContent position="popper" side="bottom" align="center" :side-offset="5">
              <SelectItem value="10">10</SelectItem>
              <SelectItem value="50">50</SelectItem>
              <SelectItem value="100">100</SelectItem>
            </SelectContent>
          </Select>
          <span class="text-sm text-gray-600 dark:text-gray-400">per page</span>
        </div>

        <!-- Pagination Controls -->
        <div v-if="totalPages > 1" class="flex justify-end">
          <Pagination>
            <PaginationContent>
              <!-- Previous Button -->
              <PaginationItem>
                <Button
                  variant="outline"
                  size="default"
                  @click="handlePageChange(currentPage - 1)"
                  :disabled="currentPage <= 1"
                  class="gap-1 pl-2.5 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
                  aria-label="Go to previous page"
                >
                  <ChevronLeft class="h-4 w-4" />
                  <span>Previous</span>
                </Button>
              </PaginationItem>

              <!-- Page Numbers -->
              <PaginationItem v-for="pageNumber in visiblePages" :key="pageNumber">
                <Button
                  :variant="currentPage === pageNumber ? 'outline' : 'ghost'"
                  size="icon"
                  @click="handlePageChange(pageNumber)"
                  :aria-current="currentPage === pageNumber ? 'page' : undefined"
                  class="h-9 w-9 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
                >
                  {{ pageNumber }}
                </Button>
              </PaginationItem>

              <!-- Next Button -->
              <PaginationItem>
                <Button
                  variant="outline"
                  size="default"
                  @click="handlePageChange(currentPage + 1)"
                  :disabled="currentPage >= totalPages"
                  class="gap-1 pr-2.5 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
                  aria-label="Go to next page"
                >
                  <span>Next</span>
                  <ChevronRight class="h-4 w-4" />
                </Button>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { Search as SearchIcon, X as XMarkIcon, ChevronLeft, ChevronRight } from 'lucide-vue-next';
import Input from '@/components/vue-ui/Input.vue';
import Button from '@/components/vue-ui/Button.vue';
import Select from '@/components/vue-ui/Select.vue';
import SelectContent from '@/components/vue-ui/SelectContent.vue';
import SelectItem from '@/components/vue-ui/SelectItem.vue';
import SelectTrigger from '@/components/vue-ui/SelectTrigger.vue';
import SelectValue from '@/components/vue-ui/SelectValue.vue';
import Pagination from '@/components/vue-ui/Pagination.vue';
import PaginationContent from '@/components/vue-ui/PaginationContent.vue';
import PaginationItem from '@/components/vue-ui/PaginationItem.vue';
import ArticleTableRow from './ArticleTableRow.vue';
import type { NewsArticle, ResearchArticle, Event } from '@/lib/clients';

// Union type for articles
type Article = NewsArticle | ResearchArticle | Event;

interface PublicationsSectionProps {
  title?: string;
  dataType?: 'news' | 'research-articles' | 'events';
}

const props = withDefaults(defineProps<PublicationsSectionProps>(), {
  title: 'All Publications',
  dataType: 'news',
});

// UI state
const currentPage = ref(1);
const pageSize = ref('10');
const activeCategory = ref('All Categories');
const sortBy = ref('date-desc');
const searchQuery = ref('');
const tableRef = ref<HTMLDivElement>();

// Data state
const articles = ref<Article[]>([]);
const searchResults = ref<Article[]>([]);
const isLoading = ref(false);
const isSearching = ref(false);
const error = ref<string | null>(null);
const totalItems = ref(0);

// Categories from API data
const categories = ref<{name: string, slug: string}[]>([]);

// Fetch categories from API
const fetchCategories = async () => {
  try {
    if (props.dataType === 'news') {
      const { newsClient } = await import('@/lib/clients');
      const categoriesData = await newsClient.getNewsCategories();
      categories.value = categoriesData.map((cat: any) => ({ name: cat.name, slug: cat.slug }));
    } else if (props.dataType === 'events') {
      const { eventsClient } = await import('@/lib/clients');
      const categoriesData = await eventsClient.getEventCategories();
      categories.value = categoriesData.map((cat: any) => ({ name: cat.name, slug: cat.slug }));
    } else {
      const { researchClient } = await import('@/lib/clients');
      const categoriesData = await researchClient.getResearchCategories();
      categories.value = categoriesData.map((cat: any) => ({ name: cat.name, slug: cat.slug }));
    }
  } catch (err) {
    console.error('❌ [PublicationsSection] Error fetching categories:', err);
    categories.value = [];
  }
};

const categoryOptions = computed(() => {
  return ['All Categories', ...categories.value.map(cat => cat.name)];
});

const sortOptions = [
  { value: 'date-desc', label: 'Latest First' },
  { value: 'date-asc', label: 'Oldest First' },
  { value: 'title-asc', label: 'Title A-Z' },
  { value: 'title-desc', label: 'Title Z-A' },
];

// Computed properties
const allArticles = computed(() => {
  return searchQuery.value.trim().length >= 2 ? searchResults.value : articles.value;
});

const currentArticles = computed(() => {
  const start = (currentPage.value - 1) * parseInt(pageSize.value);
  const end = start + parseInt(pageSize.value);
  return allArticles.value.slice(start, end);
});

const isSearchMode = computed(() => searchQuery.value.trim().length >= 2);

const totalPages = computed(() => {
  return Math.ceil(allArticles.value.length / parseInt(pageSize.value));
});

const hasActiveFilters = computed(() => {
  return (
    activeCategory.value !== 'All Categories' ||
    sortBy.value !== 'date-desc' ||
    searchQuery.value.trim() !== ''
  );
});

const visiblePages = computed(() => {
  const maxVisible = 7;
  const pages: number[] = [];

  if (totalPages.value <= maxVisible) {
    for (let i = 1; i <= totalPages.value; i++) {
      pages.push(i);
    }
  } else {
    // Simplified pagination for now
    for (let i = 1; i <= Math.min(maxVisible, totalPages.value); i++) {
      pages.push(i);
    }
  }

  return pages;
});

// Methods
const handlePageChange = (page: number) => {
  currentPage.value = page;
  // Scroll to table section with navbar offset
  if (tableRef.value) {
    const elementTop = tableRef.value.getBoundingClientRect().top + window.pageYOffset;
    const navbarHeight = 120;
    const scrollToPosition = elementTop - navbarHeight;

    window.scrollTo({
      top: scrollToPosition,
      behavior: 'smooth',
    });
  }
};

const resetFilters = () => {
  activeCategory.value = 'All Categories';
  sortBy.value = 'date-desc';
  searchQuery.value = '';
  currentPage.value = 1;
};

// Helper function to get category slug from name
const getCategorySlug = (categoryName: string): string | undefined => {
  if (categoryName === 'All Categories') return undefined;
  const category = categories.value.find(cat => cat.name === categoryName);
  return category?.slug;
};

// Real API data fetching functions
const fetchArticles = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    console.log(`🔍 [PublicationsSection] Fetching ${props.dataType} articles...`);
    
    let allArticles: Article[] = [];
    const categorySlug = getCategorySlug(activeCategory.value);
    
    if (props.dataType === 'news') {
      const { newsClient } = await import('@/lib/clients');
      const response = await newsClient.getNewsArticles({ 
        pageSize: 1000, // Get all articles
        sortBy: sortBy.value as any,
        category: categorySlug
      });
      allArticles = response.data || [];
    } else if (props.dataType === 'events') {
      const { eventsClient } = await import('@/lib/clients');
      const response = await eventsClient.getEvents({ 
        pageSize: 1000, // Get all events
        sortBy: sortBy.value as any,
        category: categorySlug
      });
      allArticles = response.events || [];
    } else {
      const { researchClient } = await import('@/lib/clients');
      const response = await researchClient.getResearchArticles({ 
        pageSize: 1000, // Get all articles  
        sortBy: sortBy.value as any,
        category: categorySlug
      });
      allArticles = response.articles || [];
    }

    console.log(`✅ [PublicationsSection] Loaded ${allArticles.length} articles for category: ${categorySlug || 'all'}`);
    
    articles.value = allArticles;
    totalItems.value = allArticles.length;
    
  } catch (err) {
    error.value = 'Failed to load articles';
    console.error('❌ [PublicationsSection] Error fetching articles:', err);
  } finally {
    isLoading.value = false;
  }
};

const performSearch = async () => {
  if (!searchQuery.value || searchQuery.value.trim().length < 2) {
    searchResults.value = [];
    isSearching.value = false;
    return;
  }

  isSearching.value = true;

  try {
    // Mock search
    await new Promise(resolve => setTimeout(resolve, 300));

    const searchTerms = searchQuery.value
      .toLowerCase()
      .split(' ')
      .filter(term => term.length > 0);
    const results = articles.value.filter((article: Article) => {
      const searchText =
        `${article.title} ${article.excerpt || ''} ${article.category}`.toLowerCase();
      return searchTerms.some(term => searchText.includes(term));
    });

    const filteredResults =
      activeCategory.value === 'All Categories'
        ? results
        : results.filter((article: Article) => article.category === activeCategory.value);

    searchResults.value = filteredResults;
  } catch (err) {
    console.error('Search failed:', err);
    searchResults.value = [];
  } finally {
    isSearching.value = false;
  }
};

// Watchers
watch([currentPage, pageSize, activeCategory, sortBy], () => {
  fetchArticles();
});

watch(
  searchQuery,
  () => {
    currentPage.value = 1;
    performSearch();
  },
  { debounce: 300 }
);

watch(pageSize, () => {
  currentPage.value = 1;
});

// Lifecycle
onMounted(async () => {
  await fetchCategories();
  await fetchArticles();
});
</script>
