import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LogOut } from "lucide-react";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { useAuth } from "@/lib/auth";

export default function NoAccessPage() {
  const { logout } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 px-4">
      <div className="absolute top-4 right-4">
        <ThemeSwitcher />
      </div>
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-3 text-center">
          <div className="flex flex-col items-center gap-2">
            <img src="/logo.png" alt="Qaynaq" className="h-16 w-auto" />
            <span className="text-xl font-bold">Qaynaq</span>
          </div>
          <CardTitle className="text-2xl">No access</CardTitle>
          <CardDescription>
            You're signed in with your identity provider, but this Qaynaq instance hasn't granted your account a role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-6">
            Ask the operator of this instance to grant your account Admin or MCP access.
          </p>
          <Button variant="outline" onClick={logout} className="w-full">
            <LogOut className="h-4 w-4 mr-1.5" />
            Sign out
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
