<template>
  <section class="pt-16 pb-20">
    <div class="container">
      <!-- Section Title -->
      <div class="text-center mb-12">
        <h2 class="text-3xl font-semibold lg:text-5xl mb-8">Establishing Our First Facility</h2>
      </div>

      <!-- Tabs Control -->
      <div class="flex justify-center mb-12">
        <div class="flex bg-gray-100 rounded-lg p-1">
          <button
            v-for="(phase, index) in phases"
            :key="index"
            @click="activeTab = index"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-md transition-colors',
              activeTab === index
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            ]"
          >
            {{ phase.name }}
          </button>
        </div>
      </div>

      <!-- Tab Content -->
      <div class="grid gap-8 lg:grid-cols-2 lg:items-center">
        <!-- Content Column -->
        <div>
          <div class="bg-white border border-gray-200 rounded-sm p-8">
            <h3 class="text-2xl font-semibold mb-6 text-gray-900">{{ phases[activeTab].title }}</h3>
            <p class="text-gray-600 mb-8">
              {{ phases[activeTab].description }}
            </p>
            
            <div class="space-y-4">
              <div 
                v-for="(item, index) in phases[activeTab].items"
                :key="index"
                class="flex items-start"
              >
                <span class="w-3 h-px bg-black mr-4 mt-3 flex-shrink-0"></span>
                <div>
                  <h4 class="font-medium text-gray-900 mb-1">{{ item.title }}</h4>
                  <p class="text-sm text-gray-600">{{ item.description }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Phase-Specific Content Column -->
        <div class="space-y-6">
          <div 
            v-for="(card, index) in phases[activeTab].rightColumn"
            :key="index"
            :class="[
              'rounded-sm p-6',
              card.type === 'primary' 
                ? 'bg-black text-white p-8' 
                : 'bg-gray-50 border border-gray-200'
            ]"
          >
            <div v-if="card.metric" class="text-4xl font-bold mb-4">{{ card.metric }}</div>
            <h3 :class="[
              'font-semibold mb-4',
              card.type === 'primary' 
                ? 'text-xl text-white' 
                : 'text-lg text-gray-900'
            ]">{{ card.title }}</h3>
            <p 
              v-if="card.type !== 'primary'"
              class="text-gray-600 text-sm"
            >
              {{ card.description }}
            </p>
            
            <!-- Checklist Items -->
            <div v-if="card.checklistItems" class="mt-4 space-y-3">
              <div 
                v-for="(item, itemIndex) in card.checklistItems"
                :key="itemIndex"
                class="flex items-start gap-3"
              >
                <!-- Status Icon -->
                <div class="flex-shrink-0 mt-0.5">
                  <div v-if="item.status === 'completed'" class="w-4 h-4 bg-black rounded-sm flex items-center justify-center">
                    <svg class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div v-else-if="item.status === 'started'" class="w-4 h-4 border-2 border-black rounded-sm relative overflow-hidden">
                    <div class="absolute inset-0 bg-black" style="clip-path: polygon(0 0, 50% 0, 50% 100%, 0 100%)"></div>
                  </div>
                  <div v-else class="w-4 h-4 border-2 border-gray-300 rounded-sm"></div>
                </div>
                
                <!-- Content -->
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium text-gray-900">{{ item.title }}</div>
                  <div v-if="item.status === 'not-started'" class="text-xs text-gray-500 mt-1">
                    Est. {{ item.estimatedDuration }} when started
                  </div>
                  <div v-else-if="item.status === 'started'" class="text-xs text-gray-500 mt-1">
                    Started {{ item.startDate }} • Expected completion {{ item.expectedCompletion }}
                  </div>
                  <div v-else-if="item.status === 'completed'" class="text-xs text-gray-500 mt-1">
                    {{ item.startDate }} - {{ item.completionDate }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Donation Button -->
      <div class="mt-12 text-center">
        <a
          href="/community/donations"
          class="inline-block px-6 py-2 text-white bg-black hover:bg-gray-800 font-semibold rounded transition-colors border border-black hover:border-gray-800"
        >
          Support Our Mission
        </a>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const activeTab = ref(0);

const phases = [
  {
    name: 'Phase I',
    title: 'Site Planning & Design', 
    description: 'Comprehensive planning and design phase to establish our first dedicated regenerative medicine facility in South Florida, focusing on optimal site selection and specialized facility design.',
    items: [
      {
        title: 'Site Selection & Feasibility',
        description: 'Identify and evaluate optimal location for our first facility with focus on patient accessibility and research infrastructure'
      },
      {
        title: 'Specialized Facility Design',
        description: 'Custom architectural plans for regenerative medicine laboratories, treatment rooms, and patient care areas'
      },
      {
        title: 'Regulatory Compliance Planning',
        description: 'Ensure all designs meet FDA, state medical facility, and local building requirements from the outset'
      },
      {
        title: 'Technology Infrastructure Planning',
        description: 'Plan for advanced research equipment, IT systems, and specialized medical technology integration'
      }
    ],
    rightColumn: [
      {
        type: 'primary',
        metric: '$350K',
        title: 'Initial Planning Budget',
        description: 'Foundation investment for site acquisition planning, architectural design, and regulatory preparation for our first facility.'
      },
      {
        type: 'secondary',
        title: 'Phase I Progress Checklist',
        checklistItems: [
          { 
            title: 'Site Selection & Feasibility',
            status: 'completed',
            startDate: 'Jan 2025',
            completionDate: 'Feb 2025'
          },
          { 
            title: 'Architectural Design',
            status: 'started',
            startDate: 'Feb 2025',
            expectedCompletion: 'May 2025'
          },
          { 
            title: 'Regulatory Planning',
            status: 'not-started',
            estimatedDuration: '2-3 months'
          },
          { 
            title: 'Final Approvals',
            status: 'not-started',
            estimatedDuration: '1-2 months'
          }
        ]
      }
    ]
  },
  {
    name: 'Phase II',
    title: 'Funding & Permits',
    description: 'Secure complete funding for our first facility and obtain all necessary permits and regulatory approvals to begin construction of this pioneering regenerative medicine center.',
    items: [
      {
        title: 'Capital Fundraising Campaign',
        description: 'Comprehensive funding strategy combining donations, grants, and partnerships to establish our first facility'
      },
      {
        title: 'Medical Facility Licensing',
        description: 'Obtain all required permits and licenses for constructing and operating our first regenerative medicine facility'
      },
      {
        title: 'Community Partnership Building',
        description: 'Establish relationships with local healthcare systems, universities, and community organizations'
      },
      {
        title: 'Operational Financial Planning',
        description: 'Develop sustainable financial models for ongoing operations of our first facility'
      }
    ],
    rightColumn: [
      {
        type: 'primary',
        metric: '$2.1M',
        title: 'First Facility Investment',
        description: 'Total capital requirement to establish our inaugural regenerative medicine facility from planning through operations.'
      },
      {
        type: 'secondary',
        title: 'Phase II Progress Checklist',
        checklistItems: [
          { 
            title: 'Capital Fundraising Campaign',
            status: 'started',
            startDate: 'Jan 2025',
            expectedCompletion: 'Jul 2025'
          },
          { 
            title: 'Medical Facility Licensing',
            status: 'not-started',
            estimatedDuration: '3-5 months'
          },
          { 
            title: 'Community Partnership Building',
            status: 'completed',
            startDate: 'Dec 2024',
            completionDate: 'Feb 2025'
          },
          { 
            title: 'Operational Financial Planning',
            status: 'not-started',
            estimatedDuration: '2-3 months'
          }
        ]
      }
    ]
  },
  {
    name: 'Phase III',
    title: 'Construction & Setup',
    description: 'Construct our first regenerative medicine facility with specialized laboratories, treatment areas, and advanced equipment to serve South Florida patients.',
    items: [
      {
        title: 'Facility Construction',
        description: 'Complete construction of our first regenerative medicine facility including specialized laboratory and clinical spaces'
      },
      {
        title: 'Advanced Equipment Installation',
        description: 'Install and calibrate cutting-edge regenerative medicine equipment and research technology systems'
      },
      {
        title: 'Technology Systems Integration',
        description: 'Integrate IT infrastructure, security systems, and operational management platforms for seamless operations'
      },
      {
        title: 'Team Building & Training',
        description: 'Recruit and train founding team of researchers, clinicians, and support staff for our first facility'
      }
    ],
    rightColumn: [
      {
        type: 'primary',
        metric: '$1.2M',
        title: 'Construction & Equipment',
        description: 'Primary investment for building construction, specialized equipment procurement, and technology installation for our first facility.'
      },
      {
        type: 'secondary',
        title: 'Phase III Progress Checklist',
        checklistItems: [
          { 
            title: 'Facility Construction',
            status: 'not-started',
            estimatedDuration: '8-12 months'
          },
          { 
            title: 'Advanced Equipment Installation',
            status: 'not-started',
            estimatedDuration: '3-4 months'
          },
          { 
            title: 'Technology Systems Integration',
            status: 'not-started',
            estimatedDuration: '2-3 months'
          },
          { 
            title: 'Team Building & Training',
            status: 'not-started',
            estimatedDuration: '4-6 months'
          }
        ]
      }
    ]
  },
  {
    name: 'Phase IV',
    title: 'Operations Launch',
    description: 'Commission and launch our first facility with full operational capability, establishing South Florida\'s premier destination for regenerative medicine research and treatment.',
    items: [
      {
        title: 'Facility Commissioning',
        description: 'Complete testing, certification, and validation of all systems and equipment for our first operational facility'
      },
      {
        title: 'Inaugural Research Programs',
        description: 'Launch initial clinical research studies and establish partnerships for our first facility operations'
      },
      {
        title: 'Community Outreach Launch',
        description: 'Initiate education programs and community health services from our first facility'
      },
      {
        title: 'Full Operational Capacity',
        description: 'Achieve complete operational capability with integrated research, treatment, and community programs'
      }
    ],
    rightColumn: [
      {
        type: 'primary',
        metric: '$350K',
        title: 'Operational Launch Budget',
        description: 'Final investment for commissioning, certification, program development, and achieving full operational status for our first facility.'
      },
      {
        type: 'secondary',
        title: 'Phase IV Progress Checklist',
        checklistItems: [
          { 
            title: 'Facility Commissioning',
            status: 'not-started',
            estimatedDuration: '2-3 months'
          },
          { 
            title: 'Inaugural Research Programs',
            status: 'not-started',
            estimatedDuration: '3-4 months'
          },
          { 
            title: 'Community Outreach Launch',
            status: 'not-started',
            estimatedDuration: '2-3 months'
          },
          { 
            title: 'Full Operational Capacity',
            status: 'not-started',
            estimatedDuration: '3-4 months'
          }
        ]
      }
    ]
  }
];
</script>