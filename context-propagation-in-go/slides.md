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

Hi, I'm Redowan, **@rednafi** on the web.

<v-clicks>

I oscillate between product and platform work at DoorDash/Wolt.

Interested in databases, streaming, o11y, and failure modes of large-scale distsys.

Writing at **rednafi.com**.

<div class="about-links">github.com/rednafi<br>linkedin.com/in/rednafi</div>

</v-clicks>

</div>
<img class="about-gopher" src="/img/gopher-hiking.svg" alt="" />
</div>

<!--
[click] I oscillate between product and platform work, [click] and I care most about how distributed systems fail at scale.
-->
---
layout: quote
quote: "A distributed system is one in which the failure of a computer you didn't even know existed can render your own computer unusable."
author: Leslie Lamport, 1987
image: /img/gopher-network.svg
---

<!--
Lamport wrote this in an email at DEC in 1987. A timeout works the same way as a failure here. A deadline set three services up the call chain can stop work in my service, and the log line is all I get.
-->
---
layout: two-cols
class: wide-left
kicker: The problem
title: The log says a context stopped, but not why
---

<Terminal title="order-service" :lines="[
  'ERROR place order: context canceled',
  'ERROR fetch menu: context deadline exceeded',
  'ERROR place order: context canceled',
  'ERROR charge card: context canceled',
  'ERROR fetch menu: context deadline exceeded',
  'ERROR place order: context canceled',
]" />

::right::

<div class="col-h">Any of these can cause it</div>

<v-clicks>

- The client hung up
- A deadline set by a caller ran out
- The server started to shut down
- Some code called `cancel()`

</v-clicks>

<p v-click class="muted">Often the cancel started in a service or a library I have never read.</p>

<img v-click src="/img/gopher-mistake.svg" class="gopher sm" />

<!--
Errors like these are common in any Go service that talks to other services.

[click] It can mean the client went away, [click] that a timeout in some caller fired, [click] that the pod is shutting down, [click] or that some code called cancel on purpose.

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

The error says that a context stopped. It doesn't say who stopped it or why.
-->

---
layout: default
title: Overview
---

<div class="overview">
<div class="ov-item" v-click>
  <div class="ov-num">01</div>
  <div class="ov-topic">What's a context?</div>
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
  <div class="ov-topic">Enforcing the rules with linters</div>
</div>
</div>

<!--
[click] First, what a context is, and the bugs I see around it. That part defines the problem.

The rest is advice. Each slide is one habit, with the bug it prevents. [click] Some are about passing the context down, [click] some about recording why it stopped, [click] and some about values. [click] The last part shows linters that check the habits for you.
-->

---
layout: section
index: "01"
kicker: Part 1
title: What's a context?
---

---
layout: default
kicker: What's a context?
title: When the caller gives up, the work keeps going
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
- The inventory and payment services do work that nobody reads

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
kicker: What's a context?
title: Go can't stop a goroutine from the outside
---

<v-clicks>

- A goroutine has no handle and no kill method
- So the work has to be told to stop, and then stop itself
- Every goroutine on the request needs the same signal
- The signal also needs a time limit, so work can stop when time runs out

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

[click] A request usually runs in many goroutines: the handler, a fan-out, the database driver. All of them need the same signal. [click] The signal also needs a time limit, so that work stops on its own when time runs out, with nobody calling cancel.

[click] Google wrote the context package for this.
-->

---
layout: default
kicker: What's a context?
title: context.Context
---

<div class="pull pull-top">
Package context defines the Context type, which carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
<cite>context package doc</cite>
</div>

<div v-click="1">
<Code gutter="19rem">

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

<Mark :line="2" v-click="2">when to give up</Mark>
<Mark :line="3" v-click="3">closed when the work should stop</Mark>
<Mark :line="4" v-click="4">Canceled or DeadlineExceeded</Mark>
<Mark :line="5" v-click="5">data about the request</Mark>

</Code>
</div>

<div class="points" style="--n: 2">
<v-clicks at="6">

- There is no `Cancel` method, so the code that gets the signal can't send it
- A context never changes. To add a deadline or a value, you derive a child

</v-clicks>
</div>

<!--
The package doc says what a context is for. Note "between processes". I come back to that in two slides.

[click] This is the whole interface. [click] Deadline tells the work when to give up. [click] Done is a channel that closes when the work should stop. Any number of goroutines can wait on it. [click] Err tells you why it stopped, and it has only two answers: Canceled or DeadlineExceeded. Part 3 is about that limit. [click] Value carries data about the request, such as a trace ID.

