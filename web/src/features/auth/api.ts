import { authMeOptions, authMeQueryKey } from "@/api";

/** Shared query key and options for the current session. */
export const sessionQueryKey = authMeQueryKey();
export const sessionOptions = authMeOptions;

export type { MeResponse } from "@/api";
