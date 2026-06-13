import { Link, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  Loader2,
  BrainCircuit,
  ArrowRight,
  ArrowLeft,
  Plus,
  Trash2,
  Workflow,
} from "lucide-react";
import { useToast } from "@/components/toast";
import { FlowBuilder } from "@/components/flow-builder/flow-builder";
import { createFlow, validateFlow, tryFlow } from "@/lib/api";
import {
  getComponent,
  getFlowCatalog,
  type FlowCatalog,
} from "@/components/flow-components/registry";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";

export interface StreamNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
  status?: string;
}

type FlowType = null | "mcp_tool" | "automation";

interface McpParameter {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

interface McpAnnotations {
  title: string;
  read_only_hint: boolean;
  destructive_hint: boolean;
  idempotent_hint: boolean;
  open_world_hint: boolean;
}

const DEFAULT_MCP_ANNOTATIONS: McpAnnotations = {
  title: "",
  read_only_hint: false,
  destructive_hint: false,
  idempotent_hint: false,
  open_world_hint: true,
};

const PARAMETER_TYPES = ["string", "number", "boolean", "array", "object"];

function FlowTypeSelector({
  onSelect,
}: {
  onSelect: (type: FlowType) => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-bold">Create New Flow</h1>
        <p className="text-muted-foreground mt-1">
          What would you like to build?
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-3xl w-full px-6">
        <button
          onClick={() => onSelect("mcp_tool")}
          className="group relative flex flex-col items-start p-8 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
        >
          <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary mb-4 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-6">
            <BrainCircuit className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-semibold mb-2">MCP Tool</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Build a tool for AI assistants. Define parameters and connect to any
            API, database, or service - instantly callable by Claude, Cursor,
            and AI agents.
          </p>
          <ArrowRight className="absolute top-8 right-6 h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors" />
        </button>

        <button
          onClick={() => onSelect("automation")}
          className="group relative flex flex-col items-start p-8 rounded-xl border-2 border-border bg-card hover:border-primary hover:shadow-lg transition-all duration-200 text-left"
        >
          <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-primary/10 text-primary mb-4 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6">
            <Workflow className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-semibold mb-2">Automation</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Move, transform, and route data between systems or orchestrate AI
            workflows that call your MCP tools. Connect 66+ sources and
            destinations with any trigger.
          </p>
          <ArrowRight className="absolute top-8 right-6 h-5 w-5 text-muted-foreground/40 group-hover:text-primary transition-colors" />
        </button>
      </div>
      <p className="text-sm text-muted-foreground mt-8">
        Looking for ready-made flows?{" "}
        <Link to="/flows/templates" className="text-primary hover:underline">
          Browse Templates
        </Link>
      </p>
    </div>
  );
}

function McpToolForm({
  onBack,
  onContinue,
}: {
  onBack: () => void;
  onContinue: (data: {
    name: string;
    description: string;
    parameters: McpParameter[];
    annotations: McpAnnotations;
  }) => void;
}) {
  const [toolName, setToolName] = useState("");
  const [description, setDescription] = useState("");
  const [parameters, setParameters] = useState<McpParameter[]>([]);
  const [annotations, setAnnotations] = useState<McpAnnotations>(
    DEFAULT_MCP_ANNOTATIONS,
  );
  const [nameError, setNameError] = useState("");

  const updateAnnotation = <K extends keyof McpAnnotations>(
    key: K,
    value: McpAnnotations[K],
  ) => {
    setAnnotations((prev) => {
      const next = { ...prev, [key]: value };
      if (key === "read_only_hint") {
        next.destructive_hint = false;
        next.idempotent_hint = false;
      }
      return next;
    });
  };

  const addParameter = () => {
    setParameters([
      ...parameters,
      { name: "", type: "string", required: false, description: "" },
    ]);
  };

  const updateParameter = (
    index: number,
    field: keyof McpParameter,
    value: any,
  ) => {
    const updated = [...parameters];
    updated[index] = { ...updated[index], [field]: value };
    setParameters(updated);
  };

  const removeParameter = (index: number) => {
    setParameters(parameters.filter((_, i) => i !== index));
  };

  const validateName = (value: string) => {
    if (value && !/^[a-zA-Z0-9_-]*$/.test(value)) {
      setNameError("Only letters, numbers, underscores, and hyphens allowed");
    } else {
      setNameError("");
    }
  };

  const canContinue =
    toolName.trim() !== "" && description.trim() !== "" && !nameError;

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <div className="w-full max-w-xl px-6">
        <button
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </button>

        <div className="mb-8">
          <h1 className="text-2xl font-bold">Configure MCP Tool</h1>
          <p className="text-muted-foreground mt-1">
            Define how AI assistants will discover and call this tool
          </p>
        </div>

        <div className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="tool-name">Tool Name</Label>
            <Input
              id="tool-name"
              value={toolName}
              onChange={(e) => {
                setToolName(e.target.value);
                validateName(e.target.value);
              }}
              placeholder="e.g. query_customers, check_inventory"
              className={nameError ? "border-destructive" : ""}
            />
            {nameError && (
              <p className="text-xs text-destructive">{nameError}</p>
            )}
            <p className="text-xs text-muted-foreground">
              The identifier AI assistants use to call this tool
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="tool-description">Description</Label>
            <Textarea
              id="tool-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="e.g. Query the customer database by name, email, or ID and return matching records"
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              Helps AI assistants understand when and how to use this tool
            </p>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Input Parameters</Label>
              <Button
                variant="outline"
                size="sm"
                onClick={addParameter}
                className="h-7 text-xs"
              >
                <Plus className="h-3 w-3 mr-1" />
                Add Parameter
              </Button>
            </div>

            {parameters.length === 0 && (
              <p className="text-sm text-muted-foreground py-4 text-center border border-dashed rounded-lg">
                No parameters defined yet. Add parameters that AI assistants
                will pass when calling this tool.
              </p>
            )}

            {parameters.map((param, index) => (
              <div
                key={index}
                className="flex items-start gap-3 p-3 rounded-lg border bg-muted/30"
              >
                <div className="flex-1 space-y-2">
                  <div className="grid grid-cols-[1fr_120px] gap-2">
                    <Input
                      value={param.name}
                      onChange={(e) =>
                        updateParameter(index, "name", e.target.value)
                      }
                      placeholder="Parameter name"
                      className="h-8 text-sm"
                    />
                    <Select
                      value={param.type}
                      onValueChange={(val) =>
                        updateParameter(index, "type", val)
                      }
                    >
                      <SelectTrigger className="h-8 text-sm">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PARAMETER_TYPES.map((t) => (
                          <SelectItem key={t} value={t}>
                            {t}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <Input
                    value={param.description}
                    onChange={(e) =>
                      updateParameter(index, "description", e.target.value)
                    }
                    placeholder="Description"
                    className="h-8 text-sm"
                  />
                  <div className="flex items-center gap-2">
                    <Checkbox
                      checked={param.required}
                      onCheckedChange={(checked) =>
                        updateParameter(index, "required", !!checked)
                      }
                      className="h-3.5 w-3.5"
                    />
                    <span className="text-xs text-muted-foreground">
                      Required
                    </span>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeParameter(index)}
                  className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>

          <div className="space-y-3 pt-2 border-t">
            <div>
              <Label>
                Annotations{" "}
                <span className="text-muted-foreground font-normal">
                  (optional)
                </span>
              </Label>
              <p className="text-xs text-muted-foreground mt-1">
                Behavioral hints that help AI clients decide whether to
                auto-approve calls. Advisory only - never used as security
                boundaries.
              </p>
            </div>

            <div className="space-y-2">
              <Label
                htmlFor="tool-display-title"
                className="text-xs font-normal text-muted-foreground"
              >
                Display Title
              </Label>
              <Input
                id="tool-display-title"
                value={annotations.title}
                onChange={(e) => updateAnnotation("title", e.target.value)}
                placeholder="Human-readable title (e.g. Customer Lookup)"
                className="h-8 text-sm"
              />
            </div>

            <div className="space-y-2 pt-1">
              <label className="flex items-start gap-2 text-sm cursor-pointer">
                <Checkbox
                  checked={annotations.read_only_hint}
                  onCheckedChange={(checked) =>
                    updateAnnotation("read_only_hint", !!checked)
                  }
                  className="mt-0.5 h-3.5 w-3.5"
                />
                <span>
                  <span className="font-medium">Read-only</span>
                  <span className="block text-xs text-muted-foreground">
                    Tool does not modify its environment
                  </span>
                </span>
              </label>

              <label className="flex items-start gap-2 text-sm cursor-pointer">
                <Checkbox
                  checked={annotations.destructive_hint}
                  onCheckedChange={(checked) =>
                    updateAnnotation("destructive_hint", !!checked)
                  }
                  className="mt-0.5 h-3.5 w-3.5"
                  disabled={annotations.read_only_hint}
                />
                <span>
                  <span className="font-medium">Destructive</span>
                  <span className="block text-xs text-muted-foreground">
                    May perform destructive updates (only meaningful when not
                    read-only)
                  </span>
                </span>
              </label>

              <label className="flex items-start gap-2 text-sm cursor-pointer">
                <Checkbox
                  checked={annotations.idempotent_hint}
                  onCheckedChange={(checked) =>
                    updateAnnotation("idempotent_hint", !!checked)
                  }
                  className="mt-0.5 h-3.5 w-3.5"
                  disabled={annotations.read_only_hint}
                />
                <span>
                  <span className="font-medium">Idempotent</span>
                  <span className="block text-xs text-muted-foreground">
                    Repeated calls with the same arguments have no additional
                    effect
                  </span>
                </span>
              </label>

              <label className="flex items-start gap-2 text-sm cursor-pointer">
                <Checkbox
                  checked={annotations.open_world_hint}
                  onCheckedChange={(checked) =>
                    updateAnnotation("open_world_hint", !!checked)
                  }
                  className="mt-0.5 h-3.5 w-3.5"
                />
                <span>
                  <span className="font-medium">Open world</span>
                  <span className="block text-xs text-muted-foreground">
                    Interacts with external entities (e.g. web search). Disable
                    for closed-domain tools
                  </span>
                </span>
              </label>
            </div>
          </div>

          <Button
            onClick={() =>
              onContinue({
                name: toolName.trim(),
                description: description.trim(),
                parameters,
                annotations: {
                  ...annotations,
                  title: annotations.title.trim(),
                },
              })
            }
            disabled={!canContinue}
            className="w-full"
            size="lg"
          >
            Continue to Flow Builder
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}
export default function NewStreamPage() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [catalog, setCatalog] =
    useState<FlowCatalog | null>(null);
  const [flowType, setFlowType] = useState<FlowType>(null);
  const [mcpInitialData, setMcpInitialData] = useState<{
    name: string;
    status: string;
    nodes: StreamNodeData[];
  } | null>(null);

  useEffect(() => {
    setCatalog(getFlowCatalog());
  }, []);

  const handleMcpContinue = (data: {
    name: string;
    description: string;
    parameters: McpParameter[];
    annotations: McpAnnotations;
  }) => {
    const mcpTool = getComponent("input", "mcp_tool");
    if (!mcpTool) {
      console.error("mcp_tool component not found in registry");
      return;
    }

    const isReadOnly = data.annotations.read_only_hint;
    const config = {
      name: data.name,
      description: data.description,
      input_schema: data.parameters,
      annotations: {
        ...(data.annotations.title ? { title: data.annotations.title } : {}),
        read_only_hint: isReadOnly,
        destructive_hint: isReadOnly ? false : data.annotations.destructive_hint,
        idempotent_hint: isReadOnly ? false : data.annotations.idempotent_hint,
        open_world_hint: data.annotations.open_world_hint,
      },
    };

    const configYaml = mcpTool.serialize(config as never);

    setMcpInitialData({
      name: data.name,
      status: "active",
      nodes: [
        {
          label: data.name,
          type: "input",
          componentId: "mcp_tool",
          component: "mcp_tool",
          configYaml,
        },
        {
          label: "mcp_tool_response",
          type: "output",
          componentId: "sync_response",
          component: "sync_response",
          configYaml: "",
        },
      ],
    });
  };

  const handleValidateStream = async (data: {
    name: string;
    status: string;
    bufferId?: number;
    nodes: StreamNodeData[];
  }) => {
    const inputNode = data.nodes.find((node) => node.type === "input");
    const processorNodes = data.nodes.filter(
      (node) => node.type === "processor",
    );
    const outputNode = data.nodes.find((node) => node.type === "output");
    if (
      !inputNode ||
      !outputNode ||
      !inputNode.componentId ||
      !outputNode.componentId
    ) {
      return {
        valid: false,
        error: "Flow must have an input and output with components selected.",
      };
    }
    const inputComponent = catalog?.input.find(
      (c) => c.id === inputNode.componentId,
    );
    const outputComponent = catalog?.output.find(
      (c) => c.id === outputNode.componentId,
    );
    if (!inputComponent || !outputComponent) {
      return {
        valid: false,
        error: "Selected components not found in available schemas.",
      };
    }
    return validateFlow({
      input_component: inputComponent.component,
      input_label: inputNode.label,
      input_config: inputNode.configYaml || "",
      output_component: outputComponent.component,
      output_label: outputNode.label,
      output_config: outputNode.configYaml || "",
      processors: processorNodes.map((node) => {
        const comp = catalog?.processor.find(
          (c) => c.id === node.componentId,
        );
        return {
          label: node.label,
          component: comp?.component || node.componentId || "",
          config: node.configYaml || "",
        };
      }),
    });
  };

  const handleTryStream = async (data: {
    processors: Array<{ label: string; component: string; config: string }>;
    messages: Array<{ content: string }>;
  }) => {
    return tryFlow(data);
  };

  const handleSaveStream = async (data: {
    name: string;
    status: string;
    bufferId?: number;
    nodes: StreamNodeData[];
    builderState: string;
    isReady: boolean;
  }) => {
    setIsSubmitting(true);
    try {
      const inputNode = data.nodes.find((node) => node.type === "input");
      const processorNodes = data.nodes.filter(
        (node) => node.type === "processor",
      );
      const outputNode = data.nodes.find((node) => node.type === "output");

      if (!inputNode || !outputNode) {
        throw new Error("Flow must have at least one input and one output");
      }

      if (!inputNode.componentId || !outputNode.componentId) {
        throw new Error("Input and output nodes must have components selected");
      }

      const inputComponent = catalog?.input.find(
        (c) => c.id === inputNode.componentId,
      );
      const outputComponent = catalog?.output.find(
        (c) => c.id === outputNode.componentId,
      );

      if (!inputComponent || !outputComponent) {
        throw new Error("Selected components not found in available schemas");
      }

      const processors = processorNodes.map((node) => {
        if (!node.componentId) {
          throw new Error(
            `Processor node "${node.label}" must have a component selected`,
          );
        }

        const processorComponent = catalog?.processor.find(
          (c) => c.id === node.componentId,
        );
        if (!processorComponent) {
          throw new Error(
            `Processor component not found for node "${node.label}"`,
          );
        }

        return {
          label: node.label,
          component: processorComponent.component,
          config: node.configYaml || "",
        };
      });

      const streamData = {
        name: data.name,
        status: data.status,
        input_component: inputComponent.component,
        input_label: inputNode.label,
        input_config: inputNode.configYaml || "",
        output_component: outputComponent.component,
        output_label: outputNode.label,
        output_config: outputNode.configYaml || "",
        buffer_id: data.bufferId,
        is_ready: data.isReady,
        builder_state: data.builderState,
        processors: processors,
      };

      await createFlow(streamData);
      addToast({
        id: "stream-created",
        title: "Flow Created",
        description: `${data.name} has been created successfully.`,
        variant: "success",
      });
      navigate("/flows");
    } catch (error) {
      console.error("Error creating stream:", error);
      addToast({
        id: "stream-creation-error",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create flow.",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!catalog) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (flowType === null) {
    return <FlowTypeSelector onSelect={setFlowType} />;
  }

  if (flowType === "mcp_tool" && !mcpInitialData) {
    return (
      <McpToolForm
        onBack={() => setFlowType(null)}
        onContinue={handleMcpContinue}
      />
    );
  }

  if (isSubmitting) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">
          {flowType === "mcp_tool" ? "Build MCP Tool" : "Add New Flow"}
        </h1>
        <p className="text-muted-foreground">
          {flowType === "mcp_tool"
            ? "Add processors to transform data between input and response"
            : "Design your data processing pipeline visually"}
        </p>
      </div>

      <FlowBuilder
        catalog={catalog}
        initialData={mcpInitialData || undefined}
        onSave={handleSaveStream}
        onValidate={handleValidateStream}
        onTry={handleTryStream}
      />
    </div>
  );
}
