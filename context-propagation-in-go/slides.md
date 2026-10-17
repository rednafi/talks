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
This talk is about context.Context: why Go has it, how it moves through a program, how code gets it wrong, and the advice and tools I use to keep it working.
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
Before I start: this code opens a PDF of the slides, so there's no need to take photos.
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
layout: two-cols
class: wide-left
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

::right::

<div v-click class="col-h">Any of these can stop the context</div>

<v-clicks>

- The client hung up
- A caller's deadline passed
- The server started to shut down
- Code called `cancel()`

</v-clicks>

<p v-click class="muted">The cancel often comes from another service or from a library.</p>

<img src="/img/gopher-mistake.svg" class="gopher sm" />

<!--
[click] Errors like these are common in any Go service that talks to other services.

[click] Any of these can stop the context. [click] The client went away, [click] a caller's deadline passed, [click] the server started to shut down, [click] or some code called cancel.

[click] The log line doesn't tell me which one. Most of the time the cancel started in a service or a library I have never opened.
-->

---
layout: quote
quote: "I had a case when on production I got random \"context canceled\" log messages. And in the case like that you don't even know where to dig and how to investigate it further."
author: bullgare, golang/go#51365, 2022
image: /img/gopher-scientist.svg
---

<!--
This comment is from golang/go#51365, the proposal that led to the cause APIs. I have had the same case many times.
-->

---
layout: default
title: Overview
---

<div class="side-gopher">
<div class="overview">
<div class="ov-item" v-click>
  <div class="ov-num">01</div>
  <div class="ov-topic">Context primer</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">02</div>
  <div class="ov-topic">Passing a context down</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">03</div>
  <div class="ov-topic">Recording why a context stopped</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">04</div>
  <div class="ov-topic">Context values</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">05</div>
  <div class="ov-topic">Enforcing the rules</div>
</div>
</div>
<img src="/img/gopher-rocket.svg" class="gopher" />
</div>

<!--
[click] First, what a context is, and the bugs I see around it.

After that, each slide is one rule and the bug it prevents. [click] Some are about passing the context down, [click] some about recording why it stopped, [click] and some about values. [click] The last part shows tools that check the rules for you.
-->

---
layout: section
index: "01"
kicker: Part 1
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
  after-label="Work for a client that has left"
  :ticks="[{ at: 0, label: '0' }, { at: 500, label: '500ms' }, { at: 1000, label: '1s' }, { at: 1500, label: '1.5s' }]"
  :rows="[
    { label: 'HTTP handler', from: 0, to: 1550 },
    { label: 'db query', from: 60, to: 720 },
    { label: 'inventory RPC', from: 740, to: 1220 },
    { label: 'payment RPC', from: 1240, to: 1550 },
  ]"
/>

<div class="points" style="--n: 2">

- The query holds a database connection the whole time
- The inventory and payment services do work that nobody uses

</div>

<!--
This is one request. The handler queries the database, then calls inventory, then calls payments.

At 500 milliseconds the client gives up. The user closed the app, or a load balancer timed out.

Nothing tells the handler, so all the work after the line is wasted.

The query keeps its connection. Two other services do work for a client that has left. During an incident, clients time out and retry, and this extra work adds to the load that caused the timeouts.
-->

---
layout: two-cols
class: wide-left
kicker: Context primer
title: Go can't stop a goroutine from the outside
---

<v-clicks>

- A goroutine has no kill method
- So you send the work a signal, and the work stops itself
- All goroutines for one request need the same signal
- The signal also needs a time limit

</v-clicks>

<div v-click class="pull">
When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.
<cite>Sameer Ajmani, Go blog, 2014</cite>
</div>

::right::

<img src="/img/gopher-sage.svg" class="gopher" />

<!--
The runtime can't stop that work for us.

[click] Go has no goroutine ID that you can kill. [click] A goroutine stops when its function returns. So to stop work, you tell it to stop, and it returns.

[click] A request usually runs in many goroutines: the handler, its worker goroutines, and the database driver. All of them need the same signal. [click] The signal also needs a time limit, so that work stops on its own when time runs out, even if no code calls cancel.

[click] Google wrote the context package for this.
-->

