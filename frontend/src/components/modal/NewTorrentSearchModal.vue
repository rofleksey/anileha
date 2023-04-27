<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card" style="width: 700px">
      <q-card-section>
        <div class="text-h6">Search new torrent</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-stepper
          v-model="step"
          header-nav
          ref="stepper"
          color="primary"
          style="width: 100%"
          animated
        >
          <q-step
            :name="1"
            title="Select torrents"
            icon="search"
            :done="step > 1"
            :header-nav="step > 1"
          >
            <q-input
              v-model="searchQuery"
              :loading="postLoading"
              debounce="1000"
              filled
              placeholder="Search"
            >
              <template v-slot:append>
                <q-icon name="search"/>
              </template>
            </q-input>
            <q-list>
              <template v-for="item in searchResults" :key="item.id">
                <q-item clickable v-ripple @click="onSelectItem(item)">
                  <q-item-section>
                    <q-item-label>{{ item.title }}</q-item-label>
                    <q-item-label caption>{{ item.size }}</q-item-label>
                  </q-item-section>

                  <q-item-section side top>
                    <q-item-label caption>{{ item.date }}</q-item-label>
                    <q-item-label caption>{{ item.seeders }} seeders</q-item-label>
                  </q-item-section>
                </q-item>

                <q-separator spaced inset/>
              </template>
            </q-list>
          </q-step>

          <q-step
            :name="2"
            title="Submit"
            icon="description"
            :done="step > 2"
            :header-nav="step > 2"
          >
            <q-toggle
              v-model="auto"
              label="Automatic Torrent"
            />
            <q-input
              ref="audioRef"
              v-model="audioLang"
              label="Audio Language (e.g. jpn)"
              :disable="!auto"
              :rules="auto ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>
            <q-input
              ref="subRef"
              v-model="subLang"
              label="Subtitle Language (e.g. eng)"
              :disable="!auto"
              :rules="auto ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>

            <q-stepper-navigation>
              <q-btn
                :disable="postLoading"
                @click="onOKClick"
                :loading="postLoading"
                color="primary"
                label="Submit"/>
            </q-stepper-navigation>
          </q-step>
        </q-stepper>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {ref, watch} from 'vue';
import {postNewTorrentFromSearch, postSearchTorrents} from 'src/lib/post-api';
import {showError} from 'src/lib/util';
import {AutoTorrent, SearchResult} from 'src/lib/api-types';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  seriesId: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const audioRef = ref<any>(null);
const subRef = ref<any>(null);

const postLoading = ref(false);
const searchResults = ref<SearchResult[]>([]);
const searchQuery = ref('');
const selectedItem = ref<SearchResult | undefined>();
const step = ref(1);
const auto = ref(false);
const audioLang = ref('jpn');
const subLang = ref('eng');

watch(searchQuery, () => {
  postLoading.value = true;
  postSearchTorrents(searchQuery.value, 0)
    .then((results) => {
      searchResults.value = results;
    })
    .catch((e) => {
      showError('Failed search torrents', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
})

function onSelectItem(item: SearchResult) {
  selectedItem.value = item;
  step.value = 2;
  console.log(item);
  console.log(step.value);
}

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  const itemValue = selectedItem.value;
  if (!itemValue || !audioRef.value?.validate() || !subRef.value?.validate()) {
    return
  }
  let autoTorrent: AutoTorrent | undefined;
  if (auto.value) {
    autoTorrent = {
      audioLang: audioLang.value,
      subLang: subLang.value,
    }
  }
  postLoading.value = true;
  postNewTorrentFromSearch(props.seriesId, itemValue.id, itemValue.provider, autoTorrent)
    .then(() => {
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to add torrent', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
