import {onMounted, onUnmounted} from 'vue';

export function useInterval(f: () => void, interval: number) {
  let intervalRef: NodeJS.Timeout | number | undefined;

  onMounted(() => {
    intervalRef = setInterval(f, interval)
  });

  function stopInterval() {
    clearInterval(intervalRef);
  }

  onUnmounted(stopInterval);

  return stopInterval;
}
