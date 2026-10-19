<script setup lang="ts">
// The context tree for one request, drawn by hand.
// Boxes are contexts. Ovals are the work that runs under them.
// Clicks: 1 WithCancel, 2 WithTimeout branch, 3 WithValue branch,
// 4 WithoutCancel branch, 5 the timeout fires, 6 the client hangs up.
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'
import { gen, base, toPaths, arrow } from '../lib/rough'

const { $clicks } = useSlideContext()
// Click 1 shows the root, so the tree runs one click behind the slide.
const c = computed(() => $clicks.value - 1)

type N = { id: string; x: number; y: number; label: string; ctx?: boolean; at: number; stop?: number }
const BW = 156, BH = 56, EW = 144, EH = 60
const nodes: N[] = [
  { id: 'root', x: 320, y: 32, label: 'Background()', ctx: true, at: 0 },
  { id: 'req', x: 320, y: 128, label: 'r.Context()', ctx: true, at: 0, stop: 6 },
  { id: 't', x: 160, y: 228, label: 'WithTimeout\n2s', ctx: true, at: 2, stop: 5 },
  { id: 'v', x: 480, y: 228, label: 'WithValue\ntrace ID', ctx: true, at: 3, stop: 6 },
  { id: 'db', x: 80, y: 332, label: 'db query', at: 2, stop: 5 },
  { id: 'rpc', x: 240, y: 332, label: 'inventory\nRPC', at: 2, stop: 5 },
  { id: 'log', x: 400, y: 332, label: 'logger', at: 3, stop: 6 },
  { id: 'wc', x: 560, y: 332, label: 'WithoutCancel', ctx: true, at: 4 },
  { id: 'audit', x: 560, y: 432, label: 'audit write', at: 4 },
]
const by = Object.fromEntries(nodes.map(n => [n.id, n]))
const edges = [['root', 'req'], ['req', 't'], ['req', 'v'], ['t', 'db'], ['t', 'rpc'], ['v', 'log'], ['v', 'wc'], ['wc', 'audit']]

const half = (n: N) => (n.ctx ? BH : EH) / 2
const opt = (seed: number) => base(seed, { roughness: 0.9, strokeWidth: 1.5 })

// Outlines, edges and stop marks are made once.
const outline = Object.fromEntries(nodes.map((n, i) => [n.id, toPaths([
  n.ctx
    ? gen.rectangle(n.x - BW / 2, n.y - BH / 2, BW, BH, opt(i * 17 + 3))
    : gen.ellipse(n.x, n.y, EW, EH, opt(i * 17 + 3)),
])]))
const cross = Object.fromEntries(nodes.map((n, i) => {
  const x = n.x + (n.ctx ? BW / 2 : EW / 2 - 14), y = n.y - half(n) + 2, r = 8
  return [n.id, toPaths([
    gen.line(x - r, y - r, x + r, y + r, base(i * 5 + 1, { strokeWidth: 2.4 })),
    gen.line(x + r, y - r, x - r, y + r, base(i * 5 + 2, { strokeWidth: 2.4 })),
  ])]
}))
const edgeOf = Object.fromEntries(edges.map(([a, b], i) => {
  const A = by[a], B = by[b]
  const bend = A.x === B.x ? 0 : (B.x < A.x ? 12 : -12)
  return [a + b, toPaths(arrow(A.x, A.y + half(A) + 3, B.x, B.y - half(B) - 4, i * 23 + 5, { strokeWidth: 1.4, head: 10, roughness: 0.8 }, bend))]
}))

const shown = (n: N) => c.value >= n.at
const stopped = (n: N) => n.stop !== undefined && c.value >= n.stop
const lines = (n: N) => n.label.split('\n')
</script>

<template>
  <svg class="ctree" viewBox="0 0 640 470" preserveAspectRatio="xMidYMin meet" xmlns="http://www.w3.org/2000/svg">
    <g v-for="[a, b] in edges" :key="a + b" class="ct-edge" :class="{ on: shown(by[b]), stop: stopped(by[b]) }">
      <path v-for="(p, j) in edgeOf[a + b]" :key="j" :d="p.d" :stroke-width="p.sw" />
    </g>
    <g v-for="n in nodes" :key="n.id" class="ct-node" :class="{ on: shown(n), ctx: n.ctx, stop: stopped(n) }">
      <rect v-if="n.ctx" class="ct-bg" :x="n.x - BW / 2 + 2" :y="n.y - BH / 2 + 2" :width="BW - 4" :height="BH - 4" rx="3" />
      <ellipse v-else class="ct-bg" :cx="n.x" :cy="n.y" :rx="EW / 2 - 3" :ry="EH / 2 - 3" />
      <path v-for="(p, j) in outline[n.id]" :key="j" class="ct-ol" :d="p.d" :stroke-width="p.sw" />
      <text :x="n.x" :y="n.y - (lines(n).length - 1) * 11" dominant-baseline="central">
        <tspan v-for="(l, j) in lines(n)" :key="j" :x="n.x" :dy="j ? 22 : 0">{{ l }}</tspan>
      </text>
      <g class="ct-x"><path v-for="(p, j) in cross[n.id]" :key="j" :d="p.d" :stroke-width="p.sw" /></g>
    </g>
 <text class="ct-note" :class="{ on: c >= 1 }" x="306" y="82" text-anchor="end">the server calls WithCancel</text>
    <text class="ct-ev end" :class="{ on: c === 5 }" x="70" y="190">timeout fires</text>
    <text class="ct-ev start" :class="{ on: c >= 6 }" x="414" y="84">client hangs up</text>
  </svg>
</template>

<style scoped>
.ctree { width: 100%; height: auto; max-height: 25rem; overflow: visible; }
.ct-node, .ct-edge { opacity: 0; transition: opacity 0.4s ease; }
.ct-node.on, .ct-edge.on { opacity: 1; }
.ct-node { color: var(--fg-dim); }
.ct-node.ctx { color: var(--accent); }
.ct-node .ct-bg { fill: var(--code-bg); stroke: none; }
.ct-node .ct-ol { fill: none; stroke: currentColor; stroke-linecap: round; transition: color 0.4s; }
.ct-node text { font: 19px 'Virgil', var(--font-body); fill: var(--fg); text-anchor: middle; transition: fill 0.4s; }
.ct-node .ct-x { opacity: 0; transition: opacity 0.3s; color: #ff9e64; }
.ct-node .ct-x path { fill: none; stroke: currentColor; stroke-linecap: round; }
.ct-node.stop { color: color-mix(in oklab, var(--fg-dim) 55%, transparent); }
.ct-node.stop text { fill: var(--fg-dim); }
.ct-node.stop .ct-x { opacity: 1; }
.ct-edge { color: color-mix(in oklab, var(--accent) 75%, transparent); }
.ct-edge path { fill: none; stroke: currentColor; stroke-linecap: round; transition: color 0.4s; }
.ct-edge.stop { color: color-mix(in oklab, var(--fg-dim) 45%, transparent); }
.ct-ev { font: 19px 'Virgil', var(--font-body); fill: #ff9e64; opacity: 0; transition: opacity 0.3s; }
.ct-ev.on { opacity: 1; }
.ct-note { font: 17px 'Virgil', var(--font-body); fill: var(--fg-dim); opacity: 0; transition: opacity 0.3s; }
.ct-note.on { opacity: 1; }
.ct-ev.start { text-anchor: start; }
.ct-ev.end { text-anchor: start; }
</style>
