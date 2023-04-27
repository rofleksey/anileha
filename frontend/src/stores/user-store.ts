import {defineStore} from 'pinia';
import {ref} from 'vue';
import {User} from 'src/lib/api-types';

const localStorageKey = 'anileha-user';

export const useUserStore = defineStore('user', () => {
  const localUserStr = localStorage.getItem(localStorageKey);

  const user = ref<User | null>(localUserStr ? JSON.parse(localUserStr) : null);

  function setUser(newUser: User | null) {
    user.value = newUser;

    if (newUser) {
      localStorage.setItem(localStorageKey, JSON.stringify(newUser))
    } else {
      localStorage.removeItem(localStorageKey);
    }
  }

  return {user, setUser}
})
