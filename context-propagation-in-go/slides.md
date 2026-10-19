---
theme: slidev-theme-tahta
title: Context propagation in Go
themeConfig:
  variant: atelier
layout: cover
kicker: Redowan Delowar
subtitle: Lessons from operating large-scale distributed systems at DoorDash (Wolt)
fonts:
  sans: Hanken Grotesk
  mono: JetBrains Mono
highlighter: shiki
lineNumbers: false
canvasWidth: 1120
aspectRatio: 16/9
drawings:
  persist: false
mdc: true
transition: fade
---

<!--
This talk is about context.Context. What it is, and how to use it so that every cancel tells you why.
-->
---
layout: default
class: qr-slide
---

<div class="qr">
  <div class="qr-code">
    <img src="/img/qr-pdf.svg?v=4" alt="QR code to the PDF of these slides" />
    <img class="qr-gopher" src="/img/gopher-standing.svg" alt="" />
  </div>
  <p>Scan to get the slides</p>
  <span>https://rednafi.com/appearances</span>
</div>

<!--
This code opens the slides as a PDF. No need to take photos.
-->
---
layout: default
class: whoami
title: "$ whoami"
---

<div class="about">
<div class="about-text">

<v-clicks>

Hi, I'm Redowan, **@rednafi** on the web.

I oscillate between product and platform work at DoorDash/Wolt.

Interested in databases, streaming, o11y, and failure modes of large-scale distsys.

Writing at **rednafi.com**.

<div class="about-links">github.com/rednafi<br>linkedin.com/in/rednafi</div>

</v-clicks>

</div>
<img class="about-gopher" src="/img/gopher-hiking.svg" alt="" />
</div>

<!--
[click] Hi, I'm Redowan. [click] I oscillate between product and platform work, [click] and I care most about how distributed systems fail at scale. [click] I write at rednafi.com, [click] and these are my links.
-->
---
layout: default
clicks: 7
title: Overview
---

<Sketch :w="1000" :h="440" style="margin-top:0.4rem" :items="[
  { curve: [[596, 40], [603, 32], [610, 40]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[610, 40], [617, 32], [624, 40]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[631, 62], [638, 54], [645, 62]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[645, 62], [652, 54], [659, 62]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[846, 30], [853, 22], [860, 30]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[860, 30], [867, 22], [874, 30]], color: 'dim', sw: 1.6, roughness: 0.6 },
  { curve: [[30, 360.9], [55, 354.2], [80, 345.9], [105, 342.0], [130, 345.2], [155, 353.3], [180, 360.4], [205, 361.5], [230, 355.9], [255, 347.4], [280, 342.2], [305, 344.0], [330, 351.5], [355, 359.3], [380, 361.9], [405, 357.4], [430, 349.0], [455, 342.8], [480, 343.1], [505, 349.8], [530, 358.1], [555, 362.0], [580, 358.8], [605, 350.7], [630, 343.6], [655, 342.5], [680, 348.1], [705, 356.6], [730, 361.8], [755, 360.0], [780, 352.5], [805, 344.7], [830, 342.1], [855, 346.6], [880, 355.0], [905, 361.2], [930, 360.9], [955, 354.2], [980, 345.9]], color: 'line', roughness: 1.6 },
  { curve: [[90, 385.5], [115, 393.4], [140, 395.8], [165, 390.5], [190, 382.6], [215, 380.2], [240, 385.5], [265, 393.4], [290, 395.8], [315, 390.5], [340, 382.6], [365, 380.2], [390, 385.5], [415, 393.4], [440, 395.8], [465, 390.5], [490, 382.6], [515, 380.2], [540, 385.5], [565, 393.4], [590, 395.8], [615, 390.5], [640, 382.6], [665, 380.2], [690, 385.5], [715, 393.4], [740, 395.8], [765, 390.5], [790, 382.6], [815, 380.2], [840, 385.5], [865, 393.4], [890, 395.8], [915, 390.5]], color: 'line', roughness: 1.6 },
  { curve: [[170, 427.1], [195, 419.3], [220, 415.0], [245, 419.8], [270, 427.4], [295, 428.1], [320, 420.9], [345, 415.1], [370, 418.2], [395, 426.2], [420, 428.7], [445, 422.6], [470, 415.7], [495, 416.9], [520, 424.7], [545, 429.0], [570, 424.2], [595, 416.6], [620, 415.9], [645, 423.1], [670, 428.9], [695, 425.8], [720, 417.8], [745, 415.3], [770, 421.4], [795, 428.3], [820, 427.1]], color: 'line', roughness: 1.6 },
  { curve: [[0, 210.0], [25, 227.5], [50, 240.0], [75, 244.0], [100, 238.1], [125, 224.3], [150, 206.3], [175, 189.4], [200, 178.4], [225, 176.4], [250, 184.1], [275, 199.1], [300, 217.3], [325, 233.4], [350, 242.8], [375, 242.8], [400, 233.4], [425, 217.3], [450, 199.1], [475, 184.1], [500, 176.4], [525, 178.4], [550, 189.4], [575, 206.3], [600, 224.3], [625, 238.1], [650, 244.0], [675, 240.0], [700, 227.5], [725, 210.0], [750, 192.5], [775, 180.0], [800, 176.0], [825, 181.9], [850, 195.7], [875, 213.7], [900, 230.6], [925, 241.6], [950, 243.6], [975, 235.9], [1000, 220.9]], color: 'info', sw: 2.2, roughness: 1.2 },
  { ellipse: [130, 382, 18, 18], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [150, 407, 12, 12], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [452, 397, 16, 16], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [470, 424, 10, 10], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [705, 382, 14, 14], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [725, 404, 10, 10], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [880, 402, 16, 16], color: 'dim', nobg: true, roughness: 0.8 },
  { ellipse: [70, 244.0, 70, 70], text: '01', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 1 },
  { line: [70, 206.0, 70, 166.0], color: 'dim', dashed: true, at: 1 },
  { text: 'A log line with\nno cause', x: 70, y: 132.0, size: 24, color: 'fg', at: 1 },
  { ellipse: [210, 176.4, 70, 70], text: '02', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 2 },
  { line: [210, 214.4, 210, 254.39999999999998], color: 'dim', dashed: true, at: 2 },
  { text: 'Context primer', x: 210, y: 288.4, size: 24, color: 'fg', at: 2 },
  { ellipse: [350, 242.8, 70, 70], text: '03', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 3 },
  { line: [350, 204.8, 350, 164.8], color: 'dim', dashed: true, at: 3 },
  { text: 'Passing a\ncontext down', x: 350, y: 130.8, size: 24, color: 'fg', at: 3 },
  { ellipse: [500, 176.4, 70, 70], text: '04', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 4 },
  { line: [500, 214.4, 500, 254.39999999999998], color: 'dim', dashed: true, at: 4 },
  { text: 'Context values', x: 500, y: 288.4, size: 24, color: 'fg', at: 4 },
  { ellipse: [640, 242.8, 70, 70], text: '05', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 5 },
  { line: [640, 204.8, 640, 164.8], color: 'dim', dashed: true, at: 5 },
  { text: 'Giving every\ncancel a cause', x: 640, y: 130.8, size: 24, color: 'fg', at: 5 },
  { ellipse: [780, 178.4, 70, 70], text: '06', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 6 },
  { line: [780, 216.4, 780, 256.4], color: 'dim', dashed: true, at: 6 },
  { text: 'Enforcing\nthe rules', x: 780, y: 290.4, size: 24, color: 'fg', at: 6 },
  { ellipse: [920, 240.0, 70, 70], text: '07', size: 24, color: 'accent', fill: true, fs: 'cross-hatch', at: 7 },
  { line: [920, 202.0, 920, 162.0], color: 'dim', dashed: true, at: 7 },
  { text: 'Checklist\nand limits', x: 920, y: 128.0, size: 24, color: 'fg', at: 7 },
]" />

<!--
[click] First, the problem: a log line that doesn't say why the context stopped.

[click] Then a short primer. What a context carries, and how it ends. [click] Then five habits for passing it down. [click] Then values, which move down the same tree. [click] Then the core of the talk. Every cancel gets a cause, and the cause gets to the log. [click] Then tools that check all of this. [click] Last, a checklist for code review, and the cancels that still have no cause.

