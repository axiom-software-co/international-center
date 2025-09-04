<template>
  <div :class="className">
    <!-- Flexbox container for proper centering -->
    <div class="flex flex-wrap justify-center gap-6">
      <Card
        v-for="(physician, index) in physicians"
        :key="index"
        class="rounded border border-border dark:border-border overflow-hidden transition-colors group bg-card dark:bg-card w-full sm:w-[calc(50%-12px)] lg:w-[calc(33.333%-16px)] xl:w-[calc(25%-18px)] max-w-[320px]"
      >
        <!-- Image with 1:1 aspect ratio (square) -->
        <div class="aspect-square overflow-hidden bg-gray-100 dark:bg-gray-800">
          <img :src="physician.image" :alt="physician.name" class="w-full h-full object-cover" />
        </div>

        <!-- Content section -->
        <div class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex-1 min-w-0">
              <h3 class="text-base font-semibold text-card-foreground truncate">
                {{ physician.name }}
              </h3>
              <p class="text-sm text-muted-foreground">{{ physician.title }}</p>
            </div>
            <!-- Social profile icons -->
            <div class="flex items-center gap-2 ml-3">
              <a
                v-if="physician.linkedin"
                :href="physician.linkedin"
                target="_blank"
                rel="noopener noreferrer"
                class="flex-shrink-0 text-muted-foreground hover:text-blue-600 transition-colors"
                :aria-label="`${physician.name} LinkedIn profile`"
              >
                <Linkedin class="w-6 h-6" />
              </a>
              <a
                v-if="physician.indeed"
                :href="physician.indeed"
                target="_blank"
                rel="noopener noreferrer"
                class="flex-shrink-0 text-muted-foreground hover:text-blue-700 transition-colors"
                :aria-label="`${physician.name} Indeed profile`"
              >
                <FileUser class="w-6 h-6" />
              </a>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import Card from '@/components/vue-ui/Card.vue';
import { Linkedin, FileUser } from 'lucide-vue-next';

interface TeamProfilesProps {
  className?: string;
}

const props = withDefaults(defineProps<TeamProfilesProps>(), {
  className: '',
});

interface Physician {
  name: string;
  title: string;
  image: string;
  linkedin: string | null;
  indeed: string | null;
  certifications: string;
  education: string;
  experience: string;
  specialties: string;
}

const physicians: Physician[] = [
  {
    name: 'Dr. Michael Rodriguez, MD',
    title: 'Medical Director & Founder',
    image: 'https://placehold.co/400x400/000000/FFFFFF/png?text=DR',
    linkedin: 'https://linkedin.com/in/example', // Placeholder LinkedIn URL
    indeed: 'https://indeed.com/r/example', // Placeholder Indeed URL
    certifications: 'Board Certified in Internal Medicine & Regenerative Medicine',
    education: 'MD, University of Miami Miller School of Medicine',
    experience: '15+ Years Clinical Experience',
    specialties: 'Specializes in PRP & Exosome Therapy',
  },
  {
    name: 'Dr. Sarah Chen, MD',
    title: 'Associate Physician',
    image: 'https://placehold.co/400x400/000000/FFFFFF/png?text=SC',
    linkedin: 'https://linkedin.com/in/example', // Placeholder LinkedIn URL
    indeed: null, // No Indeed profile
    certifications: 'Board Certified in Family Medicine & Sports Medicine',
    education: 'MD, Florida International University',
    experience: '12+ Years Clinical Experience',
    specialties: 'Specializes in Sports Injuries & Peptide Therapy',
  },
];
</script>