---
layout: default
kicker: Context primer
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

<div class="points" style="--n: 2">
<v-clicks at="7">

- Only the code that creates a context can cancel it
- A context is read-only. To add a deadline or a value, make a child context

</v-clicks>
</div>

<!--
[click] The package doc says what a context is for. Note "between processes". I come back to that in two slides.

[click] This is the whole interface. [click] Deadline tells the work when to give up. [click] Done is a channel that closes when the work should stop. Any number of goroutines can wait on it. [click] Err tells you why it stopped, and it has only two answers: Canceled or DeadlineExceeded. Part 3 is about that limit. [click] Value carries data about the request, such as a trace ID.

[click] The interface has no Cancel method. The function that gets a cancel signal is usually not the one that sends it, so a child can't cancel its parent. [click] A context is also read-only. You change it by wrapping it.
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
<p class="muted" v-click="7">When the client hangs up, all work below the request stops, except the WithoutCancel branch.</p>
</div>
</div>

<!--
[click] Each child has one parent, so the contexts in a request form a tree, with Background at the root. The server gives each request its own context.

[click] WithCancel returns a child and a function that cancels it. The request context works like this: the server cancels it when the client goes away.

[click] WithTimeout and WithDeadline do the same with a timer. Here the handler gives the query and the RPC two seconds.

[click] WithValue adds one value, here a trace ID, and the logger reads it.

[click] WithoutCancel, from Go 1.21, keeps the values and drops the stop signal. Part 2 uses it for work that has to finish.

[click] When the timer fires, the query and the RPC stop. The rest of the tree keeps going. A cancel moves down the tree and never up.

[click] When the client hangs up, the whole request stops. Only the branch under WithoutCancel keeps going, and it still has the trace ID. A value lookup goes up, from the child to the root.
-->
---
layout: diagram
kicker: Context primer
title: A context travels inside a process and between processes
note: "<span style='color:var(--accent)'>In-process</span>: your code passes <code>ctx</code> from function to function. <span style='color:var(--cat-2)'>Distributed</span>: gRPC and OpenTelemetry copy its deadline and trace into headers. They can copy only what your code passed down."
---

<TwoLevels />

<!--
"Context propagation" can mean two things.

The in-process part is the arrows inside each service. Your code passes ctx from the handler to the service to the client. Most of this talk is about this part.

The distributed part is the arrow between the two services. The payments client makes a gRPC call to the payment service. The gRPC client and the OpenTelemetry middleware read the deadline and the trace from the ctx you give them, and write them into headers. The server reads the headers and builds a new context.

The distributed part depends on the in-process part. If placeOrder starts from context.Background, the payments client gets a context with no deadline and no trace, and has nothing to send.
-->

---
layout: default
kicker: Context primer
title: Each layer passes ctx to the next
---

<Code gutter="17rem">

<div v-click>

```go
// Do: pass ctx down through every layer.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    order, err := h.orders.Get(ctx, r.PathValue("id"))
    // ...
}
```

</div>

<Mark :line="3" v-click>the handler sets the time limit</Mark>

<div v-click>

```go

func (s *OrderService) Get(ctx context.Context, id string) (Order, error) {
    return s.store.Find(ctx, id)
}
```

</div>

<Mark :block="2" :line="3" v-click>the service passes ctx to the store</Mark>

<div v-click>

```go

func (s *OrderStore) Find(ctx context.Context, id string) (Order, error) {
    row := s.db.QueryRowContext(ctx, findOrderSQL, id)
    // ...
}
```

</div>

<Mark :block="3" :line="3" v-click>the driver waits on ctx.Done()</Mark>

</Code>

<!--
[click] This is the in-process part, in code.

[click] The handler takes the request context, which the server cancels when the client goes away, and adds a two-second timeout.

[click] The service sets no timeout. [click] It passes ctx to the store. Most layers in a service look like this.

[click] The store gives ctx to database/sql. [click] The driver is the code that waits on Done. If the client hangs up or two seconds pass, the query stops.

If any layer drops ctx, every layer under it loses the signal.
-->

