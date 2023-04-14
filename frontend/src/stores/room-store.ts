import {defineStore} from 'pinia';
import {ref} from 'vue';
import {nanoid} from 'nanoid';

export const useRoomStore = defineStore('room', () => {
  const roomId = ref<string>(nanoid(6));

  function setRoomId(newId: string) {
    roomId.value = newId;
  }

  return {roomId, setRoomId}
})
