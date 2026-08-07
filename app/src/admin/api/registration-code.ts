import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";
import { CommonResponse } from "@/api/common.ts";

// Registration code admin API client.

export type RegistrationCode = {
  id: number;
  code: string;
  quota: number;
  max_uses: number;
  used_count: number;
  note: string;
  created_at: string;
  disabled: boolean;
};

export type RegistrationCodeListResponse = CommonResponse & {
  data: RegistrationCode[];
};

export type GenerateRegCodeForm = {
  number: number;
  quota: number;
  max_uses: number;
  note: string;
};

export type GenerateRegCodeResponse = CommonResponse & {
  data: string[];
};

export type RegCodeStateResponse = CommonResponse & {
  required: boolean;
};

export async function listRegistrationCodes(): Promise<RegistrationCodeListResponse> {
  try {
    const r = await axios.get("/admin/registration-code/list");
    return r.data as RegistrationCodeListResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), data: [] };
  }
}

export async function generateRegistrationCodes(
  form: GenerateRegCodeForm,
): Promise<GenerateRegCodeResponse> {
  try {
    const r = await axios.post("/admin/registration-code/generate", form);
    return r.data as GenerateRegCodeResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), data: [] };
  }
}

export async function getRegistrationCodeState(): Promise<RegCodeStateResponse> {
  try {
    const r = await axios.get("/admin/registration-code/state");
    return r.data as RegCodeStateResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), required: false };
  }
}

export async function setRegistrationCodeState(
  required: boolean,
): Promise<RegCodeStateResponse> {
  try {
    const r = await axios.post("/admin/registration-code/state", { required });
    return r.data as RegCodeStateResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e), required };
  }
}

export async function disableRegistrationCode(id: number): Promise<CommonResponse> {
  try {
    const r = await axios.post("/admin/registration-code/disable", { id });
    return r.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function deleteRegistrationCode(
  id: number,
): Promise<CommonResponse> {
  try {
    const r = await axios.get(`/admin/registration-code/delete/${id}`);
    return r.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}
