import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";
import { CommonResponse } from "@/api/common.ts";

// Community API client.
//
// Mirrors the REST endpoints exposed by `community/router.go` on the backend.
// All calls are authenticated via the global axios Authorization header
// set up in `conf/api.ts`.

export type Visibility =
  | "public"
  | "members"
  | "roles"
  | "whitelist";

export type SendPermission = "everyone" | "admins" | "whitelist";

export type Channel = {
  id: number;
  name: string;
  topic: string;
  visibility: Visibility;
  send_permission: SendPermission;
  visible_roles?: string[];
  visible_users?: number[];
  senders?: number[];
  created_at?: string;
  updated_at?: string;
};

export type Message = {
  id: number;
  channel_id: number;
  user_id: number;
  username: string;
  avatar?: string;
  content: string;
  created_at: string;
  edited_at?: string | null;
};

export type ChannelListResponse = CommonResponse & {
  data: Channel[];
};

export type ChannelResponse = CommonResponse & {
  data?: Channel;
};

export type MessageListResponse = CommonResponse & {
  data: Message[];
};

export type MessageResponse = CommonResponse & {
  data?: Message;
};

export type ChannelForm = {
  name: string;
  topic: string;
  visibility: Visibility;
  send_permission: SendPermission;
  visible_roles?: string[];
  visible_users?: number[];
  senders?: number[];
};

export async function listChannels(): Promise<ChannelListResponse> {
  try {
    const r = await axios.get("/community/channels");
    return r.data as ChannelListResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), data: [] };
  }
}

export async function createChannel(
  form: ChannelForm,
): Promise<ChannelResponse> {
  try {
    const r = await axios.post("/community/channels", form);
    return r.data as ChannelResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function getChannel(id: number): Promise<ChannelResponse> {
  try {
    const r = await axios.get(`/community/channels/${id}`);
    return r.data as ChannelResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function updateChannel(
  id: number,
  form: ChannelForm,
): Promise<ChannelResponse> {
  try {
    const r = await axios.post(`/community/channels/${id}`, form);
    return r.data as ChannelResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function deleteChannel(id: number): Promise<CommonResponse> {
  try {
    const r = await axios.delete(`/community/channels/${id}`);
    return r.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function listMessages(
  channelId: number,
  limit = 200,
): Promise<MessageListResponse> {
  try {
    const r = await axios.get(
      `/community/channels/${channelId}/messages?limit=${limit}`,
    );
    return r.data as MessageListResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), data: [] };
  }
}

export async function postMessage(
  channelId: number,
  content: string,
): Promise<MessageResponse> {
  try {
    const r = await axios.post(
      `/community/channels/${channelId}/messages`,
      { content },
    );
    return r.data as MessageResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function editMessage(
  messageId: number,
  content: string,
): Promise<CommonResponse> {
  try {
    const r = await axios.post(`/community/messages/${messageId}`, {
      content,
    });
    return r.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function deleteMessage(
  messageId: number,
): Promise<CommonResponse> {
  try {
    const r = await axios.delete(`/community/messages/${messageId}`);
    return r.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

// Build a websocket URL pointing at /community/ws on the same host that
// serves the REST API.
export function communityWsUrl(): string {
  // backendEndpoint is e.g. "/api" or "https://api.host.com"
  const endpoint =
    (axios.defaults.baseURL as string | undefined) || "/api";
  const path = `${endpoint}/community/ws`;
  if (path.startsWith("http://")) return `ws://${path.slice(7)}`;
  if (path.startsWith("https://")) return `wss://${path.slice(8)}`;
  if (path.startsWith("/")) {
    return window.location.protocol === "https:"
      ? `wss://${window.location.host}${path}`
      : `ws://${window.location.host}${path}`;
  }
  return path.replace(/^http/, "ws");
}
