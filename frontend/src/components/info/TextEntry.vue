<script setup>
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useUserStore } from "../../stores/user";

const props = defineProps({
  entry: {
    type: Object,
    required: true,
  },
});

const userStore = useUserStore();

const accessDetails = computed(() => {
  return props.entry.details.filter((it) => !it.admin || userStore.isAdmin);
});
</script>

<template>
  <div
    :class="{ entry: true, withBg: entry.bg }"
    :style="entry.bg ? { backgroundImage: `url(${entry.bg})` } : {}"
  >
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
  margin-top: 8px;
  margin-bottom: 8px;
  border-bottom: 1px solid hsla(0, 0%, 100%, 0.05);
  padding: 1em;
}

.entry.withBg {
  background-blend-mode: darken;
  background-color: rgba(0, 0, 0, 0.75);
  background-position: center center;
  background-repeat: no-repeat;
  background-size: cover;
}

.entry > * {
  padding: 2px;
}

.title {
  font-size: 18px;
}

.subtext {
  font-size: 13px;
  color: hsla(0, 0%, 100%, 0.5);
}

.subtitle.interactive {
  font-size: 13px;
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
