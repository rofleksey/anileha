<script setup>
import { useUserStore } from "../stores/user";
import { ref } from "vue";
import axios from "axios";
import { notify } from "@kyvg/vue3-notification";

defineProps({
  modelValue: {
    type: Boolean,
    required: true
  }
});

const userRef = ref("");
const passRef = ref("");

const userStore = useUserStore();

const remoteLogin = (vue) => {
  const user = userRef.value;
  const pass = passRef.value;
  if (user.trim().length === 0 || pass.trim().length === 0) {
    notify({
      title: "Failed to login",
      text: "Either username or password is blank",
      type: "error"
    });
    return;
  }
  axios
    .post("http://localhost:5000/user/login", {
      user,
      pass
    })
    .then(() => {
      userStore.setUser(user);
      notify({
        title: "Login OK",
        type: "success"
      });
      vue.$emit("update:modelValue", false);
    })
    .catch((err) => {
      notify({
        title: "Failed to login",
        text: err?.response?.data?.error ?? ""
      });
    });
};
</script>

<template>
  <Transition name="fade">
    <div v-if="modelValue" class="dialog-wrapper">
      <div
        class="dialog"
        v-click-outside="() => $emit('update:modelValue', false)"
      >
        <div class="dialog-window-bar">
          <h2>Sign in</h2>
          <button
            @click="() => $emit('update:modelValue', false)"
            rel="closeBtn"
            class="close"
          >
            ×
          </button>
        </div>
        <div class="dialog-content scroll-container">
          <div class="form">
            <input
              v-model="userRef"
              type="text"
              name="username"
              placeholder="Username"
            />
            <input
              v-model="passRef"
              type="password"
              name="password"
              placeholder="Password"
            />
            <div class="actions">
              <button @click="() => remoteLogin(this)">sign in</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.dialog-wrapper {
  z-index: 100;
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  top: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  margin: 0;
}

/* dont remove, it's used by Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.dialog {
  width: 300px;
  flex: none;
  padding: 1rem;
  background: #1c1f22;
  border-radius: 2px;
  display: flex;
  flex-direction: column;
  margin: 0 auto;
}

.dialog-window-bar {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.dialog-content {
  max-height: 100%;
  padding-right: 5px;
}

h2 {
  font-size: 1.2rem;
  font-weight: 300;
}

.dialog-window-bar button {
  margin-left: auto;
}

button {
  font: inherit;
  color: inherit;
  background-color: rgba(255, 255, 255, 0.05);
  border: 1px solid hsla(0, 100%, 100%, 0);
  padding: 0.5em 1em;
  border-radius: 0.25em;
  box-shadow: inset 0 0 0 0 hsl(0deg 0% 100% / 5%);
  transition: box-shadow 0.3s, border 0.15s ease-in;
  outline: none;
  min-height: 2em;
  overflow: hidden;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  cursor: default;
}

button:hover {
  border: 1px solid hsla(0, 100%, 100%, 0.075);
  -webkit-animation: none;
  animation: none;
  cursor: pointer;
}

input {
  font: inherit;
  line-height: 1.5;
  color: inherit;
  border: none;
  display: block;
  background: rgba(255, 255, 255, 0.05);
  box-sizing: border-box;
  width: 100%;
  padding: 12px 16px;
  border-radius: 2px;
  resize: none;
  margin-top: 10px;
  margin-bottom: 10px;
}

input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.075);
  -webkit-animation: flash 0.6s;
  animation: flash 0.6s;
}

@keyframes flash {
  0% {
    box-shadow: 0 0 4px 4px rgb(255 255 255 / 8%);
  }
}

.actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}
</style>
