import { useEffect, useState, useMemo, useCallback, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { registerBloblangLanguage } from "@/lib/bloblang-language";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from "@/components/ui/resizable";
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
import {
  Plus,
  ChevronRight,
  ChevronDown,
  Folder,
  FolderOpen,
  FileText,
  Trash2,
  Save,
  FilePlus,
} from "lucide-react";
import { useToast } from "@/components/toast";
import { FileEntry } from "@/lib/entities";
import { fetchFiles, fetchFile, createFile, updateFile, deleteFile } from "@/lib/api";
import { cn } from "@/lib/utils";

type TreeNode = {
  name: string;
  path: string;
  isFolder: boolean;
  children: TreeNode[];
  file?: FileEntry;
};

function buildTree(files: FileEntry[]): TreeNode[] {
  const root: TreeNode[] = [];

  for (const file of files) {
    const parts = file.key.split("/");
    let current = root;

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      const path = parts.slice(0, i + 1).join("/");
      const isLast = i === parts.length - 1;

      let existing = current.find((n) => n.name === part && n.isFolder === !isLast);

      if (!existing) {
        existing = {
          name: part,
          path,
          isFolder: !isLast,
          children: [],
          file: isLast ? file : undefined,
        };
        current.push(existing);
      }

      if (!isLast) {
        current = existing.children;
      }
    }
  }

  const sortNodes = (nodes: TreeNode[]): TreeNode[] => {
    return nodes.sort((a, b) => {
      if (a.isFolder && !b.isFolder) return -1;
      if (!a.isFolder && b.isFolder) return 1;
      return a.name.localeCompare(b.name);
    }).map((n) => ({
      ...n,
      children: sortNodes(n.children),
    }));
  };

  return sortNodes(root);
}

