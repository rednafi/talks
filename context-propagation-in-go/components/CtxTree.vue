<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'
import { gen, base, toPaths, arrow } from '../lib/rough'

// Clicks: 1 WithCancel, 2 WithTimeout branch, 3 WithValue branch,
// 4 WithoutCancel branch, 5 the timeout fires, 6 the client hangs up.
const { $clicks } = useSlideContext()
// Click 1 shows the root, so the tree runs one click behind the slide.
const c = computed(() => $clicks.value - 1)

type N = { id: string; x: number; y: number; label: string; ctx?: boolean; at: number; stop?: number }
const W = 134, H = 40
const nodes: N[] = [
  { id: 'root', x: 300, y: 24, label: 'Background()', ctx: true, at: 0 },
  { id: 'req', x: 300, y: 112, label: 'request ctx', ctx: true, at: 0, stop: 6 },
  { id: 't', x: 147, y: 204, label: 'WithTimeout(2s)', ctx: true, at: 2, stop: 5 },
  { id: 'v', x: 453, y: 204, label: 'WithValue(trace)', ctx: true, at: 3, stop: 6 },
  { id: 'db', x: 72, y: 296, label: 'db query', at: 2, stop: 5 },
  { id: 'rpc', x: 222, y: 296, label: 'inventory RPC', at: 2, stop: 5 },
  { id: 'log', x: 378, y: 296, label: 'logger', at: 3, stop: 6 },
  { id: 'wc', x: 528, y: 296, label: 'WithoutCancel', ctx: true, at: 4 },
  { id: 'audit', x: 528, y: 388, label: 'audit write', at: 4 },
]
const by = Object.fromEntries(nodes.map(n => [n.id, n]))
const edges = [['root', 'req'], ['req', 't'], ['req', 'v'], ['t', 'db'], ['t', 'rpc'], ['v', 'log'], ['v', 'wc'], ['wc', 'audit']]

// Hand-drawn outlines and edges, made once.
const box = Object.fromEntries(nodes.map((n, i) => [n.id,
  toPaths([gen.rectangle(n.x - W / 2, n.y - H / 2, W, H, base(i * 17 + 3, { roughness: 1 }))])]))
const edgeOf = (a: N, b: N, i: number) => {
  const y1 = a.y + H / 2 + 2, y2 = b.y - H / 2 - 3
  const bend = a.x === b.x ? 0 : (b.x < a.x ? 14 : -14)
  return toPaths(arrow(a.x, y1, b.x, y2, i * 23 + 5, { strokeWidth: 1.5, head: 9 }, bend))
}

const shown = (n: N) => c.value >= n.at
const stopped = (n: N) => n.stop !== undefined && c.value >= n.stop
const path = (a: N, b: N) => {
  const y1 = a.y + H / 2, y2 = b.y - H / 2, m = (y1 + y2) / 2
  return `M${a.x},${y1} C${a.x},${m} ${b.x},${m} ${b.x},${y2 - 4}`
}
</script>

<template>
  <svg class="ctree" viewBox="0 0 600 412" preserveAspectRatio="xMidYMin meet" xmlns="http://www.w3.org/2000/svg">
    <g v-for="([a, b], i) in edges" :key="a + b" class="edge" :class="{ on: shown(by[b]), stop: stopped(by[b]) }">
      <path v-for="(p, j) in edgeOf(by[a], by[b], i)" :key="j" :d="p.d" :stroke-width="p.sw" />
    </g>
    <g v-for="n in nodes" :key="n.id" class="node" :class="{ on: shown(n), ctx: n.ctx, stop: stopped(n), fresh: c === n.at && n.at > 0 }">
      <rect :x="n.x - W / 2 + 2" :y="n.y - H / 2 + 2" :width="W - 4" :height="H - 4" rx="3" />
      <path v-for="(p, j) in box[n.id]" :key="j" class="outline" :d="p.d" :stroke-width="p.sw" />
      <text :x="n.x" :y="n.y + 7" :textLength="n.label.length > 9 ? W - 12 : undefined" lengthAdjust="spacingAndGlyphs">{{ n.label }}</text>
    </g>
    <g class="event" :class="{ on: c === 5 }">
      <text x="136" y="172" class="end">timeout fires</text>
    </g>
    <g class="event" :class="{ on: c >= 6 }">
      <text x="380" y="100" class="start">client hangs up</text>
    </g>
  </svg>
</template>

<style scoped>
.ctree { width: 100%; height: auto; max-height: 25rem; overflow: visible; }
.node, .edge { opacity: 0; transition: opacity 0.4s ease, transform 0.4s ease; }
.node { transform: translateY(6px); transform-box: fill-box; }
.node.on, .edge.on { opacity: 1; transform: none; }
.node rect { fill: var(--code-bg); stroke: none; transition: fill 0.4s; }
.node { color: var(--fg-dim); }
.node.ctx { color: var(--accent); }
.node .outline { fill: none; stroke: currentColor; stroke-linecap: round; transition: color 0.4s; }
.node text { font: 21px 'Virgil', var(--font-body); fill: var(--fg); text-anchor: middle; transition: fill 0.4s; }
.edge { color: var(--accent); }
.edge path { fill: none; stroke: currentColor; stroke-linecap: round; transition: color 0.4s; }
.node.stop { color: var(--bad); }
.node.stop rect { fill: color-mix(in srgb, var(--bad) 14%, var(--code-bg)); }
.node.stop text { fill: var(--fg-dim); }
.edge.stop { color: var(--bad); }
.edge.stop path { opacity: 0.6; }
.event { opacity: 0; transition: opacity 0.3s; }
.event.on { opacity: 1; }
.event text.start { text-anchor: start; }
.event text.end { text-anchor: end; }
.event text { font: 17px 'Virgil', var(--font-body); fill: var(--bad); text-anchor: middle; letter-spacing: 0.04em; }
</style>
