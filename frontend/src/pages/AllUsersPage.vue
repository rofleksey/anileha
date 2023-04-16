<template>
  <q-page class="full-width" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat label="Users" />
    </q-toolbar>
    <UserTable
      :data="data"
      :loading="dataLoading"/>
    <q-page-sticky position="bottom-right" :offset="[18, 18]" >
      <q-btn fab icon="add" color="accent" @click="openNewUserModal"/>
    </q-page-sticky>
  </q-page>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {User} from 'src/lib/api-types';
import {fetchAllUsers} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import UserTable from 'components/UserTable.vue';
import {useQuasar} from 'quasar';
import NewUserModal from 'components/modal/NewUserModal.vue';

const quasar = useQuasar();
const dataLoading = ref(false);
const data = ref<User[]>([]);

function openNewUserModal() {
  quasar.dialog({
    component: NewUserModal,
  }).onOk(() => {
    refreshData();
  });
}

function refreshData() {
  dataLoading.value = true;
  fetchAllUsers()
    .then((newUsers) => {
      data.value = newUsers;
    })
    .catch((e) => {
      showError('Failed to fetch users', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

onMounted(() => {
  refreshData();
})
</script>

<style lang="sass" scoped>

</style>
