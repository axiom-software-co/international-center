<template>
  <div v-if="topics && topics.length > 0" :class="cn('rounded border bg-card p-6', className)">
    <h3 class="mb-4 text-lg font-semibold">Related Topics</h3>
    <div class="space-y-3">
      <template v-for="(topic, index) in topics" :key="topic.title">
        <div class="group cursor-pointer">
          <div class="flex items-start gap-3">
            <ArrowRight class="h-4 w-4 text-muted-foreground mt-0.5 transition-colors" />
            <div class="flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-white transition-colors">
                {{ topic.title }}
              </p>
              <p class="text-sm text-muted-foreground mt-1">{{ topic.description }}</p>
            </div>
          </div>
        </div>
        <Separator v-if="index < topics.length - 1" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Tag } from 'lucide-vue-next';
import Separator from '@/components/vue-ui/Separator.vue';
import { cn } from '@/lib/utils';



interface RelatedTopic {
  title: string;
  description: string;
}

interface NewsRelatedTopicsProps {
  topics: RelatedTopic[];
  className?: string;
}

defineProps<NewsRelatedTopicsProps>();
</script>