Each slide is one rule and the bug it prevents. Take the ones that fit your code.
-->
---
layout: two-cols
class: wide-left
clicks: 7
title: The log doesn't say why the context stopped
---

<div v-click>

<Terminal title="order-service" :lines="[
  'ERROR place order: context canceled',
  'ERROR fetch menu: context deadline exceeded',
  'ERROR place order: context canceled',
  'ERROR charge card: context canceled',
  'ERROR fetch menu: context deadline exceeded',
  'ERROR place order: context canceled',
]" />

</div>

<img src="/img/gopher-mistake.svg" class="gopher" style="max-height:8.5rem; margin-top:1.2rem" />

::right::

<Sketch :w="440" :h="400" :items="[
  { ellipse: [335, 160, 130, 86], text: 'ctx', size: 26, color: 'accent', at: 2 },
  { text: 'Any of these can stop it', x: 0, y: 8, anchor: 'start', size: 19, color: 'dim', at: 2 },
  { box: [0, 34, 192, 50], text: 'the client hangs up', size: 17, at: 3 },
  { arrow: [198, 60, 274, 136], bend: -14, color: 'dim', at: 3 },
  { box: [0, 104, 192, 50], text: &quot;a caller's deadline&quot;, size: 17, at: 4 },
  { arrow: [198, 130, 270, 152], bend: -6, color: 'dim', at: 4 },
  { box: [0, 174, 192, 50], text: 'the server shuts down', size: 16, at: 5 },
  { arrow: [198, 198, 270, 170], bend: 6, color: 'dim', at: 5 },
  { box: [0, 244, 192, 50], text: 'code calls cancel()', size: 17, at: 6 },
  { arrow: [198, 268, 276, 186], bend: 14, color: 'dim', at: 6 },
  { arrow: [335, 206, 335, 270], color: 'dont', at: 2 },
  { text: 'context canceled', x: 335, y: 294, size: 19, mono: true, color: 'dont', at: 2 },
  { text: 'no reason given', x: 262, y: 356, size: 24, anchor: 'start', color: 'accent', at: 7 },
  { arrow: [250, 352, 200, 306], bend: 10, color: 'accent', at: 7 },
]" />

<p v-click="7" class="muted">The cancel often comes from another service or from a library.</p>

<!--
[click] I see logs like this in most Go services that call other services.

[click] Any of these can stop the context. [click] The client went away. [click] A caller's deadline passed. [click] The server started to shut down. This stops a request only if you connect the signal context to the server. BaseContext does this. [click] Or some code called cancel.

[click] The log doesn't say which one. The cancel often started in a library or a service I've never read.
-->
---
layout: section
index: "02"
kicker: Part 2
title: Context primer
---
---
layout: default
kicker: Context primer
title: The work continues after the client leaves
---

<Bars
  :total="1600"
  :click="false"
  :marker="{ at: 500, label: 'client hangs up' }"
  after="waste"
  after-label="work for a client that hung up"
  :ticks="[{ at: 0, label: '0' }, { at: 500, label: '500ms' }, { at: 1000, label: '1s' }, { at: 1500, label: '1.5s' }]"
  :rows="[
    { label: 'HTTP handler', from: 0, to: 1550 },
    { label: 'db query', from: 60, to: 720 },
    { label: 'inventory RPC', from: 740, to: 1220 },
    { label: 'payment RPC', from: 1240, to: 1550 },
  ]"
  :notes="[
    { row: 1, at: 640, text: 'holds a db connection' },
    { row: 2, at: 1150, text: 'work that nobody uses' },
  ]"
/>

<!--
This is one request. The handler queries the database, then calls inventory, then payments.

At 500 milliseconds the client gives up. Maybe the user closed the app. The server cancels r.Context, but no code here checks it. So everything after the line is wasted.

The query holds its connection. Two services do work that nobody reads. During an incident, clients retry, and this waste adds to the load.

Go can't stop a goroutine from the outside. The work has to be told to stop, and then stop itself. That's what context is for.
-->
---
layout: default
kicker: Context primer
clicks: 8
title: context.Context
---

<div v-click="1" class="pull pull-top">
Package context defines the Context type, which carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
<cite>context package doc</cite>
</div>

<div v-click="2">
<Code gutter="19rem">

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

<Mark :line="2" v-click="3">when to give up</Mark>
<Mark :line="3" v-click="4">closes when the work should stop</Mark>
<Mark :line="4" v-click="5">Canceled or DeadlineExceeded</Mark>
<Mark :line="5" v-click="6">data about the request</Mark>

</Code>
</div>

<Sketch :w="1000" :h="120" style="margin-top:1.2rem" :items="[
  { box: [0, 30, 150, 52], text: 'parent', color: 'accent', size: 20, at: 7 },
  { box: [250, 30, 150, 52], text: 'child', size: 20, at: 7 },
  { arrow: [154, 56, 246, 56], at: 7 },
  { text: 'holds cancel()', x: 75, y: 104, size: 18, color: 'dim', mono: true, at: 7 },
  { arrow: [325, 26, 80, 24], bend: 34, color: 'dont', dashed: true, at: 7 },
  { cross: [205, 2, 10], at: 7 },
  { text: &quot;can't cancel\nits parent&quot;, x: 420, y: 56, size: 19, anchor: 'start', color: 'dont', at: 7 },
  { box: [640, 30, 110, 52], text: 'ctx', size: 20, at: 8 },
  { arrow: [754, 56, 846, 56], color: 'accent', at: 8 },
  { text: 'WithValue', x: 800, y: 34, size: 17, mono: true, color: 'accent', at: 8 },
  { box: [850, 30, 150, 52], text: 'new child', color: 'accent', size: 20, at: 8 },
  { text: 'a change makes a new child', x: 820, y: 104, size: 17, color: 'dim', at: 8 },
]" />

<!--
[click] The package doc says what a context is for. Note "between processes". We'll get there.

[click] This is the whole interface. [click] Deadline says when to give up. [click] Done is a channel. It closes when the work should stop. [click] Err says how the context ended. It has only two answers. The next slide shows where they come from. [click] Value carries request data, like a trace ID.

[click] There's no Cancel method. Only the code that made the context can cancel it, so a child can't cancel its parent. [click] A context is read-only. You change it by wrapping it in a child.
-->
---
layout: default
kicker: Context primer
clicks: 6
title: A context ends in one of two ways
---

<Code gutter="16rem">

<div v-click>

```go
ctx, cancel := context.WithCancel(parent)
cancel()
ctx.Err() // context.Canceled
```

</div>

<Mark :line="3" v-click>someone called cancel</Mark>

<div v-click>

```go

ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()
<-ctx.Done()
ctx.Err() // context.DeadlineExceeded
```

</div>

<Mark :block="2" :line="5" v-click>the deadline passed</Mark>

</Code>

<Sketch :w="1000" :h="140" style="margin-top:1.2rem" :items="[
  { box: [0, 4, 210, 38], text: 'the client hangs up', size: 17, at: 5 },
  { box: [0, 50, 210, 38], text: 'SIGTERM', size: 17, at: 5 },
  { box: [0, 96, 210, 38], text: 'code calls cancel()', size: 17, at: 5 },
  { arrow: [214, 23, 290, 62], bend: -8, at: 5 },
  { arrow: [214, 69, 290, 69], at: 5 },
  { arrow: [214, 115, 290, 76], bend: 8, at: 5 },
  { text: 'context.Canceled', x: 300, y: 69, size: 20, mono: true, anchor: 'start', color: 'accent', at: 5 },
  { box: [560, 24, 210, 40], text: 'the timer runs out', size: 17, at: 6 },
  { box: [560, 76, 210, 40], text: &quot;a parent's timer runs out&quot;, size: 16, at: 6 },
  { arrow: [774, 44, 812, 64], bend: -6, at: 6 },
  { arrow: [774, 96, 812, 76], bend: 6, at: 6 },
  { text: 'DeadlineExceeded', x: 822, y: 70, size: 20, mono: true, anchor: 'start', color: 'accent', at: 6 },
]" />

<!--
[click] A context ends in one of two ways. One is a call to cancel. [click] Then Err returns context.Canceled.

[click] The other is a deadline. [click] When it passes, Err returns context.DeadlineExceeded.

