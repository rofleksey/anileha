<script setup>
import SquareButton from "../input/SquareButton.vue";
import { ref } from "vue";

const shown = ref(false);

function show() {
  shown.value = true;
}

function hide() {
  shown.value = false;
}

defineExpose({
  show,
  hide,
});

defineProps({
  title: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["submit"]);
</script>

<template>
  <Transition name="fade">
    <div
      v-if="shown"
      @keydown.enter="() => emit('submit')"
      @keydown.esc="() => (shown = false)"
      class="dialog-wrapper"
    >
      <div class="dialog" v-click-outside="() => (shown = false)">
        <div class="dialog-window-bar">
          <h2>{{ title }}</h2>
          <SquareButton
            @pressed="() => (shown = false)"
            rel="closeBtn"
            class="close"
            text="×"
          />
        </div>
        <div class="dialog-content scroll-container">
          <div class="form">
            <slot />
            <div class="actions">
              <slot name="actions" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.dialog-wrapper {
  z-index: 100;
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  top: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  margin: 0;
}

/* dont remove, it's used by Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.dialog {
  width: 300px;
  flex: none;
  padding: 1rem;
  background: #1c1f22;
  border-radius: 2px;
  display: flex;
  flex-direction: column;
  margin: 0 auto;
}

.dialog-window-bar {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.dialog-content {
  max-height: 100%;
  padding-right: 5px;
}

h2 {
  font-size: 1.2rem;
  font-weight: 300;
}

.dialog-window-bar button {
  margin-left: auto;
}

.actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}
</style>
