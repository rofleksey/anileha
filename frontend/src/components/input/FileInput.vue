<script setup>
import { ref, computed } from "vue";
import { notify } from "@kyvg/vue3-notification";
import prettyBytes from "pretty-bytes";

const props = defineProps({
  hint: {
    type: String,
    default: "Select file",
  },
  fileType: {
    type: String,
    default: "",
  },
  multiple: {
    type: Boolean,
    default: false,
  },
  maxFileSize: {
    type: Number,
    default: 5 * 1024 * 1024,
  },
});
const emit = defineEmits(["select"]);

const inputRef = ref(null);
const fileCount = ref(0);
const hintText = computed(() => {
  if (fileCount.value === 0) {
    return props.hint;
  } else {
    return `Selected ${fileCount.value} files`;
  }
});

function onChange() {
  console.log(inputRef.value.files);
  const files = [];
  for (const file of inputRef.value.files) {
    if (file.size > props.maxFileSize) {
      notify({
        title: "Failed to select files",
        text: `Files should be less than ${prettyBytes(props.maxFileSize)}`,
        type: "error",
      });
      return;
    }
    if (!file.type || !file.type.startsWith(props.fileType)) {
      notify({
        title: "Failed to select files",
        text: "Invalid file type",
        type: "error",
      });
      return;
    }
    files.push(file);
  }
  inputRef.value.value = null;
  fileCount.value = files.length;
  emit("select", files);
}
</script>

<template>
  <div class="file-input" @click="() => inputRef.click()">
    <span>{{ hintText }}</span>
    <input
      :multiple="multiple"
      class="actual-input"
      ref="inputRef"
      @change="onChange"
      type="file"
    />
  </div>
</template>

<style scoped>
.file-input {
  line-height: 1.5;
  color: white;
  border: none;
  display: block;
  background-color: rgba(100, 188, 255, 0.05);
  box-sizing: border-box;
  width: 100%;
  padding: 12px 16px;
  border-radius: 2px;
  resize: none;
  margin-top: 10px;
  margin-bottom: 10px;
  transition: background-color 0.15s ease;
}
.file-input:hover {
  cursor: pointer;
  background-color: rgba(100, 188, 255, 0.2);
}
.actual-input {
  display: none;
}
</style>