These are the two lines from the start of the talk. [click] Canceled means some code called cancel. The HTTP server does it when the client disconnects. signal.NotifyContext does it on SIGTERM. Since Go 1.26, it also sets a cause that names the signal. Or your own code does it. [click] DeadlineExceeded means a timer ran out.

A child gets its parent's error. So the error doesn't say which context ended, or why.
-->
---
layout: default
kicker: Context primer
clicks: 7
title: Each With function makes a child, so contexts form a tree
---

<div class="tree-grid">
<div class="apis tight">
<div class="api" v-click="1"><code>Background()</code><span>The empty root context</span></div>
<div class="api" v-click="2"><code>WithCancel(parent)</code><span>Stop when the code calls cancel</span></div>
<div class="api" v-click="3"><code>WithTimeout(parent, d)</code><span>Stop after a duration. <code>WithDeadline</code> stops at a time</span></div>
<div class="api" v-click="4"><code>WithValue(parent, k, v)</code><span>Add one value</span></div>
<div class="api" v-click="5"><code>WithoutCancel(parent)</code><span>Keep the values and drop the stop signal (Go 1.21)</span></div>
</div>
<CtxTree />
<div class="stack">
<p class="muted" v-click="[6, 7]">The 2s timeout stops only the branch below it.</p>
<p class="muted" v-click="7">When the client hangs up, everything below <code>r.Context()</code> stops. The WithoutCancel branch keeps going.</p>
</div>
</div>

<!--
[click] Each child has one parent, so the contexts in a request form a tree. Background is the root.

[click] WithCancel returns a child and a cancel function. The net/http server calls it for each request. That child is r.Context(). The server cancels it when the client hangs up.

[click] WithTimeout does the same with a timer. Here the handler gives the query and the RPC two seconds.

[click] WithValue adds one value. Here it's a trace ID for the logger.

[click] WithoutCancel keeps the values and drops the cancel. It's from Go 1.21, and Part 3 uses it.

[click] When the timer fires, the query and the RPC stop. The rest keeps going. A cancel moves down the tree, never up.

[click] When the client hangs up, the whole request stops. Only the WithoutCancel branch keeps going, and it still has the trace ID.
-->
---
layout: diagram
kicker: Context primer
title: A context travels inside a process and between processes
note: "<span style='color:var(--accent)'>In-process</span>: your code passes <code>ctx</code> from function to function. <span style='color:var(--cat-2)'>Distributed</span>: gRPC and OpenTelemetry copy its deadline and trace into headers. They can copy only what your code passed down."
---

<Sketch :w="1000" :h="240" :items="[
  { box: [0, 10, 440, 200], color: 'line', dashed: true, nobg: true },
  { text: 'order service', x: 18, y: 36, size: 18, anchor: 'start', color: 'dim' },
  { box: [14, 80, 126, 76], text: 'HTTP\nhandler', size: 21 },
  { box: [160, 80, 126, 76], text: 'placeOrder', size: 21 },
  { box: [306, 80, 120, 76], text: 'payments\nclient', size: 21 },
  { arrow: [142, 118, 156, 118], color: 'accent' },
  { arrow: [288, 118, 302, 118], color: 'accent' },
  { text: 'ctx', x: 150, y: 178, size: 19, color: 'accent' },
  { text: 'ctx', x: 295, y: 178, size: 19, color: 'accent' },
  { arrow: [430, 118, 568, 118], color: 'info', sw: 2.4 },
  { text: 'gRPC call', x: 500, y: 86, size: 21, color: 'info' },
  { text: 'traceparent\ngrpc-timeout', x: 500, y: 160, size: 18, mono: true, color: 'info' },
  { box: [560, 10, 440, 200], color: 'line', dashed: true, nobg: true },
  { text: 'payment service', x: 578, y: 36, size: 18, anchor: 'start', color: 'dim' },
  { box: [574, 80, 126, 76], text: 'gRPC\nhandler', size: 21 },
  { box: [720, 80, 126, 76], text: 'charge', size: 21 },
  { box: [866, 80, 120, 76], text: 'PSP\nclient', size: 21 },
  { arrow: [702, 118, 716, 118], color: 'accent' },
  { arrow: [848, 118, 862, 118], color: 'accent' },
  { text: 'ctx', x: 710, y: 178, size: 19, color: 'accent' },
  { text: 'ctx', x: 855, y: 178, size: 19, color: 'accent' },
]" />

<!--
"Context propagation" means two things.

Inside a process, your code passes ctx from the handler to the service to the client. Most of this talk is about that part.

Between processes, the gRPC client and OpenTelemetry read the deadline and the trace from ctx. They write them into headers. The server reads the headers and builds a new context.

The second part needs the first. If placeOrder starts from context.Background, the payments client has no deadline and no trace to send.
-->
---
layout: default
kicker: Context primer
clicks: 6
title: A cancel stops only the code that checks the context
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: ignore ctx in the loop.
for _, img := range images {
    resize(img) // 200ms each
}
```

</div>

<Mark :line="3" v-click>keeps going after a cancel</Mark>

<div v-click>

```go

// Do: check ctx before each image.
for _, img := range images {
    if err := ctx.Err(); err != nil {
        return err
    }
    resize(img)
}
```

</div>

<Mark :block="2" :line="4" v-click>stops before the next image</Mark>

</Code>

<Sketch :w="1000" :h="86" style="margin-top:0.4rem" :items="[
  { box: [0, 20, 140, 50], text: 'cancel()', size: 19, mono: true, at: 5 },
  { arrow: [144, 45, 236, 45], at: 5 },
  { text: 'closes', x: 190, y: 27, size: 16, color: 'dim', at: 5 },
  { ellipse: [300, 45, 120, 58], text: 'Done', size: 21, color: 'accent', at: 5 },
  { arrow: [364, 36, 480, 12], bend: -8, color: 'dont', at: 6 },
  { arrow: [364, 54, 480, 76], bend: 8, color: 'do', at: 6 },
  { text: 'a loop that never checks keeps going', x: 492, y: 12, size: 21, anchor: 'start', color: 'dont', at: 6 },
  { text: 'a loop that checks ctx stops', x: 492, y: 76, size: 21, anchor: 'start', color: 'do', at: 6 },
]" />

<!--
[click] The context only tells code to stop. The code has to stop itself. This loop runs inside a function that takes ctx, but it never looks at it. [click] If the client hangs up after the first image, the loop resizes all the others anyway.

[click] The fix is one check per item. [click] Once ctx is canceled, the loop returns before the next image.

[click] A cancel only closes a channel. [click] So every loop and every wait checks ctx, or passes it to code that does. The HTTP client, database/sql, and exec.CommandContext all check it for you.
-->
---
layout: default
kicker: Context primer
clicks: 4
title: Most context bugs I see are in these groups
---

<Sketch :w="1000" :h="330" style="margin-top:0.6rem" :items="[
  { text: '1', x: 8, y: 24, size: 40, anchor: 'start', color: 'accent', at: 1 },
  { box: [8, 104, 90, 30], color: 'accent', fill: true, fs: 'cross-hatch', at: 1 },
  { box: [98, 104, 126, 30], color: 'dont', fill: true, at: 1 },
  { line: [98, 80, 98, 152], color: 'info', dashed: true, at: 1 },
  { text: 'client hangs up', x: 104, y: 70, size: 15, anchor: 'start', color: 'info', at: 1 },
  { text: 'Work that\nruns too long', x: 8, y: 222, size: 23, anchor: 'start', color: 'fg', at: 1 },
  { text: 'The client hangs up, but\nthe query continues.', x: 8, y: 296, size: 17, anchor: 'start', color: 'dim', at: 1 },
  { text: '2', x: 263, y: 24, size: 40, anchor: 'start', color: 'accent', at: 2 },
  { box: [263, 104, 80, 30], color: 'accent', fill: true, fs: 'cross-hatch', at: 2 },
  { box: [343, 104, 136, 30], color: 'dim', dashed: true, nobg: true, at: 2 },
  { line: [343, 80, 343, 152], color: 'info', dashed: true, at: 2 },
  { text: 'handler returns', x: 349, y: 70, size: 15, anchor: 'start', color: 'info', at: 2 },
  { cross: [411, 119, 11], at: 2 },
  { text: 'Work that\nstops too early', x: 263, y: 222, size: 23, anchor: 'start', color: 'fg', at: 2 },
  { text: 'The audit write fails\nwhen the handler returns.', x: 263, y: 296, size: 17, anchor: 'start', color: 'dim', at: 2 },
  { text: '3', x: 518, y: 24, size: 40, anchor: 'start', color: 'accent', at: 3 },
  { box: [518, 60, 170, 40], text: 'user: alice', size: 17, mono: true, tc: 'dim', at: 3 },
  { box: [558, 104, 176, 40], text: 'user: audit-bot', size: 17, mono: true, color: 'dont', at: 3 },
  { text: 'A value that is\nwrong or hidden', x: 518, y: 222, size: 23, anchor: 'start', color: 'fg', at: 3 },
  { text: &quot;Another package's key\nhides the user.&quot;, x: 518, y: 296, size: 17, anchor: 'start', color: 'dim', at: 3 },
  { text: '4', x: 773, y: 24, size: 40, anchor: 'start', color: 'accent', at: 4 },
  { box: [773, 98, 196, 42], text: 'context canceled', size: 17, mono: true, tc: 'dont', at: 4 },
  { text: '?', x: 979, y: 82, size: 48, color: 'accent', at: 4 },
  { text: 'A cancel with\nno cause', x: 773, y: 222, size: 23, anchor: 'start', color: 'fg', at: 4 },
  { text: 'The log says only\ncontext canceled.', x: 773, y: 296, size: 17, anchor: 'start', color: 'dim', at: 4 },
]" />

<!--
These are the bugs I see most. Each part after this fixes one or more of them.

[click] Work runs too long. A layer didn't pass ctx, or never checks it. [click] Work stops too early. Something that has to outlive the request got the request's context. [click] A value that's wrong or hidden. Two packages use the same key, or a required input hides in the context. The compiler can't catch either. [click] A cancel with no cause. That's the log line from the start.
-->
---
layout: section
index: "03"
kicker: Part 3
title: Passing a context down
---
---
layout: default
kicker: Passing a context down
clicks: 6
title: Use the caller's context
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: start a new context here.
return s.store.Find(context.Background(), id)
```

