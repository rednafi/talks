import { chromium } from 'playwright-chromium'
const b = await chromium.launch()
for (const [w,h,path] of [[1440,900,'/41'],[2560,1440,'/41'],[1280,700,'/41'],[1600,900,'/presenter/41']]) {
  const ctx = await b.newContext({ viewport: { width: w, height: h } }); const p = await ctx.newPage()
  await p.goto('http://localhost:3030'+path, { waitUntil: 'networkidle' }); await p.waitForTimeout(1500)
  const res = await p.evaluate(() => [...document.querySelectorAll('.refs a')].map(a => { const r=a.getBoundingClientRect(); const e=document.elementFromPoint(r.x+Math.min(20,r.width/2), r.y+r.height/2); return (e===a||a.contains(e)) ? 'ok' : (e? e.tagName+'.'+String(e.className).slice(0,40):'none') }))
  console.log(w,h,path, [...new Set(res)].join(' | '), res.length)
  await ctx.close()
}
await b.close()
