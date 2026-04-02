import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { CacheForm } from "@/components/cache-form/cache-form";
import { fetchCache, updateCache } from "@/lib/api";
import * as yaml from "js-yaml";

export default function EditCachePage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { addToast } = useToast();
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [cacheData, setCacheData] = useState<{
    label: string;
    component: string;
    config: any;
  } | null>(null);

  useEffect(() => {
    async function loadCache() {
      try {
        setIsLoading(true);
        const cache = await fetchCache(id || "");

        let parsedConfig = {};
        try {
          parsedConfig = yaml.load(cache.config) || {};
        } catch (error) {
          console.error("Failed to parse cache config YAML:", error);
        }

        setCacheData({
          label: cache.label,
          component: cache.component,
          config: parsedConfig,
        });
      } catch (error) {
        console.error("Error loading cache:", error);
        addToast({
          id: "fetch-error",
          title: "Error Loading Cache",
          description:
            error instanceof Error ? error.message : "An unknown error occurred",
          variant: "error",
        });
        navigate("/caches");
      } finally {
        setIsLoading(false);
      }
    }

    loadCache();
  }, [id, navigate, addToast]);

  const handleSaveCache = async (data: {
    label: string;
    component: string;
    config: any;
  }) => {
    setIsSubmitting(true);
    try {
      await updateCache(id || "", data);

      addToast({
        id: "cache-updated",
        title: "Cache Updated",
        description: `Cache "${data.label}" has been updated successfully.`,
        variant: "success",
      });

      navigate("/caches");
    } catch (error) {
      console.error("Error updating cache:", error);
      addToast({
        id: "cache-update-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to update cache.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate("/caches");
  };

  if (isLoading) {
    return (
      <div className="p-6 flex justify-center items-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit Cache</h1>
        <p className="text-muted-foreground">
          Modify your cache resource configuration
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        cacheData && (
          <CacheForm
            initialData={cacheData}
            onSubmit={handleSaveCache}
            onCancel={handleCancel}
          />
        )
      )}
    </div>
  );
}
