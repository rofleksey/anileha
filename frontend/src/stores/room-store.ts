import {defineStore} from 'pinia';
import {ref} from 'vue';
import {nanoid} from 'nanoid';

export const useRoomStore = defineStore('room', () => {
  const roomId = ref<string>(nanoid(6));
  const episodeId = ref<number | null>(null);

  function setRoomId(newId: string) {
    roomId.value = newId;
  }

  function setEpisodeId(newId: number | null) {
    episodeId.value = newId;
  }

  return {roomId, episodeId, setRoomId, setEpisodeId}
})
