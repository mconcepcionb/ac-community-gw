import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../api/swagger.yaml",
  output: {
    path: "src/api/generated",
  },
  plugins: [
    "@hey-api/client-fetch",
    "@hey-api/typescript",
    "@hey-api/sdk",
    "@tanstack/react-query",
    "zod",
  ],
});
