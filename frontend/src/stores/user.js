import {defineStore} from "pinia";

export const useUserStore = defineStore({
  id: "user",
  state: () => ({
    user: "admin",
  }),
  getters: {
    isAdmin() {
      return this.user === "admin";
    },
  },
  actions: {
    setUser(username) {
      this.user = username;
    },
  },
});
