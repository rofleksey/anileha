<script setup>
import { useListStore } from "../stores/list";
import { useUserStore } from "../stores/user";
import SearchBar from "../components/input/SearchBar.vue";
import TextList from "../components/info/TextList.vue";
import { onMounted, ref } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { getAllSeries } from "../api/api";
import AddIcon from "../components/AddIcon.vue";
import SeriesModal from "../components/modal/SeriesModal.vue";

const listStore = useListStore();
const userStore = useUserStore();
const seriesModal = ref(null);

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
    <div class="search-row">
      <SearchBar />
      <AddIcon v-if="userStore.isAdmin" @click="() => seriesModal.show()" />
    </div>
    <TextList :entries="listStore.entries" />
    <SeriesModal ref="seriesModal" />
  </div>
</template>

<style scoped>
.search {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  flex: none;
}
.search-row {
  display: flex;
  align-items: center;
}
</style>
