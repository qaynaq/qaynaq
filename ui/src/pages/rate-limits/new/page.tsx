import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import {
  ResourceForm,
  type ResourceFormSubmit,
} from "@/components/resource-form/resource-form";
import { createRateLimit } from "@/lib/api";

export default function NewRateLimitPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSave = async (data: ResourceFormSubmit) => {
    setIsSubmitting(true);
    try {
      await createRateLimit(data);
      addToast({
        id: "rate-limit-created",
        title: "Rate Limit Created",
        description: `Rate limit "${data.label}" has been created successfully.`,
        variant: "success",
      });
      navigate("/rate-limits");
    } catch (error) {
      console.error("Error creating rate limit:", error);
      addToast({
        id: "rate-limit-creation-error",
        title: "Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to create rate limit.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Add New Rate Limit</h1>
        <p className="text-muted-foreground">
          Configure a rate limit resource for your data processing pipelines
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <ResourceForm
          category="rate_limit"
          resourceLabel="Rate Limit"
          onSubmit={handleSave}
          onCancel={() => navigate("/rate-limits")}
        />
      )}
    </div>
  );
}
