import {
  getDev,
  getRestApi,
  getTokenField,
  getWebsocketApi,
} from "@/conf/env.ts";
import { syncSiteInfo } from "@/admin/api/info.ts";
import { setAxiosConfig } from "@/conf/api.ts";
import { version as _version } from "./version.json";

export const version: string = _version; // version of the current build
export const dev: boolean = getDev(); // is in development mode (for debugging, in localhost origin)
export const deploy: boolean = true; // is production environment (for api endpoint)
export const tokenField = getTokenField(deploy); // token field name for storing token

export const apiEndpoint: string = getRestApi(deploy); // api endpoint for rest api calls
export const getWebsocketEndpoint = (): string => getWebsocketApi(deploy); // dynamic websocket endpoint (reflects backend endpoint changes)
export const websocketEndpoint: string = getWebsocketApi(deploy); // initial websocket endpoint (for backward compat)

setAxiosConfig({
  endpoint: apiEndpoint,
  token: tokenField,
});

syncSiteInfo();
