<script setup>
// Quote slide with an optional image on the right (frontmatter `image:`).
import { useSlideContext } from '@slidev/client'
const { $frontmatter } = useSlideContext()
</script>

<template>
  <SlideFrame mode="statement">
    <div class="q-wrap" :class="{ 'has-img': $frontmatter.image }">
      <div class="q-text">
        <div class="quote-mark mb-3">&ldquo;</div>
        <div class="quote-body q-body" v-html="$frontmatter.quote" />
        <div v-if="$frontmatter.author" class="kicker mt-8">{{ $frontmatter.author }}</div>
      </div>
      <img v-if="$frontmatter.image" :src="$frontmatter.image" class="q-img" alt="" />
    </div>
  </SlideFrame>
</template>

<style scoped>
.q-wrap { display: grid; grid-template-columns: minmax(0, 1fr); align-items: center; gap: 3.5rem; width: 100%; }
.q-wrap.has-img { grid-template-columns: minmax(0, 1.7fr) minmax(0, 1fr); }
.q-body { font-size: 2.35rem; line-height: 1.22; max-width: 26ch; }
.q-img { justify-self: center; max-height: 19rem; width: auto; }
</style>
