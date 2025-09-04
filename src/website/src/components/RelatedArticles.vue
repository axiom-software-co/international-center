<template>
  <div v-if="items && items.length > 0" :class="cn('rounded border bg-card p-6', className)">
    <h3 class="mb-4 text-lg font-semibold">{{ title }}</h3>
    <div class="space-y-3">
      <template v-for="(item, index) in items" :key="item.title">
        <a :href="item.href || '#'" class="flex items-start gap-3 group">
          <ArrowRight class="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium group-hover:text-primary transition-colors">
              {{ item.title }}
            </p>
            <div class="flex items-center gap-2 mt-1">
              <p v-if="item.metadata.left" class="text-sm text-muted-foreground">
                {{ item.metadata.left }}
              </p>
              <span v-if="item.metadata.left && item.metadata.right" class="text-muted-foreground"
                >•</span
              >
              <p v-if="item.metadata.right" class="text-sm text-muted-foreground">
                {{ item.metadata.right }}
              </p>
            </div>
          </div>
        </a>
        <Separator v-if="index < items.length - 1" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next';
import Separator from '@/components/vue-ui/Separator.vue';
import { cn } from '@/lib/utils';



interface RelatedArticleItem {
  title: string;
  metadata: {
    left?: string;
    right?: string;
  };
  href?: string;
}

interface RelatedArticlesProps {
  title: string;
  items: RelatedArticleItem[];
  className?: string;
}

defineProps<RelatedArticlesProps>();
</script>
