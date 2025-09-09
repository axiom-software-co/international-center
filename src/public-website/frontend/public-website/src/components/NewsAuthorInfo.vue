<template>
  <div :class="cn('rounded border bg-card p-6', className)">
    <h3 class="mb-4 text-lg font-semibold">About the Author</h3>
    <div class="flex items-start gap-4">
      <div class="flex-shrink-0">
        <img
          v-if="author.avatar"
          :src="author.avatar"
          :alt="`${author.name} avatar`"
          class="h-12 w-12 rounded-full object-cover"
        />
        <div
          v-else
          class="h-12 w-12 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center"
        >
          <User class="h-6 w-6 text-gray-500 dark:text-gray-400" />
        </div>
      </div>
      <div class="flex-1 min-w-0">
        <p class="font-medium text-gray-900 dark:text-white">{{ author.name }}</p>
        <p class="text-sm text-muted-foreground mb-3">{{ author.role }}</p>
        <div class="space-y-1 text-sm text-muted-foreground">
          <p>Published: {{ formattedDate }}</p>
          <p>Reading Time: {{ readingTime }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Calendar, User, Clock } from 'lucide-vue-next';
import { cn } from '@/lib/utils';



interface NewsAuthorInfoProps {
  author: {
    name: string;
    role: string;
    avatar?: string;
  };
  publishedDate: string;
  readingTime: string;
  className?: string;
}

const props = defineProps<NewsAuthorInfoProps>();

const formattedDate = computed(() => {
  return new Date(props.publishedDate).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
});
</script>
