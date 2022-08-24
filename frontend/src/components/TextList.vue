<script setup>
import TextEntry from "@/components/TextEntry.vue";

defineProps({
  entries: {
    type: Array,
    required: true,
  },
});
</script>

<template>
  <nav class="text">
    <div class="list">
      <TextEntry v-for="entry in entries" :entry="entry" :key="entry.id" />
    </div>
  </nav>
</template>

<style scoped>
.text {
  margin-top: 0 !important;
  text-align: left;
}
</style>
