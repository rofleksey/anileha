<script setup>
import { RouterView } from "vue-router";
import Header from "./components/Header.vue";</script>

<template>
  <section class="main-column">
    <Header />
    <main>
      <RouterView v-slot="{ Component, route }">
        <Transition name="fade" mode="out-in">
          <component :is="Component" :key="route.path" />
        </Transition>
      </RouterView>
    </main>
    <footer class="main-footer"></footer>
  </section>
  <notifications classes="notification" />
</template>

<style scoped>
.main-column {
  position: relative;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 1rem);
  width: 600px;
  margin: 0 auto;
}

main {
  width: 100%;
  margin: 2rem 0;
  overflow-y: scroll;
}

/* do not remove, used for transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
