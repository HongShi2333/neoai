import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";
import { CommonResponse } from "@/api/common.ts";
import { Channel } from "@/api/community.ts";

// Admin-scoped community API.
// These endpoints bypass visibility filtering so admins can manage
// channels they wouldn't normally "see" through the user-facing list.

export type AdminChannelListResponse = CommonResponse & {
  data: Channel[];
};

// List ALL channels (admin only). Reuses the same /community/channels
// endpoint — the backend returns the full list when the caller is admin.
export async function adminListChannels(): Promise<AdminChannelListResponse> {
  try {
    const response = await axios.get("/community/channels?all=1");
    return response.data as AdminChannelListResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), data: [] };
  }
}
