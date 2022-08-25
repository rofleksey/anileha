<script setup>
import { useListStore } from "../stores/list";
import SearchBar from "../components/input/SearchBar.vue";
import TextList from "../components/info/TextList.vue";
import { onMounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { getAllTorrents } from "../api/api";

const listStore = useListStore();

onMounted(() => {
  listStore.setData([]);
  getAllTorrents()
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
