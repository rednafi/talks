// Shared roughjs helpers, so every drawn shape in the deck has the same
// Excalidraw-like hand. Paths use currentColor; CSS sets the theme color.
import rough from 'roughjs'

export const gen = rough.generator()

export type RPath = { d: string; sw: number; hach: boolean; solid: boolean }

// Softer than the Excalidraw default, so the sketches sit with the theme's
// clean type instead of fighting it.
export const base = (seed: number, o: Record<string, any> = {}) => ({
  seed,
  roughness: 1.1,
  bowing: 1,
  stroke: 'currentColor',
  strokeWidth: 1.7,
  fillStyle: 'hachure',
  hachureGap: 6,
  hachureAngle: -41,
  fillWeight: 1.05,
  ...o,
})

export const toPaths = (ds: any[]): RPath[] =>
  ds.flatMap((d) => gen.toPaths(d)).map((p) => ({
    d: p.d,
    sw: p.strokeWidth,
    hach: Math.abs(p.strokeWidth - 1.05) < 0.01,
    solid: !!p.fill && p.fill !== 'none',
  }))

// A line with an open arrowhead at (x2, y2). `bend` curves it sideways.
export function arrow(x1: number, y1: number, x2: number, y2: number, seed: number, o: Record<string, any> = {}, bend = 0) {
  const opt = base(seed, o)
  let fx = x1, fy = y1
  let body
  if (bend) {
    const mx = (x1 + x2) / 2, my = (y1 + y2) / 2
    const len = Math.hypot(x2 - x1, y2 - y1) || 1
    fx = mx - ((y2 - y1) / len) * bend
    fy = my + ((x2 - x1) / len) * bend
    body = gen.curve([[x1, y1], [fx, fy], [x2, y2]], opt)
  } else {
    body = gen.line(x1, y1, x2, y2, opt)
  }
  const a = Math.atan2(y2 - fy, x2 - fx), l = o.head ?? 12, s = 0.45
  const ho = { ...opt, strokeLineDash: undefined, seed: seed + 3 }
  return [
    body,
    gen.line(x2, y2, x2 - l * Math.cos(a - s), y2 - l * Math.sin(a - s), ho),
    gen.line(x2, y2, x2 - l * Math.cos(a + s), y2 - l * Math.sin(a + s), ho),
  ]
}
