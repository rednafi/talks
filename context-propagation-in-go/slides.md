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
title: Overview
---

<div class="side-gopher">
<div class="overview">
<div class="ov-item" v-click>
  <div class="ov-num">01</div>
  <div class="ov-topic">A log line with no cause</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">02</div>
  <div class="ov-topic">Context primer</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">03</div>
  <div class="ov-topic">Passing a context down</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">04</div>
  <div class="ov-topic">Giving every cancel a cause</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">05</div>
  <div class="ov-topic">Context values</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">06</div>
  <div class="ov-topic">Enforcing the rules</div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">07</div>
  <div class="ov-topic">Checklist and limits</div>
</div>
</div>
<img src="/img/gopher-rocket.svg" class="gopher" />
</div>

<!--
[click] First, the problem: a log line that doesn't say why the context stopped.

[click] Then a short primer. What a context carries, and how it ends. [click] Then four habits for passing it down. [click] Then the core of the talk. Every cancel gets a cause, and the cause gets to the log. [click] Then values. [click] Then tools that check all of this. [click] Last, a checklist for code review, and the cancels that still have no cause.

Each slide is one rule and the bug it prevents. Take the ones that fit your code.
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
[click] I see logs like this in most Go services that call other services.

[click] Any of these can stop the context. [click] The client went away. [click] A caller's deadline passed. [click] The server started to shut down. [click] Or some code called cancel.

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
This is one request. The handler queries the database, then calls inventory, then payments.

At 500 milliseconds the client gives up. Maybe the user closed the app. Nothing tells the handler, so everything after the line is wasted.

The query holds its connection. Two services do work that nobody reads. During an incident, clients retry, and this waste adds to the load.

Go can't stop a goroutine from the outside. The work has to be told to stop, and then stop itself. That's what context is for.
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
- A context is read-only, so you add a deadline or a value with a child

</v-clicks>
</div>

<!--
[click] The package doc says what a context is for. Note "between processes". We'll get there.

[click] This is the whole interface. [click] Deadline says when to give up. [click] Done is a channel. It closes when the work should stop. [click] Err says how the context ended. It has only two answers. The next slide shows where they come from. [click] Value carries request data, like a trace ID.

[click] There's no Cancel method. Only the code that made the context can cancel it, so a child can't cancel its parent. [click] A context is read-only. You change it by wrapping it in a child.
-->

---
layout: default
kicker: Context primer
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

<div class="points" style="--n: 2">
<v-clicks>

- `Canceled`: the client disconnected, a shutdown signal arrived, or code called `cancel()`
- `DeadlineExceeded`: a timeout ran out, in this context or in a parent

</v-clicks>
</div>

<!--
[click] A context ends in one of two ways. One is a call to cancel. [click] Then Err returns context.Canceled.

[click] The other is a deadline. [click] When it passes, Err returns context.DeadlineExceeded.

These are the two lines from the start of the talk. [click] Canceled means some code called cancel. The HTTP server does it when the client disconnects. signal.NotifyContext does it on SIGTERM. Or your own code does it. [click] DeadlineExceeded means a timer ran out.

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
<p class="muted" v-click="7">When the client hangs up, everything below the request stops. The WithoutCancel branch keeps going.</p>
</div>
</div>

<!--
[click] Each child has one parent, so the contexts in a request form a tree. Background is the root. The server gives each request its own context.

[click] WithCancel returns a child and a cancel function. The server cancels the request context when the client goes away.

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

<TwoLevels />

<!--
"Context propagation" means two things.

Inside a process, your code passes ctx from the handler to the service to the client. Most of this talk is about that part.

Between processes, the gRPC client and OpenTelemetry read the deadline and the trace from ctx. They write them into headers. The server reads the headers and builds a new context.

The second part needs the first. If placeOrder starts from context.Background, the payments client has no deadline and no trace to send.
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
[click] The context only tells code to stop. The code has to stop itself. This function takes ctx but never looks at it. [click] If the client hangs up after the first image, the loop resizes all the others anyway.

[click] The fix is one check per item. [click] Once ctx is canceled, the loop returns before the next image.

[click] A cancel only closes a channel. [click] So every loop and every wait checks ctx, or passes it to code that does. The HTTP client, database/sql, and exec.CommandContext all check it for you.
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
These are the bugs I see most. Each part after this fixes one or more of them.

[click] Work runs too long. A layer didn't pass ctx, or never checks it. [click] Work stops too early. Something that has to outlive the request got the request's context. [click] A cancel with no cause. That's the log line from the start. [click] A value that's wrong or hidden. Two packages use the same key, or a required input hides in the context. The compiler can't catch either.
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

