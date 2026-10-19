<script setup>
// An arrow from the right gutter of a <Code> panel to the end of one line.
//   <Mark :line="3">first cancel wins</Mark>
// `line` is 1-based inside the closest code block before the mark.
// `block` (1-based) picks another block inside the same <Code>.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { arrow, toPaths } from '../lib/rough'

const props = defineProps({
  line: { type: [Number, String], required: true },
  block: { type: [Number, String], default: 0 },
})

const root = ref(null)
const pos = ref({ left: 0, top: 0, arrow: 40, label: 160, ready: false })
let ro
let raf = 0

function targetPre(el, body) {
  const pres = [...body.querySelectorAll('pre')]
  if (Number(props.block) > 0)
    return pres[Number(props.block) - 1]
  let last = null
  for (const p of pres) {
    if (p.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING)
      last = p
  }
  return last ?? pres[pres.length - 1]
}

const labelEl = ref(null)
let cv
function labelWidth() {
  const el = labelEl.value
  if (!el)
    return 0
  cv ??= document.createElement('canvas')
  const g = cv.getContext('2d')
  const cs = getComputedStyle(el)
  g.font = `${cs.fontWeight} ${cs.fontSize} ${cs.fontFamily}`
  return g.measureText(el.textContent.trim()).width
}

function place() {
  const el = root.value
  if (!el)
    return
  const body = el.closest('.codebox-body')
  const box = el.closest('.codebox')
  if (!body || !box)
    return
  const pre = targetPre(el, body)
  if (!pre)
    return
  const lines = pre.querySelectorAll('.line')
  const ln = lines[Number(props.line) - 1]
  if (!ln)
    return
  const b = body.getBoundingClientRect()
  const s = b.width / body.offsetWidth || 1
  const r = ln.getBoundingClientRect()
  const p = pre.getBoundingClientRect()
  const x = box.getBoundingClientRect()
  const lineEnd = (r.right - b.left) / s
  const codeRight = (p.right - b.left) / s
  const boxRight = (x.right - b.left) / s
  const top = (r.top - b.top) / s + r.height / s / 2
  // Labels line up in one column just past the longest code line, so they
  // never sit on top of the code. The arrow fills the gap to that column.
  let codeEnd = 0
  for (const l of body.querySelectorAll('pre .line'))
    codeEnd = Math.max(codeEnd, (l.getBoundingClientRect().right - b.left) / s)
  // If the label won't fit on one line in that column, pull the column left
  // until it does, but keep at least a short arrow.
  const start = lineEnd + 10
  const need = labelWidth() + 4
  const labelX = Math.max(start + 46, Math.min(codeEnd + 48, boxRight - need))
  const arrowW = labelX - start - 6
  pos.value = {
    left: start,
    top,
    arrow: arrowW,
    label: Math.max(120, boxRight - labelX),
    ready: true,
  }
}

// A hand-drawn arrow that points left, at the code line.
const seed = Math.round(Number(props.line) * 13 + Number(props.block) * 7)
const paths = computed(() => {
  const w = pos.value.arrow
  if (w < 8)
    return []
  return toPaths(arrow(w, 8, 2, 8, seed, { strokeWidth: 1.6, roughness: 0.9, head: 9 }, 1.5))
})

function schedule() {
  cancelAnimationFrame(raf)
  raf = requestAnimationFrame(place)
}

onMounted(() => {
  schedule()
  setTimeout(schedule, 300)
  setTimeout(schedule, 1200)
  document.fonts?.ready.then(schedule)
  const body = root.value?.closest('.codebox')
  if (body) {
    ro = new ResizeObserver(schedule)
    ro.observe(body)
  }
})
onBeforeUnmount(() => {
  ro?.disconnect()
  cancelAnimationFrame(raf)
})
</script>

<template>
  <div
    ref="root" class="mark"
    :style="{ left: `${pos.left}px`, top: `${pos.top}px`, visibility: pos.ready ? undefined : 'hidden' }"
  >
    <svg class="mark-arrow" :width="pos.arrow" height="16" :viewBox="`0 0 ${pos.arrow} 16`">
      <path v-for="(p, i) in paths" :key="i" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" />
    </svg>
    <span ref="labelEl" class="mark-label" :style="{ width: `${pos.label}px` }"><slot /></span>
  </div>
</template>
