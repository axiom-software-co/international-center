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
        <div v-if="!isLoading && !error && relatedEvents.length > 0" class="pt-16 lg:pt-20 pb-8 lg:pb-12">
          <div class="container">
            <div class="mb-4 lg:mb-6">
              <h2 class="text-xl lg:text-2xl font-semibold text-gray-900 dark:text-white">
                {{ relatedEventsTitle }}
              </h2>
            </div>

            <div class="grid gap-4 md:gap-6 lg:gap-8 md:grid-cols-2 lg:grid-cols-3">
              <EventCard
                v-for="(event, index) in relatedEvents"
                :key="event.id"
                :event="event"
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
import { ref, computed, onMounted } from 'vue';
import EventBreadcrumb from './EventBreadcrumb.vue';
import EventContent from './EventContent.vue';
import EventDetails from './EventDetails.vue';
import EventContact from './EventContact.vue';
import UnifiedContentCTA from '../UnifiedContentCTA.vue';
import EventCard from '../EventCard.vue';

interface CommunityEvent {
  id: string;
  title: string;
  slug: string;
  description: string;
  content?: string;
  category: string;
  date: string;
  time: string;
  location: string;
  status: 'Open' | 'Registration Required' | 'Full' | 'Cancelled';
  featured_image?: string;
  capacity?: number;
  registered?: number;
}

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

// Reactive state
const event = ref<CommunityEvent | null>(null);
const relatedEvents = ref<CommunityEvent[]>([]);
const isLoading = ref(false);
const error = ref<string | null>(null);
const relatedLoading = ref(false);

// Removed static data - now using API

// Computed properties - Transform API event data for template
const eventData = computed<EventPageData | null>(() => {
  if (!event.value) return null;
  
  return {
    id: event.value.id,
    title: event.value.title,
    slug: event.value.slug,
    description: event.value.description,
    content: event.value.content,
    heroImage: {
      src: event.value.featured_image || 'https://placehold.co/800x600/e5e7eb/6b7280/png?text=Event+Image',
      alt: event.value.title
    },
    eventDetails: {
      eventDate: event.value.date,
      eventTime: event.value.time,
      location: event.value.location,
      capacity: event.value.capacity,
      registered: event.value.registered,
      status: event.value.status
    },
    category: event.value.category
  };
});

const relatedEventsTitle = computed(() => {
  return eventData.value?.category ? `More ${eventData.value.category} Events` : 'Related Events';
});

// Get slug from current URL
const getSlugFromUrl = (): string => {
  if (typeof window === 'undefined') return '';
  const pathParts = window.location.pathname.split('/');
  return pathParts[pathParts.length - 1] || '';
};

// Client-side data loading following API pattern
onMounted(async () => {
  try {
    isLoading.value = true;
    error.value = null;

    const slug = getSlugFromUrl();
    if (!slug) {
      throw new Error('Event slug not found');
    }

    console.log(`🔍 [DynamicEventPage] Loading event: ${slug}`);

    // Dynamic import for client-side code splitting
    const { eventsClient } = await import('../../lib/clients');
    
    const eventResult = await eventsClient.getEventBySlug(slug);
    
    if (eventResult) {
      // Transform API event data to match component interface
      event.value = {
        id: eventResult.id.toString(),
        title: eventResult.title,
        slug: eventResult.slug,
        description: eventResult.excerpt || eventResult.meta_description || '',
        content: eventResult.content || '',
        category: eventResult.category || 'Event',
        date: eventResult.event_date,
        time: eventResult.event_time || '',
        location: eventResult.location || '',
        status: 'Open', // Default status, adjust based on your event schema
        featured_image: eventResult.featured_image || '',
        capacity: eventResult.capacity || undefined,
        registered: undefined // Add if available in your schema
      };
      
      console.log('✅ [DynamicEventPage] Event loaded:', eventResult);

      // Fetch related events from same category (excluding current event)
      if (eventResult.category) {
        try {
          relatedLoading.value = true;
          console.log(`🔍 [DynamicEventPage] Loading related events for category: ${eventResult.category}`);
          
          // Get all events and filter client-side for more reliable results  
          const relatedResult = await eventsClient.getEvents({
            pageSize: 50, // Get more events to filter from
            sortBy: 'date-asc' // Show upcoming events first
          });
          
          console.log(`🔍 [DynamicEventPage] Related events API response:`, relatedResult);
          
          if (relatedResult.data) {
            // Filter by category and exclude current event
            const filteredRelated = relatedResult.data
              .filter(relatedEvent => 
                relatedEvent.id !== eventResult.id && 
                relatedEvent.category === eventResult.category
              )
              .slice(0, 3);
            
            // If no events in same category, show recent events instead
            if (filteredRelated.length === 0) {
              const recentEvents = relatedResult.data
                .filter(relatedEvent => relatedEvent.id !== eventResult.id)
                .slice(0, 3);
              relatedEvents.value = recentEvents.map(e => ({
                id: e.id.toString(),
                title: e.title,
                slug: e.slug,
                description: e.excerpt || e.meta_description || '',
                category: e.category || 'Event',
                date: e.event_date,
                time: e.event_time || '',
                location: e.location || '',
                status: 'Open' as any,
                featured_image: e.featured_image || '',
                capacity: e.capacity || undefined,
                registered: undefined
              }));
              console.log(`✅ [DynamicEventPage] No same-category events, showing recent: ${recentEvents.length}`);
            } else {
              relatedEvents.value = filteredRelated.map(e => ({
                id: e.id.toString(),
                title: e.title,
                slug: e.slug,
                description: e.excerpt || e.meta_description || '',
                category: e.category || 'Event',
                date: e.event_date,
                time: e.event_time || '',
                location: e.location || '',
                status: 'Open' as any,
                featured_image: e.featured_image || '',
                capacity: e.capacity || undefined,
                registered: undefined
              }));
              console.log(`✅ [DynamicEventPage] Related events loaded: ${filteredRelated.length}`, filteredRelated);
            }
          } else {
            console.warn('⚠️ [DynamicEventPage] No data in related events response');
          }
        } catch (relatedErr) {
          console.warn('⚠️ [DynamicEventPage] Failed to load related events:', relatedErr);
          // Don't set error, just continue without related events
        } finally {
          relatedLoading.value = false;
        }
      } else {
        console.log('ℹ️ [DynamicEventPage] No category found for current event, skipping related events');
      }
    } else {
      throw new Error(`Event not found: ${slug}`);
    }
  } catch (err: any) {
    console.error('❌ [DynamicEventPage] Failed to load event:', err.message);
    error.value = err.message || 'Failed to load event data';
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
/* Component-specific styles if needed */
</style>