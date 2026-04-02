import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { BufferForm } from "@/components/buffer-form/buffer-form";
import { createBuffer } from "@/lib/api";

export default function NewBufferPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSaveBuffer = async (data: {
    label: string;
    component: string;
    config: any;
  }) => {
    setIsSubmitting(true);
    try {
      await createBuffer(data);

      addToast({
        id: "buffer-created",
        title: "Buffer Created",
        description: `Buffer "${data.label}" has been created successfully.`,
        variant: "success",
      });

      navigate("/buffers");
    } catch (error) {
      console.error("Error creating buffer:", error);
      addToast({
        id: "buffer-creation-error",
        title: "Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to create buffer.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    navigate("/buffers");
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Add New Buffer</h1>
        <p className="text-muted-foreground">
          Configure a buffer resource for your data processing pipelines
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <BufferForm onSubmit={handleSaveBuffer} onCancel={handleCancel} />
      )}
    </div>
  );
}
