<template>
  <div class="interactive-overlay" @click="stopInterval"
       :style="{display: overlayVisible ? undefined : 'none'}">
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
import {onMounted, onUnmounted, ref} from 'vue';

let checkInterval: NodeJS.Timeout | number | undefined

const overlayVisible = ref(checkInteractive());

function checkInteractive() {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  return navigator.userActivation?.hasBeenActive ?? false;
}

function stopInterval() {
  overlayVisible.value = false;
  clearInterval(checkInterval);
}

onMounted(() => {
  checkInterval = setInterval(checkInteractive, 100);
});

onUnmounted(() => {
  stopInterval();
});
</script>

<style lang="sass" scoped>
.interactive-overlay
  position: fixed
  width: 100%
  height: 100%
  top: 0
  left: 0
  background: rgba(0, 0, 0, 0.5)
  display: flex
  align-items: center
  justify-content: center
  text-align: center
  color: white
  font-family: monospace
  font-size: 5em
  z-index: 10000
</style>
