<script setup>
import { RouterLink } from "vue-router";

defineProps({
  entry: {
    type: Object,
    required: true
  }
});
</script>

<template>
  <RouterLink :to="entry.link">
    <div class="entry">
      <div class="title">
        <span>{{ entry.title }}</span>
      </div>
      <div class="subtext">
        <template v-for="(detail, index) in entry.details" :key="detail.id">
          <span class="subtitle">
            {{ detail.text }}
          </span>

          <span v-if="index !== entry.details.length - 1" class="delimiter">
            •
          </span>
        </template>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
.entry {
  overflow: hidden;
  margin: 0;
  border-bottom: 1px solid hsla(0, 0%, 100%, 0.05);
}

.entry:hover {
  cursor: pointer;
}

.entry {
  padding: 1em;
}

.entry > * {
  padding: 2px;
}

.title {
  font-size: 20px;
}

.subtext {
  font-size: 14px;
  color: hsla(0, 0%, 100%, 0.5);
}

.title,
.subtext {
  transition: color 0.15s;
}

.entry:hover .title,
.entry:hover .subtext {
  color: white;
}
</style>
