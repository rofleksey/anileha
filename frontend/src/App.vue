<template>
  <q-layout view="hHh lpR fFf" class="bg-grey-10">
    <q-header elevated class="bg-grey-10 text-white q-py-xs" height-hint="58">
      <q-toolbar>
        <q-btn
          flat
          dense
          round
          @click="toggleLeftDrawer"
          aria-label="Menu"
          icon="menu"
        />

        <q-btn flat no-caps no-wrap class="q-ml-xs" v-if="$q.screen.gt.xs" @click="router.push('/')">
          <q-icon name="forklift" color="deep-purple-5" size="28px"/>
          <q-toolbar-title shrink class="text-weight-bold">
            AniLeha
          </q-toolbar-title>
        </q-btn>

        <q-space/>

        <div class="YL__toolbar-input-container row no-wrap">
          <q-input dense outlined square v-model="search" placeholder="Search" class="bg-gray-7 col"/>
          <q-btn class="YL__toolbar-input-btn" color="grey-7" text-color="white" icon="search" unelevated/>
        </div>

        <q-space/>

        <div class="q-gutter-sm row items-center no-wrap">
          <q-btn round dense flat color="grey-13" icon="notifications">
            <q-tooltip>Notifications</q-tooltip>
          </q-btn>
          <q-btn round flat @click="openLoginModal" :icon="curUser ? 'person_off' : 'person'">
            <q-tooltip>Account</q-tooltip>
          </q-btn>
        </div>
      </q-toolbar>
    </q-header>

    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
      class="bg-grey-10"
      :width="240"
    >
      <q-scroll-area class="fit bg-gray-10">
        <q-list padding>
          <q-item
            v-for="link in userLinks"
            :key="link.text"
            @click="router.push(link.page)"
            v-ripple
            clickable>
            <q-item-section avatar>
              <q-icon color="gray-12" :name="link.icon"/>
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ link.text }}</q-item-label>
            </q-item-section>
          </q-item>

          <template v-if="curUser?.isAdmin">
            <q-separator class="q-my-md"/>

            <q-item
              v-for="link in adminLinks"
              :key="link.text"
              @click="router.push(link.page)"
              v-ripple
              clickable>
              <q-item-section avatar>
                <q-icon color="gray-12" :name="link.icon"/>
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ link.text }}</q-item-label>
              </q-item-section>
            </q-item>
          </template>

          <q-separator class="q-mt-md q-mb-lg"/>

          <div class="q-px-md text-grey-13">
            <div class="row items-center q-gutter-x-sm q-gutter-y-xs">
              <a
                v-for="button in buttons1"
                :key="button.text"
                class="YL__drawer-footer-link"
                href="javascript:void(0)"
              >
                {{ button.text }}
              </a>
            </div>
          </div>
        </q-list>
      </q-scroll-area>
    </q-drawer>

    <q-page-container>
      <router-view v-slot="{ Component }">
        <keep-alive include="RoomPage">
          <component :is="Component"></component>
        </keep-alive>
      </router-view>

    </q-page-container>

    <!--    <q-ajax-bar-->
    <!--      position="bottom"-->
    <!--      color="accent"-->
    <!--      size="10px"-->
    <!--    />-->
  </q-layout>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, ref} from 'vue'
import {useRouter} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import {fetchMyself} from 'src/lib/get-api';
import LoginModal from 'components/modal/LoginModal.vue';
import {useQuasar} from 'quasar';
import {User} from 'src/lib/api-types';

const quasar = useQuasar();
const router = useRouter();
const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

interface LinkItem {
  icon: string;
  text: string;
  page: string;
}

interface ButtonItem {
  text: string;
}

const userLinks = ref<LinkItem[]>([
  {icon: 'home', text: 'Home', page: '/'},
  {icon: 'video_library', text: 'Series', page: '/series'},
]);

const adminLinks = ref<LinkItem[]>([
  {icon: 'download', text: 'Torrents', page: '/torrents'},
  {icon: 'settings_backup_restore', text: 'Conversions', page: '/conversions'},
]);

const buttons1 = ref<ButtonItem[]>([
  {text: 'About'},
]);

const leftDrawerOpen = ref(false)
const search = ref('')

function toggleLeftDrawer() {
  leftDrawerOpen.value = !leftDrawerOpen.value
}

function openLoginModal() {
  quasar.dialog({
    component: LoginModal,
  }).onOk(() => {
    console.log('login success');
  });
}

onMounted(() => {
  fetchMyself().then((user) => {
    userStore.setUser(user);
  }).catch((e) => {
    console.error(e);
    userStore.setUser(null);
  })
})
</script>

<style lang="sass">
.YL
  &__toolbar-input-container
    min-width: 100px
    width: 55%

  &__toolbar-input-btn
    border-radius: 0
    border-style: solid
    border-width: 1px 1px 1px 0
    border-color: rgba(0, 0, 0, .24)
    max-width: 60px
    width: 100%

  &__drawer-footer-link
    color: inherit
    text-decoration: none
    font-weight: 500
    font-size: .75rem

    &:hover
      color: #FFF
</style>
