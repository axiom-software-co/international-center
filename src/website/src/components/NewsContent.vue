<template>
  <article class="news-content" role="article">
    <header class="news-header space-y-4">
      <div class="flex flex-wrap gap-2">
        <Badge
          v-if="showFeatured && article.featured"
          variant="outline"
          class="featured-badge px-3 py-1 text-sm font-medium border border-blue-300 bg-blue-50 text-blue-700 rounded"
        >
          Featured
        </Badge>
      </div>

      <h1 class="news-title text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white leading-tight">
        {{ article.title }}
      </h1>
      
      <p class="news-summary text-lg text-gray-600 dark:text-gray-300 leading-relaxed font-medium">
        {{ article.summary }}
      </p>
    </header>

    <div v-if="article.content" class="news-content-body prose dark:prose-invert max-w-none mt-8">
      <div v-html="sanitizedContent" class="content-html"></div>
    </div>

    <aside v-if="showAuthor || showPublishedDate" class="news-metadata bg-gray-50 dark:bg-gray-800 rounded-lg p-6 mt-8">
      <div v-if="showAuthor && article.author_name" class="news-author mb-2">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          <strong>Author:</strong> {{ article.author_name }}
        </p>
      </div>
      
      <div v-if="showPublishedDate && (article.published_at || article.created_on)" class="news-date">
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Published: {{ formatDate(article.published_at || article.created_on) }}
        </p>
      </div>
    </aside>

    <div v-if="showImage && article.image_url" class="news-image mt-8">
      <img 
        :src="article.image_url" 
        :alt="article.title" 
        loading="lazy"
        class="w-full h-auto rounded-lg shadow-lg"
      />
    </div>

    <div v-if="showCTA" class="news-cta bg-gray-50 dark:bg-gray-800 rounded-lg p-6 text-center mt-8">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        Interested in Learning More?
      </h3>
      <p class="text-gray-600 dark:text-gray-300 mb-4">
        Discover how our healthcare services can help you achieve optimal health and wellness.
      </p>
      <div class="flex flex-col sm:flex-row gap-3 justify-center">
        <a
          href="/services"
          class="inline-flex items-center justify-center px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
        >
          Explore Our Services
        </a>
        <a
          href="/company/contact"
          class="inline-flex items-center justify-center px-6 py-3 border border-gray-300 hover:border-gray-400 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors"
        >
          Contact Us
        </a>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import Badge from '@/components/vue-ui/Badge.vue';
import type { NewsArticle } from '@/lib/clients/news/types';

interface NewsContentProps {
  article: NewsArticle;
  showAuthor?: boolean;
  showPublishedDate?: boolean;
  showFeatured?: boolean;
  showImage?: boolean;
  showCTA?: boolean;
}

const props = withDefaults(defineProps<NewsContentProps>(), {
  showAuthor: true,
  showPublishedDate: true,
  showFeatured: true,
  showImage: true,
  showCTA: true,
});

// HTML sanitization to prevent XSS attacks
const sanitizedContent = computed(() => {
  if (!props.article.content) return '';
  
  // Basic sanitization - remove script tags and other dangerous elements
  const allowedTags = [
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'p', 'br', 'strong', 'b', 'em', 'i', 'u',
    'ul', 'ol', 'li', 'blockquote', 'a',
    'img', 'div', 'span'
  ];
  
  // Remove script tags and other dangerous elements
  let content = props.article.content
    .replace(/<script[^>]*>.*?<\/script>/gsi, '') // Remove script tags
    .replace(/<iframe[^>]*>.*?<\/iframe>/gsi, '') // Remove iframe tags
    .replace(/javascript:/gi, '') // Remove javascript: URLs
    .replace(/on\w+="[^"]*"/gi, '') // Remove event handlers like onclick
    .replace(/on\w+='[^']*'/gi, ''); // Remove event handlers with single quotes
  
  // Create a temporary DOM element for further sanitization
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = content;
  
  // Remove any remaining dangerous attributes
  const allElements = tempDiv.querySelectorAll('*');
  allElements.forEach(el => {
    // Remove dangerous attributes
    const dangerousAttrs = ['onload', 'onerror', 'onclick', 'onmouseover', 'onfocus', 'onblur'];
    dangerousAttrs.forEach(attr => {
      if (el.hasAttribute(attr)) {
        el.removeAttribute(attr);
      }
    });
    
    // Only allow specific tags
    if (!allowedTags.includes(el.tagName.toLowerCase())) {
      // Replace disallowed tags with their content
      const parent = el.parentNode;
      if (parent) {
        while (el.firstChild) {
          parent.insertBefore(el.firstChild, el);
        }
        parent.removeChild(el);
      }
    }
  });
  
  return tempDiv.innerHTML;
});

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};
</script>

<style scoped>
.news-content {
  max-width: 800px;
  margin: 0 auto;
  line-height: 1.6;
}

.news-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #1f2937;
}

.news-summary {
  font-size: 1.25rem;
  color: #6b7280;
  margin-bottom: 2rem;
}

.content-html :deep(h2) {
  font-size: 1.875rem;
  font-weight: 600;
  margin: 2rem 0 1rem;
  color: #374151;
}

.content-html :deep(h3) {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 1.5rem 0 0.75rem;
  color: #374151;
}

.content-html :deep(p) {
  margin-bottom: 1rem;
  color: #4b5563;
}

.content-html :deep(ul), 
.content-html :deep(ol) {
  margin: 1rem 0;
  padding-left: 1.5rem;
}

.content-html :deep(li) {
  margin-bottom: 0.5rem;
  color: #4b5563;
}

.content-html :deep(blockquote) {
  border-left: 4px solid #e5e7eb;
  margin: 1.5rem 0;
  padding-left: 1rem;
  font-style: italic;
  color: #6b7280;
}

.featured-badge {
  display: inline-block;
}

.news-image img {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}

@media (prefers-color-scheme: dark) {
  .news-title {
    color: #f9fafb;
  }
  
  .content-html :deep(h2),
  .content-html :deep(h3) {
    color: #e5e7eb;
  }
  
  .content-html :deep(p),
  .content-html :deep(li) {
    color: #d1d5db;
  }
  
  .content-html :deep(blockquote) {
    border-left-color: #4b5563;
    color: #9ca3af;
  }
}
</style>
