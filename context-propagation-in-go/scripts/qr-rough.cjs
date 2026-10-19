// Hand-drawn QR code: each run of dark modules is a rough, pencil-filled box.
// Usage: node scripts/qr-rough.cjs public/img/qr-pdf.svg
const QR = require('qrcode')
const rough = require('roughjs/bundled/rough.cjs.js')
const url = 'https://cdn.jsdelivr.net/gh/rednafi/talks@main/context-propagation-in-go/slides.pdf'
const q = QR.create(url, { errorCorrectionLevel: 'H' })
const n = q.modules.size, d = q.modules.data, U = 12, m = 3, S = (n + 2 * m) * U
const fg = '#e9ecf5', ring = '#8590ff'
const gen = rough.generator()
let seed = 7
const isFinder = (r, c) => (r < 7 && c < 7) || (r < 7 && c >= n - 7) || (r >= n - 7 && c < 7)
const c0 = n / 2, hole = n * 0.13
const inHole = (r, c) => Math.hypot(r + 0.5 - c0, c + 0.5 - c0) < hole + 0.9
const dark = (r, c) => d[r * n + c] && !isFinder(r, c) && !inHole(r, c)
let out = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${S} ${S}" fill="none" stroke-linecap="round">`
const draw = (dr) => {
  for (const p of gen.toPaths(dr)) out += `<path d="${p.d.replace(/-?\d+\.\d+/g, (v) => Math.round(+v))}" stroke="${p.stroke}" stroke-width="${p.strokeWidth}" fill="${p.fill || 'none'}"/>`
}
const P = (x) => (x + m) * U
for (let r = 0; r < n; r++) {
  for (let c = 0; c < n; c++) {
    if (!dark(r, c)) continue
    let e = c
    while (e + 1 < n && dark(r, e + 1)) e++
    const w = (e - c + 1) * U
    draw(gen.rectangle(P(c) + 1, P(r) + 1, w - 2, U - 2, {
      seed: seed++, roughness: 0.7, bowing: 1, stroke: fg, strokeWidth: 1.6,
      fill: fg, fillStyle: 'zigzag', fillWeight: 2, hachureGap: 2.8, hachureAngle: -41,
    }))
    c = e
  }
}
for (const [r, c] of [[0, 0], [0, n - 7], [n - 7, 0]]) {
  draw(gen.rectangle(P(c) + U / 2, P(r) + U / 2, 6 * U, 6 * U, { seed: seed++, roughness: 0.6, bowing: 0.5, disableMultiStroke: true, stroke: ring, strokeWidth: U * 0.9 }))
  draw(gen.rectangle(P(c) + 2 * U, P(r) + 2 * U, 3 * U, 3 * U, {
    seed: seed++, roughness: 0.6, stroke: fg, strokeWidth: 2, fill: fg, fillStyle: 'cross-hatch', fillWeight: 3, hachureGap: 2.2,
  }))
}
out += '</svg>'
require('fs').writeFileSync(process.argv[2], out)
console.log(n, out.length)
