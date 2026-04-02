import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useToast } from "@/components/toast";
import { RateLimit } from "@/lib/entities";
import { fetchRateLimits, deleteRateLimit } from "@/lib/api";
import { columns } from "@/components/rate-limit-columns";

export default function RateLimitsPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [rateLimits, setRateLimits] = useState<RateLimit[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleRowClick = (rateLimit: RateLimit) => {
    navigate(`/rate-limits/${rateLimit.id}/edit`);
  };

  const handleAddNew = () => {
    navigate("/rate-limits/new");
  };

  const handleDelete = async (rateLimit: RateLimit) => {
    try {
      await deleteRateLimit(rateLimit.id);
      setRateLimits(rateLimits.filter((rl) => rl.id !== rateLimit.id));

      addToast({
        id: "rate-limit-deleted",
        title: "Rate Limit Deleted",
        description: `${rateLimit.label} has been deleted successfully.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "rate-limit-delete-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to delete rate limit.",
        variant: "error",
      });
    }
  };

  useEffect(() => {
    const loadRateLimits = async () => {
      try {
        setLoading(true);
        const data = await fetchRateLimits();
        setRateLimits(data);
        setError(null);
      } catch (err) {
        setError("Failed to fetch rate limits");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadRateLimits();
  }, []);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Rate Limits</h1>
          <p className="text-muted-foreground mt-1">
            Manage rate limit resources for your data processing pipelines
          </p>
        </div>
        <Button onClick={handleAddNew}>
          <Plus className="mr-2 h-4 w-4" />
          Add New
        </Button>
      </div>

      {loading ? (
        <p>Loading rate limits...</p>
      ) : error ? (
        <p className="text-red-500">{error}</p>
      ) : rateLimits.length === 0 ? (
        <div className="flex flex-col items-center justify-center min-h-[400px] border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg">
          <div className="text-center p-8">
            <h3 className="text-lg font-semibold mb-2">No rate limits yet</h3>
            <p className="text-muted-foreground mb-4">
              Get started by creating your first rate limit resource
            </p>
            <Button onClick={handleAddNew}>
              <Plus className="mr-2 h-4 w-4" />
              Add New Rate Limit
            </Button>
          </div>
        </div>
      ) : (
        <DataTable
          data={rateLimits}
          columns={columns()}
          onEdit={handleRowClick}
          onDelete={handleDelete}
        />
      )}
    </div>
  );
}
