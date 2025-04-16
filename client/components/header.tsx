import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-24 items-center">
        <div className="w-80">
          <a className="flex items-center space-x-2 px-8" href="/">
            <span className="font-bold text-3xl">PipeSlicer CI</span>
          </a>
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="w-full max-w-xl">
            <div className="relative">
              <Search className="absolute left-4 top-4 h-6 w-6 text-muted-foreground" />
              <Input
                placeholder="Search repositories..."
                className="pl-12 h-14 text-lg"
              />
            </div>
          </div>
        </div>
        <div className="w-80" />
      </div>
    </header>
  );
} 