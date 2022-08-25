import { defineStore } from "pinia";

export const useTextStore = defineStore({
  id: "text",
  state: () => ({
    text: ""
  }),
  getters: {},
  actions: {
    setText(text) {
      this.text = text;
    }
  }
});
