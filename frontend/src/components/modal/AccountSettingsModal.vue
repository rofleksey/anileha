<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide" @keyup.enter="onOKClick">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Settings</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="nameRef"
          v-model="userName"
          label="Name"/>
        <q-input
          ref="emailRef"
          v-model="email"
          label="Email"/>
        <q-input
          ref="passRef"
          v-model="pass"
          type="password"
          label="Password"/>
        <q-input
          ref="repeatPassRef"
          v-model="repeatPass"
          type="password"
          label="Repeat Password"/>
        <q-file
          ref="imageRef"
          v-model="image"
          label="Avatar"
          accept="image/jpeg,image/png"
          max-file-size="8388608">
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
          icon="logout"
          @click="onLogoutClick"/>
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
import {computed, onMounted, ref, watch} from 'vue';
import {postAccountAvatar, postLogout, postModifyAccount} from 'src/lib/post-api';
import {showError, showHint, showSuccess} from 'src/lib/util';
import {useUserStore} from 'stores/user-store';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

defineEmits([
  ...useDialogPluginComponent.emits
])

const userStore = useUserStore();
const user = computed(() => userStore.user);

const nameRef = ref<any>(null);
const emailRef = ref<any>(null);
const passRef = ref<any>(null);
const repeatPassRef = ref<any>(null);
const imageRef = ref<any>(null);

const postLoading = ref(false);
const userName = ref(user.value?.name ?? '');
const email = ref(user.value?.email ?? '');
const pass = ref('');
const repeatPass = ref('');
const image = ref<File | null>(null);

watch(image, () => {
  if (!image.value) {
    return
  }
  postLoading.value = true;
  postAccountAvatar(image.value)
    .then((newThumb) => {
      showSuccess('Avatar uploaded');
      const curUser = user.value
      if (curUser) {
        userStore.setUser({
          ...curUser,
          thumb: newThumb || curUser.thumb || '',
        });
      }
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to upload avatar', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
})

function onLogoutClick() {
  postLoading.value = true;
  postLogout()
    .then(() => {
      showSuccess('Logout success');
      userStore.setUser(null);
      onDialogOK(true);
    })
    .catch((e) => {
      showError('Failed to logout', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  if (!userName.value && !pass.value && !email.value) {
    showHint('No changes')
    return;
  }
  if (pass.value && pass.value !== repeatPass.value) {
    showHint('Passwords are not equal')
    return
  }
  postLoading.value = true;
  postModifyAccount(userName.value, pass.value, email.value)
    .then(() => {
      showSuccess('Settings applied');
      const curUser = user.value
      if (curUser) {
        userStore.setUser({
          ...curUser,
          name: userName.value || curUser.name || '',
          email: email.value || curUser.email || '',
        });
      }
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to apply settings', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
