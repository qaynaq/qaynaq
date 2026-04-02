import React, { useEffect, useState } from "react";
import { Sidebar } from "@/components/sidebar";
import { Search, MenuIcon, Github, LogOut } from "lucide-react";
import { CommandPalette } from "@/components/command-palette";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";
import { ThemeSwitcher } from "@/components/theme-switcher";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const REPOSITORY_URL = "https://github.com/qaynaq/qaynaq";

export default function ClientLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const [commandPaletteOpen, setCommandPaletteOpen] = React.useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);
  const { authType, logout } = useAuth();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setCommandPaletteOpen(true);
      }
    };

    const handleSidebarToggle = () => {
      setIsCollapsed(
        document.documentElement.classList.contains("sidebar-collapsed"),
      );
    };

    // Create a MutationObserver to watch for class changes on documentElement
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === "class") {
          handleSidebarToggle();
        }
      });
    });

    observer.observe(document.documentElement, { attributes: true });
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      observer.disconnect();
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const toggleSidebar = () => {
    document.documentElement.classList.toggle("sidebar-collapsed");
  };

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="main-content flex-1 min-w-0 ml-[220px] transition-all duration-300 overflow-hidden">
        <header
          className="h-16 flex items-center justify-between px-6 fixed top-0 z-10 rounded-b-lg shadow-sm bg-white dark:bg-gray-950 transition-all duration-300"
          style={{
            left: isCollapsed ? "calc(60px + 0.5rem)" : "calc(220px + 0.5rem)",
            right: "0.5rem",
          }}
        >
          <div className="flex items-center space-x-4">
            <button
              onClick={toggleSidebar}
              className="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800"
              aria-label="Toggle sidebar"
            >
              <MenuIcon className="h-5 w-5" />
            </button>
            <div className="h-6 w-px bg-gray-200 dark:bg-gray-800"></div>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="search"
                placeholder="Search"
                className="h-9 w-full rounded-md border border-input bg-background pl-8 pr-16 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 cursor-pointer"
                onClick={() => setCommandPaletteOpen(true)}
                readOnly
              />
              <div
                className="absolute right-2.5 top-2 text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded cursor-pointer"
                onClick={() => setCommandPaletteOpen(true)}
              >
                ⌘K
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <TooltipProvider>
              <ThemeSwitcher />
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" asChild>
                    <a
                      href={REPOSITORY_URL}
                      target="_blank"
                      rel="noopener noreferrer"
                      aria-label="GitHub Repository"
                    >
                      <Github className="h-5 w-5" />
                    </a>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>View GitHub Repository</p>
                </TooltipContent>
              </Tooltip>
              {authType !== "none" && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" onClick={logout}>
                      <LogOut className="h-5 w-5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Logout</p>
                  </TooltipContent>
                </Tooltip>
              )}
            </TooltipProvider>
          </div>
        </header>
        <main className="pt-16 px-2 min-w-0 overflow-x-hidden">{children}</main>
      </div>
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
      />
    </div>
  );
}
