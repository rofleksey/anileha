<script setup>
import { useEpisodesStore } from "../stores/episodes";
import SearchBar from "@/components/SearchBar.vue";
import TextList from "@/components/TextList.vue";
import axios from "axios";
import { onMounted } from "vue";
import { format as timeAgoFormat } from "timeago.js";
import { notify } from "@kyvg/vue3-notification";
import prettyBytes from "pretty-bytes";
import durationFormat from "format-duration";
import { useRoute } from "vue-router";

const episodes = useEpisodesStore();
const route = useRoute();

onMounted(() => {
  axios(`http://localhost:5000/series/${route.params.id}/episodes`)
    .then(({ data }) => {
      const episodesData = data.map((ep) => ({
        id: ep.id,
        title: ep.name,
        link: "/",
        details: [
          {
            id: "created_at",
            text: timeAgoFormat(new Date(ep.createdAt))
          },
          {
            id: "duration",
            text: durationFormat(ep.durationSec * 1000)
          },
          {
            id: "length",
            text: prettyBytes(ep.length)
          }
        ]
      }));
      episodes.setData(episodesData);
    })
    .catch(() => {
      notify({
        text: "Failed to get episodes",
        type: "error"
      });
    });
});
</script>

<template>
  <div class="search">
    <SearchBar />
    <TextList :entries="episodes.entries" />
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
