import {onMounted, onUnmounted} from 'vue';
import {useInterval} from 'src/lib/composables';

export interface IWebSocketParams {
  url: string
  onConnect?: () => void
  onMessage: (type: string, msg: any) => void
  onDisconnect?: (e: Event | null) => void
}

export interface WebSocketMessage<T> {
  type: string;
  message: T;
}

export function useWebSocket(params: IWebSocketParams) {
  let ws: WebSocket | undefined;
  let connectingWs: WebSocket | undefined;

  function sendWs<T>(type: string, msg: T) {
    const wsMsg: WebSocketMessage<any> = {
      type,
      message: msg
    }
    ws?.send(JSON.stringify(wsMsg));
  }

  function reconnect() {
    console.log('websocket reconnecting...');

    try {
      ws?.close();
      connectingWs = new WebSocket(params.url);

      connectingWs.onopen = function () {
        ws = connectingWs;
        params.onConnect && params.onConnect();
        console.log('websocket connected');
      }

      connectingWs.onmessage = function (e) {
        const data = JSON.parse(e.data) as WebSocketMessage<never>;
        params.onMessage(data.type, data.message)
      };

      connectingWs.onclose = function (e) {
        params.onDisconnect && params.onDisconnect(null);
        ws = undefined;
        console.warn('websocket closed', e)
      };

      connectingWs.onerror = function (e) {
        params.onDisconnect && params.onDisconnect(e);
        ws = undefined;
        console.warn('websocket error', e);
      };
    } catch (e) {
      console.error('websocket connect error', e);
    }
  }

  const stopInterval = useInterval(() => {
    if (!ws) {
      connectingWs?.close();
      reconnect();
    }
  }, 3000);

  onMounted(() => {
    reconnect();
  })

  onUnmounted(() => {
    stopInterval();
    connectingWs?.close();
    ws?.close();
  })

  return {
    sendWs
  };
}