<div class="points" style="--n: 1">
<v-clicks>

- Every blocking wait needs a `ctx.Done()` case, including channel sends

</v-clicks>
</div>

<!--
[click] This retry loop passes ctx to the call, but still ignores it. time.Sleep doesn't take a context. If the request is canceled during the backoff, the loop sleeps anyway.

[click] A select waits on the timer and on Done. [click] A cancel ends the wait.

[click] Use a select like this for any wait. A send, a receive, a timer. Returning ctx.Err is the usual rule. Part 4 adds the cause.
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

- `context.Background()` drops the cancel too, but also the trace ID and log fields

</v-clicks>
</div>

<!--
[click] The handler starts an audit write in a goroutine and returns. The server cancels r.Context when the handler returns. So the write fails after the user got a 201. Tests often miss it, because the fake store is fast.

[click] [click] WithoutCancel keeps the values and drops the cancel and the deadline. [click] So the work has no time limit. Give it its own timeout.

[click] Background drops the cancel too, but you lose the trace ID and log fields. In production I send this work to a pool, so shutdown can wait for it. My post on supervised fire-and-forget has the details.
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
  after-label="The time the child asked for, cut at 2s"
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
title: Giving every cancel a cause
---

---
layout: default
kicker: Giving every cancel a cause
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
-->

---
layout: default
kicker: Giving every cancel a cause
title: Stack a cancel cause on the timeout
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: rely on WithTimeoutCause alone.
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
<Mark :block="2" :line="7" v-click>deferred last, so it runs first</Mark>

</Code>

<!--
[click] There's a catch. WithTimeoutCause uses its cause only when the timer fires. Its cancel function takes no cause. So the normal return gives a plain context canceled, and most requests return normally. Sameer Ajmani's proposal, issue 56661, says this. It also gives the fix: stack two contexts.

[click] Now every path gets a cause.

[click] The outer context takes a cause on the normal return and on each error path. [click] The inner context owns the timer. On timeout, Err still returns DeadlineExceeded, and the deadline still goes to the next service.

[click] Defers run in reverse. cancelCause runs first and sets the cause. Swap the two defer lines, and you lose it.

My post also shows a time.AfterFunc version. It's simpler, but the next service gets no deadline. This is easy to get wrong at every call site. Part 6 puts it in a package.
-->
---
layout: default
kicker: Giving every cancel a cause
title: When ctx ends, return its cause
---

<p v-click class="take" style="margin-top:0">A cause moves down to child contexts. Errors move up to the caller. So return the cause.</p>

<Code gutter="15rem">

<div v-click>

```go
// Don't: return ctx.Err() from the retry wait.
select {
case <-time.After(backoff(attempt)):
case <-ctx.Done():
    return ctx.Err()
}
```

</div>

<Mark :line="5" v-click="2">context deadline exceeded</Mark>

<div v-click>

```go

// Do: return the cause.
select {
case <-time.After(backoff(attempt)):
case <-ctx.Done():
    return context.Cause(ctx)
}
```

</div>

<Mark :block="2" :line="6" v-click>the 5s timeout cause from placeOrder</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- If a call such as `db.QueryContext` fails after ctx ends, return `context.Cause(ctx)` too
- With no cause set, `context.Cause(ctx)` returns `ctx.Err()`

</v-clicks>
</div>

<!--
[click] placeOrder set a cause on its context. Code below placeOrder runs under a child of that context, so it can read the cause. The handler sits above placeOrder. Its ctx never sees that cause. The only way up is the returned error.

[click] This is the retry wait from Part 3. It returns ctx.Err, so the caller gets only context deadline exceeded.

[click] Return context.Cause instead. [click] Now the caller gets placeOrder's reason. errors.Is still matches DeadlineExceeded, because the cause wraps it. The style guide says a function usually returns ctx.Err when ctx ends. Callers that check for it keep working. net/http does the same since Go 1.23.

[click] A library call works the same way. database/sql returns ctx.Err when ctx ends, so check ctx.Err and return the cause. [click] If nobody set a cause, Cause returns ctx.Err. So you lose nothing.
-->
---
layout: default
kicker: Giving every cancel a cause
title: Log the error once, in the handler
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: log in the store and return the error too.
slog.ErrorContext(ctx, "find order failed", "err", err)
return Order{}, fmt.Errorf("find order: %w", err)
```

</div>

<Mark :line="2" v-click="1">one more line for the same failure</Mark>

<div v-click>

```go

