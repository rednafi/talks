<script setup>
// An arrow from the right gutter of a <Code> panel to the end of one line.
//   <Mark :line="3">first cancel wins</Mark>
// `line` is 1-based inside the closest code block before the mark.
// `block` (1-based) picks another block inside the same <Code>.
import { onBeforeUnmount, onMounted, ref } from 'vue'

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
  const start = lineEnd + 10
  const labelX = Math.max(codeRight + 14, start + 28)
  pos.value = {
    left: start,
    top,
    arrow: labelX - start - 6,
    label: Math.max(80, boxRight - labelX),
    ready: true,
  }
}

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
    <svg class="mark-arrow" :width="pos.arrow" height="14" :viewBox="`0 0 ${pos.arrow} 14`">
      <line x1="2" y1="7" :x2="pos.arrow" y2="7" />
      <polyline points="9,1.5 2,7 9,12.5" />
    </svg>
    <span class="mark-label" :style="{ width: `${pos.label}px` }"><slot /></span>
  </div>
</template>
