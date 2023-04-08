<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Pick audio stream</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-select
          outlined
          option-value="index"
          option-label="label"
          v-model="model"
          :options="data"
          map-options
          label="Stream"/>
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
import {computed, onMounted, ref} from 'vue';
import {BaseStream} from 'src/lib/api-types';
import prettyBytes from 'pretty-bytes';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  streams: BaseStream[];
  curIndex: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

interface StreamEntry {
  index: number;
  label: string;
}

const model = ref<StreamEntry | undefined>();

const data = computed(() => {
  return props.streams.map((stream) => {
    let name = stream.name.trim();
    if (name.length === 0) {
      name = '<blank>';
    }

    const prefix = `#${stream.index} - ${name}`;

    name = `${prefix} (${prettyBytes(stream.size)})`;

    return {
      index: stream.index,
      label: name,
    }
  })
})

onMounted(() => {
  model.value = data.value.find((it) => it.index === props.curIndex);
})

function onOKClick() {
  onDialogOK({
    stream: model.value?.index
  });
}
</script>

<style lang="sass" scoped>

</style>
