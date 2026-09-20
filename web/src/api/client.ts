import { toApiError } from "./errors";
import { client } from "./generated/client.gen";

const baseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

client.setConfig({
  baseUrl,
  credentials: "include",
});

client.interceptors.error.use((error, response) => toApiError(error, response?.status));

export { client };
