<template>
  <section class="pt-12 pb-20">
    <div class="container">
      <div class="flex flex-col items-center gap-6 text-center mb-12">
        <Badge variant="outline" class="mb-4">
          {{ subheading }}
        </Badge>
        <h2 class="text-3xl font-semibold lg:text-5xl">{{ heading }}</h2>
        <p class="text-muted-foreground lg:text-lg max-w-2xl">
          Discover how our regenerative medicine treatments have helped patients throughout South
          Florida reclaim their health and vitality.
        </p>
      </div>

      <!-- Desktop Grid View -->
      <div class="hidden lg:grid lg:grid-cols-3 gap-8">
        <Card
          v-for="(testimonial, idx) in displayTestimonials.slice(0, 6)"
          :key="idx"
          class="rounded border p-6 transition-colors group"
        >
          <CardContent class="p-6">
            <!-- Rating -->
            <div class="flex items-center gap-1 mb-4">
              <Star
                v-for="i in 5"
                :key="i"
                :class="`size-4 ${i <= (testimonial.rating || 5) ? 'text-yellow-400' : 'text-gray-300'}`"
                :fill="i <= (testimonial.rating || 5) ? 'currentColor' : 'none'"
              />
            </div>

            <!-- Content -->
            <blockquote class="text-sm leading-relaxed text-foreground/80 mb-4">
              "{{ testimonial.content }}"
            </blockquote>

            <!-- Treatment Badge -->
            <Badge variant="outline" class="mb-4">
              {{ testimonial.treatment }}
            </Badge>
          </CardContent>
          <CardFooter class="px-6 pb-6">
            <div class="flex items-center gap-4 w-full">
              <Avatar class="size-12 ring-2 ring-background">
                <AvatarImage :src="testimonial.avatar" :alt="testimonial.name" />
                <AvatarFallback>
                  {{
                    testimonial.name
                      .split(' ')
                      .map(n => n[0])
                      .join('')
                  }}
                </AvatarFallback>
              </Avatar>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <p class="font-semibold text-sm transition-colors">{{ testimonial.name }}</p>
                  <span v-if="testimonial.age" class="text-muted-foreground text-xs"
                    >Age {{ testimonial.age }}</span
                  >
                </div>
                <p class="text-xs text-muted-foreground">{{ testimonial.condition }}</p>
                <p class="text-xs text-muted-foreground">{{ testimonial.location }}</p>
              </div>
            </div>
          </CardFooter>
        </Card>
      </div>

      <!-- Mobile Carousel View -->
      <div class="lg:hidden">
        <!-- Simple responsive grid for mobile without complex carousel -->
        <div class="grid gap-6 md:grid-cols-2">
          <Card v-for="(testimonial, idx) in displayTestimonials" :key="idx" class="h-full">
            <CardContent class="p-6">
              <!-- Rating -->
              <div class="flex items-center gap-1 mb-4">
                <Star
                  v-for="i in 5"
                  :key="i"
                  :class="`size-4 ${i <= (testimonial.rating || 5) ? 'text-yellow-400' : 'text-gray-300'}`"
                  :fill="i <= (testimonial.rating || 5) ? 'currentColor' : 'none'"
                />
              </div>

              <!-- Content -->
              <blockquote class="text-sm leading-relaxed text-foreground/80 mb-4">
                "{{ testimonial.content }}"
              </blockquote>

              <!-- Treatment Badge -->
              <Badge variant="outline" class="mb-4">
                {{ testimonial.treatment }}
              </Badge>
            </CardContent>
            <CardFooter class="px-6 pb-6">
              <div class="flex items-center gap-4 w-full">
                <Avatar class="size-12 ring-2 ring-background">
                  <AvatarImage :src="testimonial.avatar" :alt="testimonial.name" />
                  <AvatarFallback>
                    {{
                      testimonial.name
                        .split(' ')
                        .map(n => n[0])
                        .join('')
                    }}
                  </AvatarFallback>
                </Avatar>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <p class="font-semibold text-sm transition-colors">{{ testimonial.name }}</p>
                    <span v-if="testimonial.age" class="text-muted-foreground text-xs"
                      >Age {{ testimonial.age }}</span
                    >
                  </div>
                  <p class="text-xs text-muted-foreground">{{ testimonial.condition }}</p>
                  <p class="text-xs text-muted-foreground">{{ testimonial.location }}</p>
                </div>
              </div>
            </CardFooter>
          </Card>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Star } from 'lucide-vue-next';
import {
  Badge,
  Card,
  CardContent,
  CardFooter,
  Avatar,
  AvatarImage,
  AvatarFallback,
} from '@/components/vue-ui';

interface Testimonial {
  name: string;
  age?: number;
  condition: string;
  avatar: string;
  content: string;
  rating?: number;
  treatment: string;
  location: string;
}

interface Props {
  heading?: string;
  subheading?: string;
  testimonials?: Testimonial[];
}

const props = withDefaults(defineProps<Props>(), {
  heading: 'Patient Success Stories',
  subheading: 'Real Results from Real People',
  testimonials: () => [
    {
      name: 'Sarah M.',
      age: 52,
      condition: 'Chronic Knee Pain',
      avatar: 'https://placehold.co/150x150/000000/FFFFFF/png?text=SM',
      content:
        'After years of knee pain from arthritis, I was skeptical about trying regenerative medicine. The PRP treatment I received was life-changing. Within 6 weeks, I was hiking again pain-free. The mobile service made it so convenient.',
      rating: 5,
      treatment: 'PRP Therapy',
      location: 'Boca Raton, FL',
    },
    {
      name: 'Michael R.',
      age: 45,
      condition: 'Sports Injury',
      avatar: 'https://placehold.co/150x150/000000/FFFFFF/png?text=MR',
      content:
        "As a weekend warrior who tore his rotator cuff, I thought my tennis days were over. The exosome therapy helped accelerate my healing beyond what traditional treatments offered. I'm back on the court stronger than ever.",
      rating: 5,
      treatment: 'Exosome Therapy',
      location: 'West Palm Beach, FL',
    },
    {
      name: 'Linda K.',
      age: 58,
      condition: 'Chronic Fatigue',
      avatar: 'https://placehold.co/150x150/000000/FFFFFF/png?text=LK',
      content:
        'The peptide therapy and IV treatments have completely transformed my energy levels. I feel 20 years younger. The physicians are incredibly knowledgeable and the at-home service is perfect for my busy schedule.',
      rating: 5,
      treatment: 'Peptide Therapy + IV',
      location: 'Delray Beach, FL',
    },
    {
      name: 'Robert D.',
      age: 62,
      condition: 'Lower Back Pain',
      avatar: 'https://placehold.co/150x150/000000/FFFFFF/png?text=RD',
      content:
        'I was considering back surgery when a friend recommended International Center. The regenerative treatments have eliminated my chronic pain. I can now play with my grandchildren without constant discomfort.',
      rating: 5,
      treatment: 'PRP + Exosome Therapy',
      location: 'Fort Lauderdale, FL',
    },
    {
      name: 'Jennifer L.',
      age: 39,
      condition: 'Joint Pain',
      avatar: 'https://placehold.co/150x150/000000/FFFFFF/png?text=JL',
      content:
        'The professionalism and expertise of the entire team is outstanding. They explained every step of the treatment process and made me feel comfortable throughout. The results speak for themselves.',
      rating: 5,
      treatment: 'PRP Therapy',
      location: 'Miami, FL',
    },
  ],
});

const displayTestimonials = computed(() => props.testimonials || []);
</script>
