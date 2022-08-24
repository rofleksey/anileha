import {defineStore} from "pinia";

export const useJsonStore = defineStore({
  id: "json",
  state: () => ({
    data: {},
  }),
  getters: {},
  actions: {
    setData(data) {
      this.data = data;
    },
  },
});
