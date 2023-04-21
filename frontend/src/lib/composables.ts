import {onMounted, onUnmounted, ref} from 'vue';
import MobileDetect from 'mobile-detect'

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


let mb: MobileDetect | undefined

export function useMobileDetect() {
  if (!mb) {
    mb = new MobileDetect(window.navigator.userAgent);
  }
  return mb;
}
