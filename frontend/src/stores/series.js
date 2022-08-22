import { defineStore } from "pinia";

export const useSeriesStore = defineStore({
  id: "series",
  state: () => ({
    entries: []
  }),
  getters: {},
  actions: {
    setData(data) {
      this.entries.length = 0;
      this.entries.push(...data);
    }
  }
});
