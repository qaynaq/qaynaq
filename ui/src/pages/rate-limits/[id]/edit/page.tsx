import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { RateLimitForm } from "@/components/rate-limit-form/rate-limit-form";
import { fetchRateLimit, updateRateLimit } from "@/lib/api";
import { RateLimit } from "@/lib/entities";
import * as yaml from "js-yaml";

export default function EditRateLimitPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { addToast } = useToast();
  const [rateLimit, setRateLimit] = useState<RateLimit | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadRateLimit = async () => {
      if (!id) {
        setError("Invalid rate limit ID");
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        const data = await fetchRateLimit(id);
        setRateLimit(data);
        setError(null);
      } catch (err) {
        console.error("Error loading rate limit:", err);
        setError("Failed to load rate limit");
        addToast({
          id: "rate-limit-load-error",
          title: "Error",
          description:
            err instanceof Error ? err.message : "Failed to load rate limit.",
          variant: "error",
        });
      } finally {
        setIsLoading(false);
      }
    };

    loadRateLimit();
  }, [id, addToast]);

  const handleSaveRateLimit = async (data: {
    label: string;
    component: string;
    config: any;
  }) => {
    if (!id) return;

    setIsSubmitting(true);
    try {
      await updateRateLimit(id, data);

      addToast({
        id: "rate-limit-updated",
        title: "Rate Limit Updated",
        description: `Rate limit "${data.label}" has been updated successfully.`,
        variant: "success",
      });

      navigate("/rate-limits");
    } catch (error) {
      console.error("Error updating rate limit:", error);
      addToast({
        id: "rate-limit-update-error",
        title: "Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to update rate limit.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate("/rate-limits");
  };

  if (isLoading) {
    return (
      <div className="p-6">
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  if (error || !rateLimit) {
    return (
      <div className="p-6">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-red-600 mb-2">Error</h2>
          <p className="text-muted-foreground">
            {error || "Rate limit not found"}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit Rate Limit</h1>
        <p className="text-muted-foreground">
          Update the configuration for {rateLimit.label}
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <RateLimitForm
          initialData={{
            label: rateLimit.label,
            component: rateLimit.component,
            config: rateLimit.config ? yaml.load(rateLimit.config) : {},
          }}
          onSubmit={handleSaveRateLimit}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
}
