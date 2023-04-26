<template>
  <div class="container">
    <h1>Forbidden</h1>
    <h6>You don't have permissions to access this page</h6>
    <video muted autoplay loop :src="ForbiddenWebm"></video>
  </div>
</template>

<script setup>
import ForbiddenWebm from 'assets/forbidden.webm';
</script>


<style lang="sass" scoped>
.container
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
</style>
