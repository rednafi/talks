<script setup lang="ts">
// Hand-drawn diagrams in the Excalidraw style. roughjs draws the shapes and
// Virgil (the Excalidraw font) sets the labels. Colors come from the theme.
// Each item can appear on a click with `at`, like the other reveals.
//
// Items (coordinates are in viewBox units):
//   { box: [x, y, w, h], text?, color?, tc?, fill?, dashed?, at? }
//   { ellipse: [cx, cy, w, h], text?, color?, tc?, fill?, dashed?, at? }
//   { diamond: [cx, cy, w, h], text?, color?, tc?, at? }
//   { arrow: [x1, y1, x2, y2], bend?, color?, dashed?, at? }
//   { line: [x1, y1, x2, y2], color?, dashed?, at? }
//   { text: 'label', x, y, size?, anchor?, color?, mono?, at? }
//   { cross: [cx, cy, r], color?, at? }
//   { check: [cx, cy, r], color?, at? }
// color (stroke) and tc (text): fg | dim | line | accent | info | dont | do
// `fill: true` hachures the shape in its stroke color; `fs: 'cross-hatch'` cross-fills it.
// All text, `mono` too, is set in Virgil, so a diagram reads as one drawing.
// `until` hides the item again from that click on.
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'
import { gen, base, toPaths, arrow } from '../lib/rough'

const props = defineProps({
  w: { type: Number, default: 600 },
  h: { type: Number, default: 400 },
  items: { type: Array, required: true },
  height: { type: String, default: '' },
})

const { $clicks } = useSlideContext()

const shapes = computed(() => (props.items as any[]).map((it, i) => {
  const seed = (it.seed ?? i + 1) * 7
  const o = base(seed, {
    strokeWidth: it.sw ?? 1.7,
    strokeLineDash: it.dashed ? [7, 6] : undefined,
    fill: it.fill ? 'currentColor' : undefined,
    roughness: it.roughness ?? 1.1,
    fillStyle: it.fs ?? 'hachure',
  })
  let ds: any[] = []
  let bg: { kind: string; a: number[] } | null = null
  let label: { x: number; y: number; lines: string[] } | null = null
  const lines = (t: any) => String(t).split('\n')
  let color = it.color ?? 'fg'
  if (it.box) {
    const [x, y, w, h] = it.box
    ds = [gen.rectangle(x, y, w, h, o)]
    bg = { kind: 'rect', a: [x + 2, y + 2, w - 4, h - 4] }
    if (it.text) label = { x: x + w / 2, y: y + h / 2, lines: lines(it.text) }
    color = it.color ?? 'dim'
  } else if (it.ellipse) {
    const [cx, cy, w, h] = it.ellipse
    ds = [gen.ellipse(cx, cy, w, h, o)]
    bg = { kind: 'ellipse', a: [cx, cy, w / 2 - 2, h / 2 - 2] }
    if (it.text) label = { x: cx, y: cy, lines: lines(it.text) }
    color = it.color ?? 'dim'
  } else if (it.diamond) {
    const [cx, cy, w, h] = it.diamond
    const pts: [number, number][] = [[cx, cy - h / 2], [cx + w / 2, cy], [cx, cy + h / 2], [cx - w / 2, cy]]
    ds = [gen.polygon(pts, o)]
    bg = { kind: 'poly', a: pts.flat() }
    if (it.text) label = { x: cx, y: cy, lines: lines(it.text) }
    color = it.color ?? 'dim'
  } else if (it.arrow) {
    const [x1, y1, x2, y2] = it.arrow
    ds = arrow(x1, y1, x2, y2, seed, { ...o, fill: undefined }, it.bend ?? 0)
    color = it.color ?? 'dim'
  } else if (it.line) {
    const [x1, y1, x2, y2] = it.line
    ds = [gen.line(x1, y1, x2, y2, o)]
    color = it.color ?? 'dim'
  } else if (it.cross) {
    const [cx, cy, r] = it.cross
    ds = [gen.line(cx - r, cy - r, cx + r, cy + r, o), gen.line(cx + r, cy - r, cx - r, cy + r, o)]
    color = it.color ?? 'dont'
  } else if (it.check) {
    const [cx, cy, r] = it.check
    ds = [gen.linearPath([[cx - r, cy], [cx - r * 0.3, cy + r * 0.7], [cx + r, cy - r * 0.8]], o)]
    color = it.color ?? 'do'
  } else if (it.text !== undefined) {
    label = { x: it.x, y: it.y, lines: lines(it.text) }
  }
  return {
    key: i,
    at: it.at ?? 0,
    until: it.until ?? Infinity,
    color,
    tc: it.tc ?? (it.text !== undefined && !it.box && !it.ellipse && !it.diamond ? color : 'fg'),
    size: it.size ?? 20,
    anchor: it.anchor ?? 'middle',
    mono: !!it.mono,
    filled: !!it.fill,
    bg: it.nobg ? null : bg,
    paths: toPaths(ds),
    label,
  }
}))

