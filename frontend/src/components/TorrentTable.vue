<template>
  <q-table
    @row-click="onRowClick"
    style="width: 100%"
    :title="props.title"
    :rows="props.data"
    :columns="columns"
    row-key="id"
    :loading="props.loading">
    <template v-slot:body-cell-status="props">
      <q-td :props="props">
        <q-icon v-if="props.value === 'idle'" class="text-orange" name="stop" size="2rem"/>
        <q-icon v-else-if="props.value === 'error'" class="text-red" name="error" size="2rem"/>
        <q-icon v-else-if="props.value === 'ready'" class="text-green" name="done" size="2rem"/>
        <q-circular-progress
          v-else
          show-value
          style="margin: 0"
          class="text-light-blue q-ma-md"
          :value="props.row.progress.progress"
          track-color="grey-9"
          size="lg"
          color="light-blue"
        />
      </q-td>
    </template>
    <template v-slot:body-cell-size="props">
      <q-td :props="props" v-if="props.row.status === 'download'">
        {{ prettyBytes(props.value) }} / {{ prettyBytes(props.row.totalDownloadLength) }}
        ({{ prettyBytes(props.row.totalLength) }})
      </q-td>
      <q-td :props="props" v-else>
        {{ prettyBytes(props.row.totalDownloadLength) }} ({{ prettyBytes(props.row.totalLength) }})
      </q-td>
    </template>
    <template v-slot:body-cell-eta="props">
      <q-td :props="props" v-if="props.row.status === 'download'">
        {{ durationFormat(props.row.progress.eta * 1000) }} ({{ durationFormat(props.row.progress.elapsed * 1000) }}
        elapsed)
      </q-td>
      <q-td :props="props" v-else-if="props.row.status === 'idle'">
        -
      </q-td>
      <q-td :props="props" v-else>
        done in {{ durationFormat(props.row.progress.elapsed * 1000) }}
      </q-td>
    </template>
    <template v-slot:body-cell-speed="props">
      <q-td :props="props" v-if="props.row.status === 'download'">
        {{ prettyBytes(props.row.progress.speed) }}ps
      </q-td>
      <q-td :props="props" v-else-if="props.row.status === 'idle'">
        -
      </q-td>
      <q-td :props="props" v-else>
        avg {{ prettyBytes(props.row.totalDownloadLength / props.row.progress.elapsed) }}ps
      </q-td>
    </template>
  </q-table>
</template>

<script setup lang="ts">
import prettyBytes from 'pretty-bytes';
import durationFormat from 'format-duration';
import {Torrent} from 'src/lib/api-types';
import {useRouter} from 'vue-router';
import {QuasarColumnType} from 'src/lib/util';

const router = useRouter();

interface Props {
  title?: string;
  data: Torrent[];
  loading: boolean;
}

const props = defineProps<Props>()

const columns: QuasarColumnType[] = [
  {
    name: 'name',
    label: 'Name',
    field: 'name',
    align: 'left',
    sortable: true,
  },
  {
    name: 'status',
    label: 'Status',
    field: 'status',
    align: 'left',
    sortable: true,
  },
  {
    name: 'size',
    label: 'Size',
    field: 'bytesRead',
    align: 'left',
  },
  {
    name: 'eta',
    label: 'ETA',
    field: 'progress',
    align: 'left',
  },
  {
    name: 'speed',
    label: 'Speed',
    field: 'progress',
    align: 'left',
  }
]

function onRowClick(e: any, torrent: Torrent) {
  console.log(torrent.id);
  router.push(`/torrent/${torrent.id}/download`)
}
</script>

<style lang="sass" scoped>

</style>
