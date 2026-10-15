// Usage: node scripts/shots.mjs [from] [to] [clicks]
// Screenshots slides from the running dev server into shots/.
import { chromium } from 'playwright-chromium'
import { mkdirSync } from 'node:fs'
const [from = 1, to = 40, clicks = 99] = process.argv.slice(2).map(Number)
mkdirSync('shots', { recursive: true })
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1600, height: 900 } })
for (let n = from; n <= to; n++) {
  const r = await p.goto(`http://localhost:3030/${n}?clicks=${clicks}`, { waitUntil: 'networkidle' })
  if (!r || !r.ok()) break
  await p.waitForTimeout(1600)
  await p.screenshot({ path: `shots/${String(n).padStart(2, '0')}.png` })
}
await b.close()
