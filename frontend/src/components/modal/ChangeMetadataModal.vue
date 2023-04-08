<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Change metadata</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="episodeRef"
          v-model="episode"
          label="Episode"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          v-model="season"
          label="Season"/>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          flat
          round
          icon="done"
          @click="onOKClick"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {onMounted, ref} from 'vue';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

defineEmits([
  ...useDialogPluginComponent.emits
])

const episodeRef = ref<any>(null);

interface Props {
  curEpisode: string;
  curSeason: string;
}

const props = defineProps<Props>()

const episode = ref('');
const season = ref('');

function onOKClick() {
  if (!episodeRef.value?.validate()) {
    return
  }
  onDialogOK({
    episode: episode.value,
    season: season.value
  });
}

onMounted(() => {
  episode.value = props.curEpisode;
  season.value = props.curSeason;
})
</script>

<style lang="sass" scoped>

</style>
