import { updateDocumentTitle, updateFavicon } from "@/utils/dom.ts";
import { setMemory } from "@/utils/memory.ts";

export let appName =
  localStorage.getItem("app_name") || import.meta.env.VITE_APP_NAME || "NeoAI";
export let appLogo =
  localStorage.getItem("app_logo") ||
  import.meta.env.VITE_APP_LOGO ||
  "/favicon.ico";
export let blobEndpoint =
  localStorage.getItem("blob_endpoint") ||
  import.meta.env.VITE_BLOB_ENDPOINT ||
  "https://blob.neoai.dev";
export let docsEndpoint =
  localStorage.getItem("docs_url") ||
  import.meta.env.VITE_DOCS_ENDPOINT ||
  "https://neoai.dev";
export let buyLink =
  localStorage.getItem("buy_link") || import.meta.env.VITE_BUY_LINK || "";

export const useDeeptrain = !!import.meta.env.VITE_USE_DEEPTRAIN;
export const backendEndpoint = import.meta.env.VITE_BACKEND_ENDPOINT || "/api";
export const deeptrainEndpoint =
  import.meta.env.VITE_DEEPTRAIN_ENDPOINT || "https://deeptrain.net";
export const deeptrainAppName = import.meta.env.VITE_DEEPTRAIN_APP || "neoai";
export const deeptrainApiEndpoint =
  import.meta.env.VITE_DEEPTRAIN_API_ENDPOINT || "https://api.deeptrain.net";

updateDocumentTitle(import.meta.env.VITE_APP_NAME);
updateFavicon(import.meta.env.VITE_APP_LOGO);

export function getDev(): boolean {
  /**
   * return if the current environment is development
   */
  return window.location.hostname === "localhost";
}

export function getRestApi(deploy: boolean): string {
  /**
   * return the REST API address.
   *
   * NeoAI URL contract (single source of truth):
   *   - Backend ALWAYS mounts API routes under `/api`.
   *   - Frontend axios baseURL must therefore end with `/api`.
   *
   *   VITE_BACKEND_ENDPOINT can be:
   *     - "/api"              (same-origin, default)
   *     - "https://api.host"   (cross-origin) → normalised to "https://api.host/api"
   *     - "https://api.host/api" (already correct) → kept as-is
   *
   *   In dev mode we talk to the local Go server on :8094 — also under /api.
   */
  if (!deploy) return "http://localhost:8094/api";

  let endpoint = backendEndpoint;
  // Strip trailing slash for consistency.
  endpoint = endpoint.replace(/\/+$/, "");
  // If the user supplied just a host (no /api suffix), append /api.
  if (!endpoint.endsWith("/api")) {
    endpoint = endpoint + "/api";
  }
  return endpoint;
}

export function getWebsocketApi(deploy: boolean): string {
  /**
   * return the WebSocket API address (derived from getRestApi so the
   * /api contract stays in exactly one place).
   */
  const rest = getRestApi(deploy);
  if (rest.startsWith("http://")) return `ws://${rest.slice(7)}`;
  if (rest.startsWith("https://")) return `wss://${rest.slice(8)}`;
  // relative path
  return location.protocol === "https:"
    ? `wss://${location.host}${rest}`
    : `ws://${location.host}${rest}`;
}

export function getTokenField(deploy: boolean): string {
  /**
   * return the token field name in localStorage
   */
  return deploy ? "token" : "token-dev";
}

export function setAppName(name: string): void {
  /**
   * set the app name in localStorage
   */
  name = name.trim() || "NeoAI";
  setMemory("app_name", name);
  appName = name;

  updateDocumentTitle(name);
}

export function setAppLogo(logo: string): void {
  /**
   * set the app logo in localStorage
   */
  logo = logo.trim() || "/favicon.ico";
  setMemory("app_logo", logo);
  appLogo = logo;

  updateFavicon(logo);
}

export function setDocsUrl(url: string): void {
  /**
   * set the docs url in localStorage
   */
  url = url.trim() || "https://neoai.dev";
  setMemory("docs_url", url);
  docsEndpoint = url;
}

export function setBlobEndpoint(endpoint: string): void {
  /**
   * set the blob endpoint in localStorage
   */
  endpoint = endpoint.trim() || "https://blob.neoai.dev";
  setMemory("blob_endpoint", endpoint);
  blobEndpoint = endpoint;
}

export function setBuyLink(link: string): void {
  /**
   * set the buy link in localStorage
   */
  link = link.trim() || "";
  setMemory("buy_link", link);
  buyLink = link;
}

