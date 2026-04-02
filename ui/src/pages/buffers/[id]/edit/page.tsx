import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { BufferForm } from "@/components/buffer-form/buffer-form";
import { fetchBuffer, updateBuffer } from "@/lib/api";
import { Buffer } from "@/lib/entities";
import * as yaml from "js-yaml";

export default function EditBufferPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { addToast } = useToast();
  const [buffer, setBuffer] = useState<Buffer | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadBuffer = async () => {
      if (!id) {
        setError("Invalid buffer ID");
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        const data = await fetchBuffer(id);
        setBuffer(data);
        setError(null);
      } catch (err) {
        console.error("Error loading buffer:", err);
        setError("Failed to load buffer");
        addToast({
          id: "buffer-load-error",
          title: "Error",
          description:
            err instanceof Error ? err.message : "Failed to load buffer.",
          variant: "error",
        });
      } finally {
        setIsLoading(false);
      }
    };

    loadBuffer();
  }, [id, addToast]);

  const handleSaveBuffer = async (data: {
    label: string;
    component: string;
    config: any;
  }) => {
    if (!id) return;

    setIsSubmitting(true);
    try {
      await updateBuffer(id, data);

      addToast({
        id: "buffer-updated",
        title: "Buffer Updated",
        description: `Buffer "${data.label}" has been updated successfully.`,
        variant: "success",
      });

      navigate("/buffers");
    } catch (error) {
      console.error("Error updating buffer:", error);
      addToast({
        id: "buffer-update-error",
        title: "Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to update buffer.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate("/buffers");
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

  if (error || !buffer) {
    return (
      <div className="p-6">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-red-600 mb-2">Error</h2>
          <p className="text-muted-foreground">
            {error || "Buffer not found"}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit Buffer</h1>
        <p className="text-muted-foreground">
          Update the configuration for {buffer.label}
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <BufferForm
          initialData={{
            label: buffer.label,
            component: buffer.component,
            config: buffer.config ? yaml.load(buffer.config) : {},
          }}
          onSubmit={handleSaveBuffer}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
}
