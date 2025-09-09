<template>
  <div>
    <!-- Error State -->
    <div v-if="error" class="text-center py-12">
      <div class="max-w-md mx-auto">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Event Temporarily Unavailable</h3>
        <p class="text-gray-600 mb-4">
          We're experiencing technical difficulties. Please try again later.
        </p>
        <a
          href="/community/events"
          class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Browse All Events
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
              <div class="md:sticky md:top-4">
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
    <div v-else-if="eventData">
      <section class="pb-0">
        <EventBreadcrumb
          :eventName="eventData.title"
          :title="eventData.title"
          :category="eventData.category"
        />

        <div class="container article-page-container">
          <div class="mt-8 grid gap-12 md:grid-cols-12 md:gap-8">
            <div class="order-2 md:order-none md:col-span-7 md:col-start-1 lg:col-span-8">
              <article class="prose dark:prose-invert mx-auto">
                <div>
                  <img
                    :src="eventData.heroImage.src"
                    :alt="eventData.heroImage.alt"
                    class="mb-8 mt-0 aspect-video w-full rounded object-cover"
                  />
                </div>

                <EventContent :event="eventData" />
              </article>
            </div>

            <div class="order-1 md:order-none md:col-span-5 lg:col-span-4">
              <div class="md:sticky md:top-4">
                <aside id="event-page-aside">
                  <EventDetails
                    :eventDate="eventData.eventDetails.eventDate"
                    :eventTime="eventData.eventDetails.eventTime"
                    :location="eventData.eventDetails.location"
                    :capacity="eventData.eventDetails.capacity"
                    :registered="eventData.eventDetails.registered"
                    :status="eventData.eventDetails.status"
                  />

                  <EventContact class="mt-8" />
                </aside>
              </div>
            </div>
          </div>
        </div>

        <!-- Related Events Section -->
        <div v-if="!isLoading && !error && relatedEventsFiltered.length > 0" class="pt-16 lg:pt-20 pb-8 lg:pb-12">
          <div class="container">
            <div class="mb-4 lg:mb-6">
              <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
                {{ relatedEventsTitle }}
              </h2>
            </div>

            <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
              <EventCard
                v-for="(relatedEvent, index) in relatedEventsFiltered"
                :key="relatedEvent.id"
                :event="relatedEvent"
                :index="index"
              />
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
import { ref, computed } from 'vue';
import { useEvent, useEvents } from '@/composables/';
import { getEventSlugFromUrl } from '../../lib/utils/url';
import { generateEventImageUrl } from '../../lib/utils/content';
import EventBreadcrumb from './EventBreadcrumb.vue';
import EventContent from './EventContent.vue';
import EventDetails from './EventDetails.vue';
import EventContact from './EventContact.vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import EventCard from '../EventCard.vue';
import type { Event } from '../../lib/clients';

interface EventPageData {
  id: string;
  title: string;
  slug: string;
  description: string;
  content?: string;
  heroImage: {
    src: string;
    alt: string;
  };
  eventDetails: {
    eventDate: string;
    eventTime: string;
    location: string;
    capacity?: number;
    registered?: number;
    status: string;
  };
  category?: string;
}

// Use composables for data fetching
const currentSlug = ref(getEventSlugFromUrl());
const { event, loading: eventLoading, error: eventError } = useEvent(currentSlug);
const { events: relatedEvents, loading: relatedLoading, error: relatedError } = useEvents({
  category: computed(() => event.value?.category),
  pageSize: 3,
  enabled: computed(() => !!event.value?.category),
  immediate: false
});

// Computed loading and error states
const isLoading = computed(() => eventLoading.value);
const error = computed(() => eventError.value);
const relatedEventsFiltered = computed(() => 
  relatedEvents.value.filter(e => e.id !== event.value?.id)
);

// Transform event data to match the expected structure
const eventData = computed<EventPageData | null>(() => {
  if (!event.value) return null;
  
  return {
    id: event.value.id,
    title: event.value.title,
    slug: event.value.slug,
    description: event.value.excerpt || event.value.meta_description || '',
    content: event.value.content,
    heroImage: {
      src: generateEventImageUrl(event.value.featured_image, event.value.title),
      alt: event.value.title
    },
    eventDetails: {
      eventDate: event.value.event_date,
      eventTime: event.value.event_time,
      location: event.value.location,
      capacity: event.value.capacity,
      registered: undefined, // Add if available in your schema
      status: event.value.status
    },
    category: event.value.category
  };
});

const relatedEventsTitle = computed(() => {
  return eventData.value?.category ? `More ${eventData.value.category} Events` : 'Related Events';
});
</script>

<style scoped>
/* Component-specific styles if needed */
</style>