[click] The interface has no Cancel method. The function that gets a cancel signal is usually not the one that sends it, so a child can't cancel its parent. [click] A context is also read-only. You change it by wrapping it.
-->

---
layout: default
kicker: What's a context?
clicks: 6
title: Each With function wraps a parent, so contexts form a tree
---

<div class="tree-grid">
<div class="apis tight">
<div class="api"><code>Background()</code><span>The root, with no deadline and no cancel</span></div>
<div class="api" v-click="1"><code>WithCancel(parent)</code><span>Stop when the code calls cancel</span></div>
<div class="api" v-click="2"><code>WithTimeout(parent, d)</code><span>Stop after a duration. <code>WithDeadline</code> stops at a time</span></div>
<div class="api" v-click="3"><code>WithValue(parent, k, v)</code><span>Add one value</span></div>
<div class="api" v-click="4"><code>WithoutCancel(parent)</code><span>Keep the values and drop the stop signal (Go 1.21)</span></div>
</div>
<CtxTree />
<div class="stack">
<p class="muted" v-click="[5, 6]">The 2s timeout stops only the branch below it.</p>
<p class="muted" v-click="6">When the client hangs up, everything below the request stops, except the WithoutCancel branch.</p>
</div>
</div>

<!--
Each child has one parent, so the contexts in a request form a tree, with Background at the root. The server gives each request its own context.

[click] WithCancel returns a child and a function that cancels it. The request context works like this: the server cancels it when the client goes away.

[click] WithTimeout and WithDeadline do the same with a timer. Here the handler gives the query and the RPC two seconds.

[click] WithValue adds one value, here a trace ID, and the logger reads it.

[click] WithoutCancel, from Go 1.21, keeps the values and drops the stop signal. Part 2 uses it for work that has to finish.

[click] When the timer fires, the query and the RPC stop. The rest of the tree keeps going. A cancel moves down the tree and never up.

[click] When the client hangs up, the whole request stops. Only the branch under WithoutCancel keeps going, and it still has the trace ID. A value lookup goes up, from the child to the root.
-->
---
layout: diagram
kicker: What's a context?
title: A context travels inside a process and between processes
note: "<span style='color:var(--accent)'>In-process</span>: your code passes <code>ctx</code> from function to function. <span style='color:var(--cat-2)'>Distributed</span>: gRPC and OpenTelemetry copy it into headers. They can copy only what your code passed down."
---

<TwoLevels />

<!--
"Context propagation" can mean two things.

In-process propagation is the arrows inside each service. Your code passes ctx from the handler to the service to the client. Most of this talk is about this part.

Distributed propagation is the arrow between the two services. The payments client makes a gRPC call to the payment service. The gRPC client and the OpenTelemetry middleware read the deadline and the trace from the ctx you give them, and write them into headers. The server reads the headers and builds a new context.

The distributed part depends on the in-process part. If placeOrder starts from context.Background, the payments client gets a context with no deadline and no trace, and has nothing to send.
-->

---
layout: default
kicker: What's a context?
title: Each layer passes ctx to the next
---

<Code gutter="17rem">

```go
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    order, err := h.orders.Get(ctx, r.PathValue("id"))
    // ...
}
```

<Mark :line="2" v-click>the handler sets the budget</Mark>

<div v-click>

```go

func (s *OrderService) Get(ctx context.Context, id string) (Order, error) {
    return s.store.Find(ctx, id)
}
```

</div>

<Mark :block="2" :line="3" v-click>the service passes it on</Mark>

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
This is in-process propagation in code.

[click] The handler takes the request context, which the server cancels when the client goes away, and adds a two-second budget.

[click] The service doesn't care about time. [click] It passes ctx on. Most layers in a service look like this.

[click] The store gives ctx to database/sql. [click] The driver is the code that waits on Done. If the client hangs up or two seconds pass, the query stops.

If any layer drops ctx, every layer under it loses the signal.
-->

---
layout: default
kicker: What's a context?
title: Canceling a context doesn't stop code that never checks it
---

<Code gutter="15rem">

```go
// Ignores ctx.
func resizeAll(ctx context.Context, images []Image) error {
    for _, img := range images {
        resize(img) // 200ms each
    }
    return nil
}
```

<Mark :line="4" v-click>keeps going after a cancel</Mark>

<div v-click>

