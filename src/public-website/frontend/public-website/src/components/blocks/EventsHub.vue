<template>
  <div>
    <!-- Featured Event Section -->
    <div
      v-if="featuredEvent"
      class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded overflow-hidden mb-8 lg:mb-12"
    >
      <a :href="featuredEventHref" class="block group featured-card">
        <div class="flex flex-col lg:flex-row">
          <!-- Content - Left Side -->
          <div class="flex-1 p-6 md:p-8 lg:p-10 xl:p-12 flex flex-col justify-center">
            <div class="space-y-4">
              <p
                class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide"
              >
                {{ featuredEvent.category }}
              </p>
              <h2
                class="text-2xl md:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white leading-tight group-hover:text-gray-700 dark:group-hover:text-gray-300 transition-colors"
              >
                {{ featuredEvent.title }}
              </h2>
              <div class="space-y-2">
                <div class="flex items-center gap-3 text-gray-500 dark:text-gray-400">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
                  </svg>
                  <span class="text-sm font-medium">{{ formatEventDate(featuredEvent.date) }}</span>
                  <span class="text-gray-400 dark:text-gray-500">•</span>
                  <span class="text-sm">{{ featuredEvent.time }}</span>
                </div>
                <div class="flex items-center gap-3 text-gray-500 dark:text-gray-400">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  </svg>
                  <span class="text-sm">{{ featuredEvent.location }}</span>
                </div>
              </div>
            </div>
          </div>
          <!-- Image - Right Side -->
          <div class="lg:w-1/2 aspect-[16/9] lg:aspect-auto">
            <div class="w-full h-full p-6">
              <img
                v-if="featuredEvent.featured_image"
                :src="featuredEvent.featured_image"
                :alt="featuredEvent.title"
                class="w-full h-full object-cover rounded"
                loading="lazy"
              />
              <div
                v-else
                class="w-full h-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center rounded"
              >
                <div class="w-16 h-16 text-gray-300 dark:text-gray-600">
                  <svg fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                    <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </a>
    </div>

    <!-- Event Categories -->
    <div
      v-for="category in eventCategories"
      :key="category.title"
      class="content-category bg-white dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700 p-4 lg:p-6 mb-8 lg:mb-12"
    >
      <div class="mb-4 lg:mb-6">
        <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
          {{ category.title }}
        </h2>
      </div>

      <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
        <EventCard
          v-for="(event, index) in category.events.slice(0, 3)"
          :key="event.id"
          :event="event"
          :index="index"
        />

        <!-- Placeholder cards for missing content -->
        <div
          v-if="category.events.length === 0"
          class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden transition-all duration-200 bg-gray-50 dark:bg-gray-800/50 h-full"
        >
          <div
            class="aspect-[3/2] lg:aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
          >
            <div class="w-12 h-12 text-gray-300 dark:text-gray-600 opacity-50">
              <svg fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
              </svg>
            </div>
          </div>
          <div class="p-4 lg:p-5">
            <div class="space-y-2 mb-4 lg:mb-6">
              <div class="text-center text-gray-400 dark:text-gray-500">
                <p class="text-sm font-medium">More events coming soon</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Newsletter CTA Section -->
    <div class="pt-8 lg:pt-12">
      <UnifiedContentCTA />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import EventCard from '../EventCard.vue';

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
  featured?: boolean;
}

interface EventCategory {
  id?: string | number;
  title: string;
  events: CommunityEvent[];
}

// Static data for demonstration
const featuredEvent = ref<CommunityEvent>({
  id: 'featured-1',
  title: 'Regenerative Medicine: Future of Healthcare',
  slug: 'regenerative-medicine-future-healthcare',
  description: 'Join leading researchers and clinicians for an in-depth exploration of cutting-edge regenerative medicine techniques and their potential to transform patient care.',
  category: 'Research Presentation',
  date: '2025-09-15',
  time: '2:00 PM - 4:00 PM',
  location: 'International Center Main Auditorium',
  status: 'Registration Required',
  featured_image: 'https://placehold.co/800x600/000000/FFFFFF/png?text=Featured+Event',
  capacity: 150,
  registered: 89,
  featured: true
});

