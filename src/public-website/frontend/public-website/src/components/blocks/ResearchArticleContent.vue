<template>
  <div v-html="article.content"></div>
</template>

<script setup lang="ts">
interface Article {
  content?: string;
}

interface Props {
  article: Article;
}

const props = defineProps<Props>();
</script>