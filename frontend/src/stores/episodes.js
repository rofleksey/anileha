import { defineStore } from "pinia";

export const useEpisodesStore = defineStore({
  id: "episodes",
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
