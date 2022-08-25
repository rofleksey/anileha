<script setup>
import { ref, toRaw } from "vue";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";
import TextInput from "../input/TextInput.vue";
import SquareButton from "../input/SquareButton.vue";
import BaseModal from "./BaseModal.vue";
import FileInput from "../input/FileInput.vue";

const nameRef = ref("");
const filesRef = ref([]);
const baseModal = ref(null);

function show() {
  baseModal.value.show();
}

function hide() {
  baseModal.value.hide();
}

defineExpose({
  show,
  hide,
});

const createSeries = () => {
  const name = nameRef.value;
  if (name.trim().length === 0) {
    notify({
      title: "Failed to create series",
      text: "Name is blank",
      type: "error",
    });
    return;
  }
  if (filesRef.value.length === 0) {
    notify({
      title: "Failed to create series",
      text: "Thumbnail is not selected",
      type: "error",
    });
    return;
  }
  const unwrappedFiles = filesRef.value.map((proxy) => toRaw(proxy));
  const formData = new FormData();
  formData.append("name", name);
  formData.append("thumb", unwrappedFiles[0]);
  axios({
    method: "post",
    url: "/admin/series",
    data: formData,
    headers: { "Content-Type": "multipart/form-data" },
  })
    .then(() => {
      notify({
        title: "Created",
        type: "success",
      });
      baseModal.value.hide();
    })
    .catch((err) => {
      notify({
        title: "Failed to create series",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
};
</script>

<template>
  <BaseModal title="New Series" ref="baseModal" @submit="createSeries">
    <FileInput
      hint="Select thumbnail"
      type="image"
      @select="(val) => (filesRef = val)"
    />
    <TextInput v-model="nameRef" type="text" placeholder="Name" />
    <template #actions>
      <SquareButton @click="createSeries" text="create" />
    </template>
  </BaseModal>
</template>
