/**
 * Global WebSocket Client Manager for Real-Time Event Bus
 */

export interface WSMessage<T = any> {
  type: string;
  topic?: string;
  data: T;
  timestamp?: number;
}

type EventHandler<T = any> = (data: T, rawMsg: WSMessage<T>) => void;

class WebSocketManager {
  private ws: WebSocket | null = null;
  private url: string = "";
  private reconnectTimer: any = null;
  private heartbeatTimer: any = null;
  private isManualClosed = false;
  private handlers = new Map<string, Set<EventHandler>>();

  constructor() {
    this.initUrl();
  }

  private initUrl() {
    const isHttps = window.location.protocol === "https:";
    const wsProto = isHttps ? "wss:" : "ws:";
    const host = window.location.host;
    this.url = `${wsProto}//${host}/ws`;
  }

  /**
   * 启动 WebSocket 连接
   */
  public connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.isManualClosed = false;
    this.initUrl();

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        // console.log("[WebSocket] 连接成功:", this.url);
        this.clearReconnect();
        this.startHeartbeat();
      };

      this.ws.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data);
          if (raw.type === "PONG") {
            return;
          }
          this.dispatch(raw.type, raw.data, raw);
          if (raw.topic) {
            this.dispatch(`topic:${raw.topic}`, raw.data, raw);
          }
        } catch (e) {
          // ignore non-json messages
        }
      };

      this.ws.onerror = (err) => {
        console.warn("[WebSocket] 遇到连接错误:", err);
      };

      this.ws.onclose = () => {
        this.stopHeartbeat();
        if (!this.isManualClosed) {
          this.scheduleReconnect();
        }
      };
    } catch (err) {
      console.warn("[WebSocket] 初始化连接失败:", err);
      this.scheduleReconnect();
    }
  }

  /**
   * 订阅指定类型的事件
   */
  public on<T = any>(type: string, handler: EventHandler<T>) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);

    // 确保已连接
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.connect();
    }

    return () => this.off(type, handler);
  }

  /**
   * 取消订阅事件
   */
  public off<T = any>(type: string, handler: EventHandler<T>) {
    const set = this.handlers.get(type);
    if (set) {
      set.delete(handler);
      if (set.size === 0) {
        this.handlers.delete(type);
      }
    }
  }

  /**
   * 发送上行数据
   */
  public send(type: string, data: any = {}, topic: string = "") {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const msg: WSMessage = {
        type,
        topic,
        data,
        timestamp: Date.now()
      };
      this.ws.send(JSON.stringify(msg));
    }
  }

  /**
   * 分发事件到各监听器
   */
  private dispatch(type: string, data: any, raw: WSMessage) {
    const set = this.handlers.get(type);
    if (set) {
      set.forEach((handler) => {
        try {
          handler(data, raw);
        } catch (err) {
          console.error(`[WebSocket] 事件处理异常 (${type}):`, err);
        }
      });
    }
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: "PING", timestamp: Date.now() }));
      }
    }, 20000);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect() {
    this.clearReconnect();
    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, 3000);
  }

  private clearReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  public close() {
    this.isManualClosed = true;
    this.stopHeartbeat();
    this.clearReconnect();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const wsManager = new WebSocketManager();
