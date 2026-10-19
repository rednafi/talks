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
layout: default
title: Overview
---

<div class="side-gopher">
<div class="overview">
<div class="ov-item" v-click>
  <div class="ov-num">01</div>
  <div><div class="ov-topic">Context primer</div><div class="ov-desc">What a context carries, the two ways it ends, and the context tree</div></div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">02</div>
  <div><div class="ov-topic">Passing a context down</div><div class="ov-desc">Use the caller's ctx, pass it to each call, wait on Done, and detach background work</div></div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">03</div>
  <div><div class="ov-topic">Giving every cancel a cause</div><div class="ov-desc">Set a cause, wrap the sentinel, return the cause, and log it once</div></div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">04</div>
  <div><div class="ov-topic">Context values</div><div class="ov-desc">What goes in a context, and how to hide the key</div></div>
</div>
<div class="ov-item" v-click>
  <div class="ov-num">05</div>
  <div><div class="ov-topic">Enforcing the rules</div><div class="ov-desc">Linters, and a wrapper that makes every cancel take a cause</div></div>
</div>
</div>
<img src="/img/gopher-rocket.svg" class="gopher" />
</div>

<!--
[click] First, a short primer: what a context carries, and the two ways it ends.

[click] Then four habits for passing it down. [click] Then the core of the talk: every cancel gets a cause, and the cause reaches the log. [click] Then values. [click] Last, tools that check all of this for you.

Each slide is one rule and the bug it prevents, so you can take the ones that fit your code.
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

Nothing tells the handler, so all the work after the line is wasted. Go has no way to stop a goroutine from the outside. The work has to be told to stop, and then stop itself. The context package does the telling.

The query keeps its connection. Two other services do work for a client that has left. During an incident, clients time out and retry, and this extra work adds to the load that caused the timeouts.
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
[click] The package doc says what a context is for. Note "between processes". I come back to that in a few slides.

[click] This is the whole interface. [click] Deadline tells the work when to give up. [click] Done is a channel that closes when the work should stop. Any number of goroutines can wait on it. Nothing stops on its own: code has to watch Done, or pass ctx to code that does, such as the HTTP client or database/sql. [click] Err tells you how it ended, and it has only two answers: Canceled or DeadlineExceeded. The next slide shows where they come from. [click] Value carries data about the request, such as a trace ID.

[click] The interface has no Cancel method. The function that gets a cancel signal is usually not the one that sends it, so a child can't cancel its parent. [click] A context is also read-only. You change it by wrapping it.
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
[click] A context ends in one of two ways. The first is a call to cancel. [click] After that, Err returns context.Canceled.

[click] The second is a deadline. [click] When it passes, Err returns context.DeadlineExceeded.

These are the two lines from the start of the talk. [click] context canceled means some code called cancel: the HTTP server when the client disconnects, signal.NotifyContext when SIGTERM arrives, or your own code. [click] deadline exceeded means a timer ran out. A child gets the same error as its parent, so the error doesn't say which context ended, or why.
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
- In a loop, check `ctx.Err()` before each item

</v-clicks>
</div>

<!--
[click] This is a retry loop that passes ctx to the call, and still ignores it. time.Sleep doesn't take a context. If the request is canceled during the backoff, the loop sleeps anyway, retries, fails at once, and sleeps again.

[click] A select waits on the timer and on Done. [click] A cancel ends the wait.

[click] Use a select like this for any wait: a send, a receive, a timer. [click] In a loop, check ctx.Err before each item. The context only tells code to stop. The code has to stop itself. Returning ctx.Err is the usual rule, and Part 3 adds the cause to it.
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
layout: section
index: "03"
kicker: Part 3
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
[click] ctx.Err has only two answers, Canceled or DeadlineExceeded. Wrapping the error with fmt.Errorf tells you where the code was. It doesn't tell you why the context stopped. The code that cancels knows the cause, and since Go 1.20 it can record it. WithCancelCause returns a cancel function that takes an error.

[click] The deferred call sets the cause for the normal return. It counts only if nothing canceled the context first. [click] On the error path, I cancel with a cause that wraps the original error. Both causes also wrap context.Canceled. The next slide shows why.

[click] context.Cause reads the cause back. [click] Only the first cancel counts, so the error cause stays and the deferred call does nothing. [click] ctx.Err doesn't change. [click] The cause moves down the tree with the cancel, so a goroutine still working under this ctx, or the code that logs its failure, can read it.

A colleague at DoorDash debugged a case where a task ran with the request context. After a successful request, it logged context canceled, and people looked for a client disconnect. With a "completed" cause, the log line says the request had finished.
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
[click] Timeouts get a cause too. Go 1.21 added WithTimeoutCause: you give the cause when you create the context, and the context uses it when the timer fires. Here the cause is a plain errors.New. [click] So it doesn't wrap DeadlineExceeded.

[click] Since Go 1.23, the HTTP client returns context.Cause instead of the sentinel. So err is the cause. [click] And this check is false. A retry policy or an alert that checks DeadlineExceeded stops matching when you add causes.

