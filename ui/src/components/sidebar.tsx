import { Link, useLocation } from "react-router-dom";
import {
  LayoutDashboard,
  Cpu,
  MemoryStick,
  Layers,
  Waypoints,
  KeyRound,
  Link2,
  Gauge,
  FileText,
  Settings,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useEffect, useState } from "react";
import { QuickCreate } from "./quick-create";

const menuItems = [
  {
    section: "General",
    items: [
      {
        name: "Dashboard",
        href: "/",
        icon: LayoutDashboard,
      },
      {
        name: "Flows",
        href: "/flows",
        icon: Waypoints,
      },
      {
        name: "Workers",
        href: "/workers",
        icon: Cpu,
      },
    ],
  },
  {
    section: "Other",
    items: [
      {
        name: "Secrets",
        href: "/secrets",
        icon: KeyRound,
      },
      {
        name: "Connections",
        href: "/connections",
        icon: Link2,
      },
      {
        name: "Caches",
        href: "/caches",
        icon: MemoryStick,
      },
      {
        name: "Rate Limits",
        href: "/rate-limits",
        icon: Gauge,
      },
      {
        name: "Buffers",
        href: "/buffers",
        icon: Layers,
      },
      {
        name: "Files",
        href: "/files",
        icon: FileText,
      },
    ],
  },
  {
    section: "Dashboard",
    items: [
      {
        name: "Settings",
        href: "/settings",
        icon: Settings,
      },
    ],
  },
];

export function Sidebar() {
  const location = useLocation();
  const pathname = location.pathname;
  const [isCollapsed, setIsCollapsed] = useState(false);

  useEffect(() => {
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

    return () => observer.disconnect();
  }, []);

  return (
    <aside
      className={cn(
        "sidebar bg-gray-50 dark:bg-gray-900 fixed top-2 bottom-2 overflow-y-auto border border-gray-200 dark:border-gray-800 rounded-xl transition-all duration-300 z-20",
        isCollapsed ? "w-[60px] left-2" : "w-[220px] left-2",
      )}
    >
      <div className="h-14 flex items-center px-4 border-b border-gray-200 dark:border-gray-800">
        <div className="flex items-center gap-2">
          <img src="/logo.png" alt="Qaynaq" className="h-6 w-6 flex-shrink-0 rounded" />
          {!isCollapsed && (
            <h1 className="text-lg font-semibold" style={{ fontFamily: "'Unbounded', sans-serif" }}>Qaynaq</h1>
          )}
        </div>
      </div>

      {/* Quick Create Button */}
      <div className="px-3 py-3 border-b border-gray-200 dark:border-gray-800">
        <QuickCreate isCollapsed={isCollapsed} />
      </div>

      <TooltipProvider delayDuration={300}>
        <div className="px-3 py-2">
          {menuItems.map((section, sectionIndex) => (
            <div key={section.section} className="mb-6">
              {!isCollapsed ? (
                <h2 className="text-xs font-semibold text-muted-foreground mb-2 px-2">
                  {section.section}
                </h2>
              ) : sectionIndex > 0 ? (
                <div className="h-px bg-gray-200 dark:bg-gray-800 my-4 mx-2"></div>
              ) : null}
              <div className="space-y-1">
                {section.items.map((item) => {
                  const isActive =
                    pathname === item.href ||
                    (item.href !== "/" && pathname.startsWith(item.href));

                  const menuButton = (
                    <Button
                      key={item.href}
                      variant="ghost"
                      asChild
                      className={cn(
                        "w-full justify-start text-sm font-medium",
                        isCollapsed && "px-2",
                        isActive
                          ? "bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-50"
                          : "text-muted-foreground",
                      )}
                    >
                      <Link
                        to={item.href}
                        className={cn(
                          "flex items-center gap-3 rounded-lg px-3 py-2 text-gray-500 transition-all hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-50",
                          isActive &&
                            "bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-gray-50",
                          isCollapsed && "justify-center",
                        )}
                      >
                        <item.icon
                          className={cn(
                            "h-4 w-4",
                            isCollapsed ? "mx-auto" : "mr-2",
                          )}
                        />
                        {!isCollapsed && <span>{item.name}</span>}
                      </Link>
                    </Button>
                  );

                  return isCollapsed ? (
                    <Tooltip key={item.href}>
                      <TooltipTrigger asChild>{menuButton}</TooltipTrigger>
                      <TooltipContent side="right">{item.name}</TooltipContent>
                    </Tooltip>
                  ) : (
                    menuButton
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      </TooltipProvider>
    </aside>
  );
}
