<script setup>
// A small hand-drawn timeline: one bar per row, a vertical marker, and what
// happens to each bar after the marker.
//   after="waste": the part past the marker shows as wasted work (on click).
//   after="stop":  bars end at the marker; the part they skip shows as an outline.
//   after="clip":  like stop, used for a deadline that a child can't extend.
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'
import { gen, base, toPaths, arrow } from '../lib/rough'

const props = defineProps({
  rows: { type: Array, required: true }, // [{ label, from, to }]
  total: { type: Number, required: true },
  marker: { type: Object, required: true }, // { at, label }
  after: { type: String, default: 'waste' },
  afterLabel: { type: String, default: '' },
  ticks: { type: Array, default: () => [] }, // [{ at, label }]
  click: { type: Boolean, default: true },
  notes: { type: Array, default: () => [] }, // [{ row, at, text }] handwritten, below the axis
})
const { $clicks } = useSlideContext()
const tailOn = computed(() => !props.click || $clicks.value >= 1)

const W = 1000, LAB = 190, TRACK = W - LAB - 10, ROW = 46, BAR = 26, TOP = 34
const x = (v) => LAB + (v / props.total) * TRACK
const NOTE = computed(() => (props.notes.length ? 70 : 0))
// A waste key sits under the axis. A clip or stop label goes inside the
// outlined tail instead, so it takes no extra room.
const keyed = computed(() => !!props.afterLabel && props.after === 'waste')
const H = computed(() => TOP + props.rows.length * ROW + 34 + (keyed.value ? 40 : 0) + NOTE.value)

const drawn = computed(() => props.rows.map((r, i) => {
  const y = TOP + i * ROW
  const cut = Math.min(Math.max(props.marker.at, r.from), r.to)
  const head = cut > r.from
    ? toPaths([gen.rectangle(x(r.from), y, x(cut) - x(r.from), BAR, base(i * 11 + 1, { fill: 'currentColor', fillStyle: 'cross-hatch', hachureGap: 7, roughness: 1 }))])
    : []
  let tail = []
  if (r.to > cut) {
    const w = x(r.to) - x(cut)
    tail = props.after === 'waste'
      ? toPaths([gen.rectangle(x(cut), y, w, BAR, base(i * 11 + 5, { fill: 'currentColor', roughness: 1 }))])
      : toPaths([gen.rectangle(x(cut), y, w, BAR, base(i * 11 + 5, { strokeLineDash: [7, 6], roughness: 1 }))])
  }
  return { label: r.label, y, head, tail }
}))

const axisY = computed(() => TOP + props.rows.length * ROW + 4)
const axis = computed(() => toPaths([gen.line(LAB, axisY.value, W - 10, axisY.value, base(91, { roughness: 0.7, strokeWidth: 1.2 }))]))
const noteY = computed(() => axisY.value + 34 + (keyed.value ? 40 : 0) + 34)
const inTail = computed(() => {
  if (!props.afterLabel || keyed.value)
    return null
  const i = props.rows.length - 1, r = props.rows[i]
  const cut = Math.min(Math.max(props.marker.at, r.from), r.to)
  return r.to > cut ? { x: (x(cut) + x(r.to)) / 2, y: TOP + i * ROW + BAR / 2 } : null
})
const noted = computed(() => props.notes.map((n, i) => ({
  ...n,
  x: x(n.at),
  paths: toPaths(arrow(x(n.at) + 6, noteY.value - 22, x(n.at), TOP + n.row * ROW + BAR + 4, i * 31 + 9, { strokeWidth: 1.5, head: 10 }, 18)),
})))
const mark = computed(() => toPaths([gen.line(x(props.marker.at), 18, x(props.marker.at), axisY.value, base(77, { strokeWidth: 2.2, roughness: 1.2 }))]))
const swatch = computed(() => toPaths([gen.rectangle(0, axisY.value + 44, 34, 18, base(55, props.after === 'waste' ? { fill: 'currentColor' } : { strokeLineDash: [6, 5] }))]))
</script>

<template>
  <svg class="bars" :class="`after-${after}`" :viewBox="`0 0 ${W} ${H}`" xmlns="http://www.w3.org/2000/svg">
    <g v-for="(r, i) in drawn" :key="i">
      <text class="lab" :x="0" :y="r.y + BAR / 2" dominant-baseline="central">{{ r.label }}</text>
      <g class="head"><path v-for="(p, j) in r.head" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" :class="{ hach: p.hach }" /></g>
      <g class="tail" :class="{ on: tailOn }"><path v-for="(p, j) in r.tail" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" :class="{ hach: p.hach }" /></g>
    </g>
    <g class="axis"><path v-for="(p, j) in axis" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" /></g>
    <text v-for="t in ticks.filter(t => t.at !== marker.at)" :key="t.at" class="tick" :x="x(t.at)" :y="axisY + 20" :text-anchor="t.at >= total ? 'end' : (t.at <= 0 ? 'start' : 'middle')">{{ t.label }}</text>
    <g class="mk"><path v-for="(p, j) in mark" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" /></g>
    <text class="mk-label" :x="x(marker.at) + 10" y="12">{{ marker.label }}</text>
    <text v-if="inTail" class="tail in-tail" :class="{ on: tailOn }" :x="inTail.x" :y="inTail.y" text-anchor="middle" dominant-baseline="central">{{ afterLabel }}</text>
    <g v-if="keyed" class="tail key" :class="{ on: tailOn }">
      <path v-for="(p, j) in swatch" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" :class="{ hach: p.hach }" />
      <text class="key-label" x="48" :y="axisY + 54" dominant-baseline="central">{{ afterLabel }}</text>
    </g>
    <g v-for="(n, i) in noted" :key="'n' + i" class="note">
      <path v-for="(p, j) in n.paths" :key="j" :d="p.d" fill="none" stroke="currentColor" :stroke-width="p.sw" />
      <text :x="n.x + 6" :y="noteY" text-anchor="middle">{{ n.text }}</text>
    </g>
  </svg>
</template>

<style scoped>
.bars { display: block; width: 100%; height: auto; margin-top: 1.4rem; overflow: visible; }
.bars path { stroke-linecap: round; stroke-linejoin: round; }
.lab { font: 20px 'Virgil', var(--font-body); fill: var(--fg); }
.head { color: var(--accent); }
.head path.hach { opacity: 0.55; }
.tail { opacity: 0; transition: opacity 0.4s ease; }
.tail.on { opacity: 1; }
.after-waste .tail { color: var(--dont); }
.after-stop .tail, .after-clip .tail { color: var(--fg-dim); }
.tail path.hach { opacity: 0.6; }
.axis { color: var(--line); }
.tick { font: 17px 'Virgil', var(--font-body); fill: var(--fg-dim); }
.mk { color: var(--cat-2); }
.mk-label { font: 20px 'Virgil', var(--font-body); fill: var(--cat-2); }
.key .key-label { font: 19px 'Virgil', var(--font-body); fill: var(--fg-dim); }
.key { color: inherit; }
.in-tail { font: 19px 'Virgil', var(--font-body); fill: var(--fg-dim); }
.note { color: var(--dont); }
.note path { opacity: 0.8; }
.note text { font: 20px 'Virgil', var(--font-body); fill: var(--dont); }
.after-waste .key { color: var(--dont); }
.after-stop .key, .after-clip .key { color: var(--fg-dim); }
</style>