---
layout: default
kicker: Context primer
title: A cancel stops only the code that checks the context
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: ignore ctx in the loop.
func resizeAll(ctx context.Context, images []Image) error {
    for _, img := range images {
        resize(img) // 200ms each
    }
    return nil
}
```

</div>

<Mark :line="4" v-click>keeps going after a cancel</Mark>

<div v-click>

```go

// Do: check ctx before each image.
func resizeAll(ctx context.Context, images []Image) error {
    for _, img := range images {
        if err := ctx.Err(); err != nil {
            return err
        }
        resize(img)
    }
    return nil
}
```

</div>

<Mark :block="2" :line="5" v-click>stops before the next image</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- `cancel()` only closes the `Done` channel
- Make every loop and every wait check `ctx`, or pass it to code that checks it

</v-clicks>
</div>

<!--
[click] Cancellation in Go is cooperative. The context tells the code to stop, and the code has to do the stopping itself.

This function resizes a list of images and takes ctx, but never looks at it. [click] If the client hangs up after the first image, the loop still resizes all the others.

[click] The fix is one check per item. [click] When the context is canceled, the loop returns before the next image.

[click] A cancel only closes a channel. [click] Code that waits has to select on ctx.Done, and code that calls something slow has to pass ctx to it. If no layer checks, nothing stops.
-->

---
layout: default
kicker: Context primer
title: The standard library stops when the context ends
---

<Code gutter="18rem">

<div v-click>

```go
// Do: pass ctx to standard library calls.
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := http.DefaultClient.Do(req)

rows, err := db.QueryContext(ctx, listOrdersSQL, userID)

