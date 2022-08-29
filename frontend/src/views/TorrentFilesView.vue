<script setup>
import { useListStore } from "../stores/list";
import SearchBar from "../components/input/SearchBar.vue";
import TextList from "../components/info/TextList.vue";
import { onMounted, ref } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { useRoute } from "vue-router";
import { getTorrentFilesByTorrentId } from "../api/api";
import DownloadIcon from "../components/modal/icons/DownloadIcon.vue";
import ConvertIcon from "../components/modal/icons/ConvertIcon.vue";
import StartTorrentModal from "../components/modal/StartTorrentModal.vue";
import StartConversionModal from "../components/modal/StartConversionModal.vue";

const listStore = useListStore();
const startTorrentModal = ref(null);
const startConversionModal = ref(null);
const route = useRoute();

onMounted(() => {
  listStore.setData([]);
  getTorrentFilesByTorrentId(route.params.id)
    .then((data) => {
      listStore.setData(data);
    })
    .catch((err) => {
      notify({
        title: "Failed to get torrent files",
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
      <DownloadIcon @click="() => startTorrentModal.show()" />
      <ConvertIcon @click="() => startConversionModal.show()" />
    </div>
    <TextList :entries="listStore.entries" />
    <StartTorrentModal :torrent-id="route.params.id" ref="startTorrentModal" />
    <StartConversionModal
      :torrent-id="route.params.id"
      ref="startConversionModal"
    />
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
