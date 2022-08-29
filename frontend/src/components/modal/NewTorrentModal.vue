<script setup>
import { ref, toRaw } from "vue";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";
import SquareButton from "../input/SquareButton.vue";
import BaseModal from "./BaseModal.vue";
import FileInput from "../input/FileInput.vue";
import CheckBoxInput from "../input/CheckBoxInput.vue";

const autoRef = ref(true);
const filesRef = ref([]);
const baseModal = ref(null);

function show() {
  baseModal.value.show();
}

function hide() {
  baseModal.value.hide();
}

const props = defineProps({
  seriesId: {
    type: Number,
    required: true,
  },
});

defineExpose({
  show,
  hide,
});

const addTorrent = () => {
  if (filesRef.value.length === 0) {
    notify({
      title: "Failed to add torrent",
      text: "Files are not selected",
      type: "error",
    });
    return;
  }
  const unwrappedFiles = filesRef.value.map((proxy) => toRaw(proxy));
  const formData = new FormData();
  formData.append("seriesId", props.seriesId);
  formData.append("auto", autoRef.value);
  unwrappedFiles.forEach((file) => {
    formData.append(`files`, file);
  });
  axios({
    method: "post",
    url: "/admin/torrent",
    data: formData,
    headers: { "Content-Type": "multipart/form-data" },
  })
    .then(() => {
      notify({
        title: "Added",
        type: "success",
      });
      baseModal.value.hide();
    })
    .catch((err) => {
      notify({
        title: "Failed to add torrent",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
};
</script>

<template>
  <BaseModal title="Add torrent" ref="baseModal" @submit="addTorrent">
    <FileInput
      hint="Select torrent files"
      file-type="application/x-bittorrent"
      :multiple="true"
      @select="(val) => (filesRef = val)"
    />
    <CheckBoxInput v-model="autoRef" text="Auto download/convert" />
    <template #actions>
      <SquareButton @click="addTorrent" text="add" />
    </template>
  </BaseModal>
</template>
