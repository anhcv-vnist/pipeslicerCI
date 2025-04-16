'use client';

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { GitBranch, Settings, Container } from "lucide-react";

const navigation = [
  {
    name: "Repositories",
    href: "/dashboard/repositories",
    icon: GitBranch,
  },
  {
    name: "Image Builder",
    href: "/dashboard/image-builder",
    icon: Container,
  },
  {
    name: "Settings",
    href: "/dashboard/settings",
    icon: Settings,
  },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <div className="flex h-screen w-80 flex-col border-r bg-background">
      <nav className="flex-1 space-y-3 p-8">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "flex items-center gap-5 rounded-lg px-6 py-4 text-lg font-medium transition-colors",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-muted"
              )}
            >
              <item.icon className="h-6 w-6" />
              {item.name}
            </Link>
          );
        })}
      </nav>
    </div>
  );
} 