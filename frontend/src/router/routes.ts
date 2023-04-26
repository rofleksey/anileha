import {RouteRecordRaw} from 'vue-router';
import {useUserStore} from 'stores/user-store';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('pages/EmptyPage.vue'),
  },
  {
    path: '/series',
    name: 'all-series',
    component: () => import('pages/AllSeriesPage.vue'),
  },
  {
    path: '/torrents',
    name: 'all-torrents',
    component: () => import('pages/AllTorrentsPage.vue'),
    beforeEnter: (to, from, next) => {
      const userStore = useUserStore();
      if (!userStore.user?.roles?.includes('admin')) {
        next({ path: "/forbidden" });
      } else next();
    },
  },
  {
    path: '/conversions',
    name: 'all-conversions',
    component: () => import('pages/AllConversionsPage.vue'),
    beforeEnter: (to, from, next) => {
      const userStore = useUserStore();
      if (!userStore.user?.roles?.includes('admin')) {
        next({ path: "/forbidden" });
      } else next();
    },
  },
  {
    path: '/users',
    name: 'all-users',
    component: () => import('pages/AllUsersPage.vue'),
    beforeEnter: (to, from, next) => {
      const userStore = useUserStore();
      if (!userStore.user?.roles?.includes('owner')) {
        next({ path: "/forbidden" });
      } else next();
    },
  },
  {
    path: '/series/:seriesId',
    name: 'series-parent',
    component: () => import('pages/series/SeriesParentPage.vue'),
    children: [
      {
        path: 'episodes',
        name: 'series-episodes',
        component: () => import('pages/series/SeriesEpisodesPage.vue'),
      },
      {
        path: 'torrents',
        name: 'series-torrents',
        component: () => import('pages/series/SeriesTorrentsPage.vue'),
        beforeEnter: (to, from, next) => {
          const userStore = useUserStore();
          if (!userStore.user?.roles?.includes('admin')) {
            next({ path: "/forbidden" });
          } else next();
        },
      },
      {
        path: 'conversions',
        name: 'series-conversions',
        component: () => import('pages/series/SeriesConversionsPage.vue'),
        beforeEnter: (to, from, next) => {
          const userStore = useUserStore();
          if (!userStore.user?.roles?.includes('admin')) {
            next({ path: "/forbidden" });
          } else next();
        },
      }
    ]
  },
  {
    path: '/torrent/:torrentId',
    name: 'torrent-parent',
    component: () => import('pages/torrent/TorrentParentPage.vue'),
    beforeEnter: (to, from, next) => {
      const userStore = useUserStore();
      if (!userStore.user?.roles?.includes('admin')) {
        next({ path: "/forbidden" });
      } else next();
    },
    children: [
      {
        path: 'download',
        name: 'torrent-download',
        component: () => import('pages/torrent/TorrentDownloadPage.vue'),
      },
      {
        path: 'convert',
        name: 'torrent-convert',
        component: () => import('pages/torrent/TorrentConvertPage.vue'),
      },
    ]
  },
  {
    path: '/watch/:episodeId',
    name: 'watch',
    component: () => import('pages/WatchPage.vue'),
  },
  {
    path: '/room',
    name: 'room',
    component: () => import('pages/RoomPage.vue'),
    beforeEnter: (to, from, next) => {
      const userStore = useUserStore();
      if (!userStore.user) {
        next({ path: "/unauthorized" });
      } else next();
    },
  },
  {
    path: '/about',
    name: 'about',
    component: () => import('pages/AboutPage.vue'),
  },
  {
    path: '/forbidden',
    name: 'forbidden',
    component: () => import('pages/ForbiddenPage.vue'),
  },
  {
    path: '/unauthorized',
    name: 'unauthorized',
    component: () => import('pages/UnauthorizedPage.vue'),
  },
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
];

export default routes;
