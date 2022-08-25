<script setup>
import { useListStore } from "../stores/list";
import SearchBar from "@/components/SearchBar.vue";
import TextList from "@/components/TextList.vue";
import { onMounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { getAllSeries } from "../api/api";

const listStore = useListStore();

onMounted(() => {
  listStore.setData([]);
  getAllSeries()
    .then((data) => {
      listStore.setData(data);
    })
    .catch((err) => {
      notify({
        title: "Failed to get series",
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
