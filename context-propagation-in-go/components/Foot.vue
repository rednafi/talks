<script setup>
// Project override of tahta's Foot.
// Section dividers don't count as slides, so the counter skips them:
// "7 / 31" means the seventh content slide out of 31.
import { computed, toValue } from 'vue'
import { useSlideContext } from '@slidev/client'

const props = defineProps({ mark: { type: Boolean, default: true } })
const { $frontmatter, $page, $nav, $slidev } = useSlideContext()

const strip = (s) => String(s ?? '').replace(/<[^>]*>/g, '')
const left = computed(() => strip($frontmatter.foot ?? $slidev?.configs?.title ?? ''))

const slides = computed(() => toValue($nav?.slides) || toValue($slidev?.nav?.slides) || [])
const counted = (s) => (s?.meta?.slide?.frontmatter?.layout ?? '') !== 'section'
const isSection = computed(() => $frontmatter.layout === 'section')
const isCover = computed(() => $frontmatter.layout === 'cover')
const pos = computed(() => {
  const page = Number(toValue($page)) || 1
  return slides.value.slice(0, page).filter(counted).length
})
const total = computed(() => slides.value.filter(counted).length)
</script>

<template>
  <div class="foot">
    <span class="foot-left">{{ left }}</span>
    <span v-if="!isSection && !isCover" class="foot-pg"><span class="pg">{{ pos }}</span> / {{ total }}</span>
  </div>
</template>