const shown = (s: { at: number; until: number }) => $clicks.value >= s.at && $clicks.value < s.until
</script>

<template>
  <svg class="sketch" :viewBox="`0 0 ${w} ${h}`" :style="height ? { height } : {}" xmlns="http://www.w3.org/2000/svg">
    <g v-for="s in shapes" :key="s.key" :class="['sk', `sk-${s.color}`, { on: shown(s) }]">
      <template v-if="s.bg">
        <rect v-if="s.bg.kind === 'rect'" class="sk-bg" :x="s.bg.a[0]" :y="s.bg.a[1]" :width="s.bg.a[2]" :height="s.bg.a[3]" rx="3" />
        <ellipse v-else-if="s.bg.kind === 'ellipse'" class="sk-bg" :cx="s.bg.a[0]" :cy="s.bg.a[1]" :rx="s.bg.a[2]" :ry="s.bg.a[3]" />
        <polygon v-else class="sk-bg" :points="s.bg.a.join(' ')" />
      </template>
      <path
        v-for="(p, j) in s.paths" :key="j" :d="p.d"
        :fill="p.solid ? 'currentColor' : 'none'" stroke="currentColor" :stroke-width="p.sw"
        :class="{ hach: p.hach }"
      />
      <text
        v-if="s.label"
        :class="[`tc-${s.tc}`, { mono: s.mono }]"
        :x="s.label.x"
        :y="s.label.y - ((s.label.lines.length - 1) * s.size * 1.22) / 2"
        :font-size="s.size"
        :text-anchor="s.anchor"
        dominant-baseline="central"
      >
        <tspan v-for="(l, j) in s.label.lines" :key="j" :x="s.label.x" :dy="j ? s.size * 1.22 : 0">{{ l }}</tspan>
      </text>
    </g>
  </svg>
</template>

<style scoped>
.sketch { display: block; width: 100%; height: auto; overflow: visible; }
.sk { opacity: 0; transition: opacity 0.35s ease; }
.sk.on { opacity: 1; }
.sk path { stroke-linecap: round; stroke-linejoin: round; }
.sk path.hach { opacity: 0.45; }
.sk-bg { fill: var(--code-bg); stroke: none; }
.sk text { font-family: 'Virgil', var(--font-body); stroke: none; }
.sk text.mono { font-family: 'Virgil', var(--font-body); }
.sk-fg { color: color-mix(in oklab, var(--fg) 82%, transparent); }
.sk-dim { color: color-mix(in oklab, var(--fg-dim) 85%, transparent); }
.sk-line { color: var(--line); }
.sk-accent { color: var(--accent); }
.sk-info { color: var(--cat-2); }
.sk-dont { color: #ff9e64; }
.sk-do { color: #9ece6a; }
.tc-fg { fill: var(--fg); }
.tc-dim { fill: var(--fg-dim); }
.tc-line { fill: var(--fg-dim); }
.tc-accent { fill: var(--accent); }
.tc-info { fill: var(--cat-2); }
.tc-dont { fill: #ff9e64; }
.tc-do { fill: #9ece6a; }
</style>