out, err := exec.CommandContext(ctx, "pg_dump", dbName).Output()
```

</div>

<Mark :line="6" v-click>stops the request</Mark>
<Mark :line="8" v-click>the driver gets ctx and stops the query</Mark>
<Mark :line="10" v-click>kills the process</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- The HTTP server cancels `r.Context()` when the client disconnects
- `signal.NotifyContext` cancels a context when a signal such as `SIGTERM` arrives

</v-clicks>
</div>

<!--
[click] Give the standard library a context and it stops on its own.

[click] The HTTP client stops the request. [click] database/sql passes ctx to the driver, and drivers such as pgx stop the query when ctx is canceled. [click] CommandContext kills the child process.

The standard library also sends the signal. [click] The HTTP server cancels the request context when the client disconnects. [click] NotifyContext turns a signal such as SIGTERM into a cancel in main.

All of this is documented, and has been for years. I still see many context canceled errors in my logs.
-->

---
layout: default
kicker: Context primer
title: Go has documented the context rules since 2014
---

<div class="timeline">
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">Jul 2014</div><div class="tl-title">Go blog</div>
    <div class="tl-desc">Go Concurrency Patterns: Context. Pass ctx as the first argument.</div>
  </div>
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">Aug 2016</div><div class="tl-title">Go 1.7</div>
    <div class="tl-desc">context moves into the standard library.</div>
  </div>
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">Feb 2021</div><div class="tl-title">Go blog</div>
    <div class="tl-desc">Contexts and structs. Don't store a context in a struct.</div>
  </div>
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">Nov 2022</div><div class="tl-title">Google style guide</div>
    <div class="tl-desc">Rules for passing contexts and for custom Context types.</div>
  </div>
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">2023</div><div class="tl-title">Go 1.20, 1.21</div>
    <div class="tl-desc">Causes and WithoutCancel.</div>
  </div>
</div>

<!--
The rules are not new. [click] The Go blog had them in 2014. [click] context joined the standard library in Go 1.7. [click] The Go blog added a post on structs in 2021. [click] Google published its Go style guide in November 2022. [click] Causes and WithoutCancel came in 2023.

I still find the same bugs in code review every month. The docs explain how to pass a context. They say much less about how context code fails in production.
-->

---
layout: default
kicker: Context primer
title: Most context bugs I see are in these groups
---

<div class="three" style="grid-template-columns: repeat(4, minmax(0, 1fr)); column-gap: 1.8rem">
<div class="fail" v-click>
  <div class="fail-num">1</div>
  <div class="fail-title">Work that runs too long</div>
  <div class="fail-why">The client left, but the query continues.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">2</div>
  <div class="fail-title">Work that stops too early</div>
  <div class="fail-why">The audit write fails when the handler returns.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">3</div>
  <div class="fail-title">A cancel with no cause</div>
  <div class="fail-why">The log says only context canceled.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">4</div>
  <div class="fail-title">A value that is wrong or hidden</div>
  <div class="fail-why">Another package's key hides the user.</div>
</div>
</div>

<!--
These are the problems. Everything after this slide is advice that prevents one of them.

[click] In the first group, work doesn't stop. A layer didn't pass ctx, or never checks it. [click] In the second, work stops too early, because something that has to outlive the request got the request's context. [click] The third group is the log line from the start: a cancel with no cause. [click] The fourth is about values. Two packages use the same key and the second hides the first, or a required input hides in the context. The compiler can't catch either.

Each piece of advice works on its own, so you can take the ones that fit your code.
-->

---
layout: section
index: "02"
kicker: Part 2
title: Passing a context down
---

---
layout: default
kicker: Passing a context down
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

<div class="points" style="--n: 2">
<v-clicks>

- Use `context.Background()` only in `main` and `init`
- In tests, use `t.Context()` (Go 1.24)
- Use `context.TODO()` where no context is passed in yet
- Never pass a `nil` context

</v-clicks>
</div>

<!--
[click] This is the bug I see most. context.Background starts a new tree in the middle of a request. Nothing above this line can stop the query, the trace splits in two, and the next service gets no deadline.

[click] The fix is to pass the ctx you got. Name it ctx and make it the first argument. Sameer Ajmani wrote in 2014 that Google requires this on every function between incoming and outgoing requests.

[click] Background belongs at the root of the tree: main and init. [click] In tests, use t.Context, added in Go 1.24. It is canceled before the test's cleanup functions run. [click] When you migrate old code and don't have a context yet, TODO marks the place, and you can grep for it. [click] And never pass nil.
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
[click] This Notifier got its context from whatever request built it. A caller with a 200 millisecond timeout can't give it to Send. When the stored context is canceled, every later Send fails with context canceled, because the request that built the Notifier ended.

[click] Take ctx as a parameter on each method. The caller's deadline and cancel then apply to the call.

[click] The Go blog post on structs says the same.
-->

---
layout: default
kicker: Passing a context down
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

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- Every blocking wait needs a `ctx.Done()` case, including channel sends
- Return `ctx.Err()`, so the caller can tell a cancel from a timeout

</v-clicks>
</div>

<!--
[click] This is a retry loop that passes ctx to the call, and still ignores it. time.Sleep doesn't take a context. If the request is canceled during the backoff, the loop sleeps anyway, retries, fails at once, and sleeps again.

[click] A select waits on the timer and on Done. [click] A cancel ends the wait.

[click] Use a select like this for any wait: a send, a receive, a timer. [click] Return ctx.Err, so the caller sees Canceled or DeadlineExceeded and can stop its own retries. Since Go 1.23, the garbage collector also frees a time.After timer that nobody reads, if go.mod says go 1.23 or later.
-->

---
layout: default
kicker: Passing a context down
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

<div class="points" style="--n: 1">
<v-clicks>

- `context.Background()` also drops the cancel, but it loses the trace ID and the log fields

</v-clicks>
</div>

<!--
[click] The handler starts an audit write in a goroutine and returns. The server cancels r.Context when the handler returns, so the write fails after the user already got a 201. In tests the fake audit store is fast, so the write finishes before the handler returns. In production it fails.

[click] [click] WithoutCancel, from Go 1.21, keeps the values and drops the stop signal and the deadline. [click] That leaves the work with no time limit, so put your own timeout on top.

[click] Background would drop the cancel too, but it loses the trace ID and every log field. In production I send this work to a pool instead of a plain go statement, so that shutdown can wait for it. My post on supervised fire-and-forget has the details.
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
  :total="5000"
  :click="false"
  :marker="{ at: 2000, label: 'parent deadline' }"
  after="clip"
  after-label="The time the child asked for, cut at 2s"
  :ticks="[{ at: 0, label: '0' }, { at: 2000, label: '2s' }, { at: 5000, label: '5s' }]"
  :rows="[
    { label: 'parent (2s)', from: 0, to: 2000 },
    { label: 'child (10s)', from: 0, to: 5000 },
  ]"
/>

</div>

<p v-click class="take">Set the deadline once, in the handler. Give a lower layer its own timeout only when it needs less time.</p>

<!--
[click] Here is a parent and a child. [click] The parent gets two seconds. [click] The child asks for ten.

[click] The child's deadline is still two seconds from now. A child's deadline is the earlier of its own and its parent's, so a deadline can only get shorter as it moves down the tree.

[click] The dashed part is the time the child asked for and never gets.

[click] In a real service, the parent is the handler and the child is a function that the handler calls. A lower layer can't give itself more time. So set the deadline once, where the request or the job starts. Give an inner call a shorter timeout when it needs one, for example a cache lookup that should give up fast. If some work can't be stopped partway, check ctx.Deadline before you start it.
-->
---
layout: section
index: "03"
kicker: Part 3
title: Recording why a context stopped
---

---
layout: default
kicker: Recording why a context stopped
title: Cancel with a cause that says why
---

<Code gutter="15rem">

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

<Mark :line="3" v-click="1">cancel takes a cause (Go 1.20)</Mark>
<Mark :line="4" v-click>cause on normal return</Mark>
<Mark :line="7" v-click>cause for this error</Mark>

</Code>

<div v-click>

<Terminal :lines="[{cmd: 'context.Cause(ctx)'}, 'inventory ord-123: connection refused: context canceled']" />

</div>

<div class="points" style="--n: 3">
<v-clicks>

- Only the first cancel sets the cause
- `ctx.Err()` still returns `context.Canceled`
- Work still running under `ctx` can read the cause

</v-clicks>
</div>

<!--
[click] ctx.Err has only two answers, Canceled or DeadlineExceeded. Wrapping the error with fmt.Errorf tells you where the code was. It doesn't tell you why the context stopped. The code that cancels knows the cause, and since Go 1.20 it can record it. WithCancelCause returns a cancel function that takes an error.

[click] The deferred call sets the cause for the normal return. It counts only if nothing canceled the context first. [click] On the error path, I cancel with a cause that wraps the original error. Both causes also wrap context.Canceled. Three slides on, I show why.

[click] context.Cause reads the cause back. [click] Only the first cancel counts, so the error cause stays and the deferred call does nothing. [click] ctx.Err doesn't change. [click] The cause moves down the tree with the cancel, so a goroutine still working under this ctx, or the code that logs its failure, can read it.

A colleague at DoorDash debugged a case where a task ran with the request context. After a successful request, it logged context canceled, and people looked for a client disconnect. With a "completed" cause, the log line says the request had finished.
-->

---
layout: default
kicker: Recording why a context stopped
title: WithTimeoutCause sets a cause only when the timer fires
---

<Code gutter="13rem">

<div v-click>

```go
// Don't: expect this cause on every cancel.
ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second,
    fmt.Errorf("placeOrder %s: 5s timeout", id))