</div>

<Mark :line="2" v-click="1">loses the caller's deadline</Mark>

<div v-click>

```go

// Do: pass the caller's context.
return s.store.Find(ctx, id)
```

</div>

</Code>

<Sketch :w="1000" :h="112" style="margin-top:1.6rem" :items="[
  { box: [0, 8, 232, 96], at: 3 },
  { text: 'in main and init', x: 116, y: 38, size: 19, color: 'fg', at: 3 },
  { text: 'Background()', x: 116, y: 76, size: 19, mono: true, color: 'accent', at: 3 },
  { box: [256, 8, 232, 96], at: 4 },
  { text: 'in tests (Go 1.24)', x: 372, y: 38, size: 19, color: 'fg', at: 4 },
  { text: 't.Context()', x: 372, y: 76, size: 19, mono: true, color: 'accent', at: 4 },
  { box: [512, 8, 232, 96], at: 5 },
  { text: 'no ctx passed in yet', x: 628, y: 38, size: 19, color: 'fg', at: 5 },
  { text: 'TODO()', x: 628, y: 76, size: 19, mono: true, color: 'accent', at: 5 },
  { box: [768, 8, 232, 96], at: 6 },
  { text: 'never pass', x: 884, y: 38, size: 19, color: 'fg', at: 6 },
  { text: 'nil', x: 862, y: 76, size: 19, mono: true, color: 'dont', at: 6 },
  { cross: [912, 76, 12], sw: 2.4, at: 6 },
]" />

<!--
[click] This is the bug I see most. context.Background starts a new tree in the middle of a request. Callers can't stop the query anymore. The trace splits in two. The next service gets no deadline.

[click] The fix is to pass the ctx you got. Name it ctx and make it the first argument.

[click] Background belongs at the root, in main and init. [click] In tests, use t.Context, from Go 1.24. [click] When old code has no context yet, use TODO. You can grep for it later. [click] And never pass nil.
-->
---
layout: default
kicker: Passing a context down
title: Pass ctx to each call, and don't store it in a struct
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: keep the context in a struct.
type Notifier struct {
    ctx    context.Context
    client *sms.Client
}
```

</div>

<Mark :line="3" v-click="1">one deadline for every call</Mark>

<div v-click>

```go

// Do: pass the context to each call.
func (n *Notifier) Send(ctx context.Context, msg string) error {
    return n.client.Send(ctx, msg)
}
```

</div>

</Code>

<div v-click class="pull">
When you store the context in a struct, you obscure lifetime to the callers, or worse intermingle two scopes together in unpredictable ways.
<cite>Go blog, Contexts and structs, 2021</cite>
</div>

<!--
[click] This Notifier got its context from the request that built it. A caller with a 200 millisecond timeout can't give it to Send. When that old request ends, every later Send fails with context canceled.

[click] Take ctx as a parameter on each method. Then each call gets the caller's deadline and cancel.

[click] The Go blog post on structs says the same.
-->
---
layout: default
kicker: Passing a context down
clicks: 5
title: Wait on the timer and ctx.Done() together
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: sleep without a check for cancel.
time.Sleep(backoff(attempt))
```

</div>

<Mark :line="2" v-click="1">a cancel can't stop it</Mark>

<div v-click>

```go

// Do: stop the wait when ctx is done.
select {
case <-time.After(backoff(attempt)):
case <-ctx.Done():
    return ctx.Err()
}
```

</div>

<Mark :block="2" :line="5" v-click>a cancel ends the wait</Mark>
<Mark :block="2" :line="6" v-click>Part 5 returns the cause here</Mark>

</Code>

<Sketch :w="1000" :h="118" style="margin-top:1.2rem" :items="[
  { text: 'a timer', x: 110, y: 10, size: 18, color: 'dim', at: 5 },
  { box: [0, 26, 220, 46], text: '<-time.After(d)', size: 19, mono: true, at: 5 },
  { text: 'a receive', x: 450, y: 10, size: 18, color: 'dim', at: 5 },
  { box: [340, 26, 220, 46], text: 'v := <-ch', size: 19, mono: true, at: 5 },
  { text: 'a send', x: 790, y: 10, size: 18, color: 'dim', at: 5 },
  { box: [680, 26, 220, 46], text: 'ch <- v', size: 19, mono: true, at: 5 },
  { text: '+ <-ctx.Done()', x: 110, y: 100, size: 19, mono: true, color: 'accent', at: 5 },
  { text: '+ <-ctx.Done()', x: 450, y: 100, size: 19, mono: true, color: 'accent', at: 5 },
  { text: '+ <-ctx.Done()', x: 790, y: 100, size: 19, mono: true, color: 'accent', at: 5 },
]" />

<!--
[click] This retry loop passes ctx to the call, but still ignores it. time.Sleep doesn't take a context. If the request is canceled during the backoff, the loop sleeps anyway.

[click] A select waits on the timer and on Done. [click] A cancel ends the wait.

[click] Returning ctx.Err is the usual rule, and it's correct. Part 5 swaps it for context.Cause. The errors.Is result stays the same, and the error also gets the reason.

[click] Use a select like this for any wait. A send, a receive, a timer.
-->
---
layout: default
kicker: Passing a context down
clicks: 5
title: Give background work its own context and timeout
---

<Code gutter="17rem">

<div v-click>

```go
// Don't: use the request context for background work.
go h.audit.Write(r.Context(), orderID)
```

</div>

<Mark :line="2" v-click="1">canceled when the handler returns</Mark>

<div v-click>

```go

// Do: keep the values and set a new timeout.
detached := context.WithoutCancel(r.Context())
go func() {
    ctx, cancel := context.WithTimeout(detached, 5*time.Second)
    defer cancel()
    h.audit.Write(ctx, orderID)
}()
```

</div>

<Mark :block="2" :line="3" v-click>keeps the values (Go 1.21)</Mark>
<Mark :block="2" :line="5" v-click>its own timeout</Mark>

</Code>

