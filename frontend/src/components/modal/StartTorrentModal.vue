<script setup>
import { ref } from "vue";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";
import SquareButton from "../input/SquareButton.vue";
import TextInput from "../input/TextInput.vue";
import BaseModal from "./BaseModal.vue";

const filesRef = ref("");
const baseModal = ref(null);

function show() {
  baseModal.value.show();
}

function hide() {
  baseModal.value.hide();
}

const props = defineProps({
  torrentId: {
    type: String,
    required: true,
  },
});

defineExpose({
  show,
  hide,
});

const startTorrent = () => {
  const files = filesRef.value.trim();
  if (files.length === 0) {
    notify({
      title: "Failed to start torrent",
      text: "Invalid files string",
      type: "error",
    });
    return;
  }
  axios
    .post("/admin/torrent/start", {
      torrentId: props.torrentId,
      fileIndices: files,
    })
    .then(() => {
      notify({
        title: "Started",
        type: "success",
      });
      baseModal.value.hide();
    })
    .catch((err) => {
      notify({
        title: "Failed to start torrent",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
};
</script>

<template>
  <BaseModal title="Start torrent" ref="baseModal" @submit="startTorrent">
    <TextInput v-model="filesRef" type="text" placeholder="Files" />
    <template #actions>
      <SquareButton @click="startTorrent" text="start" />
    </template>
  </BaseModal>
</template>
