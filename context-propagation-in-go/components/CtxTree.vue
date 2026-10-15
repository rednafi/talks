<script setup lang="ts">
import { computed } from 'vue'
import { useSlideContext } from '@slidev/client'

// Clicks: 1 WithCancel, 2 WithTimeout branch, 3 WithValue branch,
// 4 WithoutCancel branch, 5 the timeout fires, 6 the client hangs up.
const { $clicks } = useSlideContext()
const c = computed(() => $clicks.value)

type N = { id: string; x: number; y: number; label: string; ctx?: boolean; at: number; stop?: number }
const W = 140, H = 36
const nodes: N[] = [
  { id: 'root', x: 300, y: 24, label: 'Background()', ctx: true, at: 0 },
  { id: 'req', x: 300, y: 100, label: 'request ctx', ctx: true, at: 0, stop: 6 },
  { id: 't', x: 147, y: 180, label: 'WithTimeout(2s)', ctx: true, at: 2, stop: 5 },
  { id: 'v', x: 453, y: 180, label: 'WithValue(trace)', ctx: true, at: 3, stop: 6 },
  { id: 'db', x: 72, y: 262, label: 'db query', at: 2, stop: 5 },
  { id: 'rpc', x: 222, y: 262, label: 'inventory RPC', at: 2, stop: 5 },
  { id: 'log', x: 378, y: 262, label: 'logger', at: 3, stop: 6 },
  { id: 'wc', x: 528, y: 262, label: 'WithoutCancel', ctx: true, at: 4 },
  { id: 'audit', x: 528, y: 344, label: 'audit write', at: 4 },
]
const by = Object.fromEntries(nodes.map(n => [n.id, n]))
const edges = [['root', 'req'], ['req', 't'], ['req', 'v'], ['t', 'db'], ['t', 'rpc'], ['v', 'log'], ['v', 'wc'], ['wc', 'audit']]

const shown = (n: N) => c.value >= n.at
const stopped = (n: N) => n.stop !== undefined && c.value >= n.stop
const path = (a: N, b: N) => {
  const y1 = a.y + H / 2, y2 = b.y - H / 2, m = (y1 + y2) / 2
  return `M${a.x},${y1} C${a.x},${m} ${b.x},${m} ${b.x},${y2 - 4}`
}
</script>

<template>
  <svg class="ctree" viewBox="0 0 600 372" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <marker id="ct-arrow" viewBox="0 0 8 8" refX="4" refY="4" markerWidth="7" markerHeight="7" orient="auto">
        <path d="M0,0 L8,4 L0,8 z" fill="currentColor" />
      </marker>
    </defs>
    <g v-for="([a, b]) in edges" :key="a + b" class="edge" :class="{ on: shown(by[b]), stop: stopped(by[b]) }">
      <path :d="path(by[a], by[b])" marker-end="url(#ct-arrow)" />
    </g>
    <g v-for="n in nodes" :key="n.id" class="node" :class="{ on: shown(n), ctx: n.ctx, stop: stopped(n), fresh: c === n.at && n.at > 0 }">
      <rect :x="n.x - W / 2" :y="n.y - H / 2" :width="W" :height="H" rx="6" />
      <text :x="n.x" :y="n.y + 4.5">{{ n.label }}</text>
    </g>
    <g class="event" :class="{ on: c === 5 }">
      <text x="136" y="154" class="end">timeout fires</text>
    </g>
    <g class="event" :class="{ on: c >= 6 }">
      <text x="380" y="105" class="start">client hangs up</text>
    </g>
  </svg>
</template>

<style scoped>
.ctree { width: 100%; height: auto; overflow: visible; }
.node, .edge { opacity: 0; transition: opacity 0.4s ease, transform 0.4s ease; }
.node { transform: translateY(6px); transform-box: fill-box; }
.node.on, .edge.on { opacity: 1; transform: none; }
.node rect { fill: var(--ink-2); stroke: var(--fg-dim); stroke-width: 1; transition: stroke 0.4s, fill 0.4s; }
.node.ctx rect { stroke: var(--accent); }
.node text { font: 500 14px var(--font-mono); fill: var(--fg); text-anchor: middle; transition: fill 0.4s; }
.edge path { fill: none; stroke: var(--accent); stroke-width: 1.2; color: var(--accent); transition: stroke 0.4s, color 0.4s; }
.node.stop rect { stroke: var(--bad); fill: color-mix(in srgb, var(--bad) 14%, var(--ink-2)); }
.node.stop text { fill: var(--fg-dim); }
.edge.stop path { stroke: var(--bad); color: var(--bad); opacity: 0.6; }
.event { opacity: 0; transition: opacity 0.3s; }
.event.on { opacity: 1; }
.event text.start { text-anchor: start; }
.event text.end { text-anchor: end; }
.event text { font: 600 13px var(--font-body); fill: var(--bad); text-anchor: middle; letter-spacing: 0.04em; }
</style>
