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
        <q-icon v-if="props.value === 'created'" class="text-orange" name="schedule" size="2rem"/>
        <q-icon v-else-if="props.value === 'cancelled'" class="text-purple" name="block" size="2rem"/>
        <q-icon v-else-if="props.value === 'error'" class="text-red" name="error" size="2rem"/>
        <q-icon v-else-if="props.value === 'ready'" class="text-green" name="done" size="2rem"/>
        <q-circular-progress
          v-else
          show-value
          class="text-light-blue q-ma-md"
          :value="props.row.progress.progress"
          size="2.5rem"
          color="light-blue"
        />
      </q-td>
    </template>
    <template v-slot:body-cell-eta="props">
      <q-td :props="props" v-if="props.row.status === 'processing'">
        {{ durationFormat(props.row.progress.eta * 1000) }} ({{ durationFormat(props.row.progress.elapsed * 1000) }})
      </q-td>
      <q-td :props="props" v-else-if="props.row.status === 'created'">
        -
      </q-td>
      <q-td :props="props" v-else>
        done in {{ durationFormat(props.row.progress.elapsed * 1000) }}
      </q-td>
    </template>
    <template v-slot:body-cell-speed="props">
      <q-td :props="props" v-if="props.row.status === 'processing'">
        x{{ props.row.progress.speed }}
      </q-td>
      <q-td :props="props" v-else-if="props.row.status === 'created'">
        -
      </q-td>
      <q-td :props="props" v-else>
        avg x{{ props.row.progress.speed }}
      </q-td>
    </template>
  </q-table>
</template>

<script setup lang="ts">
import durationFormat from 'format-duration';
import {Conversion, Torrent} from 'src/lib/api-types';
import {useRouter} from 'vue-router';
import {QuasarColumnType} from 'src/lib/util';

const router = useRouter();

interface Props {
  title?: string;
  data: Conversion[];
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

function onRowClick(e: any, conversion: Conversion) {
  console.log(conversion.id);
}
</script>

<style lang="sass" scoped>

</style>
