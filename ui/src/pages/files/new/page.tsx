import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { createFile } from "@/lib/api";
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

export default function FileNewPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const { resolvedTheme } = useTheme();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    key: "",
    content: "",
  });
  const editorLanguage = useMemo(() => getLanguageFromKey(formData.key), [formData.key]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const key = formData.key.trim();

    if (!key) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "File key is required.",
        variant: "error",
      });
      return;
    }

    if (key.endsWith("/")) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: "File key must include a filename (cannot end with '/').",
        variant: "error",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      await createFile({
        key: formData.key.trim(),
        content: formData.content,
      });

      addToast({
        id: "file-created",
        title: "File Created",
        description: `File "${formData.key}" has been created successfully.`,
        variant: "success",
      });
      navigate("/files");
    } catch (error) {
      addToast({
        id: "file-creation-error",
        title: "Error Creating File",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Create New File</h1>
        <p className="text-muted-foreground">
          Create a file that can be referenced by flow components.
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
                Use forward slashes to organize files into directories (e.g.,{" "}
                <code className="bg-muted px-1 py-0.5 rounded">
                  schemas/order.json
                </code>
                ).
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
                {isSubmitting ? "Creating..." : "Create File"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
