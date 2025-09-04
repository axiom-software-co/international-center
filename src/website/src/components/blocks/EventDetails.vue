<template>
  <div class="bg-white rounded-lg border border-gray-200 p-6">
    <h3 class="text-lg font-semibold text-gray-900 mb-4">Event Details</h3>
    
    <div class="space-y-3">
      <!-- Date & Time -->
      <div class="flex items-start">
        <svg class="w-5 h-5 text-gray-500 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
        </svg>
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-900">Date & Time</p>
          <p class="text-sm text-gray-600">{{ formatEventDate(eventDate) }}</p>
          <p class="text-sm text-gray-600">{{ eventTime }}</p>
        </div>
      </div>

      <Separator />

      <!-- Location -->
      <div class="flex items-start">
        <svg class="w-5 h-5 text-gray-500 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
        </svg>
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-900">Location</p>
          <p class="text-sm text-gray-600">{{ location }}</p>
          <p class="text-sm text-gray-500">123 Medical Plaza Drive<br>South Florida, FL 33101</p>
        </div>
      </div>

      <!-- Capacity -->
      <template v-if="capacity && registered !== undefined">
        <Separator />
        
        <div class="flex items-start">
          <svg class="w-5 h-5 text-gray-500 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
          </svg>
          <div class="flex-1">
            <p class="text-sm font-medium text-gray-900">Registration</p>
            <p class="text-sm text-gray-600">{{ registered }}/{{ capacity }} registered</p>
            <div class="mt-2 w-full bg-gray-200 rounded-full h-2">
              <div 
                class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                :style="{ width: `${Math.min((registered / capacity) * 100, 100)}%` }"
              ></div>
            </div>
          </div>
        </div>
      </template>

    </div>

    <!-- Registration Button -->
    <div class="mt-6 pt-6 border-t border-gray-200">
      <button
        :disabled="isEventFull"
        :class="[
          'w-full px-4 py-3 text-sm font-medium rounded transition-colors',
          isEventFull 
            ? 'bg-gray-100 text-gray-500 cursor-not-allowed' 
            : 'bg-blue-600 text-white hover:bg-blue-700'
        ]"
      >
        {{ buttonText }}
      </button>
      <p v-if="isEventFull" class="text-xs text-gray-500 text-center mt-2">
        This event is currently full
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import Separator from '@/components/vue-ui/Separator.vue';

interface Props {
  eventDate: string;
  eventTime: string;
  location: string;
  capacity?: number;
  registered?: number;
  status: string;
}

const props = defineProps<Props>();

const isEventFull = computed(() => {
  if (!props.capacity || props.registered === undefined) return false;
  return props.registered >= props.capacity;
});

const buttonText = computed(() => {
  if (isEventFull.value) return 'Event Full';
  if (props.status === 'Registration Required') return 'Register Now';
  return 'Join Event';
});

const formatEventDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};
</script>