[click] The fix is to wrap DeadlineExceeded with %w. [click] Then errors.Is matches again, and the message still has the cause. The same goes for Canceled, which is why the causes on the last slide wrap it.
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
[click] WithTimeoutCause uses its cause only when the timer fires. The cancel function it returns is a plain CancelFunc, so the normal return, and every error path that calls cancel, gives a plain context canceled. Most requests return normally, so most of the time there is no cause. Sameer Ajmani's proposal for these functions, issue 56661, says this, and it suggests the fix: stack two contexts.

[click] This way, every path gets a cause.

[click] The outer context takes a cause on the normal return and on each error path. [click] The inner context owns the timer. On timeout, Err still returns DeadlineExceeded, and the gRPC or HTTP client still sends the deadline to the next service.

[click] Defers run last in, first out. cancelCause runs first and sets the cause. cancelTimeout then finds the context already canceled and does nothing. If you swap the two defer lines, you lose the cause.

My post also shows a time.AfterFunc version. It is simpler, but the context has no deadline, so the next service gets none. I use the stack.

This is easy to get wrong at every call site. Part 5 puts it in a package.
-->
---
layout: default
kicker: Giving every cancel a cause
title: When ctx ends, return its cause
---

<Code gutter="15rem">

<div v-click>

```go
// Don't: return only what the driver says.
row := s.db.QueryRowContext(ctx, findOrderSQL, id)
if err := row.Scan(&o.ID, &o.Status); err != nil {
    return Order{}, fmt.Errorf("find order: %w", err)
}
```

</div>

<Mark :line="4" v-click="1">says what ended, not why</Mark>

<div v-click>

```go

// Do: when ctx has ended, return its cause.
if err := row.Scan(&o.ID, &o.Status); err != nil {
    if ctx.Err() != nil {
        err = context.Cause(ctx)
    }
    return Order{}, fmt.Errorf("find order: %w", err)
}
```

</div>

<Mark :block="2" :line="5" v-click>the cause that placeOrder set</Mark>
<Mark :block="2" :line="7" v-click>errors.Is still matches</Mark>

</Code>

<div class="points" style="--n: 2">
<v-clicks>

- In a `select` on `ctx.Done()`, return `context.Cause(ctx)` the same way
- With no cause set, `context.Cause(ctx)` returns `ctx.Err()`

</v-clicks>
</div>

<!--
A cause moves down the tree, from a context to its children. An error moves up the call stack. The handler that logs sits above placeOrder, so its own ctx never sees placeOrder's cause. The error has to carry the cause up.

[click] Here the store gets an error back from the driver. It says only that ctx ended.

[click] [click] If ctx has ended, return its cause instead. [click] The cause wraps DeadlineExceeded, so errors.Is still matches. The style guide says a function usually returns ctx.Err() when its context ends, and callers that check for it keep working. net/http does the same since Go 1.23.

[click] Use the same rule in the select from Part 2. [click] If nobody set a cause, Cause returns ctx.Err(), so you lose nothing.
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
  '    err=&quot;find order: placeOrder ord-123: 5s timeout: context deadline exceeded&quot;',
]" />

</div>

<!--
[click] If every layer logs and also returns, one failure prints many lines. The style guide says the same: if you return an error, let the caller log it. My post on stacked log lines has the details.

[click] Return the error up, and log it once in the handler.

[click] The error carries the cause, so this is the log line from the start of the talk, now with the reason: placeOrder's five-second timeout fired while the store waited on the query. A dashboard can still group these with errors.Is(err, context.DeadlineExceeded).
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
Values move down the same tree as the cancel signal, so they are easy to use. A value doesn't show in any function signature, so when one is missing, nothing fails at compile time.

[click] My test: if the function can't do its job without the value, the value goes in the signature. A trace ID passes the test, because nothing breaks without one. Auth data is the hard case. http.Handler has a fixed signature, so the middleware puts the user in the request context, and the handler reads it once and passes it on as an argument.

[click] A user ID that a query needs fails the test, and so do clients and config. Every value in a context has to be optional, but not every optional value belongs there.

[click] The package doc says the same thing from the other side: values are for request-scoped data, not optional parameters.
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
[click] Values are found by key, and any package can add a key. Here two packages pick the same string. Value returns the first equal key on the way up, so the second WithValue hides the first. Alice is still in the parent context, but nobody can reach her through this one. staticcheck flags built-in key types with SA1029.

[click] An unexported key type fixes this. [click] Only this package can write ctxKey, so no other package can make an equal key.

[click] The package exports a setter and a getter. Callers get a typed User and a bool, and never see the key. Auth middleware calls WithUser, and the handler calls UserFromContext once, then passes the user on as an argument. httptrace, pprof labels, and OpenTelemetry follow the same pattern.
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
Code review catches most of these bugs, but not all of them. A tool can check most of these rules.

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
[click] Stacking contexts and wrapping context.Canceled at every call site is easy to get wrong. So put it in a small package. [click] Its cancel function takes a cause, so callers have to give one. [click] A nil cause still gets a message. [click] It wraps context.Canceled, so errors.Is keeps working. The package's WithTimeout does the stack from Part 3.

[click] Then a forbidigo rule fails the build when someone calls the standard constructors, including the Cause versions. xctx itself is the one place with a nolint comment. The context.WithTimeout calls from Part 2 become xctx.WithTimeout. The package still returns a plain context.Context, so every library works as before.

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
