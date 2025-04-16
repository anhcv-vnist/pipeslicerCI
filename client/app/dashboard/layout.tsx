import { Header } from "@/components/header";
import { Sidebar } from "@/components/sidebar";
import { Toaster } from "@/components/ui/toaster";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative flex min-h-screen flex-col">
      <Header />
      <div className="flex-1 flex">
        <Sidebar />
        <main className="flex-1 p-6">{children}</main>
      </div>
      <Toaster />
    </div>
  );
} 