```go

// Checks ctx before each image.
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

- `cancel()` only closes the `Done` channel. It doesn't stop any goroutine
- So every loop and every wait has to check `ctx`, or pass it to code that does

</v-clicks>
</div>

<!--
Cancellation in Go is cooperative. The context tells the code to stop, and the code has to do the stopping itself.

This function resizes a list of images and takes ctx, but never looks at it. [click] If the client hangs up after the first image, the loop still resizes all the others.

[click] The fix is one check per item. [click] When the context is canceled, the loop returns before the next image.

[click] A cancel only closes a channel. [click] Code that waits has to select on ctx.Done, and code that calls something slow has to pass ctx to it. If no layer checks, nothing stops.
-->

---
layout: default
kicker: What's a context?
title: The standard library stops its own work when the context ends
---

<Code gutter="18rem">

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := http.DefaultClient.Do(req)

rows, err := db.QueryContext(ctx, listOrdersSQL, userID)

out, err := exec.CommandContext(ctx, "pg_dump", dbName).Output()
```

<Mark :line="5" v-click>closes the connection</Mark>
<Mark :line="7" v-click>the driver gets ctx and stops the query</Mark>
<Mark :line="9" v-click>kills the process</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- The HTTP server cancels `r.Context()` when the client goes away
- `signal.NotifyContext` cancels a context when a signal such as `SIGTERM` arrives

</v-clicks>
</div>

<!--
The standard library does its part of the cooperation. Give it a context and it stops on its own.

[click] The HTTP client closes the connection. [click] database/sql passes ctx to the driver, and drivers such as pgx stop the query when it ends. [click] CommandContext kills the child process.

The standard library also sends the signal. [click] The HTTP server cancels the request context when the client disconnects. [click] NotifyContext turns a signal such as SIGTERM into a cancel in main.

All of this is documented, and has been for years. My logs are still full of context canceled.
-->

---
layout: default
kicker: What's a context?
title: The rules have been written down since 2014
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
    <div class="tl-desc">Rules for contexts and custom contexts, made public.</div>
  </div>
  <div class="tl-item" v-click>
    <div class="tl-dot" /><div class="tl-date">2023</div><div class="tl-title">Go 1.20, 1.21</div>
    <div class="tl-desc">Causes and WithoutCancel.</div>
  </div>
</div>

<!--
The rules are not new. [click] The Go blog had them in 2014. [click] context joined the standard library in Go 1.7. [click] The Go blog added a post on structs in 2021. [click] Google published its Go style guide in November 2022. [click] Causes and WithoutCancel came in 2023.

I still find the same bugs in code review every month. The docs explain how to pass a context. They say much less about how it breaks in production.
-->

---
layout: default
kicker: What's a context?
title: Most context bugs I see fall into these groups
---

<div class="three" style="grid-template-columns: repeat(4, minmax(0, 1fr)); column-gap: 1.8rem">
<div class="fail" v-click>
  <div class="fail-num">1</div>
  <div class="fail-title">Work that doesn't stop</div>
  <div class="fail-sym">The client left, and the query keeps going.</div>
  <div class="fail-why">A layer drops the context or ignores it.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">2</div>
  <div class="fail-title">Work that stops too early</div>
  <div class="fail-sym">The audit write fails after the response.</div>
  <div class="fail-why">Background work got the request's context.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">3</div>
  <div class="fail-title">A stop that doesn't say why</div>
  <div class="fail-sym">context canceled, with no reason.</div>
  <div class="fail-why">The code that canceled didn't record one.</div>
</div>
<div class="fail" v-click>
  <div class="fail-num">4</div>
  <div class="fail-title">A value that goes missing</div>
  <div class="fail-sym">The user ID is gone, and nothing failed to compile.</div>
  <div class="fail-why">A key collided, or a required value hid in the context.</div>
</div>
</div>

<p v-click class="muted center-note">Each piece of advice that follows prevents one of these.</p>

<!--
These are the problems. Everything after this slide is advice that prevents one of them.

[click] In the first group, work doesn't stop. A layer broke the chain, or never checks. [click] In the second, work stops too early, because something that has to outlive the request got the request's context. [click] The third group is the log line from the start: a stop with no reason. [click] The fourth is about values. A value can go missing, and the compiler can't tell you.

[click] Each piece of advice works on its own, so you can take the ones that fit your code.
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
title: Pass ctx as the first argument, all the way down
---

<Code gutter="15rem">

```go
// Don't: a new root in the middle of the chain.
return s.store.Find(context.Background(), id)
```

