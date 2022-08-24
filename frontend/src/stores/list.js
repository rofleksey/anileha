import {defineStore} from "pinia";

export const useListStore = defineStore({
  id: "list",
  state: () => ({
    entries: [],
  }),
  getters: {},
  actions: {
    setData(data) {
      this.entries.splice(0);
      this.entries.push(...data);
    },
  },
});
