import {createRouter, createWebHashHistory} from "vue-router";
import AllSeriesView from "../views/AllSeriesView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "series",
      component: AllSeriesView,
    },
    {
      path: "/torrents",
      name: "torrents",
      component: () => import("../views/AllTorrentsView.vue"),
    },
    {
        path: "/convert",
        name: "conversions",
        component: () => import("../views/AllConversionsView.vue"),
    },
      {
          path: "/torrents/:id",
          name: "single torrent",
          component: () => import("../views/SingleTorrentView.vue")
      },
      {
          path: "/torrents/files/:id",
          name: "torrent files",
          component: () => import("../views/TorrentFilesView.vue")
      },
      {
          path: "/convert/:id",
          name: "single conversion",
          component: () => import("../views/SingleConversionView.vue")
      },
      {
          path: "/convert/:id/logs",
          name: "conversion logs",
          component: () => import("../views/ConversionLogsView.vue")
      },
      {
      path: "/episodes/:id",
      name: "single episode",
      component: () => import("../views/PlayerView.vue")
    },
    {
      path: "/convert/series/:id",
      name: "series conversions",
      component: () => import("../views/SeriesConversionsView.vue")
    },
    {
      path: "/torrents/series/:id",
      name: "series torrents",
      component: () => import("../views/SeriesTorrentsView.vue"),
    },
    {
      path: "/episodes/series/:id",
      name: "series episodes",
      component: () => import("../views/SeriesEpisodesView.vue"),
    },
  ],
});

export default router;
