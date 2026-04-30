import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Check, Loader2, ShieldAlert } from "lucide-react";
import {
  fetchOAuthConsentRequest,
  decideOAuthConsent,
} from "@/lib/api";
import { OAuthConsentRequest } from "@/lib/entities";

export default function OAuthConsentPage() {
  const [searchParams] = useSearchParams();
  const requestID = searchParams.get("request_id") || "";
  const [request, setRequest] = useState<OAuthConsentRequest | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState<"allow" | "deny" | null>(null);

  useEffect(() => {
    if (!requestID) {
      setLoadError("Missing request_id in URL.");
      return;
    }
    fetchOAuthConsentRequest(requestID)
      .then(setRequest)
      .catch(() => {
        setLoadError(
          "This authorization request was not found or has expired. Start the flow again from your MCP client.",
        );
      });
  }, [requestID]);

  const decide = async (decision: "allow" | "deny") => {
    setSubmitting(decision);
    try {
      const { redirect_url } = await decideOAuthConsent(requestID, decision);
      window.location.href = redirect_url;
    } catch {
      setSubmitting(null);
      setLoadError(
        "Could not record your decision. Please try again from your MCP client.",
      );
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Authorize MCP client</CardTitle>
          <CardDescription>
            An MCP client wants to connect to this Qaynaq instance on your
            behalf.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          {loadError ? (
            <div className="flex gap-2 rounded-md border border-amber-300 bg-amber-50 dark:border-amber-700 dark:bg-amber-950 p-3 text-sm text-amber-900 dark:text-amber-100">
              <ShieldAlert className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <p>{loadError}</p>
            </div>
          ) : !request ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading request...
            </div>
          ) : (
            <>
              <div className="rounded-lg bg-muted p-4 space-y-1">
                <p className="font-medium">
                  {request.client_name || request.client_id}
                </p>
                <p className="text-xs text-muted-foreground break-all">
                  Redirect:{" "}
                  <code className="bg-background px-1 py-0.5 rounded">
                    {request.redirect_uri}
                  </code>
                </p>
              </div>

              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-2">
                  Permissions
                </h3>
                <ul className="space-y-1 text-sm">
                  <li className="flex items-start gap-2">
                    <Check className="h-4 w-4 text-green-600 dark:text-green-400 mt-0.5 flex-shrink-0" />
                    <span>Call MCP tools and read tool results as you</span>
                  </li>
                </ul>
              </div>

              <p className="text-sm text-muted-foreground">
                Signed in as <strong>{request.user_email}</strong>
              </p>

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => decide("deny")}
                  disabled={submitting !== null}
                >
                  {submitting === "deny" ? "Cancelling..." : "Cancel"}
                </Button>
                <Button
                  onClick={() => decide("allow")}
                  disabled={submitting !== null}
                >
                  {submitting === "allow" ? "Allowing..." : "Allow"}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
