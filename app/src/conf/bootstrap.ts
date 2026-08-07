import {
  getDev,
  getRestApi,
  getTokenField,
  getWebsocketApi,
} from "@/conf/env.ts";
import { syncSiteInfo } from "@/admin/api/info.ts";
import { setAxiosConfig } from "@/conf/api.ts";
import { version as _version } from "./version.json";
import { setMemory } from "@/utils/memory.ts";

export const version: string = _version; // version of the current build
export const dev: boolean = getDev(); // is in development mode (for debugging, in localhost origin)
export const deploy: boolean = true; // is production environment (for api endpoint)
export const tokenField = getTokenField(deploy); // token field name for storing token

export let apiEndpoint: string = getRestApi(deploy); // api endpoint for rest api calls
export let websocketEndpoint: string = getWebsocketApi(deploy); // api endpoint for websocket calls

setAxiosConfig({
  endpoint: apiEndpoint,
  token: tokenField,
});

// NeoAI: pick up OAuth callback token from URL fragment.
//
// When the OAuth provider redirects back to the frontend, the backend
// appends `#access_token=<jwt>` to the redirect URL. We read it here,
// persist it to the token slot, and clean up the URL so users don't
// accidentally share it. Subsequent validateToken calls (e.g. from
// NavBar) will then pick it up and authenticate the user.
(function handleOAuthCallbackToken() {
  if (typeof window === "undefined") return;
  const hash = window.location.hash || "";
  const match = hash.match(/access_token=([^&]+)/);
  if (match && match[1]) {
    const token = decodeURIComponent(match[1]);
    if (token && token.length > 0) {
      setMemory(tokenField, token);
      // Strip the fragment so the token isn't visible in the URL bar.
      history.replaceState(null, "", window.location.pathname + window.location.search);
    }
  }
})();

syncSiteInfo();
