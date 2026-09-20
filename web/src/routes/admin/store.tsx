import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/admin/store")({
  component: StoreLayout,
});

function StoreLayout() {
  return <Outlet />;
}
