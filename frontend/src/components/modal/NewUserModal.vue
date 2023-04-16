<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide" @keyup.enter="onOKClick">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Create user</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="usernameRef"
          v-model="username"
          label="Username"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          ref="emailRef"
          v-model="email"
          label="Email"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          ref="passwordRef"
          v-model="password"
          label="Password"
          type="password"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          ref="rolesRef"
          v-model="roles"
          label="Roles (comma separated)"/>
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
import {postNewUser} from 'src/lib/post-api';
import {showError, showSuccess} from 'src/lib/util';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

defineEmits([
  ...useDialogPluginComponent.emits
])

const usernameRef = ref<any>(null);
const emailRef = ref<any>(null);
const passwordRef = ref<any>(null);

const postLoading = ref(false);
const username = ref('');
const email = ref('');
const password = ref('');
const roles = ref('');

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  if (!usernameRef.value?.validate() || !emailRef.value?.validate() || !passwordRef.value?.validate()) {
    return
  }
  postLoading.value = true;
  const roleArr = roles.value
    .split(',')
    .map((r) => r.trim())
    .filter((r) => r.length > 0);
  postNewUser(username.value, password.value, email.value, roleArr)
    .then(() => {
      showSuccess('User created');
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to create user', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
