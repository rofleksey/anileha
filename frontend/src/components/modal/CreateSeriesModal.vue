<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Add series</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="seriesTitleRef"
          v-model="seriesTitle"
          label="Title"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-file
          ref="seriesIconRef"
          v-model="seriesIcon"
          label="Icon"
          accept="image/jpeg,image/png"
          max-file-size="8388608"
          :rules="[ val => val || 'Required' ]">
          <template v-slot:prepend>
            <q-icon name="attach_file"/>
          </template>
        </q-file>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
          flat
          round
          icon="add"
          @click="onOKClick"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {ref} from 'vue';
import {postNewSeries} from 'src/lib/post-api';
import {showError} from 'src/lib/util';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

defineEmits([
  ...useDialogPluginComponent.emits
])

const seriesTitleRef = ref<any>(null);
const seriesIconRef = ref<any>(null);

const postLoading = ref(false);
const seriesTitle = ref('');
const seriesIcon = ref<File | null>(null);

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  const icon = seriesIcon.value;
  if (!icon || !seriesTitleRef.value?.validate() || !seriesIconRef.value?.validate()) {
    return
  }
  postLoading.value = true;
  postNewSeries(seriesTitle.value, icon)
    .then(() => {
      onDialogOK();
    })
    .catch((e) => {
      showError('failed to add series', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
