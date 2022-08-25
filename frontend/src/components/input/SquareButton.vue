<script setup>
defineProps({
  text: {
    type: String,
    required: true,
  },
});
const emit = defineEmits(["pressed"]);
</script>

<template>
  <button @click="() => emit('pressed')" rel="closeBtn" class="close">
    {{ text }}
  </button>
</template>

<style scoped>
button {
  font: inherit;
  color: inherit;
  background-color: rgba(255, 255, 255, 0.05);
  border: 1px solid hsla(0, 100%, 100%, 0);
  padding: 0.5em 1em;
  border-radius: 0.25em;
  box-shadow: inset 0 0 0 0 hsl(0deg 0% 100% / 5%);
  transition: box-shadow 0.3s, border 0.15s ease-in;
  outline: none;
  min-height: 2em;
  overflow: hidden;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  cursor: default;
}

button:hover {
  border: 1px solid hsla(0, 100%, 100%, 0.075);
  -webkit-animation: none;
  animation: none;
  cursor: pointer;
}
</style>