<Mark :line="2">no cancel, no deadline, no trace</Mark>

<div v-click>

```go

// Do: pass the caller's ctx.
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
This is the bug I see most. context.Background starts a new tree in the middle of a request. Nothing above this line can stop the query, the trace splits in two, and the next service gets no deadline.

[click] The fix is to pass the ctx you got. Name it ctx and make it the first argument. Sameer Ajmani wrote in 2014 that Google requires this on every function between incoming and outgoing requests.

[click] Background belongs at the root of the tree: main and init. [click] In tests, use t.Context, added in Go 1.24. It is canceled just before the test's cleanup functions run. [click] When you migrate old code and don't have a context yet, TODO marks the place, and you can grep for it. [click] And never pass nil. Use TODO.
-->

---
layout: default
kicker: Passing a context down
title: Pass ctx to each call instead of storing it in a struct
---

<Code gutter="15rem">

```go
// Don't: every call uses the context from when the struct was built.
type Notifier struct {
    ctx    context.Context
    client *sms.Client
}
```

<Mark :line="3">one deadline for every call</Mark>

<div v-click>

```go

// Do: each caller passes its own.
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
This Notifier got its context from whatever request built it. A caller with a 200 millisecond budget can't pass it in. When the stored context is canceled, every later Send fails with context canceled, because of a request that ended long ago.

[click] Take ctx as a parameter on each method. The caller's deadline and cancel then apply to the call.

[click] The Go blog post on structs says this in one sentence.
-->

---
layout: default
kicker: Passing a context down
title: Wait on the timer and ctx.Done() together
---

<Code gutter="15rem">

```go
// Don't: the retry sleeps through a cancel.
time.Sleep(backoff(attempt))
```

<Mark :line="2">can't be interrupted</Mark>

<div v-click>

```go

// Do: return on whichever comes first.
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
- Since Go 1.23, a `time.After` timer that nobody reads is garbage collected

</v-clicks>
</div>

<!--
This is a retry loop that passes ctx correctly to the call, and still ignores it. time.Sleep doesn't take a context. If the request is canceled during the backoff, the loop sleeps anyway, retries, fails at once, and sleeps again.

[click] A select waits on the timer and on Done. [click] A cancel ends the wait.

[click] Use a select like this for any wait: a send, a receive, a timer. [click] Since Go 1.23, the runtime collects a time.After timer that nobody reads, so this doesn't leak timers.
-->

---
layout: default
kicker: Passing a context down
title: Detach work that has to outlive the request, then bound it
---

<Code gutter="17rem">

```go
// Don't: canceled when the handler returns.
go h.audit.Write(r.Context(), orderID)
```

<Mark :line="2">fails after the response</Mark>

<div v-click>

```go

// Do: keep the values, drop the cancel, add a bound.
detached := context.WithoutCancel(r.Context())
h.pool.Submit(func() {
    ctx, cancel := context.WithTimeout(detached, 5*time.Second)
    defer cancel()
    h.audit.Write(ctx, orderID)
})
```

</div>

<Mark :block="2" :line="3" v-click>Go 1.21</Mark>
<Mark :block="2" :line="5" v-click>a bound of your own</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- `context.Background()` also works, and loses the trace ID and the log fields
- A pool lets shutdown wait for the work, and a bare `go` doesn't

</v-clicks>
</div>

<!--
The handler starts an audit write in a goroutine and returns. The server cancels r.Context when the handler returns, so the write fails after the user already got a 201. In tests the handler is slow and the write finishes. In production it fails.

[click] [click] WithoutCancel, from Go 1.21, keeps the values and drops the stop signal and the deadline. [click] That leaves the work with no time limit, so put your own timeout on top.

[click] Background would stop the cancel too, but it loses the trace ID and every log field. [click] I send the work to a pool instead of a plain go statement, so that shutdown can wait for it. My post on supervised fire-and-forget has the details.
-->

---
layout: default
kicker: Passing a context down
title: Set the deadline at the edge, since a child can't extend it
---

<Code gutter="16rem">

```go
// Handler: the request gets 2s.
ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
defer cancel()
```

<Mark :line="2" v-click>the edge sets the budget</Mark>

<div v-click>

```go

