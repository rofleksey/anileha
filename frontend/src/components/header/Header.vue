<script setup>
import Logo from "./Logo.vue";
import SeriesIcon from "./icons/SeriesIcon.vue";
import TorrentsIcon from "./icons/TorrentsIcon.vue";
import ConversionsIcon from "./icons/ConversionsIcon.vue";
import EpisodesIcon from "./icons/EpisodesIcon.vue";
import LoginIcon from "./icons/LoginIcon.vue";
import LogoutIcon from "./icons/LogoutIcon.vue";
import { useRoute } from "vue-router/dist/vue-router";
import { onMounted, ref } from "vue";
import LoginModal from "../modal/LoginModal.vue";
import { useUserStore } from "../../stores/user";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";

const modal = ref(null);
const route = useRoute();
const userStore = useUserStore();

onMounted(() => {
  axios("/user/me")
    .then(({ data }) => {
      console.log(data);
      userStore.setUser(data.user, data.isAdmin);
    })
    .catch(() => {
      userStore.logout();
    });
});

function logout() {
  if (window.confirm("Do you really want to logout?")) {
    axios
      .post("/user/logout")
      .then(() => {
        userStore.logout();
        notify({
          title: "Logged out",
          type: "success",
        });
      })
      .catch((err) => {
        notify({
          title: "Failed to logout",
          text: err?.response?.data?.error ?? "",
          type: "error",
        });
      });
  }
}
</script>

<template>
  <div class="header">
    <Logo />
    <RouterLink to="/">
      <SeriesIcon :selected="route.path === '/'" />
    </RouterLink>
    <RouterLink to="/torrents">
      <TorrentsIcon
        v-if="userStore.isAdmin"
        :selected="route.path.startsWith('/torrents')"
      />
    </RouterLink>
    <RouterLink to="/convert">
      <ConversionsIcon
        v-if="userStore.isAdmin"
        :selected="route.path.startsWith('/convert')"
      />
    </RouterLink>
    <RouterLink to="/episodes">
      <EpisodesIcon :selected="route.path.startsWith('/episodes')" />
    </RouterLink>
    <LoginIcon
      @click="() => modal.show()"
      v-if="userStore.user === null"
      :selected="false"
    />
    <LogoutIcon
      @click="logout"
      v-if="userStore.user !== null"
      :selected="false"
    />
    <LoginModal ref="modal" />
  </div>
</template>

<style scoped>
.header {
  padding-top: 2rem;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
