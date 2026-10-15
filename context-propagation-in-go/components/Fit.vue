<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'

// Project override of tahta's Fit.
// The theme version shrinks content with a transform but keeps its layout width,
// so a shrunk slide only fills part of the frame. This one lays the content out
// at width W / s before scaling by s, so it always spans the full frame width,
// and it may scale sparse content slides up (to MAX) so they don't look empty.
const MIN = 0.42
const MAX = 1.4

const outer = ref(null)
const inner = ref(null)
const scale = ref(1)
const width = ref(null)
const fixBottom = ref(0)
let ro
let raf = 0

function fits(s, W, H) {
  const i = inner.value
  i.style.width = `${W / s}px`
  if (i.scrollHeight * s > H + 1 || i.scrollWidth > i.clientWidth + 1)
    return false
  // Code blocks and terminals never wrap: their lines must fit too.
  for (const el of i.querySelectorAll('pre, .terminal-body, table'))
    if (el.scrollWidth > el.clientWidth + 1)
      return false
  return true
}

function fit() {
  const o = outer.value
  const i = inner.value
  if (!o || !i)
    return
  const W = o.clientWidth
  const H = o.clientHeight
  if (!W || !H)
    return
  // Only content slides grow; cover, statement and end keep their designed size.
  const max = o.closest('.l-topic > .l-body') ? MAX : 1
  let s = max
  if (!fits(max, W, H)) {
    let lo = MIN
    let hi = max
    for (let k = 0; k < 12; k++) {
      const mid = (lo + hi) / 2
      if (fits(mid, W, H))
        lo = mid
      else
        hi = mid
    }
    s = lo
  }
  width.value = `${W / s}px`
  i.style.width = width.value
  // Make the layout box match the visual box, so centering layouts still center.
  fixBottom.value = (s - 1) * i.scrollHeight
  scale.value = s
}

function schedule() {
  cancelAnimationFrame(raf)
  raf = requestAnimationFrame(fit)
}

onMounted(() => {
  fit()
  ro = new ResizeObserver(schedule)
  ro.observe(outer.value)
  schedule()
  document.fonts?.ready.then(schedule)
})
onBeforeUnmount(() => {
  ro?.disconnect()
  cancelAnimationFrame(raf)
})
</script>

<template>
  <div ref="outer" class="fit">
    <div ref="inner" class="fit-inner" :style="{ width, transform: `scale(${scale})`, transformOrigin: 'top left', marginBottom: `${fixBottom}px` }">
      <slot />
    </div>
  </div>
</template>
