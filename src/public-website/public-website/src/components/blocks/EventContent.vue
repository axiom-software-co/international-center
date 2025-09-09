<template>
  <div>
    <div class="mb-6">
      <h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
        {{ event.title }}
      </h1>
      <p class="text-lg text-gray-600 leading-relaxed">
        {{ event.description }}
      </p>
    </div>

    <div v-if="event.content" class="prose prose-lg max-w-none" v-html="event.content"></div>
    
    <div v-else class="space-y-6">
      <div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-4">Event Overview</h2>
        <p class="text-gray-700 leading-relaxed">
          {{ event.description }}
        </p>
      </div>

      <div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-4">What to Expect</h2>
        <p class="text-gray-700 leading-relaxed">
          Join us for an engaging and informative session designed to provide valuable insights and practical knowledge. 
          This event offers an excellent opportunity to learn from experts and connect with like-minded individuals in our community.
        </p>
      </div>

      <div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-4">Who Should Attend</h2>
        <p class="text-gray-700 leading-relaxed">
          This event is open to all community members interested in health, wellness, and regenerative medicine. 
          Whether you're a healthcare professional, researcher, student, or simply curious about the topic, you're welcome to join us.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
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

interface Props {
  event: EventPageData;
}

defineProps<Props>();
</script>

<style scoped>
.prose h2 {
  margin-top: 2rem;
  margin-bottom: 1rem;
}

.prose h3 {
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
}

.prose ul {
  margin: 1rem 0;
  padding-left: 1.5rem;
}

.prose li {
  margin: 0.5rem 0;
}
</style>