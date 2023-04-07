import {defineStore} from 'pinia';
import {ref} from 'vue';
import {User} from 'src/lib/api-types';

export const useUserStore = defineStore('catalog', () => {
  const user = ref<User | null>(null);

  function setUser(newUser: User | null) {
    user.value = newUser;
  }

  return {user, setUser}
})