defer cancel()
```

</div>

<Mark :line="3" v-click>used only on timeout</Mark>
<Mark :line="4" v-click>cancel() sets no cause</Mark>

</Code>

<div v-click>

| Path | `context.Cause(ctx)` |
| --- | --- |
| The timer fires | `placeOrder ord-123: 5s timeout` |
| The function returns first | `context canceled` |
| An error path calls `cancel()` | `context canceled` |

</div>

<!--
[click] Most cancels come from timeouts, so Go 1.21 added WithTimeoutCause and WithDeadlineCause. You give the cause when you create the context.

[click] The context uses that cause only when the timer fires. [click] The cancel function you get back is a plain CancelFunc. In the standard library it cancels with Canceled and a nil cause.

[click] So the cause shows up only on the timeout path. Most requests return normally. That path and every cancel() call give a plain context canceled.

Sameer Ajmani's proposal for these functions, issue 56661, says this: the returned CancelFunc doesn't set the cause, and in the common case the cause is context.Canceled.
-->

---
layout: default
kicker: Recording why a context stopped
title: Put a cancel cause on top of the timeout
---

<div v-click class="pull pull-top">
If users need to set the cause on both the timer and cancel() paths, they can stack contexts.
<cite>Sameer Ajmani, golang/go#56661, 2022</cite>
</div>

<Code gutter="15rem">

<div v-click>

```go
// Do: stack a cancel cause on the timeout.
func placeOrder(ctx context.Context, id string) error {
    ctx, cancelCause := context.WithCancelCause(ctx)
    ctx, cancelTimeout := context.WithTimeoutCause(ctx, 5*time.Second,
        fmt.Errorf("placeOrder %s: 5s timeout: %w", id, context.DeadlineExceeded))
    defer cancelTimeout()
    defer cancelCause(fmt.Errorf("placeOrder: completed: %w", context.Canceled))

    // call cancelCause(...) on each error path
}
```

</div>

<Mark :line="3" v-click>takes a cause on every path</Mark>
<Mark :line="4" v-click>has the timer and the deadline</Mark>
<Mark :line="7" v-click>deferred last, so it runs first</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- On timeout, `ctx.Err()` still returns `DeadlineExceeded`
- The next service still gets the deadline

</v-clicks>
</div>

<!--
[click] The proposal that added WithTimeoutCause suggests a fix: stack two contexts.

[click] This way, every path gets a cause.

[click] The outer context has a cancel function that takes a cause. [click] The inner context owns the timer, so it has a deadline.

[click] Defers run last in, first out. cancelCause runs first and sets the cause. cancelTimeout then finds the context already canceled and does nothing. If you swap the two defer lines, you lose the cause.

[click] The stack still acts like a plain timeout. On timeout, Err returns DeadlineExceeded. [click] And Deadline returns the deadline, so the gRPC or HTTP client sends it to the next service.

This is easy to get wrong at every call site. Part 5 puts it in a package.
-->

---
layout: default
kicker: Recording why a context stopped
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
[click] Here the cause is a plain errors.New. [click] So it doesn't wrap DeadlineExceeded.

[click] Since Go 1.23, the HTTP client returns context.Cause instead of the sentinel. So err is the cause. [click] And this check is false. A retry policy or an alert that checks DeadlineExceeded stops matching when you add causes.

[click] The fix is to wrap DeadlineExceeded with %w. [click] Then errors.Is matches again, and the message still has the cause. The timeout cause on the last slide does this.
-->

---
layout: default
kicker: Recording why a context stopped
title: Log the error and the cause as separate fields
---

<Code>

<div v-click>

```go
// Do: log the error and the cause.
if err := ctx.Err(); err != nil {
    logger.LogAttrs(ctx, slog.LevelError, "place order",
        slog.Any("err", err),
        slog.Any("cause", context.Cause(ctx)),
    )
}
```

</div>

</Code>

<div v-click>

<Terminal :lines="[
  'level=ERROR msg=&quot;place order&quot;',
  '    err=&quot;context deadline exceeded&quot;',
  '    cause=&quot;placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<div class="points" style="--n: 2">
<v-clicks>

- `err` gives the kind of failure, so dashboards can group by it
- `cause` gives the details, so the on-call engineer knows what to fix

</v-clicks>
</div>

<!--
[click] A cause helps only if someone reads it, and the code that logs is usually far from the code that canceled.

ctx.Err and context.Cause answer different questions, so I log both as separate structured fields.

[click] A dashboard can count timeouts by err, and the same log line has the cause. Compare it with the log from the start of the talk.

[click] err has few distinct values, so it is good for grouping. [click] cause is the text that the on-call engineer reads during an incident.
-->

---
layout: section
index: "04"
kicker: Part 4
title: Context values
---

---
layout: two-cols
class: even
kicker: Context values
title: Put a value in the context only if the code works without it
---

<div v-click>
<div class="col-h">Put in a context</div>

- Trace, request, and correlation IDs
- Request-scoped log fields
- Auth data, read once in the handler

</div>

<img src="/img/gopher-umbrella.svg" class="gopher sm" style="margin-top:1.5rem" />

::right::

<div v-click>
<div class="col-h">Make it a parameter</div>

- User data the function works on
- Database handles and service clients
- Configuration
- Inputs that change the result

</div>

<div v-click class="pull">
Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions.
<cite>context package doc</cite>
</div>

<!--
Values move down the same tree as the cancel signal, so they are easy to use. A value doesn't show in any function signature, so when one is missing, nothing fails at compile time.

[click] My test: if the function can't do its job without the value, the value goes in the signature. A trace ID passes the test, because nothing breaks without one. Auth data is the hard case. http.Handler has a fixed signature, so the middleware puts the user in the request context, and the handler reads it once and passes it on as an argument.

[click] A user ID that a query needs fails the test, and so does anything that changes the result.

[click] The package doc says the same. An optional parameter changes what a function does. A trace ID doesn't, so it isn't a parameter.
-->

---
layout: default
kicker: Context values
title: Use an unexported key type, so keys can't collide
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

// Do: use an unexported key type.
package auth

type ctxKey struct{}
```

