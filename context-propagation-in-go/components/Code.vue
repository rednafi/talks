<script setup>
// A code panel. Put one or more fenced blocks inside; wrap later blocks in
// <div v-click> to reveal the code part by part. The blocks join into one
// panel. `gutter` reserves room on the right for <Mark> labels.
defineProps({
  file: { type: String, default: '' },
  gutter: { type: String, default: '0px' },
})
</script>

<template>
  <div class="codebox" :style="{ '--gutter': gutter }">
    <div v-if="file" class="codebox-file">{{ file }}</div>
    <div class="codebox-body">
      <slot />
    </div>
  </div>
</template>
