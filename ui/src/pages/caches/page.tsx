import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/components/toast";
import { Cache } from "@/lib/entities";
import { fetchCaches, deleteCache } from "@/lib/api";
import { columns } from "@/components/cache-columns";

export default function CachesPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [caches, setCaches] = useState<Cache[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleRowClick = (cache: Cache) => {
    navigate(`/caches/${cache.id}/edit`);
  };

  const handleAddNew = () => {
    navigate("/caches/new");
  };

  const handleDelete = async (cache: Cache) => {
    try {
      await deleteCache(cache.id);
      setCaches(caches.filter((c) => c.id !== cache.id));

      addToast({
        id: "cache-deleted",
        title: "Cache Deleted",
        description: `${cache.label} has been deleted successfully.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "cache-delete-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to delete cache.",
        variant: "error",
      });
    }
  };

  useEffect(() => {
    const loadCaches = async () => {
      try {
        setLoading(true);
        const data = await fetchCaches();
        setCaches(data);
        setError(null);
      } catch (err) {
        setError("Failed to fetch caches");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadCaches();
  }, []);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Caches</h1>
          <p className="text-muted-foreground mt-1">
            Manage cache resources for your data processing pipelines
          </p>
        </div>
        <Button onClick={handleAddNew}>
          <Plus className="mr-2 h-4 w-4" />
          Add New
        </Button>
      </div>

      {loading ? (
        <p>Loading caches...</p>
      ) : error ? (
        <p className="text-red-500">{error}</p>
      ) : caches.length === 0 ? (
        <div className="flex flex-col items-center justify-center min-h-[400px] border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg">
          <div className="text-center p-8">
            <h3 className="text-lg font-semibold mb-2">No caches yet</h3>
            <p className="text-muted-foreground mb-4">
              Get started by creating your first cache resource
            </p>
            <Button onClick={handleAddNew}>
              <Plus className="mr-2 h-4 w-4" />
              Add New Cache
            </Button>
          </div>
        </div>
      ) : (
        <DataTable
          data={caches}
          columns={columns()}
          onEdit={handleRowClick}
          onDelete={handleDelete}
        />
      )}
    </div>
  );
}
