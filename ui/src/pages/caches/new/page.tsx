import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { CacheForm } from "@/components/cache-form/cache-form";
import { createCache } from "@/lib/api";

export default function NewCachePage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSaveCache = async (data: {
    label: string;
    component: string;
    config: any;
  }) => {
    setIsSubmitting(true);
    try {
      await createCache(data);

      addToast({
        id: "cache-created",
        title: "Cache Created",
        description: `Cache "${data.label}" has been created successfully.`,
        variant: "success",
      });

      // Navigate back to caches list
      navigate("/caches");
    } catch (error) {
      console.error("Error creating cache:", error);
      addToast({
        id: "cache-creation-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create cache.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate("/caches");
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Add New Cache</h1>
        <p className="text-muted-foreground">
          Configure a cache resource for your data processing pipelines
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <CacheForm onSubmit={handleSaveCache} onCancel={handleCancel} />
      )}
    </div>
  );
}
