<script setup>
import { useUserStore } from "../../stores/user";
import { ref } from "vue";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";
import TextInput from "../input/TextInput.vue";
import SquareButton from "../input/SquareButton.vue";
import BaseModal from "./BaseModal.vue";

const userRef = ref("");
const passRef = ref("");
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

const userStore = useUserStore();

const attemptLogin = () => {
  const user = userRef.value;
  const pass = passRef.value;
  if (user.trim().length === 0 || pass.trim().length === 0) {
    notify({
      title: "Failed to login",
      text: "Either username or password is blank",
      type: "error",
    });
    return;
  }
  axios
    .post("/user/login", {
      user,
      pass,
    })
    .then(({ data }) => {
      userStore.setUser(data.user, data.isAdmin);
      notify({
        title: "Login OK",
        type: "success",
      });
      baseModal.value.hide();
    })
    .catch((err) => {
      notify({
        title: "Failed to login",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
};
</script>

<template>
  <BaseModal title="Sign in" ref="baseModal" @submit="attemptLogin">
    <TextInput v-model="userRef" type="text" placeholder="Username" />
    <TextInput v-model="passRef" type="password" placeholder="Password" />
    <template #actions>
      <SquareButton @click="attemptLogin" text="sign in" />
    </template>
  </BaseModal>
</template>
