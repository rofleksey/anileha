<script setup>
import { RouterView } from "vue-router";
import Header from "./components/Header.vue";
</script>

<template>
  <section class="main-column">
    <Header />
    <main>
      <RouterView />
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
}
</style>
