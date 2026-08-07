import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";
import { getWebsocketEndpoint, tokenField } from "@/conf/bootstrap.ts";
import { getMemory } from "@/utils/memory.ts";

export type CommunityChannel = {
  id: number;
  name: string;
  description: string;
  topic: string;
  visibility: "public" | "private";
  visible_groups: string[];
  post_groups: string[];
  members: string[];
  posters: string[];
  position: number;
  created_by: string;
  created_at: string;
};

export type CommunityMessage = {
  id: number;
  channel_id: number;
  sender_id: number;
  sender_username: string;
  content: string;
  created_at: string;
};

export type CommonCommunityResponse = {
  status: boolean;
  error: string;
};

export type ChannelListResponse = CommonCommunityResponse & {
  data: CommunityChannel[];
};

export type ChannelResponse = CommonCommunityResponse & {
  data?: CommunityChannel;
};

export type MessageListResponse = CommonCommunityResponse & {
  data: CommunityMessage[];
};

export type SendMessageResponse = CommonCommunityResponse & {
  message?: CommunityMessage;
};

export type ChannelForm = {
  name: string;
  description: string;
  topic: string;
  visibility: "public" | "private";
  visible_groups: string[];
  post_groups: string[];
  members: string[];
  posters: string[];
  position: number;
};

function err(e: unknown, fallback = ""): string {
  return getErrorMessage(e) || fallback;
}

// ---- admin management ----

export async function listChannelsAdmin(): Promise<ChannelListResponse> {
  try {
    const resp = await axios.get("/admin/community/channel/list");
    return resp.data as ChannelListResponse;
  } catch (e) {
    return { status: false, error: err(e), data: [] };
  }
}

export async function createChannel(
  form: ChannelForm,
): Promise<ChannelResponse> {
  try {
    const resp = await axios.post("/admin/community/channel/create", form);
    return resp.data as ChannelResponse;
  } catch (e) {
    return { status: false, error: err(e) };
  }
}

export async function updateChannel(
  id: number,
  form: ChannelForm,
): Promise<ChannelResponse> {
  try {
    const resp = await axios.post(`/admin/community/channel/update/${id}`, form);
    return resp.data as ChannelResponse;
  } catch (e) {
    return { status: false, error: err(e) };
  }
}

export async function deleteChannel(
  id: number,
): Promise<CommonCommunityResponse> {
  try {
    const resp = await axios.get(`/admin/community/channel/delete/${id}`);
    return resp.data as CommonCommunityResponse;
  } catch (e) {
    return { status: false, error: err(e) };
  }
}

// ---- user-facing ----

export async function listChannels(): Promise<ChannelListResponse> {
  try {
    const resp = await axios.get("/community/channels");
    return resp.data as ChannelListResponse;
  } catch (e) {
    return { status: false, error: err(e), data: [] };
  }
}

export async function listMessages(
  channelId: number,
  before = 0,
  limit = 50,
): Promise<MessageListResponse> {
  try {
    const params = new URLSearchParams();
    if (before > 0) params.set("before", String(before));
    if (limit > 0) params.set("limit", String(limit));
    const query = params.toString();
    const resp = await axios.get(
      `/community/channel/${channelId}/messages${query ? `?${query}` : ""}`,
    );
    return resp.data as MessageListResponse;
  } catch (e) {
    return { status: false, error: err(e), data: [] };
  }
}

export async function sendMessage(
  channelId: number,
  content: string,
): Promise<SendMessageResponse> {
  try {
    const resp = await axios.post(`/community/channel/${channelId}/send`, {
      content,
    });
    return resp.data as SendMessageResponse;
  } catch (e) {
    return { status: false, error: err(e) };
  }
}

// ---- websocket ----

export type WsOutgoing = {
  type: "message" | "system";
  message: CommunityMessage;
};

export type CommunitySocket = {
  send: (content: string) => void;
  close: () => void;
};

export function openCommunitySocket(
  channelId: number,
  onMessage: (msg: CommunityMessage) => void,
  onClose?: () => void,
): CommunitySocket {
  const url = `${getWebsocketEndpoint()}/community/ws`;
  let ws: WebSocket | null = null;
  let closed = false;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;

  const connect = () => {
    ws = new WebSocket(url);

    ws.onopen = () => {
      ws?.send(
        JSON.stringify({
          token: getMemory(tokenField) || "anonymous",
          channel_id: channelId,
        }),
      );
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WsOutgoing;
        if (data && data.type === "message" && data.message) {
          onMessage(data.message);
        }
      } catch (e) {
        console.warn("[community] failed to parse ws message", e);
      }
    };

    ws.onclose = () => {
      if (closed) {
        onClose?.();
        return;
      }
      // auto-retry until explicitly closed
      retryTimer = setTimeout(() => connect(), 3000);
    };

    ws.onerror = () => {
      ws?.close();
    };
  };

  connect();

  return {
    send: (content: string) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ content }));
      }
    },
    close: () => {
      closed = true;
      if (retryTimer) clearTimeout(retryTimer);
      ws?.close();
    },
  };
}
