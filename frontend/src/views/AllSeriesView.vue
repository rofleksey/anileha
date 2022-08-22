<script setup>
import { useSeriesStore } from "../stores/series";
import SearchBar from "@/components/SearchBar.vue";
import TextList from "@/components/TextList.vue";
import axios from "axios";
import { onMounted } from "vue";
import { format } from "timeago.js";
import { notify } from "@kyvg/vue3-notification";

const series = useSeriesStore();

onMounted(() => {
  axios("http://localhost:5000/series")
    .then(({ data }) => {
      const seriesData = data.map((series) => ({
        id: series.id,
        title: series.name,
        link: `/s/${series.id}`,
        details: [
          {
            id: "updated_at",
            text: format(new Date(series.updatedAt))
          }
        ]
      }));
      series.setData(seriesData);
    })
    .catch(() => {
      notify({
        text: "Failed to get series",
        type: "error"
      });
    });
});
</script>

<template>
  <div class="search">
    <SearchBar />
    <TextList :entries="series.entries" />
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
