<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true,
  },
  type: {
    type: String,
    required: true,
  },
  placeholder: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update:modelValue"]);
</script>

<template>
  <input
    :value="modelValue"
    @input="(e) => emit('update:modelValue', e.target.value)"
    :type="type"
    :placeholder="placeholder"
  />
</template>

<style scoped>
input {
  font: inherit;
  line-height: 1.5;
  color: inherit;
  border: none;
  display: block;
  background: rgba(255, 255, 255, 0.05);
  box-sizing: border-box;
  width: 100%;
  padding: 12px 16px;
  border-radius: 2px;
  resize: none;
  margin-top: 10px;
  margin-bottom: 10px;
}

input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.075);
  -webkit-animation: flash 0.6s;
  animation: flash 0.6s;
}

@keyframes flash {
  0% {
    box-shadow: 0 0 4px 4px rgb(255 255 255 / 8%);
  }
}
</style>
