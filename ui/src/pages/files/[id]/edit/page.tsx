import { useEffect, useState, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchFile, updateFile } from "@/lib/api";
import { registerBloblangLanguage } from "@/lib/bloblang-language";

function getLanguageFromKey(key: string): string {
  const ext = key.split(".").pop()?.toLowerCase();
  const map: Record<string, string> = {
    json: "json", js: "javascript", ts: "typescript", py: "python",
    rb: "ruby", go: "go", rs: "rust", yaml: "yaml", yml: "yaml",
    xml: "xml", html: "html", css: "css", sql: "sql", sh: "shell",
    bash: "shell", md: "markdown", toml: "ini", ini: "ini",
    csv: "plaintext", txt: "plaintext", blobl: "bloblang",
  };
  return ext ? map[ext] || "plaintext" : "plaintext";
}

export default function FileEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToast } = useToast();
  const { resolvedTheme } = useTheme();
  const [loading, setLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    key: "",
    content: "",
  });
  const editorLanguage = useMemo(() => getLanguageFromKey(formData.key), [formData.key]);

  useEffect(() => {
    const loadFile = async () => {
      if (!id) return;
      try {
        setLoading(true);
        const file = await fetchFile(id);
        setFormData({
          key: file.key,
          content: file.content || "",
        });
      } catch (error) {
        addToast({
          id: "file-load-error",
          title: "Error Loading File",
          description:
            error instanceof Error
              ? error.message
              : "An unknown error occurred",
          variant: "error",
        });
        navigate("/files");
      } finally {
        setLoading(false);
      }
    };
    loadFile();
  }, [id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!id || !formData.key.trim()) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "File key is required.",
        variant: "error",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      await updateFile(id, {
        key: formData.key.trim(),
        content: formData.content,
      });

      addToast({
        id: "file-updated",
        title: "File Updated",
        description: `File "${formData.key}" has been updated successfully.`,
        variant: "success",
      });
      navigate("/files");
    } catch (error) {
      addToast({
        id: "file-update-error",
        title: "Error Updating File",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <p>Loading file...</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit File</h1>
        <p className="text-muted-foreground">
          Update the file content. Changes will take effect for new flow
          assignments.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>File Details</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="key">Key (path)</Label>
              <Input
                id="key"
                type="text"
                placeholder="e.g., schemas/order.json or scripts/transform.py"
                value={formData.key}
                onChange={(e) =>
                  setFormData({ ...formData, key: e.target.value })
                }
                className="font-mono"
                required
              />
              <p className="text-xs text-muted-foreground">
                Reference this file in flow configs as{" "}
                <code className="bg-muted px-1 py-0.5 rounded">
                  qaynaq://{formData.key || "your/file/path"}
                </code>
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="content">Content</Label>
              <div className="rounded-md border overflow-hidden">
                <Editor
                  value={formData.content}
                  language={editorLanguage}
                  theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
                  beforeMount={registerBloblangLanguage}
                  onChange={(value) =>
                    setFormData({ ...formData, content: value ?? "" })
                  }
                  height="400px"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: "on",
                    scrollBeyondLastLine: false,
                    wordWrap: "on",
                    tabSize: 2,
                    automaticLayout: true,
                    padding: { top: 8 },
                  }}
                />
              </div>
            </div>

            <div className="flex justify-end space-x-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate("/files")}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? "Updating..." : "Update File"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
