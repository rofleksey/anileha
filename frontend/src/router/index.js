import { createRouter, createWebHashHistory } from "vue-router";
import AllSeriesView from "../views/AllSeriesView.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "series",
      component: AllSeriesView
    },
    {
      path: "/s/:id",
      name: "series episodes",
      component: () => import("../views/SeriesEpisodesView.vue")
    }
  ]
});

export default router;
