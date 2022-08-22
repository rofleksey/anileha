import { defineStore } from "pinia";

export const useUserStore = defineStore({
  id: "user",
  state: () => ({
    user: null
  }),
  getters: {},
  actions: {
    setUser(username) {
      this.user = username;
    }
  }
});
