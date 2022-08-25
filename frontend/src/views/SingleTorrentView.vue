<script setup>
import { useJsonStore } from "../stores/json";
import axios from "axios";
import { onMounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { useRoute } from "vue-router/dist/vue-router";
import JsonTree from "@/components/JsonTree.vue";

const jsonStore = useJsonStore();
const route = useRoute();

onMounted(() => {
  jsonStore.setData({});
  axios(`/admin/torrent/${route.params.id}`)
    .then(({ data }) => {
      jsonStore.setData(data);
    })
    .catch((err) => {
      notify({
        title: "Failed to get torrent",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
});
</script>

<template>
  <div class="json">
    <JsonTree :data="jsonStore.data" />
  </div>
</template>

<style scoped>
.json {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  flex: none;
}
</style>
