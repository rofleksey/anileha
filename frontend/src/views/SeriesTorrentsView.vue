<script setup>
import { useListStore } from "../stores/list";
import SearchBar from "../components/input/SearchBar.vue";
import TextList from "../components/info/TextList.vue";
import { onMounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { useRoute } from "vue-router/dist/vue-router";
import { getTorrentsBySeriesId } from "../api/api";

const listStore = useListStore();
const route = useRoute();

onMounted(() => {
  listStore.setData([]);
  getTorrentsBySeriesId(route.params.id)
    .then((data) => {
      listStore.setData(data);
    })
    .catch((err) => {
      notify({
        title: "Failed to get torrents",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
});
</script>

<template>
  <div class="search">
    <SearchBar />
    <TextList :entries="listStore.entries" />
  </div>
</template>

<style scoped>
.search {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  flex: none;
}
</style>