<Sketch :w="1000" :h="124" style="margin-top:1.2rem" :items="[
  { box: [0, 4, 470, 116], color: 'dont', at: 5 },
  { text: 'Background()', x: 22, y: 30, size: 19, mono: true, anchor: 'start', color: 'dont', at: 5 },
  { check: [36, 66, 10], at: 5 },
  { text: 'drops the cancel', x: 60, y: 66, size: 19, anchor: 'start', color: 'fg', at: 5 },
  { cross: [36, 98, 9], at: 5 },
  { text: 'drops the trace ID and log fields', x: 60, y: 98, size: 19, anchor: 'start', color: 'fg', at: 5 },
  { box: [530, 4, 470, 116], color: 'do', at: 5 },
  { text: 'WithoutCancel(ctx)', x: 552, y: 30, size: 19, mono: true, anchor: 'start', color: 'do', at: 5 },
  { check: [566, 66, 10], at: 5 },
  { text: 'keeps the trace ID and log fields', x: 590, y: 66, size: 19, anchor: 'start', color: 'fg', at: 5 },
  { check: [566, 98, 10], at: 5 },
  { text: 'drops the cancel and the deadline', x: 590, y: 98, size: 19, anchor: 'start', color: 'fg', at: 5 },
]" />

<!--
[click] The handler starts an audit write in a goroutine and returns. The server cancels r.Context when the handler returns. So the write fails after the user got a 201. Tests often miss it, because the fake store is fast.

[click] [click] WithoutCancel keeps the values and drops the cancel and the deadline. [click] So the work has no time limit. Give it its own timeout.

[click] Background drops the cancel. It also drops the trace ID and the log fields. In production I send this work to a pool, so shutdown can wait for it. My post on supervised fire-and-forget has the details.
-->
---
layout: default
kicker: Passing a context down
title: A child context can't extend its parent's deadline
---

<Code gutter="16rem">

<div v-click>

```go
parent, cancelParent := context.WithTimeout(ctx, 2*time.Second)
defer cancelParent()

child, cancelChild := context.WithTimeout(parent, 10*time.Second)
defer cancelChild()

deadline, _ := child.Deadline()
```

</div>

<Mark :line="1" v-click>the parent gets 2s</Mark>
<Mark :line="4" v-click>the child asks for 10s</Mark>
<Mark :line="7" v-click>2s from now</Mark>

</Code>

<div v-click>

<Bars
  :total="10000"
  :click="false"
  :marker="{ at: 2000, label: 'parent deadline' }"
  after="clip"
  after-label="8s that the child never gets"
  :ticks="[{ at: 0, label: '0' }, { at: 2000, label: '2s' }, { at: 5000, label: '5s' }, { at: 10000, label: '10s' }]"
  :rows="[
    { label: 'parent (2s)', from: 0, to: 2000 },
    { label: 'child (10s)', from: 0, to: 10000 },
  ]"
/>

</div>

<p v-click class="take">Set the deadline once, in the handler. Give a lower layer its own timeout only when it needs less time.</p>

<!--
[click] Here's a parent and a child. [click] The parent gets two seconds. [click] The child asks for ten.

[click] The child still ends two seconds from now. A child's deadline is the earlier of its own and its parent's. So a deadline only gets shorter down the tree.

[click] The dashed part is time the child asked for and never gets.

[click] So set the deadline once, where the request starts. Give an inner call a shorter timeout when it needs one, like a cache lookup that should give up fast.
-->
---
layout: section
index: "04"
kicker: Part 4
title: Context values
---
---
layout: default
kicker: Context values
clicks: 3
title: Put a value in the context only if the code works without it
---

<Sketch :w="1000" :h="262" :items="[
  { diamond: [500, 38, 380, 74], text: 'can the code\nwork without it?', size: 19, color: 'accent', at: 1 },
  { arrow: [318, 50, 220, 92], bend: 10, color: 'dont', at: 1 },
  { text: 'no', x: 292, y: 86, size: 20, color: 'dont', at: 1 },
  { box: [0, 96, 460, 162], color: 'dont', at: 1 },
  { text: 'make it a parameter', x: 230, y: 124, size: 21, color: 'dont', at: 1 },
  { text: 'User data the function works on\nDatabase handles and service clients\nConfiguration\nBusiness logic inputs', x: 230, y: 198, size: 18, color: 'fg', at: 1 },
  { arrow: [682, 50, 780, 92], bend: -10, color: 'do', at: 2 },
  { text: 'yes', x: 706, y: 86, size: 20, color: 'do', at: 2 },
  { box: [540, 96, 460, 162], color: 'do', at: 2 },
  { text: 'put it in the context', x: 770, y: 124, size: 21, color: 'do', at: 2 },
  { text: 'Trace, request, and correlation IDs\nAuth tokens from middleware\nRequest-scoped log fields\nIdempotency keys', x: 770, y: 198, size: 18, color: 'fg', at: 2 },
]" />

<div style="display:grid; grid-template-columns:minmax(0,1fr) auto; column-gap:2rem; align-items:end">
<div v-click="3" class="pull">
Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions.
<cite>context package doc</cite>
<p class="muted" style="margin-top:0.5rem; font-style:normal">An optional value also has to be request data.</p>
</div>
<img src="/img/gopher-umbrella.svg" class="gopher" style="max-height:7rem; margin:0" />
</div>

<!--
Values move down the same tree. They don't show up in any function signature. So when one is missing, the compiler can't tell you.

[click] My test: if the function can't work without it, it goes in the signature. A user ID that a query needs fails the test. So do clients and config.

[click] A trace ID passes, because nothing breaks without it. Auth is the tricky case. http.Handler has a fixed signature, so middleware puts the user in the context. The handler reads it once and passes it on as an argument. Every value in a context should be optional. But not every optional value belongs there.

[click] The package doc adds a second rule. The value has to be request data. A function option is optional, but it isn't request data. So it stays a parameter.
-->
---
layout: default
kicker: Context values
title: Hide an unexported key behind a setter and a getter
---

<Code gutter="14rem">

<div v-click>

```go
// Don't: use a string as the key.
ctx = context.WithValue(ctx, "user", "alice")     // package auth
ctx = context.WithValue(ctx, "user", "audit-bot") // package audit
ctx.Value("user")                                 // "audit-bot"
```

</div>

<Mark :line="4" v-click="1">audit-bot hides alice</Mark>

<div v-click>

```go

// Do: keep an unexported key behind a setter and a getter.
package auth

type ctxKey struct{}

func WithUser(ctx context.Context, u User) context.Context {
    return context.WithValue(ctx, ctxKey{}, u)
}

func UserFromContext(ctx context.Context) (User, bool) {
    u, ok := ctx.Value(ctxKey{}).(User)
    return u, ok
}
```

</div>

<Mark :block="2" :line="5" v-click>only auth can make one</Mark>
<Mark :block="2" :line="12" v-click>a typed value and a bool</Mark>

</Code>

<!--
[click] Any package can add a key. Here two packages pick the same string. Value returns the first match on the way up, so the second WithValue hides the first. staticcheck flags this with SA1029.

[click] An unexported key type fixes it. [click] Only this package can write ctxKey, so no other package can make an equal key.

[click] Export a setter and a getter. Callers get a typed User and a bool, and never see the key. Middleware calls WithUser. The handler calls UserFromContext once and passes the user on. httptrace, pprof labels, and OpenTelemetry work the same way.
-->
---
layout: default
kicker: Context values
title: Don't hide a dependency in a context
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: find the transaction in the context.
func (s *OrderStore) Create(ctx context.Context, o Order) (int64, error) {
    var db DBTX = s.db
    if tx, ok := TxFromContext(ctx); ok {
        db = tx
    }
    // ...
}
```

</div>

<Mark :line="4" v-click>false if a caller forgets WithTx</Mark>
<Mark :line="3" v-click>then the store writes outside the transaction</Mark>

<div v-click>

```go

