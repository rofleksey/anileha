<template>
  <q-table
    @row-click="onRowClick"
    style="width: 100%"
    :title="props.title"
    :rows="props.data"
    :columns="columns"
    row-key="id"
    :loading="props.loading">
    <template v-slot:body-cell-roles="props">
      <q-td :props="props">
        {{ props.value.length ? props.value.join(', ') : '-' }}
      </q-td>
    </template>
  </q-table>
</template>

<script setup lang="ts">
import {User} from 'src/lib/api-types';
import {QuasarColumnType} from 'src/lib/util';

interface Props {
  title?: string;
  data: User[];
  loading: boolean;
}

const props = defineProps<Props>()

const columns: QuasarColumnType[] = [
  {
    name: 'login',
    label: 'Login',
    field: 'login',
    align: 'left',
    sortable: true,
  },
  {
    name: 'name',
    label: 'Name',
    field: 'name',
    align: 'left',
    sortable: true,
  },
  {
    name: 'email',
    label: 'Email',
    field: 'email',
    align: 'left',
    sortable: true,
  },
  {
    name: 'roles',
    label: 'Roles',
    field: 'roles',
    align: 'left',
  }
]

function onRowClick(e: any, user: User) {
  console.log(user.id);
}
</script>

<style lang="sass" scoped>

</style>
