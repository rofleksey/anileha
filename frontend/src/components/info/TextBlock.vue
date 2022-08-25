<script setup>
defineProps({
  text: {
    type: String,
    required: true,
  },
});
</script>

<template>
  <div class="text">
    {{ text }}
  </div>
</template>

<style scoped>
.text {
  background: #282c34;
  color: white;
  font-family: monospace;
  font-size: 13px;
  white-space: pre-wrap;
}
</style>