// Deeper down: asking for 10s doesn't help.
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
```

</div>

<Mark :block="2" :line="3" v-click>still ends at 2s</Mark>

</Code>

<div v-click>

<Bars
  :total="5000"
  :click="false"
  :marker="{ at: 2000, label: 'handler deadline' }"
  after="clip"
  after-label="The 10s the child asked for, cut at 2s"
  :ticks="[{ at: 0, label: '0' }, { at: 2000, label: '2s' }, { at: 5000, label: '5s' }]"
  :rows="[
    { label: 'handler ctx', from: 0, to: 2000 },
    { label: 'WithTimeout(10s)', from: 0, to: 5000 },
  ]"
/>

</div>

<p v-click class="take">Give an inner call its own timeout only when it needs less time than the request.</p>

<!--
[click] The handler is the edge. It decides how long the whole request may take, here two seconds.

[click] Deeper down, some function asks for ten seconds. [click] It still ends at two. A child's deadline is the earlier of its own and its parent's, so a deadline can only get shorter as it moves down the tree.

[click] The dashed part is the time the child asked for and never gets.

[click] So set the real deadline once, where the request or the job starts. Give an inner call a shorter timeout when it needs one, for example a cache lookup that should give up fast. If some work can't be stopped half way, check ctx.Deadline before you start it.
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

```go
func processOrder(ctx context.Context, id string) error {
    ctx, cancel := context.WithCancelCause(ctx)
    defer cancel(errors.New("processOrder: completed"))

    if err := checkInventory(ctx, id); err != nil {
        cancel(fmt.Errorf("inventory %s: %w", id, err))
        return err
    }
    // ...
}
```

<Mark :line="2">cancel takes a cause (Go 1.20)</Mark>
<Mark :line="3" v-click>default reason</Mark>
<Mark :line="6" v-click>specific reason</Mark>

</Code>

<div v-click>

<Terminal :lines="[{cmd: 'context.Cause(ctx)'}, 'inventory ord-123: connection refused']" />

</div>

<div class="points" style="--n: 3">
<v-clicks>

- The first cancel wins, and later calls do nothing
- `ctx.Err()` still returns `context.Canceled`
- `context.Cause` works on any context below this one

</v-clicks>
</div>

<!--
ctx.Err has only two answers, Canceled or DeadlineExceeded. Wrapping the error with fmt.Errorf tells you where the code was. It doesn't tell you why the context stopped. The code that cancels knows the reason, and since Go 1.20 it can record it. WithCancelCause returns a cancel function that takes an error.

[click] The deferred call sets the default reason. It counts only if nothing canceled the context first. [click] On the error path, I cancel with a specific reason that wraps the original error.

[click] context.Cause reads the reason back. [click] The first cancel wins, so the specific reason stays and the deferred call does nothing. [click] ctx.Err doesn't change. [click] And the cause moves down the tree with the cancel, so a function deep in the tree, the one that logs, can read a reason set at the top.

The "completed" cause on the deferred cancel matters. A colleague at DoorDash debugged a case where a best-effort task started with the request context logged context canceled after a successful request. People went looking for a client disconnect. With a "completed" cause, the log line says the request had finished.
-->

---
layout: default
kicker: Recording why a context stopped
title: Don't rely on WithTimeoutCause alone
---

<Code gutter="9rem">

```go
ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second,
    fmt.Errorf("processOrder %s: 5s timeout", id))
