<template>
  <q-table
    @row-click="onRowClick"
    style="width: 100%"
    :title="props.title"
    :rows="props.data"
    :columns="columns"
    :pagination="{
      rowsPerPage: 25
    }"
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
          style="margin: 0"
          class="text-light-blue q-ma-md"
          :value="props.row.progress.progress"
          track-color="grey-9"
          size="lg"
          color="light-blue"
        />
      </q-td>
    </template>
    <template v-slot:body-cell-eta="props">
      <q-td :props="props" v-if="props.row.status === 'processing'">
        {{ durationFormat(props.row.progress.eta * 1000) }} ({{ durationFormat(props.row.progress.elapsed * 1000) }}
        elapsed)
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
import {Conversion} from 'src/lib/api-types';
import {QuasarColumnType} from 'src/lib/util';
import {useQuasar} from 'quasar';
import LogsPreviewModal from 'components/modal/LogsPreviewModal.vue';

const quasar = useQuasar();

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
  quasar.dialog({
    component: LogsPreviewModal,
    componentProps: {
      conversionId: conversion.id,
    },
  });
}
</script>

<style lang="sass" scoped>

</style>
