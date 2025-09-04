#!/usr/bin/env node

const express = require('express');
const cors = require('cors');

// Mock news data
const mockNewsArticles = [
  {
    id: 1,
    title: "Revolutionary Regenerative Medicine Breakthrough",
    slug: "revolutionary-regenerative-medicine-breakthrough",
    excerpt: "Our research team has achieved a significant milestone in stem cell therapy applications.",
    description: "Our research team has achieved a significant milestone in stem cell therapy applications for treating degenerative conditions.",
    category_id: 1,
    category: "Medical Innovation", 
    category_data: { name: "Medical Innovation", id: 1 },
    featured_image: "/images/news/regenerative-medicine-expansion-hero.png",
    published_at: "2024-12-15T10:00:00Z",
    author: "Dr. Sarah Johnson",
    featured: true,
    readingTime: "5 min read"
  },
  {
    id: 2,
    title: "New Telemedicine Platform Launch",
    slug: "new-telemedicine-platform-launch",
    excerpt: "Expanding access to healthcare through innovative digital solutions.",
    description: "Expanding access to healthcare through innovative digital solutions and remote monitoring capabilities.",
    category_id: 2,
    category: "Technology",
    category_data: { name: "Technology", id: 2 },
    featured_image: "/images/news/telemedicine-revolution-hero.png",
    published_at: "2024-12-10T14:30:00Z",
    author: "International Center Team",
    featured: false,
    readingTime: "3 min read"
  },
  {
    id: 3,
    title: "Community Wellness Workshop Success",
    slug: "community-wellness-workshop-success",
    excerpt: "Over 200 participants joined our latest community health initiative.",
    description: "Over 200 participants joined our latest community health initiative focusing on preventive care and wellness education.",
    category_id: 3,
    category: "Community Health",
    category_data: { name: "Community Health", id: 3 },
    featured_image: "/images/news/community-wellness-workshops-hero.png",
    published_at: "2024-12-05T09:15:00Z",
    author: "Dr. Michael Chen",
    featured: false,
    readingTime: "4 min read"
  }
];

const mockNewsCategories = [
  { id: 1, name: "Medical Innovation", description: "Latest medical breakthroughs and innovations" },
  { id: 2, name: "Technology", description: "Healthcare technology advancements" },
  { id: 3, name: "Community Health", description: "Community outreach and health programs" }
];

// Mock research data
const mockResearchArticles = [
  {
    id: 1,
    title: "PRP Therapy Efficacy in Sports Medicine",
    slug: "prp-therapy-efficacy-sports-medicine",
    excerpt: "Clinical study demonstrates significant improvement in athletic recovery times.",
    description: "Clinical study demonstrates significant improvement in athletic recovery times using platelet-rich plasma therapy.",
    category: "Clinical Research",
    featured_image: "/images/research/prp-sports-medicine-study-hero.png",
    published_at: "2024-11-20T16:00:00Z",
    author: "Dr. Robert Williams",
    featured: true,
    readingTime: "8 min read"
  },
  {
    id: 2,
    title: "Exosome Therapy Clinical Trial Results",
    slug: "exosome-therapy-clinical-trial-results",
    excerpt: "Phase II trial shows promising results for regenerative treatment applications.",
    description: "Phase II trial shows promising results for regenerative treatment applications in tissue repair and healing.",
    category: "Clinical Research",
    featured_image: "/images/research/exosome-clinical-trial-hero.png",
    published_at: "2024-11-15T11:30:00Z",
    author: "Dr. Emily Davis",
    featured: false,
    readingTime: "10 min read"
  },
  {
    id: 3,
    title: "Corporate Wellness Program Impact Study",
    slug: "corporate-wellness-program-impact-study",
    excerpt: "Comprehensive analysis of employee health outcomes and productivity improvements.",
    description: "Comprehensive analysis of employee health outcomes and productivity improvements following implementation of our wellness programs.",
    category: "Case Studies",
    featured_image: "/images/research/corporate-wellness-case-hero.png",
    published_at: "2024-11-10T13:45:00Z",
    author: "Dr. Lisa Thompson",
    featured: false,
    readingTime: "6 min read"
  }
];

const mockResearchCategories = [
  "Clinical Research",
  "Case Studies", 
  "Laboratory Studies",
  "Medical Innovation",
  "Treatment Outcomes"
];

// Create News App
const newsApp = express();
newsApp.use(cors());
newsApp.use(express.json());

newsApp.get('/', (req, res) => {
  console.log('📥 [Mock News API] GET / called');
  res.json({
    data: mockNewsArticles,
    total: mockNewsArticles.length,
    page: 1,
    pageSize: 50,
    totalPages: 1
  });
});

newsApp.get('/categories', (req, res) => {
  console.log('📥 [Mock News API] GET /categories called');
  res.json({
    data: mockNewsCategories
  });
});

newsApp.get('/:slug', (req, res) => {
  console.log(`📥 [Mock News API] GET /${req.params.slug} called`);
  const article = mockNewsArticles.find(a => a.slug === req.params.slug);
  if (article) {
    res.json({ data: article });
  } else {
    res.status(404).json({ error: 'Article not found' });
  }
});

newsApp.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'news', timestamp: new Date().toISOString() });
});

// Create Research App
const researchApp = express();
researchApp.use(cors());
researchApp.use(express.json());

researchApp.get('/', (req, res) => {
  console.log('📥 [Mock Research API] GET / called');
  res.json({
    data: mockResearchArticles,
    total: mockResearchArticles.length,
    page: 1,
    pageSize: 50,
    totalPages: 1
  });
});

researchApp.get('/categories', (req, res) => {
  console.log('📥 [Mock Research API] GET /categories called');
  res.json({
    data: mockResearchCategories
  });
});

researchApp.get('/:slug', (req, res) => {
  console.log(`📥 [Mock Research API] GET /${req.params.slug} called`);
  const article = mockResearchArticles.find(a => a.slug === req.params.slug);
  if (article) {
    res.json({ data: article });
  } else {
    res.status(404).json({ error: 'Article not found' });
  }
});

researchApp.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'research', timestamp: new Date().toISOString() });
});

// Start servers
const newsPort = 8082;
const researchPort = 8083;

const newsServer = newsApp.listen(newsPort, () => {
  console.log(`🚀 Mock News API server running on http://localhost:${newsPort}`);
});

const researchServer = researchApp.listen(researchPort, () => {
  console.log(`🚀 Mock Research API server running on http://localhost:${researchPort}`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('📴 Shutting down mock API servers...');
  newsServer.close();
  researchServer.close();
});

process.on('SIGINT', () => {
  console.log('📴 Shutting down mock API servers...');
  newsServer.close();
  researchServer.close();
  process.exit(0);
});