function TreeItem({
  node,
  depth,
  selectedFileId,
  expandedFolders,
  onSelectFile,
  onToggleFolder,
  onCreateInFolder,
}: {
  node: TreeNode;
  depth: number;
  selectedFileId: string | null;
  expandedFolders: Set<string>;
  onSelectFile: (file: FileEntry) => void;
  onToggleFolder: (path: string) => void;
  onCreateInFolder: (folderPath: string) => void;
}) {
  const isExpanded = expandedFolders.has(node.path);
  const isSelected = node.file && node.file.id === selectedFileId;

  if (node.isFolder) {
    return (
      <div>
        <div
          className={cn(
            "group flex items-center gap-1 w-full text-left py-1 px-2 text-sm hover:bg-accent rounded-sm cursor-pointer",
          )}
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => onToggleFolder(node.path)}
        >
          {isExpanded ? (
            <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          )}
          {isExpanded ? (
            <FolderOpen className="h-4 w-4 shrink-0 text-blue-500" />
          ) : (
            <Folder className="h-4 w-4 shrink-0 text-blue-500" />
          )}
          <span className="truncate flex-1">{node.name}</span>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onCreateInFolder(node.path);
            }}
            className="hidden group-hover:flex h-5 w-5 items-center justify-center rounded-sm hover:bg-accent-foreground/10 shrink-0"
            title={`New file in ${node.name}`}
          >
            <Plus className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
        {isExpanded && (
          <div>
            {node.children.map((child) => (
              <TreeItem
                key={child.path}
                node={child}
                depth={depth + 1}
                selectedFileId={selectedFileId}
                expandedFolders={expandedFolders}
                onSelectFile={onSelectFile}
                onToggleFolder={onToggleFolder}
                onCreateInFolder={onCreateInFolder}
              />
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <button
      onClick={() => node.file && onSelectFile(node.file)}
      className={cn(
        "flex items-center gap-1 w-full text-left py-1 px-2 text-sm rounded-sm",
        isSelected
          ? "bg-accent text-accent-foreground"
          : "hover:bg-accent/50",
      )}
      style={{ paddingLeft: `${depth * 16 + 8 + 18}px` }}
    >
      <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
      <span className="truncate">{node.name}</span>
    </button>
  );
}

function getLanguageFromKey(key: string): string {
  const ext = key.split(".").pop()?.toLowerCase();
  const map: Record<string, string> = {
    json: "json",
    js: "javascript",
    ts: "typescript",
    py: "python",
    rb: "ruby",
    go: "go",
    rs: "rust",
    yaml: "yaml",
    yml: "yaml",
    xml: "xml",
    html: "html",
    css: "css",
    sql: "sql",
    sh: "shell",
    bash: "shell",
    md: "markdown",
    toml: "ini",
    ini: "ini",
    csv: "plaintext",
    txt: "plaintext",
    blobl: "bloblang",
  };
  return ext ? map[ext] || "plaintext" : "plaintext";
}

function formatSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
}

export default function FilesPage() {
  const { addToast } = useToast();
  const { resolvedTheme } = useTheme();
  const [searchParams, setSearchParams] = useSearchParams();
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<FileEntry | null>(null);
  const [isNewFile, setIsNewFile] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [loadingFile, setLoadingFile] = useState(false);
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [formData, setFormData] = useState({ key: "", content: "" });
  const [hasChanges, setHasChanges] = useState(false);
  const initialFileLoaded = useRef(false);

  const tree = useMemo(() => buildTree(files), [files]);
  const editorLanguage = useMemo(() => getLanguageFromKey(formData.key), [formData.key]);

  const expandFoldersForKey = useCallback((key: string) => {
    const parts = key.split("/");
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      for (let i = 1; i < parts.length; i++) {
        next.add(parts.slice(0, i).join("/"));
      }
      return next;
    });
  }, []);

  const loadFiles = useCallback(async () => {
    try {
      setLoading(true);
      const data = await fetchFiles();
      setFiles(data);
      return data;
    } catch {
      addToast({
        id: "files-load-error",
        title: "Error",
        description: "Failed to load files.",
        variant: "error",
      });
      return [];
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadFiles().then((data) => {
      if (initialFileLoaded.current) return;
      initialFileLoaded.current = true;

      const fileId = searchParams.get("file");
      if (fileId && data.length > 0) {
        const match = data.find((f) => f.id === fileId);
        if (match) {
          expandFoldersForKey(match.key);
          handleSelectFile(match);
        }
      }
    });
  }, [loadFiles]);

  const handleSelectFile = async (file: FileEntry) => {
    setIsNewFile(false);
    setLoadingFile(true);
    setSearchParams({ file: file.id }, { replace: true });
    try {
      const fullFile = await fetchFile(file.id);
      setSelectedFile(fullFile);
      setFormData({ key: fullFile.key, content: fullFile.content || "" });
      setHasChanges(false);
    } catch {
      addToast({
        id: "file-load-error",
        title: "Error",
        description: "Failed to load file content.",
        variant: "error",
      });
    } finally {
      setLoadingFile(false);
    }
  };

  const handleNewFile = () => {
    setSelectedFile(null);
    setIsNewFile(true);
    setFormData({ key: "", content: "" });
    setHasChanges(false);
    setSearchParams({ new: "" }, { replace: true });
  };

  const handleNewFileInFolder = (folderPath: string) => {
    setSelectedFile(null);
    setIsNewFile(true);
    setFormData({ key: `${folderPath}/`, content: "" });
    setHasChanges(false);
    setSearchParams({ new: "" }, { replace: true });
    expandFoldersForKey(`${folderPath}/placeholder`);
  };

  const handleToggleFolder = (path: string) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const handleFormChange = (field: "key" | "content", value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const handleSave = async () => {
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

    const hasConflict = files.some((f) => {
      if (!isNewFile && selectedFile?.id === f.id) return false;
      return f.key.startsWith(key + "/") || key.startsWith(f.key + "/");
    });
    if (hasConflict) {
      addToast({
        id: "validation-error",
        title: "Validation Error",
        description: `"${key}" conflicts with an existing file or folder of the same name.`,
        variant: "error",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      if (isNewFile) {
        await createFile({
          key: formData.key.trim(),
          content: formData.content,
        });
        addToast({
          id: "file-created",
          title: "File Created",
          description: `File "${formData.key}" has been created.`,
          variant: "success",
        });
      } else if (selectedFile) {
        await updateFile(selectedFile.id, {
          key: formData.key.trim(),
          content: formData.content,
        });
        addToast({
          id: "file-updated",
          title: "File Updated",
          description: `File "${formData.key}" has been updated.`,
          variant: "success",
        });
      }

      setHasChanges(false);
      setIsNewFile(false);
      const refreshed = await fetchFiles();
      setFiles(refreshed);

      const saved = refreshed.find((f) => f.key === formData.key.trim());
      if (saved) {
        const fullFile = await fetchFile(saved.id);
        setSelectedFile(fullFile);
        setFormData({ key: fullFile.key, content: fullFile.content || "" });
        setSearchParams({ file: saved.id }, { replace: true });
        expandFoldersForKey(saved.key);
      }
    } catch (error) {
      addToast({
        id: "file-save-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = () => {
    setDeleteConfirmOpen(true);
  };

  const confirmDelete = async () => {
    if (!selectedFile) return;
    try {
      await deleteFile(selectedFile.id);
      addToast({
        id: "file-deleted",
        title: "File Deleted",
        description: `File "${selectedFile.key}" has been deleted.`,
        variant: "success",
      });
      setSelectedFile(null);
      setIsNewFile(false);
      setFormData({ key: "", content: "" });
      setHasChanges(false);
      setSearchParams({}, { replace: true });
      await loadFiles();
    } catch (error) {
      addToast({
        id: "file-delete-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setDeleteConfirmOpen(false);
    }
  };

  return (
    <div className="h-[calc(100vh-theme(spacing.20))]">
      <ResizablePanelGroup direction="horizontal">
        <ResizablePanel defaultSize={25} minSize={15} maxSize={40}>
          <div className="flex flex-col h-full border-r">
            <div className="flex items-center justify-between p-3 border-b">
              <h2 className="text-sm font-semibold">Files</h2>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={handleNewFile}
                title="New File"
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
            <ScrollArea className="flex-1">
              <div className="py-1">
                {loading ? (
                  <p className="text-sm text-muted-foreground px-3 py-2">
                    Loading...
                  </p>
                ) : files.length === 0 ? (
                  <div className="text-center py-8 px-3">
                    <FileText className="h-8 w-8 mx-auto text-muted-foreground mb-2" />
                    <p className="text-sm text-muted-foreground">No files yet</p>
                    <Button
                      variant="link"
                      size="sm"
                      className="mt-1"
                      onClick={handleNewFile}
                    >
                      Create your first file
                    </Button>
                  </div>
                ) : (
                  tree.map((node) => (
                    <TreeItem
                      key={node.path}
                      node={node}
                      depth={0}
                      selectedFileId={selectedFile?.id ?? null}
                      expandedFolders={expandedFolders}
                      onSelectFile={handleSelectFile}
                      onToggleFolder={handleToggleFolder}
                      onCreateInFolder={handleNewFileInFolder}
                    />
                  ))
                )}
              </div>
            </ScrollArea>
          </div>
        </ResizablePanel>

        <ResizableHandle />

        <ResizablePanel defaultSize={75}>
          <div className="flex flex-col h-full">
            {!selectedFile && !isNewFile ? (
              <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                <FilePlus className="h-12 w-12 mb-3" />
                <p className="text-sm">Select a file or create a new one</p>
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3"
                  onClick={handleNewFile}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  New File
                </Button>
              </div>
            ) : loadingFile ? (
              <div className="flex items-center justify-center h-full">
                <p className="text-sm text-muted-foreground">Loading file...</p>
              </div>
            ) : (
              <>
                <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                    {isNewFile ? (
                      <span className="text-sm font-medium text-muted-foreground italic">
                        New File
                      </span>
                    ) : (
                      <span className="text-sm font-mono truncate">
                        {selectedFile?.key}
                        {selectedFile && (
                          <span className="text-muted-foreground ml-2 font-sans text-xs">
                            ({formatSize(selectedFile.size)})
                          </span>
                        )}
                      </span>
                    )}
                    {hasChanges && (
                      <span className="text-xs text-orange-500 font-medium">
                        (unsaved)
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-1">
                    {!isNewFile && selectedFile && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-muted-foreground hover:text-destructive"
                        onClick={handleDelete}
                        title="Delete File"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                    <Button
                      size="sm"
                      onClick={handleSave}
                      disabled={isSubmitting || (!hasChanges && !isNewFile)}
                      className="h-7"
                    >
                      <Save className="mr-1.5 h-3.5 w-3.5" />
                      {isSubmitting
                        ? "Saving..."
                        : isNewFile
                          ? "Create"
                          : "Save"}
                    </Button>
                  </div>
                </div>

                <div className="px-4 py-3 border-b">
                  <Label htmlFor="file-key" className="text-xs text-muted-foreground">
                    Key (path)
                  </Label>
                  <Input
                    id="file-key"
                    type="text"
                    placeholder="e.g., schemas/order.json"
                    value={formData.key}
                    onChange={(e) => handleFormChange("key", e.target.value)}
                    className="font-mono mt-1 h-8 text-sm"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Use forward slashes to organize files into directories (e.g.,{" "}
                    <code className="bg-muted px-1 py-0.5 rounded">
                      schemas/order.json
                    </code>
                    ).
                  </p>
                </div>

                <div className="flex-1 min-h-0">
                  <Editor
                    value={formData.content}
                    language={editorLanguage}
                    theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
                    onChange={(value) => handleFormChange("content", value ?? "")}
                    beforeMount={registerBloblangLanguage}
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
              </>
            )}
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete File</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{selectedFile?.key}"?
              <br />
              <br />
              <strong>Warning:</strong> Flows referencing this file will fail
              if the file is deleted.
              <br />
              <br />
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
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
