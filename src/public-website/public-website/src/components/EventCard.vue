<template>
  <a :href="eventHref" class="block group event-card">
    <div
      class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-all duration-200 bg-white dark:bg-gray-800 h-full hover:bg-gray-50 dark:hover:bg-gray-700/50"
    >
      <div
        class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-200 dark:bg-gray-600 relative"
      >
        <img
          v-if="event.featured_image"
          :src="event.featured_image"
          :alt="event.title"
          class="w-full h-full object-cover"
          loading="lazy"
        />
        <div
          v-else
          class="w-full h-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center"
        >
          <div class="w-12 h-12 text-gray-300 dark:text-gray-600">
            <svg fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
              <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
            </svg>
          </div>
        </div>
        
      </div>
      
      <div class="p-4 lg:p-5">
        <div class="mb-4 lg:mb-6">
          <h3
            class="text-sm lg:text-base font-semibold text-gray-900 dark:text-white line-clamp-2 group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors"
          >
            {{ event.title }}
          </h3>
        </div>
        
        <div class="space-y-2 mb-4">
          <div class="flex items-center text-xs text-gray-500 dark:text-gray-400">
            <svg class="w-3 h-3 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
            </svg>
            <span>{{ formatEventDate(event.date) }}</span>
          </div>
          
          <div class="flex items-center text-xs text-gray-500 dark:text-gray-400">
            <svg class="w-3 h-3 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{{ event.time }}</span>
          </div>
          
          <div class="flex items-center text-xs text-gray-500 dark:text-gray-400">
            <svg class="w-3 h-3 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
            </svg>
            <span class="line-clamp-1">{{ event.location }}</span>
          </div>
        </div>

        <!-- Capacity Info -->
        <div v-if="event.capacity && event.registered !== undefined" class="mt-4 pt-3 border-t border-gray-200 dark:border-gray-600">
          <div class="flex justify-between items-center text-xs text-gray-500 dark:text-gray-400">
            <span>{{ event.registered }}/{{ event.capacity }} registered</span>
            <div class="w-16 bg-gray-200 dark:bg-gray-600 rounded-full h-1.5">
              <div 
                class="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
                :style="{ width: `${Math.min((event.registered / event.capacity) * 100, 100)}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface CommunityEvent {
  id: string | number;
  title: string;
  slug: string;
  description: string;
  category: string;
  date: string;
  time: string;
  location: string;
  status: 'Open' | 'Registration Required' | 'Full' | 'Cancelled';
  featured_image?: string;
  capacity?: number;
  registered?: number;
}

interface Props {
  event: CommunityEvent;
  index?: number;
}

const props = defineProps<Props>();

const eventHref = computed(() => `/community/events/${props.event.slug}`);


const formatEventDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  });
};
</script>

<style scoped>
.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

</style>