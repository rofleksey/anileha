import {RouteRecordRaw} from 'vue-router';

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
  },
  {
    path: '/conversions',
    name: 'all-conversions',
    component: () => import('pages/AllConversionsPage.vue'),
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
      },
      {
        path: 'conversions',
        name: 'series-conversions',
        component: () => import('pages/series/SeriesConversionsPage.vue'),
      }
    ]
  },
  {
    path: '/torrent/:torrentId',
    name: 'torrent-parent',
    component: () => import('pages/torrent/TorrentParentPage.vue'),
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
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
];

export default routes;