defer cancel()
```

<Mark :line="3" v-click>takes no cause</Mark>

</Code>

<div v-click>

<Code file="context/context.go" gutter="9rem">

```go
return c, func() { c.cancel(true, Canceled, nil) }
```

</Code>

</div>

<div v-click>

| Path | `context.Cause(ctx)` |
| --- | --- |
| The timer fires | `processOrder ord-123: 5s timeout` |
| The function returns first | `context canceled` |
| An error path calls `cancel()` | `context canceled` |

</div>

<!--
Most cancels come from timeouts, so Go 1.21 added WithTimeoutCause and WithDeadlineCause. You give the cause up front, and the context uses it when the timer fires.

[click] The cancel function you get back is a plain CancelFunc. [click] This is the line in the standard library. It cancels with Canceled and a nil cause.

[click] So the cause shows up only on the timeout path. The normal return, which happens on most requests, and every explicit cancel give you a bare context canceled.

Sameer Ajmani's proposal for these functions, issue 56661, says this directly: the returned CancelFunc doesn't set the cause, and in the common case the cause is context.Canceled.
-->

---
layout: default
kicker: Recording why a context stopped
title: Stack a cancel cause on top of the timeout
---

<Code gutter="15rem">

```go
func processOrder(ctx context.Context, id string) error {
    ctx, cancelCause := context.WithCancelCause(ctx)
    ctx, cancelTimeout := context.WithTimeoutCause(ctx, 5*time.Second,
        fmt.Errorf("processOrder %s: 5s timeout", id))
    defer cancelTimeout()
    defer cancelCause(errors.New("processOrder: completed"))

    // call cancelCause(...) on each error path
}
```

<Mark :line="2" v-click>takes a cause on every path</Mark>
<Mark :line="3" v-click>owns the timer and the deadline</Mark>
<Mark :line="6" v-click>deferred last, so it runs first</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- `ctx.Err()` is `DeadlineExceeded` when the timer fires
- `ctx.Deadline()` still works, so gRPC can send it on

</v-clicks>
</div>

<div v-click class="pull">
If users need to set the cause on both the timer and cancel() paths, they can stack contexts.
<cite>Sameer Ajmani, golang/go#56661, 2022</cite>
</div>

<!--
To get a cause on every path, stack two contexts.

[click] The outer context has a cancel function that takes a cause. [click] The inner context owns the timer, so it is a real deadline context.

[click] Defers run last in, first out. cancelCause runs first and sets the reason. cancelTimeout then finds the context already canceled and does nothing. If you swap the two defer lines, you lose the cause.

[click] [click] The inner context keeps the deadline. A version with time.AfterFunc and one WithCancelCause also records causes, but its context has no deadline, so gRPC has nothing to send to the next service.

[click] The proposal that added WithTimeoutCause suggests this stack. It is a lot to get right at every call site. Part 5 shows how to put it in one helper.
-->

---
layout: default
kicker: Recording why a context stopped
title: Wrap the sentinel in every cause, so errors.Is still matches
---

<Code gutter="15rem">

```go
ctx, cancel := context.WithTimeoutCause(ctx, 20*time.Millisecond,
    errors.New("inventory: budget spent"))
defer cancel()

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
_, err := http.DefaultClient.Do(req)

errors.Is(err, context.DeadlineExceeded)
```

<Mark :line="2" v-click>no %w, so the sentinel is lost</Mark>
<Mark :line="8" v-click>false</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- Since Go 1.23, `net/http` returns `context.Cause(ctx)`, so a plain cause takes the place of the sentinel
- Retry and alert code that checks `DeadlineExceeded` stops matching

</v-clicks>
</div>

<div v-click style="margin-top: 1.5rem">

<Code gutter="15rem">

```go
fmt.Errorf("inventory: budget spent: %w", context.DeadlineExceeded)
```

<Mark :line="1" v-click>wrapped with %w, so true</Mark>

</Code>

</div>

<!--
The causes on the last few slides have a problem. A cause is an error like any other.

[click] This cause is a plain errors.New. It doesn't wrap DeadlineExceeded.

[click] So this check is false. net/http made this a real problem: since Go 1.23 the HTTP client returns context.Cause instead of the sentinel.

[click] A plain cause takes the place of the sentinel. [click] A retry policy or an alert that checks DeadlineExceeded stops working on the day you add causes.

[click] The fix is to wrap the sentinel with %w. [click] Then errors.Is matches the sentinel again, and the message still has the reason. Go 1.26 does this for signal.NotifyContext: its cause reports itself as context.Canceled.
-->

---
layout: default
kicker: Recording why a context stopped
title: Log the error and the cause as separate fields
---

<Code>

```go
if err := ctx.Err(); err != nil {
    logger.LogAttrs(ctx, slog.LevelError, "request failed",
        slog.Any("err", err),
        slog.Any("cause", context.Cause(ctx)),
    )
}
```

</Code>

<div v-click>

<Terminal :lines="[
  'level=ERROR msg=&quot;request failed&quot;',
  '    err=&quot;context deadline exceeded&quot;',
  '    cause=&quot;processOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<div class="points" style="--n: 2">
<v-clicks>

- `err` gives the kind, so dashboards can group by it
- `cause` gives the reason, so on-call knows what to do

</v-clicks>
</div>

<!--
A cause helps only if someone reads it, and the code that logs is usually far from the code that canceled.

ctx.Err and context.Cause answer different questions, so I log both as separate structured fields.

[click] A dashboard can count timeouts by err, and the same log line has the exact reason.

[click] err has few distinct values, so it is good for grouping. [click] cause is the text that on-call reads during an incident.
-->

---
layout: two-cols
class: wide-left
kicker: Recording why a context stopped
title: Give every cancel in your code a cause
---

<v-clicks>

- The `net/http` server cancels `r.Context()` with no cause when the client hangs up
- Most client libraries were written before Go 1.20
- `os/signal.NotifyContext` got a cause only in Go 1.26
- The cause stays in the process. gRPC sends only the deadline
- If every cancel in your code has a cause, a bare `context canceled` came from outside your code

</v-clicks>

::right::

<img src="/img/gopher-soldering.svg" class="gopher" />

<!--
Causes have limits. [click] The HTTP server in the current Go source still uses a plain WithCancel for the request context, so a client hang-up is still a bare context canceled.

[click] Most libraries are older than Go 1.20. [click] NotifyContext started to set "interrupt signal received" in Go 1.26. [click] And a cause never crosses to the next service. gRPC sends the deadline, not the reason, so each service has to set its own causes.

[click] So causes pay off when all of your own code sets them. Then a cancel with no cause also tells you something: it came from outside your code, and on-call can start with the caller.
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
title: Put a value in the context only if the code can run without it
---

<div class="col-h">Fine in a context</div>

<v-clicks>

- Trace, request, and correlation IDs
- Auth data set by middleware
- Request-scoped log fields
- Idempotency keys

</v-clicks>

<img src="/img/gopher-umbrella.svg" class="gopher sm" style="margin-top:1.5rem" />

::right::

<div class="col-h">Make it a parameter</div>

<v-clicks>

- User data the function works on
- Database handles and service clients
- Configuration
- Business inputs

</v-clicks>

<div v-click class="pull">
Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions.
<cite>context package doc</cite>
</div>

<!--
Values move down the same tree as the cancel signal, so they are easy to reach for. A value doesn't show in any function signature, so when one is missing, nothing fails at compile time.

My test: if the function can't do its job without the value, the value goes in the signature. Every value in a context should be optional.

A trace ID passes the test, because nothing breaks without one. A user ID that a query needs fails it. Auth data set by middleware is the common case. http.Handler has a fixed signature, so the middleware puts the user in the request context, and the handler reads it once and passes it on as an argument.

The Google style guide gives a similar list: security credentials, tracing information, deadlines, and cancellation signals.
-->

---
layout: default
kicker: Context values
title: Use an unexported key type, so keys can't collide
---

<Code gutter="14rem">

```go
// Don't: two packages pick the same string.
ctx = context.WithValue(ctx, "user", "alice")     // package auth
ctx = context.WithValue(ctx, "user", "audit-bot") // package audit
ctx.Value("user")                                 // "audit-bot"
```

<Mark :line="4">alice is gone</Mark>

<div v-click>

```go

