<script setup>
// In-process vs. distributed context propagation.
// Two services side by side. Inside each, ctx moves from function to function.
// Between them, the payments client makes a gRPC call and the context
// crosses as headers.
</script>

<template>
  <div class="tl2">
    <div class="svc">
      <div class="svc-name">order service</div>
      <div class="row">
        <div class="node">HTTP handler</div>
        <div class="h"><span>ctx</span></div>
        <div class="node">placeOrder(ctx)</div>
        <div class="h"><span>ctx</span></div>
        <div class="node">payments client</div>
      </div>
    </div>

    <div class="wire">
      <div class="wire-top">gRPC call</div>
      <div class="wire-line" />
      <div class="wire-bot">traceparent<br>grpc-timeout</div>
    </div>

    <div class="svc">
      <div class="svc-name">payment service</div>
      <div class="row">
        <div class="node">gRPC handler</div>
        <div class="h"><span>ctx</span></div>
        <div class="node">charge(ctx)</div>
        <div class="h"><span>ctx</span></div>
        <div class="node">PSP client</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tl2 { display: grid; grid-template-columns: 1fr 7.5rem 1fr; align-items: center; width: 100%; font-family: var(--font-mono); margin-top: 1.5rem; }
.svc { align-self: stretch; border: 1.5px dashed var(--line); border-radius: var(--radius); padding: 0.8rem 0.8rem 1.1rem; }
.svc-name { color: var(--fg-dim); font: 700 0.72rem/1 var(--font-body); letter-spacing: 0.1em; text-transform: uppercase; margin-bottom: 1rem; }
.row { display: grid; grid-template-columns: 1fr 2.2rem 1fr 2.2rem 1fr; align-items: center; }
.node {
  text-align: center; padding: 0.8rem 0.3rem; font-size: 0.74rem; white-space: nowrap; line-height: 1.25; color: var(--fg);
  border: 1.5px solid var(--accent); border-radius: var(--radius-sm); background: var(--code-bg);
}
.h { position: relative; height: 2px; background: var(--accent); margin: 0 0.25rem; }
.h::after { content: ""; position: absolute; right: -2px; top: -5px; border: 6px solid transparent; border-left: 8px solid var(--accent); border-right: 0; }
.h span { position: absolute; left: 50%; top: -1.25rem; transform: translateX(-50%); font-size: 0.68rem; color: var(--accent); }
.wire { position: relative; display: flex; flex-direction: column; align-items: center; padding-top: 1.7rem; }
.wire-line { position: relative; width: 100%; height: 2px; background: var(--cat-2); margin: 0.45rem 0; }
.wire-line::after { content: ""; position: absolute; right: -2px; top: -5px; border: 6px solid transparent; border-left: 9px solid var(--cat-2); border-right: 0; }
.wire-top { font: 600 0.78rem/1 var(--font-body); color: var(--cat-2); }
.wire-bot { font-size: 0.7rem; line-height: 1.35; color: var(--cat-2); text-align: center; }
</style>
