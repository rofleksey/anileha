<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true,
  },
  text: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update:modelValue"]);
</script>

<template>
  <div
    :class="{ checkbox: true, checked: modelValue }"
    @click="() => emit('update:modelValue', !modelValue)"
  >
    <svg data-v-21fba295="" viewbox="0 0 16 16" class="icon">
      <polygon data-v-21fba295="" points="4 4 4 12 12 12 12 4"></polygon>
    </svg>
    <div class="label">{{ text }}</div>
  </div>
</template>

<style scoped>
.checkbox {
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 1em;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
  transition: background 0.6s;
}
.checkbox:active {
  background: rgba(255, 255, 255, 0.05);
  transition: background 0s;
}
.checkbox:hover {
  color: white;
}
.checkbox svg {
  width: 16px;
  height: 16px;
  fill: white;
  flex: none;
  background: rgba(255, 255, 255, 0.05);
  box-shadow: inset 0 0 0 0 hsl(0deg 0% 100% / 5%);
  border-radius: 2px;
  overflow: hidden;
  margin-right: 8px;
  transition: fill 0.1s, background 0.6s, box-shadow 0.3s;
}
.checkbox:active svg {
  box-shadow: inset 0 0 0 0 hsl(0deg 0% 100% / 5%),
    0 0 0 0.25em hsl(0deg 0% 100% / 5%);
  background: rgba(255, 255, 255, 0.1);
  transition: fill 0.1s, background 0s, box-shadow 0s;
}
.checkbox:not(.checked) svg {
  fill: rgba(255, 255, 255, 0.2);
}
.label {
  transition: color 0.15s ease;
}
</style>