// Do: a key type that only this package can name.
package auth

type ctxKey struct{}
```

</div>

<Mark :block="2" :line="5" v-click>no other package can build one</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- `Value` walks up the tree and returns the first key that is equal
- Two string keys with the same text are equal

</v-clicks>
</div>

<!--
Values are found by key, and any package can add a key. Here two packages pick the same string. The second WithValue hides the first. Alice is still in the parent context, but nobody can reach her through this one.

[click] [click] An unexported key type fixes this. Only this package can write ctxKey, so no other package can make an equal key. An empty struct also doesn't allocate when it is stored in an interface.

[click] Value returns the first match on the way up. [click] Keys are compared as interfaces, and two strings with the same type and text are equal. staticcheck flags built-in key types with SA1029.
-->

---
layout: default
kicker: Context values
title: Hide the key behind a setter and a getter
---

<Code gutter="15rem">

```go
package auth

func WithUser(ctx context.Context, u User) context.Context {
    return context.WithValue(ctx, ctxKey{}, u)
}

func UserFromContext(ctx context.Context) (User, bool) {
    u, ok := ctx.Value(ctxKey{}).(User)
    return u, ok
}
```

<Mark :line="3" v-click>the only way in</Mark>
<Mark :line="8" v-click>the one type assertion</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- Callers get a typed value and a bool, and never touch the key
- A missing value gives `false`, not a panic

</v-clicks>
</div>

<!--
The key type stays unexported, and the package exports two functions. [click] One sets the value. [click] One reads it, and the type assertion is in one place.

[click] Callers get a User and a bool. [click] If the value is missing, they get false, not a panic. Auth middleware calls WithUser, and the handler calls UserFromContext once, then passes the user on as an argument.

The standard library and OpenTelemetry do the same. httptrace, pprof labels, OpenTelemetry spans, and gRPC metadata each export a setter and a getter, and none of them export the key type.
-->

---
layout: default
kicker: Context values
title: Don't hide a dependency in a context
---

<Code gutter="15rem">

```go
// Don't do this.
func (s *BookStore) Create(ctx context.Context, b Book) (int64, error) {
    var db DBTX = s.db
    if tx, ok := TxFromContext(ctx); ok {
        db = tx
    }
    // ...
}
```

<Mark :line="4" v-click>forget WithTx and this is false, with no error</Mark>

<div v-click>

```go

