<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Login</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="usernameRef"
          v-model="username"
          label="Username"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          ref="passwordRef"
          v-model="password"
          label="Password"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
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
import {ref} from 'vue';
import {postLogin} from 'src/lib/post-api';
import {showError, showSuccess} from 'src/lib/util';
import {useUserStore} from 'stores/user-store';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()
const userStore = useUserStore();

defineEmits([
  ...useDialogPluginComponent.emits
])

const usernameRef = ref<any>(null);
const passwordRef = ref<any>(null);

const postLoading = ref(false);
const username = ref('');
const password = ref('');

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  if (!usernameRef.value?.validate() || !passwordRef.value?.validate()) {
    return
  }
  postLoading.value = true;
  postLogin(username.value, password.value)
    .then((user) => {
      userStore.setUser(user);
      showSuccess('Login success');
      onDialogOK();
    })
    .catch((e) => {
      showError('failed to login', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
