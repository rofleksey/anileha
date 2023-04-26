<template>
  <div class="container">
    <h1>Unauthorized</h1>
    <h6>Log in first to access this page</h6>
    <video muted autoplay loop :src="UnauthorizedWebm"></video>
  </div>
</template>

<script setup>
import UnauthorizedWebm from 'assets/unauthorized.webm';
</script>


<style lang="sass" scoped>
.container
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
</style>