// Do this: the transaction is an argument.
err := store.Tx(ctx, func(tx book.Store) error {
    // use tx for every write
})
```

</div>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- The writes go to the pool and are no longer atomic
- The compiler can't help, because the dependency is invisible

</v-clicks>
</div>

<!--
A typed getter stops key collisions. It can't stop a required value from hiding in a context.

A reader suggested this pattern on my repository post: put the transaction in the context and let the store pick it up.

[click] If someone forgets to call WithTx, or calls it on the wrong context, the store falls back to the connection pool. Nothing returns an error, and the writes are no longer atomic.

[click] A callback that receives the transactional store makes the dependency visible in the signature.

[click] [click] Code review can catch this and most of the earlier advice. Review also misses some of it, so the last part is about tools.
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
<div class="rule" v-click><code>errorlint</code><span><code>err == context.Canceled</code>, which fails on a wrapped cause</span></div>
</div>

<!--
Most of the advice so far can be checked by a tool.

[click] go vet runs with go test, and lostcancel is on by default. [click] [click] staticcheck flags a nil context and a built-in key type. [click] [click] containedctx and contextcheck are in golangci-lint. contextcheck finds the Background in the middle of a chain. It also reports some false positives, so start it on new code. [click] errorlint finds equality checks against errors. Once causes wrap the sentinel, those checks have to use errors.Is.
-->

---
layout: default
kicker: Enforcing the rules
title: Wrap the constructors so every cancel needs a cause
---

<Code gutter="16rem">

```go
package xctx

func WithCancel(parent context.Context) (context.Context, context.CancelCauseFunc) {
    ctx, cancel := context.WithCancelCause(parent)
    return ctx, func(cause error) {
        cancel(fmt.Errorf("%w: %w", cause, context.Canceled))
    }
}
```

<Mark :line="5" v-click>the cancel func needs a reason</Mark>
<Mark :line="6" v-click>errors.Is still matches</Mark>

<div v-click>

```yaml

# .golangci.yml
linters:
  settings:
    forbidigo:
      forbid:
        - pattern: ^context\.With(Cancel|Timeout|Deadline)$
          msg: use xctx, which takes a cause
```

</div>

</Code>

<p v-click class="muted" style="max-width:none">At DoorDash, Alex Hornbake built this as <code>xctx</code> for a Go monorepo that many teams share.</p>

<!--
Stacking contexts and wrapping sentinels at every call site is easy to get wrong. So put it in a small package. [click] Its cancel function takes a cause, so you can't cancel without one. [click] It wraps the sentinel, so errors.Is keeps working. A WithTimeout in the same package does the stack from part 3.

[click] Then a forbidigo rule fails the build when someone calls the standard constructors. The package still returns a plain context.Context, so every library works as before.

[click] Alex Hornbake did this at DoorDash in early 2026. After the rollout, every cancel in that repo carries a reason, and a bare context canceled points to the client or a library.
-->

---
layout: default
kicker: Wrap-up
title: The checklist I use in code review
---

<div class="checklist">

<v-clicks>

1. Pass `ctx` first, all the way down, and don't store it in a struct
2. Wait on `ctx.Done()` in every loop and every blocking wait
3. Set the deadline at the edge, and give detached work its own timeout
4. Give every cancel a cause, including the normal return
5. Make every cause wrap `context.Canceled` or `context.DeadlineExceeded`
6. Log `ctx.Err()` and `context.Cause(ctx)` as separate fields
7. Keep values optional and keys unexported, behind a setter and a getter
8. Let linters check the rest

</v-clicks>

</div>

<!--
The checklist follows the order of the talk.

[click] [click] [click] These keep work from running too long or stopping too early. [click] [click] [click] These make context canceled say why. [click] This keeps values from turning into hidden parameters. [click] And linters keep the rest true when nobody is looking.
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
The Go blog posts are the canonical guide. My posts go deeper on causes and values. The issues show what people asked for over the years, and how the API changed. The docs, style guides, and linters back the rules in this talk. The slides are on github.com/rednafi/talks.
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
