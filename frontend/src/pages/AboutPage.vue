<template>
  <div class="container">
    <h5>AniLeha</h5>
    <div class="text-body1">Made by rofleksey (roflex)</div>
    <div class="text-body1"><a class="text-grey-13" href="https://github.com/rofleksey/anileha">source</a></div>
    <div class="text-body1"><a class="text-grey-13" href="https://t.me/rofleksey">telegram</a></div>
    <img style="" :src="IAmJpg"/>
    <audio controls style="margin-top: 10px">
      <source src="/theme.mp3" type="audio/mpeg">
      Your browser does not support the audio element.
    </audio>
  </div>
</template>

<script setup>
import IAmJpg from 'assets/i_am.jpg';
</script>


<style lang="sass" scoped>
.container
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
</style>
