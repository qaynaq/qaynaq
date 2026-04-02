import { Database, Mail, ShoppingBag, Plus } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { completeSetup } from "@/lib/api";

type WizardPath = "database" | "google-workspace" | "shopify" | "skip" | "other";

export function WelcomeOverlay({
  onSelectPath,
}: {
  onSelectPath: (path: WizardPath) => void;
}) {
  const handleSkip = async () => {
    await completeSetup();
    onSelectPath("skip");
  };

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
      <div className="max-w-[620px] w-full bg-card border rounded-xl p-10 mx-4">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-semibold mb-2">Welcome to Qaynaq</h1>
          <p className="text-sm text-muted-foreground">
            Build your first MCP tool in minutes
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-6">
          <Card
            className="cursor-pointer hover:border-primary transition-colors duration-200"
            onClick={() => onSelectPath("database")}
          >
            <CardContent className="pt-6 pb-4 text-center">
              <Database className="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
              <p className="font-medium mb-2">Database</p>
              <Badge variant="secondary" className="text-xs">
                Most popular
              </Badge>
            </CardContent>
          </Card>

          <Card
            className="cursor-pointer hover:border-primary transition-colors duration-200"
            onClick={() => onSelectPath("google-workspace")}
          >
            <CardContent className="pt-6 pb-4 text-center">
              <Mail className="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
              <p className="font-medium mb-2">Google Workspace</p>
              <Badge variant="secondary" className="text-xs">
                Tool pack
              </Badge>
            </CardContent>
          </Card>

          <Card
            className="cursor-pointer hover:border-primary transition-colors duration-200"
            onClick={() => onSelectPath("shopify")}
          >
            <CardContent className="pt-6 pb-4 text-center">
              <ShoppingBag className="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
              <p className="font-medium mb-2">Shopify</p>
              <Badge variant="secondary" className="text-xs">
                Tool pack
              </Badge>
            </CardContent>
          </Card>

          <Card
            className="cursor-pointer hover:border-primary transition-colors duration-200"
            onClick={() => onSelectPath("other")}
          >
            <CardContent className="pt-6 pb-4 text-center">
              <Plus className="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
              <p className="font-medium mb-2">Something else</p>
              <Badge variant="secondary" className="text-xs">
                Custom flow
              </Badge>
            </CardContent>
          </Card>
        </div>

        <div className="text-center">
          <button
            onClick={handleSkip}
            className="text-sm text-muted-foreground hover:text-foreground transition-colors underline"
          >
            Skip for now
          </button>
        </div>
      </div>
    </div>
  );
}
