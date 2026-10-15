const QR=require('qrcode');
const url='https://cdn.jsdelivr.net/gh/rednafi/talks@main/context-propagation-in-go/slides.pdf';
const q=QR.create(url,{errorCorrectionLevel:'H'});
const n=q.modules.size, d=q.modules.data, m=3, S=n+2*m;
const fg='#e9ecf5', ring='#8590ff', fg2='#e9ecf5', bg='#141824';
const isFinder=(r,c)=>(r<7&&c<7)||(r<7&&c>=n-7)||(r>=n-7&&c<7);
const c0=n/2, hole=n*0.13; // logo radius in modules
let out=`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${S} ${S}"><rect width="${S}" height="${S}" rx="2.2" fill="${bg}" stroke="#262c3b" stroke-width="0.15"/>`;
for(let r=0;r<n;r++)for(let c=0;c<n;c++){
  if(!d[r*n+c]||isFinder(r,c))continue;
  if(Math.hypot(r+0.5-c0,c+0.5-c0)<hole+0.9)continue;
  out+=`<circle cx="${c+m+0.5}" cy="${r+m+0.5}" r="0.43" fill="${fg}"/>`;
}
for(const [r,c] of [[0,0],[0,n-7],[n-7,0]]){
  const x=c+m,y=r+m;
  out+=`<rect x="${x+0.5}" y="${y+0.5}" width="6" height="6" rx="1.6" fill="none" stroke="${ring}" stroke-width="1"/>`;
  out+=`<rect x="${x+2}" y="${y+2}" width="3" height="3" rx="0.8" fill="${fg2}"/>`;
}
const cx=c0+m;
out+=`<circle cx="${cx}" cy="${cx}" r="${hole+0.4}" fill="${bg}"/>`;
out+='</svg>';
require('fs').writeFileSync(process.argv[2],out); console.log(n);