</div>

<Mark :block="2" :line="5" v-click>only this package can make one</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- `Value` goes up the tree and returns the value of the first equal key
- Two string keys with the same text are equal

</v-clicks>
</div>

<!--
[click] Values are found by key, and any package can add a key. Here two packages pick the same string. The second WithValue hides the first. Alice is still in the parent context, but nobody can reach her through this one.

[click] [click] An unexported key type fixes this. Only this package can write ctxKey, so no other package can make an equal key. An empty struct also doesn't allocate when it is stored in an interface.

[click] Value returns the first match on the way up. [click] Keys are compared as interfaces, and two strings with the same type and text are equal. staticcheck flags built-in key types with SA1029.
-->

---
layout: default
kicker: Context values
title: Hide the key behind a setter and a getter
---

<Code gutter="15rem">

<div v-click>

```go
// Do: keep the key behind two functions.
package auth

func WithUser(ctx context.Context, u User) context.Context {
    return context.WithValue(ctx, ctxKey{}, u)
}

func UserFromContext(ctx context.Context) (User, bool) {
    u, ok := ctx.Value(ctxKey{}).(User)
    return u, ok
}
```

</div>

<Mark :line="4" v-click>the only way to set it</Mark>
<Mark :line="9" v-click>the only type assertion</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- Callers get a typed value and a bool, and never see the key
- A missing value returns `false`