// Do: pass the transaction as an argument.
func (s *OrderStore) Create(ctx context.Context, tx DBTX, o Order) (int64, error) {
    // ...
}
```

</div>

<Mark :block="2" :line="3" v-click>callers pass it</Mark>

</Code>

<!--
[click] A typed getter stops key collisions. It can't stop a required value from hiding in a context. A reader suggested this on my repository post: put the transaction in the context, and let the store find it.

[click] If a caller forgets WithTx, ok is false. [click] Then the store falls back to the pool. Nothing fails, and the writes are no longer atomic.

[click] Pass the transaction as an argument. [click] Now it's in the signature, and the compiler checks every caller.
-->
---
layout: section
index: "05"
kicker: Part 5
title: Giving every cancel a cause
---
---
layout: default
kicker: Giving every cancel a cause
clicks: 7
title: Cancel with a cause that says why
---

<Code gutter="9rem">

<div v-click>

```go
// Do: give each cancel a cause.
func placeOrder(ctx context.Context, id string) error {
    ctx, cancel := context.WithCancelCause(ctx)
    defer cancel(fmt.Errorf("placeOrder: completed: %w", context.Canceled))

    if err := checkInventory(ctx, id); err != nil {
        cancel(fmt.Errorf("inventory %s: %w: %w", id, err, context.Canceled))
        return err
    }
    // ...
}
```

</div>

<Mark :line="3" v-click="1">Go 1.20</Mark>
<Mark :line="4" v-click>normal return</Mark>
<Mark :line="7" v-click>error path</Mark>

</Code>

<Sketch :w="1000" :h="124" style="margin-top:1rem" :items="[
  { text: 'context.Cause(ctx)', x: 0, y: 12, size: 19, mono: true, anchor: 'start', color: 'dim', at: 4 },
  { arrow: [196, 12, 250, 12], color: 'accent', at: 4 },
  { text: 'inventory ord-123: connection refused: context canceled', x: 262, y: 12, size: 19, mono: true, anchor: 'start', color: 'accent', at: 4 },
  { box: [0, 70, 150, 42], text: 'cancel(errA)', size: 17, mono: true, at: 5 },
  { check: [168, 91, 9], at: 5 },
  { box: [188, 70, 150, 42], text: 'cancel(errB)', size: 17, mono: true, tc: 'dim', at: 5 },
  { cross: [356, 91, 8], at: 5 },
  { text: 'the first cancel sets it', x: 378, y: 91, size: 18, anchor: 'start', color: 'dim', at: 5 },
  { box: [580, 70, 170, 42], text: 'Err() = Canceled', size: 17, mono: true, at: 6 },
  { box: [776, 70, 64, 42], text: 'ctx', size: 18, color: 'accent', at: 7 },
  { arrow: [844, 91, 884, 91], at: 7 },
  { text: 'children\nread errA', x: 892, y: 91, size: 17, anchor: 'start', color: 'accent', at: 7 },
]" />

<!--
[click] ctx.Err has only two answers. Wrapping the error tells you where the code was, not why the context ended. The code that cancels knows why. Since Go 1.20, it can say so. WithCancelCause gives you a cancel function that takes an error.

[click] The deferred call sets the cause for the normal return. [click] On the error path, I cancel with a cause that wraps the error. Both causes also wrap context.Canceled. The next slide shows why.

[click] context.Cause reads the cause back. [click] Only the first cancel counts. So the error cause stays, and the deferred call does nothing. [click] ctx.Err doesn't change. [click] The cause moves down the tree with the cancel. Any goroutine still working under this ctx can read it.

A colleague at DoorDash once chased a context canceled after a successful request. People looked for a client disconnect. A "completed" cause would have said the request had finished.
-->
---
layout: default
kicker: Giving every cancel a cause
title: Wrap DeadlineExceeded in the cause, so errors.Is still matches
---

<Code gutter="17rem">

<div v-click>

```go
// Don't: use a cause that doesn't wrap DeadlineExceeded.
ctx, cancel := context.WithTimeoutCause(ctx, 20*time.Millisecond,
    errors.New("inventory: 20ms timeout"))
defer cancel()

_, err := http.DefaultClient.Do(req.WithContext(ctx))
errors.Is(err, context.DeadlineExceeded)
```

</div>

<Mark :line="3" v-click>errors.New doesn't wrap DeadlineExceeded</Mark>
<Mark :line="6" v-click>since Go 1.23, err is the cause</Mark>
<Mark :line="7" v-click>false</Mark>

<div v-click>

```go

// Do: wrap DeadlineExceeded with %w.
ctx, cancel := context.WithTimeoutCause(ctx, 20*time.Millisecond,
    fmt.Errorf("inventory: 20ms timeout: %w", context.DeadlineExceeded))
```

</div>

<Mark :block="2" :line="4" v-click>errors.Is now returns true</Mark>

</Code>

<!--
[click] Timeouts get causes too. Go 1.21 added WithTimeoutCause. You give the cause up front, and the context uses it when the timer fires. Here the cause is a plain errors.New. [click] So it doesn't wrap DeadlineExceeded.

[click] Since Go 1.23, the HTTP client returns the cause instead of DeadlineExceeded. [click] So this check is false. Retry policies and alerts that check DeadlineExceeded stop working.

[click] Wrap DeadlineExceeded with %w. [click] Now errors.Is matches, and the message still has the cause. Same for Canceled. That's why the causes on the last slide wrap it.

This fixes the timer path. The next slide fixes the other paths. The Go 1.23 release notes don't say this. The change is in the net/http transport code.
-->
---
layout: default
kicker: Giving every cancel a cause
title: Stack a cancel cause on the timeout
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: stop at WithTimeoutCause.
ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second,
    fmt.Errorf("placeOrder %s: 5s timeout: %w", id, context.DeadlineExceeded))
defer cancel()
```

</div>

<Mark :line="4" v-click="1">cancel() sets no cause</Mark>

<div v-click>

```go

// Do: stack a cancel cause on the timeout.
ctx, cancelCause := context.WithCancelCause(ctx)
ctx, cancelTimeout := context.WithTimeoutCause(ctx, 5*time.Second,
    fmt.Errorf("placeOrder %s: 5s timeout: %w", id, context.DeadlineExceeded))
defer cancelTimeout()
defer cancelCause(fmt.Errorf("placeOrder: completed: %w", context.Canceled))
```

</div>

<Mark :block="2" :line="3" v-click>a cause on every other path</Mark>
<Mark :block="2" :line="4" v-click>the timer and the deadline</Mark>
<Mark :block="2" :line="7" v-click>this defer runs first</Mark>

</Code>

<!--
[click] There's a catch. WithTimeoutCause uses its cause only when the timer fires. Its cancel function takes no cause. So the normal return gives a plain context canceled, and most requests return normally. Sameer Ajmani's proposal, issue 56661, says this. It also gives the fix: stack two contexts.

[click] Now every path gets a cause.

[click] The outer context takes a cause on the normal return and on each error path. [click] The inner context owns the timer. On timeout, Err still returns DeadlineExceeded, and the deadline still goes to the next service.

[click] Defers run in reverse. cancelCause runs first and sets the cause. Swap the two defer lines, and you lose it.

If the caller's deadline is earlier than 5 seconds, your timer doesn't fire. The caller's deadline ends the context first. Then you get the caller's cause, or DeadlineExceeded with no cause. The last slide shows this again.

My post also shows a time.AfterFunc version. It's simpler, but the next service gets no deadline. This is easy to get wrong at every call site. Part 6 puts it in a package.
-->
---
layout: default
kicker: Giving every cancel a cause
clicks: 5
title: A cause moves down, so the error has to carry it up
---

<Sketch :w="1000" :h="400" height="23rem" :items="[
  { box: [150, 20, 270, 66], text: 'HTTP handler\nlogs the error', size: 21, at: 1 },
  { box: [150, 166, 270, 66], text: 'placeOrder\nsets the cause', size: 21, color: 'accent', at: 1 },
  { box: [150, 312, 270, 66], text: 'retry wait\n<-ctx.Done()', size: 21, at: 1 },
  { arrow: [285, 90, 285, 162], color: 'dim', at: 1 },
  { text: 'ctx', x: 300, y: 126, size: 17, color: 'dim', anchor: 'start', at: 1 },
  { arrow: [285, 236, 285, 308], color: 'dim', at: 1 },
  { text: 'ctx', x: 300, y: 272, size: 17, color: 'dim', anchor: 'start', at: 1 },
  { arrow: [140, 214, 140, 340], bend: -36, color: 'accent', sw: 2.6, at: 2 },
  { text: 'the cause\nmoves down', x: 66, y: 278, size: 20, color: 'accent', at: 2 },
  { arrow: [140, 184, 140, 70], bend: 36, color: 'dont', dashed: true, at: 3 },
  { cross: [114, 128, 12], color: 'dont', sw: 3, at: 3 },
  { text: 'not up', x: 48, y: 128, size: 20, color: 'dont', at: 3 },
  { arrow: [430, 344, 430, 58], bend: -60, color: 'do', sw: 2.6, at: 4 },
  { text: 'return context.Cause(ctx)', x: 500, y: 190, size: 19, mono: true, color: 'do', anchor: 'start', at: 4 },
  { text: 'the error moves up with the cause', x: 500, y: 226, size: 20, color: 'do', anchor: 'start', at: 4 },
  { arrow: [424, 46, 494, 46], color: 'do', at: 5 },
  { box: [498, 16, 490, 62], color: 'do', at: 5 },
  { text: 'level=ERROR msg=&quot;place order&quot;\nerr=&quot;placeOrder ord-123: 5s timeout: ...&quot;', x: 512, y: 47, size: 17, mono: true, anchor: 'start', at: 5 },
]" />

