// Usage: node scripts/sheet.mjs out.png 01 02 03 04  -> 2x2 contact sheet from shots/
import { chromium } from 'playwright-chromium'
import { resolve } from 'node:path'
import { writeFileSync } from 'node:fs'
const [out, ...ids] = process.argv.slice(2)
const html = `<body style="margin:0;background:#333;display:grid;grid-template-columns:1fr 1fr;gap:4px">${ids.map(i => `<img src="${i}.png" style="width:800px">`).join('')}</body>`
const b = await chromium.launch(); const p = await b.newPage({ viewport: { width: 1604, height: 908 } })
writeFileSync('shots/sheet.html', html); await p.goto('file://' + resolve('shots/sheet.html')); await p.waitForTimeout(300)
await p.screenshot({ path: out, fullPage: false }); await b.close()