</v-clicks>
</div>

<!--
[click] The key type stays unexported, and the package exports two functions. [click] One sets the value. [click] One reads it, and the type assertion is in one place.

[click] Callers get a User and a bool. [click] If the value is missing, they get false, not a panic. Auth middleware calls WithUser, and the handler calls UserFromContext once, then passes the user on as an argument.

The standard library and OpenTelemetry do the same. httptrace, pprof labels, OpenTelemetry spans, and gRPC metadata each export a setter and a getter, and none of them export the key type.
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

<Mark :block="2" :line="3" v-click>callers have to pass it</Mark>

</Code>

<!--
[click] A typed getter stops key collisions. It can't stop a required value from hiding in a context.

A reader suggested this pattern on my repository post: put the transaction in the context and let the store pick it up.

[click] If someone forgets to call WithTx, or calls it on the wrong context, ok is false. [click] The store falls back to the connection pool. Nothing returns an error, and the writes are no longer atomic.

[click] Pass the transaction as an argument. [click] Now the dependency is in the signature, and the compiler checks that every caller passes it.

Code review catches most of these bugs, but not all of them, so the last part is about tools.
-->

---
layout: section
index: "05"
kicker: Part 5
title: Enforcing the rules
---

---
layout: default
kicker: Enforcing the rules
title: Let linters catch what code review misses
---

<div class="rules">
<div class="rule" v-click><code>go vet: lostcancel</code><span>A cancel function that is never called</span></div>
<div class="rule" v-click><code>staticcheck: SA1012</code><span>A <code>nil</code> context passed to a function</span></div>
<div class="rule" v-click><code>staticcheck: SA1029</code><span>A built-in type, such as a string, used as a key</span></div>
<div class="rule" v-click><code>containedctx</code><span>A context stored in a struct field</span></div>
<div class="rule" v-click><code>contextcheck</code><span>A function that uses a context it didn't get from its caller</span></div>
<div class="rule" v-click><code>errorlint</code><span><code>err == context.Canceled</code>, which is false when a cause wraps it</span></div>
</div>

