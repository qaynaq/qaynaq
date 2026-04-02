import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Loader2,
  Check,
  ShoppingBag,
} from "lucide-react";
import { testShopifyConnection, createFlow } from "@/lib/api";
import { shopifyPack } from "@/lib/mcp-tool-templates/shopify";
import { buildFlowFromTemplate } from "@/lib/flow-builder-utils";

export type ShopifyWizardResult = {
  toolCount: number;
};

const TOOL_GROUPS = [
  { label: "Orders", ids: ["shopify_list_orders", "shopify_get_order"] },
  { label: "Products", ids: ["shopify_list_products", "shopify_get_product"] },
  { label: "Customers", ids: ["shopify_list_customers", "shopify_get_customer"] },
  { label: "Inventory", ids: ["shopify_list_inventory_items"] },
];

export function ShopifyWizard({
  onComplete,
  onBack,
}: {
  onComplete: (result: ShopifyWizardResult) => void;
  onBack: () => void;
}) {
  const [step, setStep] = useState(1);

  // Step 1 state
  const [shopName, setShopName] = useState("");
  const [accessToken, setAccessToken] = useState("");
  const [testing, setTesting] = useState(false);
  const [connectionOk, setConnectionOk] = useState(false);
  const [connectedShopName, setConnectedShopName] = useState("");
  const [connectionError, setConnectionError] = useState("");
  const [helpExpanded, setHelpExpanded] = useState(false);

  // Step 2 state
  const [selectedTools, setSelectedTools] = useState<Set<string>>(
    new Set(shopifyPack.templates.map((t) => t.id)),
  );

  // Step 3 state
  const [deploying, setDeploying] = useState(false);
  const [deployProgress, setDeployProgress] = useState({ current: 0, total: 0 });

  const handleTestConnection = async () => {
    setTesting(true);
    setConnectionError("");
    setConnectionOk(false);

    try {
      const result = await testShopifyConnection(shopName, accessToken);
      if (result.ok) {
        setConnectionOk(true);
        setConnectedShopName(result.shop_name || shopName);
      } else {
        setConnectionError(result.error || "Connection failed");
      }
    } catch {
      setConnectionError("Failed to test connection");
    } finally {
      setTesting(false);
    }
  };

  const toggleTool = (toolId: string) => {
    setSelectedTools((prev) => {
      const next = new Set(prev);
      if (next.has(toolId)) {
        next.delete(toolId);
      } else {
        next.add(toolId);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (selectedTools.size === shopifyPack.templates.length) {
      setSelectedTools(new Set());
    } else {
      setSelectedTools(new Set(shopifyPack.templates.map((t) => t.id)));
    }
  };

  const handleDeploy = async () => {
    const tools = shopifyPack.templates.filter((t) => selectedTools.has(t.id));
    setDeploying(true);
    setDeployProgress({ current: 0, total: tools.length });

    const sharedConfig: Record<string, string> = {
      shop_name: shopName,
      api_access_token: accessToken,
    };

    for (let i = 0; i < tools.length; i++) {
      setDeployProgress({ current: i + 1, total: tools.length });
      const flow = buildFlowFromTemplate(
        tools[i],
        sharedConfig,
        shopifyPack.sharedConfig,
        shopifyPack.id,
      );
      try {
        await createFlow(flow);
      } catch {
        // continue deploying remaining tools
      }
    }

    setDeploying(false);
    onComplete({ toolCount: tools.length });
  };

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
      <div className="max-w-[620px] w-full bg-card border rounded-xl p-10 mx-4">
        <div className="flex items-center gap-2 mb-6">
          <div className="flex gap-1">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`h-1.5 w-8 rounded-full ${s <= step ? "bg-primary" : "bg-muted"}`}
              />
            ))}
          </div>
          <span className="text-xs text-muted-foreground ml-2">
            Step {step} of 3
          </span>
        </div>

        {step === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">Connect Shopify</h2>
              <p className="text-sm text-muted-foreground">
                Enter your store credentials to connect
              </p>
            </div>

            <div className="space-y-4">
              <div>
                <Label className="text-sm mb-1.5 block">Store Name</Label>
                <div className="flex items-center gap-0">
                  <Input
                    value={shopName}
                    onChange={(e) => {
                      setShopName(e.target.value);
                      setConnectionOk(false);
                    }}
                    placeholder="mystore"
                    className="rounded-r-none"
                  />
                  <span className="inline-flex items-center px-3 h-9 border border-l-0 rounded-r-md bg-muted text-sm text-muted-foreground">
                    .myshopify.com
                  </span>
                </div>
              </div>

              <div>
                <Label className="text-sm mb-1.5 block">
                  Admin API Access Token
                </Label>
                <Input
                  type="password"
                  value={accessToken}
                  onChange={(e) => {
                    setAccessToken(e.target.value);
                    setConnectionOk(false);
                  }}
                  placeholder="shpat_xxxxxxxxxxxxx"
                />
              </div>
            </div>

            <button
              onClick={() => setHelpExpanded(!helpExpanded)}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {helpExpanded ? (
                <ChevronUp className="h-3 w-3" />
              ) : (
                <ChevronDown className="h-3 w-3" />
              )}
              How to create a Custom App
            </button>

            {helpExpanded && (
              <div className="text-xs text-muted-foreground space-y-1 bg-muted/50 rounded-lg p-4">
                <p>1. In your Shopify admin, go to <strong>Settings</strong> &gt; <strong>Apps and sales channels</strong></p>
                <p>2. Click <strong>Develop apps</strong> (you may need to enable developer mode first)</p>
                <p>3. Click <strong>Create an app</strong> and give it a name</p>
                <p>4. Go to <strong>Configuration</strong> &gt; <strong>Admin API integration</strong></p>
                <p>5. Select scopes: <code>read_orders</code>, <code>read_products</code>, <code>read_customers</code>, <code>read_inventory</code></p>
                <p>6. Click <strong>Install app</strong> and copy the <strong>Admin API access token</strong></p>
              </div>
            )}

            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                onClick={handleTestConnection}
                disabled={!shopName || !accessToken || testing}
              >
                {testing ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Testing...
                  </>
                ) : (
                  "Test connection"
                )}
              </Button>
              {connectionOk && (
                <div className="flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm text-green-600">
                    Connected to {connectedShopName}
                  </span>
                </div>
              )}
            </div>

            {connectionError && (
              <p className="text-sm text-destructive">{connectionError}</p>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={onBack}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button
                onClick={() => setStep(2)}
                disabled={!connectionOk}
              >
                Next
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">Choose tools</h2>
              <p className="text-sm text-muted-foreground">
                Select which Shopify tools to deploy as MCP tools
              </p>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ShoppingBag className="h-5 w-5 text-muted-foreground" />
                <span className="text-sm font-medium">Shopify Tools</span>
              </div>
              <button
                onClick={toggleAll}
                className="text-xs text-muted-foreground hover:text-foreground"
              >
                {selectedTools.size === shopifyPack.templates.length
                  ? "Deselect all"
                  : "Select all"}
              </button>
            </div>

            <div className="space-y-4">
              {TOOL_GROUPS.map((group) => (
                <div key={group.label}>
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">
                    {group.label}
                  </p>
                  <div className="space-y-1">
                    {shopifyPack.templates
                      .filter((t) => group.ids.includes(t.id))
                      .map((t) => (
                        <label
                          key={t.id}
                          className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-muted/50 cursor-pointer"
                        >
                          <Checkbox
                            checked={selectedTools.has(t.id)}
                            onCheckedChange={() => toggleTool(t.id)}
                          />
                          <div className="min-w-0">
                            <p className="text-sm font-medium truncate">
                              {t.name}
                            </p>
                            <p className="text-xs text-muted-foreground truncate">
                              {t.description}
                            </p>
                          </div>
                        </label>
                      ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(1)}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button
                onClick={() => setStep(3)}
                disabled={selectedTools.size === 0}
              >
                Next
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-1">Deploy tools</h2>
              <p className="text-sm text-muted-foreground">
                Deploy {selectedTools.size} Shopify tool
                {selectedTools.size !== 1 ? "s" : ""} to your MCP endpoint
              </p>
            </div>

            <div className="bg-muted/50 rounded-lg p-4 space-y-2">
              <div className="flex items-center gap-2">
                <ShoppingBag className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">{connectedShopName}</span>
              </div>
              <p className="text-xs text-muted-foreground">
                {selectedTools.size} tool{selectedTools.size !== 1 ? "s" : ""} selected
              </p>
            </div>

            {deploying && (
              <div className="flex items-center gap-3">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span className="text-sm">
                  Deploying {deployProgress.current}/{deployProgress.total}...
                </span>
              </div>
            )}

            <div className="flex justify-between pt-2">
              <Button variant="ghost" onClick={() => setStep(2)} disabled={deploying}>
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>
              <Button onClick={handleDeploy} disabled={deploying}>
                {deploying ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Deploying...
                  </>
                ) : (
                  <>
                    Deploy {selectedTools.size} tool
                    {selectedTools.size !== 1 ? "s" : ""}
                  </>
                )}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
