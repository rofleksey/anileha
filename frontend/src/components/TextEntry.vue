<script setup>
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useUserStore } from "../stores/user";

const props = defineProps({
  entry: {
    type: Object,
    required: true
  }
});

const userStore = useUserStore();

const accessDetails = computed(() => {
  return props.entry.details.filter((it) => !it.admin || userStore.isAdmin);
});
</script>

<template>
  <div class="entry">
    <RouterLink :to="entry.link">
      <div class="title">
        <span>{{ entry.title }}</span>
      </div>
    </RouterLink>
    <div class="subtext">
      <template v-for="(detail, index) in accessDetails" :key="detail.id">
        <RouterLink v-if="detail.link" :to="detail.link">
          <span class="subtitle interactive">
            {{ detail.text }}
          </span>
        </RouterLink>
        <span
          v-else-if="detail.onclick"
          @click="detail.onclick"
          class="subtitle interactive"
        >
          {{ detail.text }}
        </span>
        <span v-else class="subtitle">
          {{ detail.text }}
        </span>
        <span v-if="index !== accessDetails.length - 1" class="delimiter">
          •
        </span>
      </template>
    </div>
  </div>
</template>

<style scoped>
.entry {
  overflow: hidden;
  margin: 0;
  border-bottom: 1px solid hsla(0, 0%, 100%, 0.05);
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

.subtitle.interactive {
  font-size: 15px;
  color: #64bcffaa;
}

.title,
.subtext,
.subtitle {
  transition: color 0.15s;
}

.title:hover {
  color: white;
  cursor: pointer;
}

.subtitle.interactive:hover {
  color: #64bcff;
  cursor: pointer;
}
</style>
