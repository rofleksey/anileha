<script setup>
import { JsonViewer } from "vue3-json-viewer";
import "vue3-json-viewer/dist/index.css";
defineProps({
  data: {
    type: Object,
    required: true,
  },
});
</script>

<template>
  <JsonViewer theme="dark" :value="data" sort expanded :expand-depth="3" />
</template>

<style lang="scss" scoped></style>