<!--
A tool can check most of these rules.

[click] lostcancel is on by default in go vet. go test runs only some vet checks, and lostcancel isn't one of them, so run go vet or golangci-lint in CI. [click] [click] staticcheck flags a nil context and a built-in key type. [click] [click] containedctx and contextcheck are in golangci-lint. contextcheck finds the Background in the middle of a chain. It also reports some false positives, so start it on new code. [click] errorlint finds equality checks against errors. Once causes wrap context.Canceled, those checks need errors.Is.
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
    ctx, cancel := context.WithCancelCause(parent)
    return ctx, func(cause error) {
        cancel(fmt.Errorf("%w: %w", cause, context.Canceled))
    }
}
```

</div>

<Mark :line="6" v-click>the cancel func needs a cause</Mark>
<Mark :line="7" v-click>errors.Is still matches</Mark>

<div v-click>

```yaml

# Do: block the plain constructors.
# .golangci.yml
version: "2"
linters:
  enable: [forbidigo]
  settings:
    forbidigo:
      analyze-types: true
      forbid:
        - pattern: ^context\.With(Cancel|Timeout|Deadline)$
          msg: use xctx, which takes a cause
```

</div>

</Code>

<p v-click class="muted" style="max-width:none">At DoorDash, Alex Hornbake built this as <code>xctx</code> for a Go monorepo that many teams share.</p>

<!--
[click] Stacking contexts and wrapping context.Canceled at every call site is easy to get wrong. So put it in a small package. [click] Its cancel function takes a cause, so callers have to give one. [click] It wraps context.Canceled, so errors.Is keeps working. The package's WithTimeout does the stack from Part 3.

[click] Then a forbidigo rule fails the build when someone calls the standard constructors. The context.WithTimeout calls from Part 2 become xctx.WithTimeout. The package still returns a plain context.Context, so every library works as before.

[click] Alex Hornbake did this at DoorDash in early 2026. After the rollout, every cancel in that repo has a cause.
-->

---
layout: default
kicker: Wrap-up
title: The checklist I use in code review
---

<div class="side-gopher">
<div class="checklist">

<v-clicks>

1. Pass the caller's `ctx` to each call, and don't store it in a struct
2. Check `ctx` in every loop and every blocking wait
3. Give background work `WithoutCancel` and its own timeout
4. Give every cancel a cause, including the normal return
5. Make every cause wrap `context.Canceled` or `context.DeadlineExceeded`
6. Log `ctx.Err()` and `context.Cause(ctx)` as separate fields
7. Keep only optional values in a context, behind a setter and a getter
8. Run `go vet` and golangci-lint in CI, and block the plain constructors

</v-clicks>

</div>
<img src="/img/gopher-knight.svg" class="gopher" />
</div>

<!--
The checklist follows the order of the talk.

[click] [click] [click] These keep work from running too long or stopping too early. [click] [click] [click] These make context canceled say why. [click] This keeps values from turning into hidden parameters. [click] And the tools check most of the list on every pull request.
-->

---
layout: two-cols
class: wide-left
kicker: Wrap-up
title: Some cancels still have no cause
---

<v-clicks>

- The `net/http` server, when the client hangs up
- Libraries that call a plain `cancel()`
- Other services, because gRPC sends only the deadline and the cancel signal

</v-clicks>

::right::

<img src="/img/gopher-soldering.svg" class="gopher" />

<!--
Causes help only where somebody sets them.

[click] The HTTP server in the current Go source still uses a plain WithCancel for the request context, so a client disconnect gives a plain context canceled. [click] Many libraries don't set causes yet. [click] And a cause never crosses to the next service. gRPC sends the deadline and the cancel signal only, so each service has to set its own causes.

If all of your own cancels have a cause, a context canceled with no cause came from one of these, so start the search there. signal.NotifyContext sets a cause since Go 1.26.
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
Start with the Go blog posts. My posts go deeper on causes and values. The issues show what people asked for over the years, and how the API changed. The docs, style guides, and linters back the rules in this talk. The slides are on github.com/rednafi/talks.
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