const eventCategories = ref<EventCategory[]>([
  {
    id: 'educational-workshops',
    title: 'Educational Workshops',
    events: [
      {
        id: 'workshop-1',
        title: 'Understanding Stem Cell Therapy',
        slug: 'understanding-stem-cell-therapy',
        description: 'Learn about the science behind stem cell therapy and its applications in modern medicine.',
        category: 'Educational Workshop',
        date: '2025-09-22',
        time: '10:00 AM - 12:00 PM',
        location: 'Education Center Room 101',
        status: 'Open',
        capacity: 30,
        registered: 18,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=Stem+Cell+Workshop'
      },
      {
        id: 'workshop-2',
        title: 'Nutrition for Optimal Healing',
        slug: 'nutrition-optimal-healing',
        description: 'Discover how proper nutrition can enhance your body\'s natural healing processes.',
        category: 'Educational Workshop',
        date: '2025-09-29',
        time: '1:00 PM - 3:00 PM',
        location: 'Community Health Center',
        status: 'Registration Required',
        capacity: 25,
        registered: 25,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=Nutrition+Workshop'
      },
      {
        id: 'workshop-3',
        title: 'PRP Therapy Explained',
        slug: 'prp-therapy-explained',
        description: 'Comprehensive overview of Platelet-Rich Plasma therapy and patient success stories.',
        category: 'Educational Workshop',
        date: '2025-10-06',
        time: '11:00 AM - 1:00 PM',
        location: 'Medical Training Facility',
        status: 'Open',
        capacity: 40,
        registered: 12,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=PRP+Therapy'
      }
    ]
  },
  {
    id: 'community-health',
    title: 'Community Health Seminars',
    events: [
      {
        id: 'health-1',
        title: 'Preventive Care in the Modern Age',
        slug: 'preventive-care-modern-age',
        description: 'Learn about preventive healthcare strategies and early intervention techniques.',
        category: 'Community Health',
        date: '2025-09-20',
        time: '6:00 PM - 8:00 PM',
        location: 'Community Center Auditorium',
        status: 'Open',
        capacity: 100,
        registered: 45,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=Preventive+Care'
      },
      {
        id: 'health-2',
        title: 'Mental Health and Physical Wellness',
        slug: 'mental-health-physical-wellness',
        description: 'Explore the connection between mental health and physical recovery.',
        category: 'Community Health',
        date: '2025-10-04',
        time: '7:00 PM - 9:00 PM',
        location: 'Wellness Center Main Hall',
        status: 'Registration Required',
        capacity: 75,
        registered: 62,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=Mental+Health'
      }
    ]
  },
  {
    id: 'wellness-programs',
    title: 'Wellness Programs',
    events: [
      {
        id: 'wellness-1',
        title: 'Weekly Wellness Circle',
        slug: 'weekly-wellness-circle',
        description: 'Join our supportive community for weekly discussions on health and wellness topics.',
        category: 'Wellness Program',
        date: '2025-09-18',
        time: '5:30 PM - 6:30 PM',
        location: 'Wellness Center Circle Room',
        status: 'Open',
        capacity: 20,
        registered: 14,
        featured_image: 'https://placehold.co/600x400/000000/FFFFFF/png?text=Wellness+Circle'
      }
    ]
  }
]);

// Computed properties
const featuredEventHref = computed(() => `/community/events/${featuredEvent.value?.slug || ''}`);

const formatEventDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};
</script>

<style scoped>
.featured-card:hover h2 {
  color: rgb(55 65 81);
}

.dark .featured-card:hover h2 {
  color: rgb(209 213 219);
}
</style>