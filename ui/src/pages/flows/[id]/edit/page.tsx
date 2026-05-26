import React, { useEffect, useState, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useToast } from "@/components/toast";
import { FlowBuilder } from "@/components/flow-builder/flow-builder";
import { fetchStream, updateFlow, validateFlow, tryFlow } from "@/lib/api";
import { getFlowCatalog, type FlowCatalog } from "@/components/flow-components/registry";

export interface StreamNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
  status?: string;
}

export default function EditStreamPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { addToast } = useToast();

  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [streamData, setStreamData] = useState<{
    name: string;
    status: string;
    bufferId?: number;
    nodes: StreamNodeData[];
    builderState?: string;
  } | null>(null);
  const [catalog, setCatalog] = useState<FlowCatalog | null>(null);
  const loadedRef = useRef(false);

  useEffect(() => {
    async function loadDataWithSchemas() {
      if (loadedRef.current) return;
      loadedRef.current = true;

      try {
        setIsLoading(true);
        
        const schemas = getFlowCatalog();
        setCatalog(schemas);
        
        const streamResponse = await fetchStream(id || "");
        
        const nameOf = (componentId: string, type: "input" | "processor" | "output"): string =>
          schemas[type].find((c) => c.id === componentId)?.name ?? componentId;

        const nodes: StreamNodeData[] = [];

        nodes.push({
          label: streamResponse.input_label || "Input",
          type: "input",
          component: nameOf(streamResponse.input_component, "input"),
          componentId: streamResponse.input_component,
          configYaml: streamResponse.input_config || "",
          status: streamResponse.status,
        });

        if (streamResponse.processors?.length) {
          for (const processor of streamResponse.processors as any[]) {
            nodes.push({
              label: processor.label || "Processor",
              type: "processor",
              component: nameOf(processor.component, "processor"),
              componentId: processor.component,
              configYaml: processor.config || "",
              status: streamResponse.status,
            });
          }
        }

        nodes.push({
          label: streamResponse.output_label || "Output",
          type: "output",
          component: nameOf(streamResponse.output_component, "output"),
          componentId: streamResponse.output_component,
          configYaml: streamResponse.output_config || "",
          status: streamResponse.status,
        });

        setStreamData({
          name: streamResponse.name,
          status: streamResponse.status,
          bufferId: streamResponse.buffer_id,
          nodes,
          builderState: streamResponse.builder_state,
        });

        setIsLoading(false);
      } catch (error) {
        console.error("Error loading data:", error);
        addToast({
          id: "fetch-error",
          title: "Error Loading Data",
          description:
            error instanceof Error
              ? error.message
              : "An unknown error occurred",
          variant: "error",
        });
        navigate("/flows");
      }
    }

    loadDataWithSchemas();
  }, [id, navigate, addToast]);

  const handleValidateStream = async (data: { name: string; status: string; bufferId?: number; nodes: StreamNodeData[] }) => {
    const inputNode = data.nodes.find((node) => node.type === "input");
    const processorNodes = data.nodes.filter((node) => node.type === "processor");
    const outputNode = data.nodes.find((node) => node.type === "output");
    if (!inputNode || !outputNode || !inputNode.componentId || !outputNode.componentId) {
      return { valid: false, error: "Flow must have an input and output with components selected." };
    }
    const inputComponent = catalog?.input.find(c => c.id === inputNode.componentId);
    const outputComponent = catalog?.output.find(c => c.id === outputNode.componentId);
    if (!inputComponent || !outputComponent) {
      return { valid: false, error: "Selected components not found in available schemas." };
    }
    return validateFlow({
      input_component: inputComponent.component,
      input_label: inputNode.label,
      input_config: inputNode.configYaml || "",
      output_component: outputComponent.component,
      output_label: outputNode.label,
      output_config: outputNode.configYaml || "",
      processors: processorNodes.map(node => {
        const comp = catalog?.processor.find(c => c.id === node.componentId);
        return { label: node.label, component: comp?.component || node.componentId || "", config: node.configYaml || "" };
      }),
    });
  };

  const handleTryStream = async (data: { processors: Array<{ label: string; component: string; config: string }>; messages: Array<{ content: string }> }) => {
    return tryFlow(data);
  };

  const handleSaveStream = async (data: { name: string; status: string; bufferId?: number; nodes: StreamNodeData[]; builderState: string; isReady: boolean }) => {
    setIsSubmitting(true);

    try {
      const inputNode = data.nodes.find((node) => node.type === "input");
      const processorNodes = data.nodes.filter((node) => node.type === "processor");
      const outputNode = data.nodes.find((node) => node.type === "output");

      if (!inputNode || !outputNode) {
        throw new Error("Flow must have at least one input and one output");
      }

      if (!inputNode.componentId || !outputNode.componentId) {
        throw new Error("Input and output nodes must have components selected");
      }

      const inputComponent = catalog?.input.find(c => c.id === inputNode.componentId);
      const outputComponent = catalog?.output.find(c => c.id === outputNode.componentId);

      if (!inputComponent || !outputComponent) {
        throw new Error("Selected components not found in available schemas");
      }

      const processors = processorNodes.map((node) => {
        if (!node.componentId) {
          throw new Error(`Processor node "${node.label}" must have a component selected`);
        }

        const processorComponent = catalog?.processor.find(c => c.id === node.componentId);
        if (!processorComponent) {
          throw new Error(`Processor component not found for node "${node.label}"`);
        }

        return {
          label: node.label,
          component: processorComponent.component,
          config: node.configYaml || ""
        };
      });

      const updatedStreamData = {
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
        processors: processors
      };

      await updateFlow(id || "", updatedStreamData);
      addToast({
        id: "stream-updated",
        title: "Flow Updated",
        description: `${data.name} has been updated successfully.`,
        variant: "success",
      });

      navigate("/flows");
    } catch (error) {
      addToast({
        id: "stream-error",
        title: "Error Updating Flow",
        description:
          error instanceof Error ? error.message : "An unknown error occurred",
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading || !catalog) {
    return (
      <div className="p-6 flex justify-center items-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit Flow</h1>
        <p className="text-muted-foreground">
          Modify your data processing pipeline visually
        </p>
      </div>

      {isSubmitting ? (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <FlowBuilder
          catalog={catalog}
          initialData={streamData!}
          onSave={handleSaveStream}
          onValidate={handleValidateStream}
          onTry={handleTryStream}
        />
      )}
    </div>
  );
}