// Do: log once, in the handler.
if err := placeOrder(r.Context(), r.PathValue("id")); err != nil {
    slog.ErrorContext(r.Context(), "place order", "err", err)
    http.Error(w, "order failed", http.StatusInternalServerError)
    return
}
```

</div>

</Code>

<div v-click>

<Terminal :lines="[
  'level=ERROR msg=&quot;place order&quot;',
  '    err=&quot;placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<!--
[click] If every layer logs and returns, one failure prints many lines. The style guide says to let the caller log. My post on stacked log lines has more.

[click] Return the error up. Log it once, in the handler.

[click] The error carries the cause. So here's the log line from the start of the talk, now with the reason. placeOrder's five-second timeout fired during a retry wait. Dashboards can still group these with errors.Is.
-->
---
layout: section
index: "05"
kicker: Part 5
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
- Auth tokens from middleware
- Request-scoped log fields
- Idempotency keys

</div>

<img src="/img/gopher-umbrella.svg" class="gopher sm" style="margin-top:1.5rem" />

::right::

<div v-click>
<div class="col-h">Make it a parameter</div>

- User data the function works on
- Database handles and service clients
- Configuration
- Business logic inputs

</div>

<div v-click class="pull">
Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions.
<cite>context package doc</cite>
</div>

<!--
Values move down the same tree. They don't show up in any function signature. So when one is missing, the compiler can't tell you.

[click] My test: if the function can't work without it, it goes in the signature. A trace ID passes, because nothing breaks without it. Auth is the tricky case. http.Handler has a fixed signature, so middleware puts the user in the context. The handler reads it once and passes it on as an argument.

[click] A user ID that a query needs fails the test. So do clients and config. Every value in a context should be optional. But not every optional value belongs there.

[click] The package doc says the same from the other side.
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

<Mark :block="2" :line="3" v-click>callers have to pass it</Mark>

</Code>

<!--
[click] A typed getter stops key collisions. It can't stop a required value from hiding in a context. A reader suggested this on my repository post: put the transaction in the context, and let the store find it.

[click] If a caller forgets WithTx, ok is false. [click] Then the store falls back to the pool. Nothing fails, and the writes are no longer atomic.

[click] Pass the transaction as an argument. [click] Now it's in the signature, and the compiler checks every caller.
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
title: Linters can check most of these rules
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
        cancel(fmt.Errorf("%w: %w", cause, context.Canceled))
    }
}
```

</div>

<Mark :line="6" v-click>the cancel func needs a cause</Mark>
<Mark :line="7" v-click>nil still gets a cause</Mark>
<Mark :line="10" v-click>errors.Is still matches</Mark>

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

</Code>

<p v-click class="muted" style="max-width:none">At DoorDash, Alex Hornbake built this as <code>xctx</code> for a Go monorepo that many teams share.</p>

<!--
[click] Stacking contexts and wrapping Canceled at every call site is easy to get wrong. So put it in a small package. [click] Its cancel function takes a cause, so callers have to give one. [click] A nil cause still gets a message. [click] It wraps context.Canceled, so errors.Is keeps working. Its WithTimeout does the stack from Part 4.

[click] Then a forbidigo rule fails the build on the standard constructors, including the Cause versions. xctx itself is the one nolint. It still returns a plain context.Context, so every library works.

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

1. Pass `ctx` to each call, and don't store it in a struct
2. Check `ctx` in every loop and every blocking wait
3. Give background work `WithoutCancel` and a timeout
4. Give every cancel a cause, including the normal return
5. Stack `WithCancelCause` on a timeout to cover every path
6. Wrap `Canceled` or `DeadlineExceeded` in every cause
7. Return `context.Cause(ctx)`, and log once in the handler
8. Put only optional data in a context, and hide the key

</v-clicks>

</div>
<img src="/img/gopher-knight.svg" class="gopher" />
</div>

<!--
The checklist follows the order of the talk.

[click] [click] [click] These keep work from running too long or stopping too early. [click] [click] [click] [click] These make context canceled say why. [click] This keeps values from turning into hidden parameters. The linters and xctx check most of the list on every pull request.
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
Causes help only where someone sets them.

[click] The HTTP server still cancels the request context without a cause. So a client disconnect gives a plain context canceled. [click] Many libraries don't set causes yet. [click] And a cause never crosses to the next service. gRPC sends only the deadline and the cancel. Each service sets its own causes.

Once all your own cancels have causes, a plain context canceled came from one of these. Start looking there.
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
