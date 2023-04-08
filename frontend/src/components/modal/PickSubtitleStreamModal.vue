<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Pick subtitle stream</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-select
          outlined
          option-value="id"
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
import {computed, ComputedRef, onMounted, ref} from 'vue';
import {SubStream} from 'src/lib/api-types';
import prettyBytes from 'pretty-bytes';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  streams: SubStream[];
  curIndex: number | undefined;
  files: string[];
  curFile: string | undefined;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

interface StreamEntry {
  id: number | string;
  label: string;
}

const model = ref<StreamEntry | undefined>();

const data: ComputedRef<StreamEntry[]> = computed(() => {
  const integratedStreams: StreamEntry[] = props.streams.map((stream) => {
    let name = stream.name.trim();
    if (name.length === 0) {
      name = '<blank>';
    }

    const prefix = `#${stream.index} - ${name}`;

    if (stream.textLength >= 32) {
      name = `${prefix} (${Math.ceil(stream.textLength / 1000)}k chars)`
    } else {
      name = `${prefix} (${prettyBytes(stream.size)})`;
    }

    return {
      id: stream.index,
      label: name,
    }
  });

  const fileStreams: StreamEntry[] = props.files.map((fileName) => {
    return {
      id: fileName,
      label: `file: ${fileName}`,
    }
  });

  return [...integratedStreams, ...fileStreams];
})

onMounted(() => {
  if (props.curIndex !== undefined) {
    model.value = data.value.find((it) => it.id === props.curIndex);
  } else if (props.curFile !== undefined) {
    model.value = data.value.find((it) => it.id === props.curFile);
  }
})

function onOKClick() {
  const curValue = model.value;
  if (!curValue) {
    return;
  }
  const id = curValue.id;
  if (typeof id === 'string') {
    onDialogOK({
      file: id,
    });
  } else {
    onDialogOK({
      stream: id,
    });
  }
}
</script>

<style lang="sass" scoped>

</style>
