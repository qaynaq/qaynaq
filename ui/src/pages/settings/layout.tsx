import { NavLink, Outlet } from "react-router-dom";
import { cn } from "@/lib/utils";

const tabs = [
  { to: "/settings/authentication", label: "Authentication" },
  { to: "/settings/tokens", label: "API Tokens" },
  { to: "/settings/sessions", label: "Sessions" },
  { to: "/settings/clients", label: "OAuth Clients" },
];

export default function SettingsLayout() {
  return (
    <div className="p-6 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Manage application settings and integrations
        </p>
      </div>

      <div className="border-b border-gray-200 dark:border-gray-800 mb-6">
        <nav className="flex gap-1 overflow-x-auto">
          {tabs.map((tab) => (
            <NavLink
              key={tab.to}
              to={tab.to}
              className={({ isActive }) =>
                cn(
                  "px-4 py-2 text-sm font-medium border-b-2 -mb-px whitespace-nowrap transition-colors",
                  isActive
                    ? "border-foreground text-foreground"
                    : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300 dark:hover:border-gray-700",
                )
              }
            >
              {tab.label}
            </NavLink>
          ))}
        </nav>
      </div>

      <Outlet />
    </div>
  );
}
