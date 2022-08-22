<script setup>
import Logo from "@/components/icons/Logo.vue";
import SeriesIcon from "@/components/icons/SeriesIcon.vue";
import TorrentsIcon from "@/components/icons/TorrentsIcon.vue";
import ConversionsIcon from "@/components/icons/ConversionsIcon.vue";
import EpisodesIcon from "@/components/icons/EpisodesIcon.vue";
import LoginIcon from "@/components/icons/LoginIcon.vue";
import LogoutIcon from "@/components/icons/LogoutIcon.vue";
import { useRoute } from "vue-router/dist/vue-router";
import { onMounted, ref } from "vue";
import LoginModal from "@/components/LoginModal.vue";
import { useUserStore } from "../stores/user";
import axios from "axios";

const modalOpen = ref(false);

const route = useRoute();
const userStore = useUserStore();

onMounted(() => {
  axios("http://localhost:5000/user/me")
    .then(({ data }) => {
      userStore.setUser(data);
    })
    .catch(() => {
      userStore.setUser(null);
    });
});
</script>

<template>
  <div class="header">
    <Logo />
    <RouterLink to="/">
      <SeriesIcon :selected="route.path === '/'" />
    </RouterLink>
    <RouterLink to="/t">
      <TorrentsIcon v-if="userStore.user === 'admin'" :selected="false" />
    </RouterLink>
    <RouterLink to="/c">
      <ConversionsIcon v-if="userStore.user === 'admin'" :selected="false" />
    </RouterLink>
    <RouterLink to="/e">
      <EpisodesIcon
        :selected="route.path.startsWith('/s/') || route.path.startsWith('/e')"
      />
    </RouterLink>
    <LoginIcon
      @click="modalOpen = true"
      v-if="userStore.user === null"
      :selected="false"
    />
    <LogoutIcon v-if="userStore.user !== null" :selected="false" />
    <LoginModal v-model="modalOpen" />
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
