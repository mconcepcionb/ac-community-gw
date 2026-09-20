import { Link } from "@tanstack/react-router";

import { PageHeader } from "@/components/common/page-header";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const sections = [
  {
    title: "AzerothCore status",
    description: "Connected players, peak, queue and uptime.",
    to: "/azeroth/status",
  },
  {
    title: "Admin accounts",
    description: "Ban, unban and set GM levels.",
    to: "/admin/accounts",
  },
  {
    title: "Online players",
    description: "Live list of connected players.",
    to: "/admin/online",
  },
  {
    title: "Wallet",
    description: "Your points balance and order history.",
    to: "/store/wallet",
  },
] as const;

export function HomePage() {
  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="ac-community-gw"
        description="Operations console for the community gateway."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {sections.map((section) => (
          <Link key={section.to} to={section.to} className="block">
            <Card className="h-full transition-colors hover:border-foreground/30">
              <CardHeader>
                <CardTitle>{section.title}</CardTitle>
                <CardDescription>{section.description}</CardDescription>
              </CardHeader>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