<!--
A cause and an error move in opposite directions.

[click] This is the call chain. The handler calls placeOrder. placeOrder waits in the retry loop. Each one passes ctx down.

[click] placeOrder sets the cause. A cause moves down to child contexts, so the retry wait can read it.

[click] A cause doesn't move up. The handler's ctx is the parent. When placeOrder's timer fires, the handler's ctx isn't canceled, and it has no cause.

[click] Errors move up. So the returned error is the only way for the cause to reach the handler.

[click] The handler logs the error once, with the cause. The next slide shows how to return the cause. The slide after it shows the log.
-->
---
layout: default
kicker: Giving every cancel a cause
clicks: 3
title: Return context.Cause(ctx) where you return ctx.Err()
---

<Code gutter="15rem">

<div v-click>

```go
// Do: return the cause where you return ctx.Err().
select {
case <-time.After(backoff(attempt)):
case <-ctx.Done():
    return context.Cause(ctx)
}
```

</div>

<Mark :line="5" v-click>same errors.Is result, plus the reason</Mark>

</Code>

<Sketch :w="1000" :h="236" style="margin-top:1rem" :items="[
  { text: 'ctx.Err()', x: 300, y: 14, size: 19, mono: true, anchor: 'start', color: 'dim', at: 3 },
  { text: 'context.Cause(ctx)', x: 590, y: 14, size: 19, mono: true, anchor: 'start', color: 'accent', at: 3 },
  { line: [0, 34, 1000, 34], color: 'line', at: 3 },
  { text: 'errors.Is(err,\nDeadlineExceeded)', x: 0, y: 66, size: 17, mono: true, anchor: 'start', color: 'fg', at: 3 },
  { check: [312, 66, 10], at: 3 },
  { text: 'true', x: 334, y: 66, size: 18, mono: true, anchor: 'start', color: 'do', at: 3 },
  { check: [602, 66, 10], at: 3 },
  { text: 'true', x: 624, y: 66, size: 18, mono: true, anchor: 'start', color: 'do', at: 3 },
  { line: [0, 100, 1000, 100], color: 'line', at: 3 },
  { text: 'the message', x: 0, y: 134, size: 20, anchor: 'start', color: 'fg', at: 3 },
  { text: 'context deadline\nexceeded', x: 300, y: 134, size: 17, mono: true, anchor: 'start', color: 'dim', at: 3 },
  { text: 'placeOrder ord-123: 5s timeout:\ncontext deadline exceeded', x: 590, y: 134, size: 17, mono: true, anchor: 'start', color: 'accent', at: 3 },
  { line: [0, 168, 1000, 168], color: 'line', at: 3 },
  { text: 'no cause set', x: 0, y: 202, size: 20, anchor: 'start', color: 'fg', at: 3 },
  { text: 'context deadline\nexceeded', x: 300, y: 202, size: 17, mono: true, anchor: 'start', color: 'dim', at: 3 },
  { text: 'context deadline exceeded', x: 590, y: 202, size: 17, mono: true, anchor: 'start', color: 'dim', at: 3 },
  { text: 'same output', x: 990, y: 202, size: 19, anchor: 'end', color: 'do', at: 3 },
]" />

<!--
Returning ctx.Err is the usual rule, and it's fine. This is a drop-in swap.

[click] This is the retry wait from Part 3. Where it returned ctx.Err, return context.Cause.

[click] Callers that check errors.Is see the same answer, because the cause wraps DeadlineExceeded. That's why we wrap. Logs get the reason.

[click] If nobody set a cause, you get exactly ctx.Err. So the swap never loses anything. Add context with fmt.Errorf and %w as you normally would.

The style guide says to document a function that returns an error other than ctx.Err. One line in the doc comment is enough. The same goes for a library call that fails after ctx ends. database/sql returns ctx.Err, so return the cause instead. net/http already does this since Go 1.23.
-->
---
layout: default
class: log-once
kicker: Giving every cancel a cause
clicks: 4
title: Log the error once, in the handler
---

<Code gutter="12rem">

<div v-click>

```go
// Don't: log the error in placeOrder, and also return it.
if err := retry(ctx, charge); err != nil {
    slog.ErrorContext(ctx, "retry failed", "err", err)
    return err
}
```

</div>

<Mark :line="3" v-click="1">the handler logs it again</Mark>

</Code>

<div v-click class="log-bad">

<Terminal :lines="[
  'level=ERROR msg=&quot;retry failed&quot; err=&quot;placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
  'level=ERROR msg=&quot;place order&quot;  err=&quot;placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<Code gutter="12rem">

<div v-click>

```go
// Do: log once, in the handler. placeOrder only returns the error.
if err := placeOrder(ctx, id); err != nil {
    slog.ErrorContext(ctx, "place order", "err", err)
}
```

</div>

<Mark :line="3" v-click="3">the only log line</Mark>

</Code>

<div v-click="4" class="log-good">

<Terminal :lines="[
  'level=ERROR msg=&quot;place order&quot; err=&quot;placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<!--
[click] placeOrder logs the error and then returns it. The handler logs the same error again.

[click] So one failure gives two log lines. With more layers, you get more lines. The style guide says to let the caller log. My post on stacked log lines has more.

[click] placeOrder only returns the error. The handler logs it once.

[click] This is the log line from the start of the talk, now with the cause. placeOrder's 5s timeout fired during a retry wait. Dashboards can still group these lines with errors.Is.
-->
---
layout: section
index: "06"
kicker: Part 6
title: Enforcing the rules
---
---
layout: default
kicker: Enforcing the rules
clicks: 6
title: Linters can check most of these rules
---

<Sketch :w="1000" :h="350" style="margin-top:0.6rem" :items="[
  { box: [0, 6, 310, 156], at: 1 },
  { text: 'go vet: lostcancel', x: 155, y: 46, size: 20, mono: true, color: 'accent', at: 1 },
  { text: 'a cancel func that\nis never called', x: 155, y: 110, size: 19, color: 'fg', at: 1 },
  { box: [345, 6, 310, 156], at: 2 },
  { text: 'staticcheck: SA1012', x: 500, y: 46, size: 20, mono: true, color: 'accent', at: 2 },
  { text: 'a nil context passed\nto a function', x: 500, y: 110, size: 19, color: 'fg', at: 2 },
  { box: [690, 6, 310, 156], at: 3 },
  { text: 'staticcheck: SA1029', x: 845, y: 46, size: 20, mono: true, color: 'accent', at: 3 },
  { text: 'a built-in type, like a\nstring, used as a key', x: 845, y: 110, size: 19, color: 'fg', at: 3 },
  { box: [0, 186, 310, 156], at: 4 },
  { text: 'containedctx', x: 155, y: 226, size: 20, mono: true, color: 'accent', at: 4 },
  { text: 'a context stored\nin a struct field', x: 155, y: 290, size: 19, color: 'fg', at: 4 },
  { box: [345, 186, 310, 156], at: 5 },
  { text: 'contextcheck', x: 500, y: 226, size: 20, mono: true, color: 'accent', at: 5 },
  { text: &quot;a context that didn't come\nfrom the caller&quot;, x: 500, y: 290, size: 19, color: 'fg', at: 5 },
  { box: [690, 186, 310, 156], at: 6 },
  { text: 'errorlint', x: 845, y: 226, size: 20, mono: true, color: 'accent', at: 6 },
  { text: 'err == context.Canceled,\nwhich is false for a cause', x: 845, y: 290, size: 19, color: 'fg', at: 6 },
]" />

<!--
Code review catches most of these bugs. Tools catch more.

