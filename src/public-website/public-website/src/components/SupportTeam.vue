<template>
  <div :class="className">
    <!-- Flexbox container for proper centering -->
    <div class="flex flex-wrap justify-center gap-6">
      <Card
        v-for="(member, index) in supportStaff"
        :key="index"
        class="rounded border border-border dark:border-border overflow-hidden transition-colors group bg-card dark:bg-card w-full sm:w-[calc(50%-12px)] lg:w-[calc(33.333%-16px)] xl:w-[calc(25%-18px)] max-w-[320px]"
      >
        <!-- Image with 1:1 aspect ratio (square) -->
        <div class="aspect-square overflow-hidden bg-gray-100 dark:bg-gray-800">
          <img :src="member.image" :alt="member.name" class="w-full h-full object-cover" />
        </div>

        <!-- Content section -->
        <div class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex-1 min-w-0">
              <h3 class="text-base font-semibold text-card-foreground truncate">
                {{ member.name }}
              </h3>
              <p class="text-sm text-muted-foreground">{{ member.title }}</p>
            </div>
            <!-- Social profile icons -->
            <div v-if="member.linkedin || member.indeed" class="flex items-center gap-2 ml-3">
              <a
                v-if="member.linkedin"
                :href="member.linkedin"
                target="_blank"
                rel="noopener noreferrer"
                class="flex-shrink-0 text-muted-foreground hover:text-blue-600 transition-colors"
                :aria-label="`${member.name} LinkedIn profile`"
              >
                <Linkedin class="w-6 h-6" />
              </a>
              <a
                v-if="member.indeed"
                :href="member.indeed"
                target="_blank"
                rel="noopener noreferrer"
                class="flex-shrink-0 text-muted-foreground hover:text-blue-700 transition-colors"
                :aria-label="`${member.name} Indeed profile`"
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
import { User, Phone, Mail } from 'lucide-vue-next';



interface SupportTeamProps {
  className?: string;
}

const props = withDefaults(defineProps<SupportTeamProps>(), {
  className: '',
});

interface StaffMember {
  name: string;
  title: string;
  image: string;
  experience: string;
  linkedin: string | null;
  indeed: string | null;
}

const supportStaff: StaffMember[] = [
  {
    name: 'Lisa Thompson, NP',
    title: 'Nurse Practitioner',
    image: 'https://placehold.co/400x400/000000/FFFFFF/png?text=LT',
    experience: '8+ Years Experience',
    linkedin: null, // No LinkedIn profile
    indeed: 'https://indeed.com/r/example', // Has Indeed profile
  },
  {
    name: 'James Wilson, MA',
    title: 'Medical Assistant',
    image: 'https://placehold.co/400x400/000000/FFFFFF/png?text=JW',
    experience: '6+ Years Experience',
    linkedin: 'https://linkedin.com/in/example', // Has LinkedIn
    indeed: 'https://indeed.com/r/example', // Has Indeed
  },
  {
    name: 'Maria Garcia',
    title: 'Patient Coordinator',
    image: 'https://placehold.co/400x400/000000/FFFFFF/png?text=MG',
    experience: '5+ Years Experience',
    linkedin: null as string | null, // No social profiles
    indeed: null,
  },
];
</script>
