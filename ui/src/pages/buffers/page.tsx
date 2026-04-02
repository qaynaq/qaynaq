import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/components/toast";
import { Buffer } from "@/lib/entities";
import { fetchBuffers, deleteBuffer } from "@/lib/api";
import { columns } from "@/components/buffer-columns";

export default function BuffersPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [buffers, setBuffers] = useState<Buffer[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleRowClick = (buffer: Buffer) => {
    navigate(`/buffers/${buffer.id}/edit`);
  };

  const handleAddNew = () => {
    navigate("/buffers/new");
  };

  const handleDelete = async (buffer: Buffer) => {
    try {
      await deleteBuffer(buffer.id);
      setBuffers(buffers.filter((b) => b.id !== buffer.id));

      addToast({
        id: "buffer-deleted",
        title: "Buffer Deleted",
        description: `${buffer.label} has been deleted successfully.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "buffer-delete-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to delete buffer.",
        variant: "error",
      });
    }
  };

  useEffect(() => {
    const loadBuffers = async () => {
      try {
        setLoading(true);
        const data = await fetchBuffers();
        setBuffers(data);
        setError(null);
      } catch (err) {
        setError("Failed to fetch buffers");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadBuffers();
  }, []);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Buffers</h1>
          <p className="text-muted-foreground mt-1">
            Manage buffer resources for your data processing pipelines
          </p>
        </div>
        <Button onClick={handleAddNew}>
          <Plus className="mr-2 h-4 w-4" />
          Add New
        </Button>
      </div>

      {loading ? (
        <p>Loading buffers...</p>
      ) : error ? (
        <p className="text-red-500">{error}</p>
      ) : buffers.length === 0 ? (
        <div className="flex flex-col items-center justify-center min-h-[400px] border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg">
          <div className="text-center p-8">
            <h3 className="text-lg font-semibold mb-2">No buffers yet</h3>
            <p className="text-muted-foreground mb-4">
              Get started by creating your first buffer resource
            </p>
            <Button onClick={handleAddNew}>
              <Plus className="mr-2 h-4 w-4" />
              Add New Buffer
            </Button>
          </div>
        </div>
      ) : (
        <DataTable
          data={buffers}
          columns={columns()}
          onEdit={handleRowClick}
          onDelete={handleDelete}
        />
      )}
    </div>
  );
}
