import { defineStore } from "pinia";

export const useUserStore = defineStore({
  id: "user",
  state: () => ({
    user: null,
    isAdmin: false
  }),
  getters: {},
  actions: {
    setUser(username, isAdmin) {
      this.user = username;
      this.isAdmin = isAdmin;
    },
    logout() {
      this.user = null;
      this.isAdmin = null;
    }
  }
});
