<script setup lang="ts">
import { ref } from 'vue';

interface GalleryImage {
  id: number;
  src: string;
  alt: string;
}

const galleryImages = ref<GalleryImage[]>([
  {
    id: 1,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Patient+Support',
    alt: 'Volunteers providing patient support and assistance'
  },
  {
    id: 2,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Training+Session',
    alt: 'Volunteer training and orientation session'
  },
  {
    id: 3,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Community+Outreach',
    alt: 'Volunteers at community health fair event'
  },
  {
    id: 4,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Research+Support',
    alt: 'Volunteers assisting with research activities'
  },
  {
    id: 5,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Team+Meeting',
    alt: 'Volunteer team meeting and collaboration'
  },
  {
    id: 6,
    src: 'https://placehold.co/400x300/000000/ffffff?text=Administrative+Work',
    alt: 'Volunteers helping with administrative tasks'
  }
]);
</script>

<template>
  <section class="pt-16 pb-20">
    <div class="container">
      <div class="text-center mb-16">
        <h2 class="text-3xl font-semibold lg:text-5xl mb-4">Volunteer Gallery</h2>
      </div>

      <div class="max-w-6xl mx-auto">
        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <div v-for="image in galleryImages" :key="image.id" class="group">
            <img 
              :src="image.src"
              :alt="image.alt"
              class="w-full h-64 object-cover rounded-sm"
              loading="lazy"
            />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>