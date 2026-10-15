<script setup>
// A small timeline: one bar per row, a vertical marker, and what happens to
// each bar after the marker.
//   after="waste": the part past the marker shows as wasted work (on click).
//   after="stop":  bars end at the marker; the part they skip shows as an outline.
//   after="clip":  like stop, used for a deadline that a child can't extend.
import { computed } from 'vue'

const props = defineProps({
  rows: { type: Array, required: true }, // [{ label, from, to }]
  total: { type: Number, required: true },
  marker: { type: Object, required: true }, // { at, label }
  after: { type: String, default: 'waste' },
  afterLabel: { type: String, default: '' },
  ticks: { type: Array, default: () => [] }, // [{ at, label }]
  click: { type: Boolean, default: true },
})
const pct = (v) => `${(v / props.total) * 100}%`
const parts = computed(() => props.rows.map((r) => {
  const cut = Math.min(Math.max(props.marker.at, r.from), r.to)
  return {
    ...r,
    head: { left: pct(r.from), width: pct(cut - r.from) },
    tail: r.to > cut ? { left: pct(cut), width: pct(r.to - cut) } : null,
  }
}))
</script>

<template>
  <div class="bars" :class="`after-${after}`">
    <div class="bars-grid">
      <template v-for="(r, i) in parts" :key="i">
        <div class="bars-label">{{ r.label }}</div>
        <div class="bars-track">
          <div class="bar head" :style="r.head" />
          <div v-if="r.tail && click" v-click="1" class="bar tail" :style="r.tail" />
          <div v-else-if="r.tail" class="bar tail" :style="r.tail" />
        </div>
      </template>
      <div />
      <div class="bars-axis">
        <span v-for="t in ticks.filter(t => t.at !== marker.at)" :key="t.at" class="tick" :style="{ left: pct(t.at), transform: t.at >= total ? 'translateX(-100%)' : (t.at <= 0 ? 'none' : 'translateX(-50%)') }">{{ t.label }}</span>
      </div>
    </div>
    <div class="bars-overlay">
      <div class="bars-mark" :style="{ left: pct(marker.at) }">
        <span>{{ marker.label }}</span>
      </div>
    </div>
    <div v-if="afterLabel && click" v-click="1" class="bars-key"><span class="swatch" /> {{ afterLabel }}</div>
    <div v-else-if="afterLabel" class="bars-key"><span class="swatch" /> {{ afterLabel }}</div>
  </div>
</template>

<style scoped>
.bars { position: relative; width: 100%; --lab: 11.5rem; margin-top: 2.2rem; }
.bars-grid { display: grid; grid-template-columns: var(--lab) 1fr; row-gap: 0.9rem; align-items: center; }
.bars-label { font-family: var(--font-mono); font-size: 0.9rem; color: var(--fg); padding-right: 1rem; }
.bars-track { position: relative; height: 1.6rem; }
.bar { position: absolute; top: 0; bottom: 0; border-radius: 4px; }
.bar.head { background: color-mix(in oklab, var(--accent) 75%, transparent); }
.after-waste .bar.tail {
  background: repeating-linear-gradient(-45deg, color-mix(in oklab, var(--bad) 55%, transparent) 0 6px, transparent 6px 11px);
  border: 1px solid color-mix(in oklab, var(--bad) 70%, transparent);
}
.after-stop .bar.tail, .after-clip .bar.tail { border: 1.5px dashed var(--fg-dim); background: none; opacity: 0.7; }
.bars-axis { position: relative; height: 1.2rem; border-top: 1px solid var(--line); margin-top: 0.2rem; }
.tick { position: absolute; top: 0.35rem; transform: translateX(-50%); font-family: var(--font-mono); font-size: 0.72rem; color: var(--fg-dim); }
.bars-overlay { position: absolute; inset: -1.9rem 0 1.4rem var(--lab); pointer-events: none; }
.bars-mark { position: absolute; top: 0; bottom: 0; width: 0; border-left: 2px solid var(--cat-2); }
.bars-mark span { position: absolute; top: 0; left: 0.5rem; white-space: nowrap; font: 600 0.85rem/1 var(--font-body); color: var(--cat-2); }
.bars-key { margin-top: 1.4rem; font-size: 0.95rem; color: var(--fg-dim); display: flex; align-items: center; gap: 0.6rem; }
.after-waste .swatch { width: 1.6rem; height: 0.9rem; border-radius: 3px; background: repeating-linear-gradient(-45deg, color-mix(in oklab, var(--bad) 55%, transparent) 0 6px, transparent 6px 11px); border: 1px solid color-mix(in oklab, var(--bad) 70%, transparent); }
.after-stop .swatch, .after-clip .swatch { width: 1.6rem; height: 0.9rem; border-radius: 3px; border: 1.5px dashed var(--fg-dim); opacity: 0.7; }
</style>