[click] go vet has lostcancel on by default. go test skips it, so run go vet or golangci-lint in CI. [click] [click] staticcheck flags a nil context and a built-in key type. [click] [click] containedctx and contextcheck are in golangci-lint. contextcheck finds a Background in the middle of a chain. It has some false positives, so start it on new code. [click] errorlint finds == checks on errors. Once causes wrap context.Canceled, you need errors.Is.
-->
---
layout: default
kicker: Enforcing the rules
title: Wrap the constructors so every cancel needs a cause
---

<Code gutter="16rem">

<div v-click>

```go
// Do: make the cancel func take a cause.
package xctx

func WithCancel(parent context.Context) (context.Context, context.CancelCauseFunc) {
    ctx, cancel := context.WithCancelCause(parent) //nolint:forbidigo
    return ctx, func(cause error) {
        if cause == nil {
            cause = errors.New("no cause given")
        }
        if !errors.Is(cause, context.Canceled) {
            cause = fmt.Errorf("%w: %w", cause, context.Canceled)
        }
        cancel(cause)
    }
}
```

</div>

<Mark :line="6" v-click>the cancel func needs a cause</Mark>
<Mark :line="7" v-click>nil still gets a cause</Mark>
<Mark :line="10" v-click>wrap Canceled once, so errors.Is matches</Mark>

</Code>

<p v-click class="muted" style="max-width:none">At DoorDash, Alex Hornbake built this as <code>xctx</code> for a Go monorepo that many teams share.</p>

<!--
[click] Stacking contexts and wrapping Canceled at every call site is easy to get wrong. So put it in a small package. [click] Its cancel function takes a cause, so callers have to give one. [click] A nil cause still gets a message. [click] It wraps context.Canceled, so errors.Is keeps working. If the cause already wraps it, it doesn't wrap it again. Its WithTimeout does the stack from Part 5.

[click] Alex Hornbake did this at DoorDash in early 2026. It still returns a plain context.Context, so every library works.
-->
---
layout: default
kicker: Enforcing the rules
title: Block the plain constructors in CI
---

<Code gutter="16rem">

<div v-click>

```yaml
# Do: block the plain constructors in .golangci.yml.
version: "2"
linters:
  enable: [forbidigo]
  settings:
    forbidigo:
      analyze-types: true
      forbid:
        - pattern: ^context\.With(Cancel|Timeout|Deadline)(Cause)?$
```

</div>

<Mark :line="7" v-click>a renamed import still matches</Mark>
<Mark :line="9" v-click>blocks the Cause versions also</Mark>

</Code>

<div v-click>

<Terminal :lines="[
  {cmd: 'golangci-lint run'},
  'order/place.go:12:16: use of `context.WithTimeout` forbidden by pattern',
  '    `^context\\.With(Cancel|Timeout|Deadline)(Cause)?$` (forbidigo)',
]" />

</div>

<!--
[click] A package helps only if people use it. So a forbidigo rule fails the build on the standard constructors.

[click] analyze-types matches the package path. So a renamed import still matches. [click] The rule also blocks the Cause versions. So only xctx makes the stack. xctx itself is the one nolint.

[click] This is what a pull request sees. After the rollout, every cancel our own code makes has a cause.
-->
---
layout: default
kicker: Wrap-up
title: The checklist I use in code review
---

<div class="side-gopher">
<div class="checklist">

<v-clicks>

1. Pass `ctx` to each call, and don't store it in a struct
2. Check `ctx` in every loop and every blocking wait
3. Give background work `WithoutCancel` and a timeout
4. Put only optional request data in a context, and hide the key
5. Give every cancel a cause, including the normal return
6. Stack `WithCancelCause` on a timeout to cover every path
7. Wrap `Canceled` or `DeadlineExceeded` in every cause
8. Return `context.Cause(ctx)`, and log once in the handler

</v-clicks>

</div>
<img src="/img/gopher-knight.svg" class="gopher" />
</div>

<!--
The checklist follows the order of the talk.

[click] [click] [click] These keep work from running too long or stopping too early. [click] This keeps values from turning into hidden parameters. [click] [click] [click] [click] These make context canceled say why. The linters and xctx check most of the list on every pull request.
-->
---
layout: two-cols
class: wide-left
kicker: Wrap-up
clicks: 4
title: Some cancels still have no cause
---

<Sketch :w="660" :h="440" :items="[
  { box: [210, 178, 240, 84], text: 'context canceled\nwith no cause', size: 24, color: 'dont', tc: 'dont' },
  { ellipse: [140, 58, 300, 120], text: 'the net/http server,\nwhen the client hangs up', size: 23, dashed: true, at: 1 },
  { arrow: [196, 114, 250, 174], bend: -8, color: 'dont', at: 1 },
  { ellipse: [520, 58, 300, 120], text: 'libraries that call\na plain cancel()', size: 23, dashed: true, at: 2 },
  { arrow: [464, 114, 410, 174], bend: 8, color: 'dont', at: 2 },
  { ellipse: [140, 382, 300, 120], text: 'other services:\ngRPC sends no cause', size: 23, dashed: true, at: 3 },
  { arrow: [196, 326, 250, 266], bend: 8, color: 'dont', at: 3 },
  { ellipse: [520, 382, 300, 120], text: &quot;a caller's deadline\nthat fires first&quot;, size: 23, dashed: true, at: 4 },
  { arrow: [464, 326, 410, 266], bend: -8, color: 'dont', at: 4 },
]" />

::right::

<img src="/img/gopher-soldering.svg" class="gopher" style="max-height:10rem; margin-top:6rem" />

<!--
Causes help only where someone sets them.

[click] The HTTP server still cancels the request context without a cause. So a client disconnect gives a plain context canceled. [click] Many libraries don't set causes yet. [click] And a cause never crosses to the next service. gRPC sends only the deadline and the cancel. Each service sets its own causes. [click] And if a caller's deadline is earlier than yours, your timer never fires. You get the caller's cause, or a plain DeadlineExceeded.

When all your own cancels have causes, an error with no cause came from one of these. Start looking there.
-->
---
layout: default
kicker: References
title: Further reading
---

<div class="refs">

#### Go blog

- [Go Concurrency Patterns: Context](https://go.dev/blog/context), Sameer Ajmani, 2014
- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines), Sameer Ajmani, 2014
- [Contexts and structs](https://go.dev/blog/context-and-structs), Jean Barkhuysen and Matt T. Proud, 2021

#### My posts

- [What canceled my Go context?](https://rednafi.com/go/context-cancellation-cause/)
- [Avoiding collisions in Go context keys](https://rednafi.com/go/avoid-context-key-collisions/)
- [What belongs in Go's context values?](https://rednafi.com/shards/2026/03/what-belongs-in-go-context-values/)
- [Is passing user ID through context an antipattern?](https://rednafi.com/shards/2026/03/user-id-through-context/)
- [Stacked log lines considered harmful](https://rednafi.com/shards/2026/04/no-stacked-loglines/)

#### Go issues

- [#26356: context: ease debugging of where a context was canceled?](https://github.com/golang/go/issues/26356)
- [#40221: context: add WithoutCancel](https://github.com/golang/go/issues/40221)
- [#51365: context: add APIs for writing and reading cancelation cause](https://github.com/golang/go/issues/51365)
- [#56661: context: add APIs for setting a cancelation cause when deadline or timer expires](https://github.com/golang/go/issues/56661)

#### Docs and style guides

- [Package context](https://pkg.go.dev/context)
- [Go 1.20 release notes: context](https://go.dev/doc/go1.20#context)
- [Go 1.21 release notes: context](https://go.dev/doc/go1.21#context)
- [Google Go style guide: Contexts](https://google.github.io/styleguide/go/decisions#contexts)
- [Google Go style guide: Custom contexts](https://google.github.io/styleguide/go/decisions#custom-contexts)

#### Linters

- [go vet: lostcancel](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/lostcancel)
- [staticcheck: SA1012, SA1029](https://staticcheck.dev/docs/checks/)
- [golangci-lint: containedctx, contextcheck, errorlint, forbidigo](https://golangci-lint.run/docs/linters/)

</div>

<!--
Start with the Go blog posts. My posts go deeper on causes, values, and logging. The slides are on github.com/rednafi/talks.
-->
---
layout: end
title: Thanks
subtitle: Questions?
contact: rednafi.com
---

<img src="/img/gopher-party.svg" class="gopher sm" style="margin-top:1.5rem" />

<!--
Thanks. I'm happy to take questions.
-->
