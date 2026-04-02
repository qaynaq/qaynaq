import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DataTable } from "@/components/data-table";
import { Plus, KeyRound, Eye, EyeOff, Lock } from "lucide-react";
import { useToast } from "@/components/toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Secret } from "@/lib/entities";
import { fetchSecrets, deleteSecret, createSecret } from "@/lib/api";
import { useRelativeTime } from "@/lib/utils";

// Component for relative time display
const RelativeTime = ({ dateString }: { dateString: string }) => {
  const relativeTime = useRelativeTime(dateString);
  return <span>{relativeTime}</span>;
};

export default function SecretsPage() {
  const { addToast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showValue, setShowValue] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [secretToDelete, setSecretToDelete] = useState<Secret | null>(null);
  const [formData, setFormData] = useState({
    key: "",
    value: "",
  });

  const handleDelete = async (secret: Secret) => {
    if (!secret || !secret.key) {
      addToast({
        id: "secret-delete-error",
        title: "Error Deleting Secret",
        description: "Invalid secret data",
        variant: "error",
      });
      return;
    }

    setSecretToDelete(secret);
    setDeleteConfirmOpen(true);
  };

  const confirmDelete = async () => {
    if (!secretToDelete) return;

    try {
      await deleteSecret(secretToDelete.key);
      setSecrets(secrets.filter((s) => s.key !== secretToDelete.key));
      addToast({
        id: "secret-deleted",
        title: "Secret Deleted",
        description: `Secret "${secretToDelete.key}" has been deleted successfully.`,
        variant: "success",
      });
    } catch (error) {
      addToast({
        id: "secret-delete-error",
        title: "Error Deleting Secret",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setDeleteConfirmOpen(false);
      setSecretToDelete(null);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirmOpen(false);
    setSecretToDelete(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.key.trim() || !formData.value.trim()) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "Both key and value are required.",
        variant: "error",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      const newSecret = await createSecret({
        key: formData.key.trim(),
        value: formData.value.trim(),
      });

      setSecrets([...secrets, newSecret]);
      setFormData({ key: "", value: "" });
      setShowValue(false);
      setIsModalOpen(false);

      setTimeout(() => {
        addToast({
          id: "secret-created",
          title: "Secret Created",
          description: `Secret "${newSecret.key}" has been created successfully.`,
          variant: "success",
        });
      }, 100);
    } catch (error) {
      addToast({
        id: "secret-creation-error",
        title: "Error Creating Secret",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setFormData({ key: "", value: "" });
    setShowValue(false);
    setIsModalOpen(false);
  };

  const handleModalOpenChange = (open: boolean) => {
    setIsModalOpen(open);
    if (!open) {
      setFormData({ key: "", value: "" });
      setShowValue(false);
    }
  };

  const columns = [
    {
      key: "key" as keyof Secret,
      title: "Key",
      render: (value: string) => (
        <div className="flex items-center gap-2">
          <KeyRound className="h-4 w-4 text-muted-foreground" />
          <span className="font-medium">{value || "Unknown"}</span>
        </div>
      ),
    },
    {
      key: "createdAt" as keyof Secret,
      title: "Created",
      render: (value: string) =>
        value && value !== "Unknown" ? (
          <RelativeTime dateString={value} />
        ) : (
          "Unknown"
        ),
    },
  ];

  useEffect(() => {
    const loadSecrets = async () => {
      try {
        setLoading(true);
        const data = await fetchSecrets();
        const validSecrets = data.filter((secret: any) => secret && secret.key);
        setSecrets(validSecrets);
        setError(null);
      } catch (err) {
        setError("Failed to fetch secrets");
      } finally {
        setLoading(false);
      }
    };
    loadSecrets();
  }, []);

  // Auto-open create dialog if ?create=true query parameter is present
  useEffect(() => {
    if (searchParams.get("create") === "true") {
      setIsModalOpen(true);
      setSearchParams({}, { replace: true }); // Clear the query param
    }
  }, [searchParams, setSearchParams]);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Secrets</h1>
          <p className="text-muted-foreground">
            Manage your application secrets and configuration values
          </p>
        </div>
        <Dialog open={isModalOpen} onOpenChange={handleModalOpenChange}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add New Secret
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Add New Secret</DialogTitle>
            </DialogHeader>
            <form
              onSubmit={handleSubmit}
              className="space-y-4"
              autoComplete="off"
            >
              <div className="space-y-2">
                <Label htmlFor="key">Key</Label>
                <div className="relative">
                  <KeyRound className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="key"
                    name="key"
                    type="text"
                    placeholder="Enter secret key (e.g., API_KEY, DATABASE_URL)"
                    value={formData.key}
                    onChange={(e) =>
                      setFormData({ ...formData, key: e.target.value })
                    }
                    className="pl-10"
                    autoComplete="off"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="value">Value</Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="value"
                    name="value"
                    type={showValue ? "text" : "password"}
                    placeholder="Enter secret value"
                    value={formData.value}
                    onChange={(e) =>
                      setFormData({ ...formData, value: e.target.value })
                    }
                    className="pl-10 pr-10"
                    autoComplete="new-password"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowValue(!showValue)}
                    className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                  >
                    {showValue ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>

              <div className="flex justify-end space-x-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleCancel}
                  disabled={isSubmitting}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "Creating..." : "Create Secret"}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Secrets</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p>Loading secrets...</p>
          ) : error ? (
            <p className="text-red-500">{error}</p>
          ) : (
            <DataTable
              data={secrets}
              columns={columns}
              onDelete={handleDelete}
              getRowId={(secret) => secret?.key || `unknown-${Math.random()}`}
            />
          )}
        </CardContent>
      </Card>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Secret</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the secret "{secretToDelete?.key}
              "?
              <br />
              <br />
              <strong>Warning:</strong> Please ensure that no flows are using
              this secret, otherwise flows that depend on this secret might
              not work as expected.
              <br />
              <br />
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={cancelDelete}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
