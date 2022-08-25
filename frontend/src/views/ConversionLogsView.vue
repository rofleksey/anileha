<script setup>
import { useTextStore } from "../stores/text";
import axios from "axios";
import { onMounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { useRoute } from "vue-router/dist/vue-router";
import TextBlock from "../components/info/TextBlock.vue";

const textStore = useTextStore();
const route = useRoute();

onMounted(() => {
  textStore.setText("");
  axios(`/admin/convert/${route.params.id}/logs`)
    .then(({ data }) => {
      textStore.setText(data);
    })
    .catch((err) => {
      notify({
        title: "Failed to get conversion logs",
        text: err?.response?.data?.error ?? "",
        type: "error"
      });
    });
});
</script>

<template>
  <div class="text-block">
    <TextBlock :text="textStore.text" />
  </div>
</template>

<style scoped>
.text-block {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  flex: none;
}
</style>
