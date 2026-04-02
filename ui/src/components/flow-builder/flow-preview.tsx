import { useMemo, useState, useEffect } from "react";
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  type Node,
  type Edge,
  type ColorMode,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { v4 as uuidv4 } from "uuid";
import * as yaml from "js-yaml";
import { Flow } from "@/lib/entities";
import {
  componentSchemas as rawComponentSchemas,
  componentLists,
} from "@/lib/component-schemas";
import type { AllComponentSchemas } from "./node-config-panel";

import { InputNode } from "./nodes/input-node";
import { ProcessorNode } from "./nodes/processor-node";
import { OutputNode } from "./nodes/output-node";
import { BranchGroupNode } from "./nodes/branch-group-node";
import { CatchGroupNode } from "./nodes/catch-group-node";
import { ChildProcessorNode } from "./nodes/child-processor-node";
import { BrokerGroupNode } from "./nodes/broker-group-node";
import { BrokerInputGroupNode } from "./nodes/broker-input-group-node";
import { ChildOutputNode } from "./nodes/child-output-node";
import { ChildInputNode } from "./nodes/child-input-node";
import { SwitchGroupNode } from "./nodes/switch-group-node";
import { SwitchProcessorGroupNode } from "./nodes/switch-processor-group-node";
import { SwitchCaseStartNode } from "./nodes/switch-case-start-node";
import { ChildCaseNode } from "./nodes/child-case-node";
import { PipelineEdge } from "./edges/pipeline-edge";

const nodeTypes = {
  inputNode: InputNode,
  processorNode: ProcessorNode,
  outputNode: OutputNode,
  branchGroupNode: BranchGroupNode,
  catchGroupNode: CatchGroupNode,
  childProcessorNode: ChildProcessorNode,
  brokerGroupNode: BrokerGroupNode,
  brokerInputGroupNode: BrokerInputGroupNode,
  childOutputNode: ChildOutputNode,
  childInputNode: ChildInputNode,
  switchGroupNode: SwitchGroupNode,
  switchProcessorGroupNode: SwitchProcessorGroupNode,
  switchCaseStartNode: SwitchCaseStartNode,
  childCaseNode: ChildCaseNode,
};

const edgeTypes = {
  pipeline: PipelineEdge,
};

const NODE_SPACING_X = 350;
const NODE_Y = 200;
const GROUP_WIDTH = 220;
const CHILD_NODE_WIDTH = 160;
const CHILD_NODE_HEIGHT = 72;
const CHILD_GAP_Y = 35;
const CHILD_X = Math.round((GROUP_WIDTH - CHILD_NODE_WIDTH) / 2);
const GROUP_BOTTOM_PAD = 50;


const CATCH_CHILD_Y_START = 68;
const CATCH_GROUP_MIN_HEIGHT = 140;

function calcCatchGroupHeight(childCount: number): number {
  if (childCount === 0) return CATCH_GROUP_MIN_HEIGHT;
  return CATCH_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

const BROKER_CHILD_Y_START = 96;
const BROKER_GROUP_MIN_HEIGHT = 170;

function calcBrokerGroupHeight(childCount: number): number {
  if (childCount === 0) return BROKER_GROUP_MIN_HEIGHT;
  return BROKER_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

const BROKER_INPUT_CHILD_Y_START = 68;
const BROKER_INPUT_GROUP_MIN_HEIGHT = 140;

function calcBrokerInputGroupHeight(childCount: number): number {
  if (childCount === 0) return BROKER_INPUT_GROUP_MIN_HEIGHT;
  return BROKER_INPUT_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

const SWITCH_CHILD_Y_START = 68;
const SWITCH_GROUP_MIN_HEIGHT = 140;

function calcSwitchGroupHeight(childCount: number): number {
  if (childCount === 0) return SWITCH_GROUP_MIN_HEIGHT;
  return SWITCH_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

const transformComponentSchemas = (): AllComponentSchemas => {
  const allSchemas: AllComponentSchemas = {
    input: [],
    processor: [],
    output: [],
  };

  for (const typeKey of ["input", "pipeline", "output"] as const) {
    const list = componentLists[typeKey] || [];
    const targetTypeForApp = typeKey === "pipeline" ? "processor" : typeKey;

    let schemaCategory:
      | typeof rawComponentSchemas.input
      | typeof rawComponentSchemas.pipeline
      | typeof rawComponentSchemas.output
      | undefined;
    if (typeKey === "input") schemaCategory = rawComponentSchemas.input;
    else if (typeKey === "pipeline") schemaCategory = rawComponentSchemas.pipeline;
    else if (typeKey === "output") schemaCategory = rawComponentSchemas.output;

    list.forEach((componentName: string) => {
      const rawSchema = schemaCategory?.[componentName as keyof typeof schemaCategory];
      if (rawSchema) {
        allSchemas[targetTypeForApp].push({
          id: componentName,
          name: (rawSchema as any).title || componentName,
          component: componentName,
          type: targetTypeForApp,
          schema: rawSchema,
        });
      }
    });
  }
  return allSchemas;
};

interface FlowPreviewProps {
  flow: Flow;
}

function FlowPreviewContent({ flow }: FlowPreviewProps) {
  const componentSchemas = useMemo(() => transformComponentSchemas(), []);
  const [colorMode, setColorMode] = useState<ColorMode>("light");

  useEffect(() => {
    const isDark = document.documentElement.classList.contains("dark");
    setColorMode(isDark ? "dark" : "light");
    const observer = new MutationObserver(() => {
      setColorMode(document.documentElement.classList.contains("dark") ? "dark" : "light");
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
    return () => observer.disconnect();
  }, []);

  const getComponentDisplayName = (componentId: string, type: "input" | "processor" | "output"): string => {
    const component = componentSchemas[type].find((c) => c.id === componentId);
    return component ? `${component.name} (${component.component})` : componentId;
  };

  const { nodes, edges } = useMemo(() => {
    const flowNodes: Node[] = [];
    const flowEdges: Edge[] = [];
    let xPos = 0;
    let prevNodeIds: string[] = [];

    const connectFromPrev = (targetId: string) => {
      for (const pid of prevNodeIds) {
        flowEdges.push({
          id: `e-${pid}-${targetId}`,
          source: pid,
          target: targetId,
          type: "pipeline",
        });
      }
    };

    const inputId = uuidv4();

    if (flow.input_component === "broker") {
      let childConfigs: any[] = [];
      if (flow.input_config?.trim()) {
        try {
          const brokerConfig = yaml.load(flow.input_config) as any;
          if (brokerConfig && Array.isArray(brokerConfig.inputs)) childConfigs = brokerConfig.inputs;
        } catch {}
      }

      const childCount = childConfigs.length;
      const groupHeight = calcBrokerInputGroupHeight(childCount);

      flowNodes.push({
        id: inputId,
        type: "brokerInputGroupNode",
        position: { x: xPos, y: NODE_Y - 25 },
        style: { width: GROUP_WIDTH, height: groupHeight },
        data: {
          label: flow.input_label || "broker",
          type: "input",
          componentId: "broker",
          component: "broker",
          nodeId: inputId,
          isGroup: true,
          childCount,
          configYaml: "",
          readOnly: true,
        },
      });

      childConfigs.forEach((inputObj, i) => {
        const componentName = Object.keys(inputObj).find((k) => k !== "label") || Object.keys(inputObj)[0];
        const config = inputObj[componentName];
        const childLabel = inputObj.label as string | undefined;
        const schema = componentSchemas.input.find(
          (o) => o.component === componentName || o.id === componentName
        );
        const childId = uuidv4();
        flowNodes.push({
          id: childId,
          type: "childInputNode",
          position: {
            x: CHILD_X,
            y: BROKER_INPUT_CHILD_Y_START + i * (CHILD_NODE_HEIGHT + CHILD_GAP_Y),
          },
          parentId: inputId,
          extent: "parent" as const,
          data: {
            label: childLabel || schema?.id || componentName,
            type: "input",
            componentId: schema?.id || componentName,
            component: componentName,
            configYaml:
              typeof config === "string"
                ? config
                : config && Object.keys(config).length > 0
                  ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                  : "",
            nodeId: childId,
            readOnly: true,
          },
        });
      });
    } else {
      flowNodes.push({
        id: inputId,
        type: "inputNode",
        position: { x: xPos, y: NODE_Y },
        data: {
          label: flow.input_label || "Input",
          type: "input",
          componentId: flow.input_component,
          component: getComponentDisplayName(flow.input_component, "input"),
          configYaml: flow.input_config,
          nodeId: inputId,
          readOnly: true,
        },
      });
    }

    prevNodeIds = [inputId];
    xPos += NODE_SPACING_X;

    if (flow.processors && flow.processors.length > 0) {
      for (const proc of flow.processors) {
        const id = uuidv4();

        if (proc.component === "branch") {
          let branchConfig: any = {};
          if (proc.config?.trim()) {
            try { branchConfig = yaml.load(proc.config) || {}; } catch {}
          }
          const childConfigs: any[] = Array.isArray(branchConfig.processors) ? branchConfig.processors : [];

          const childCount = childConfigs.length;
          const groupHeight = calcCatchGroupHeight(childCount);

          flowNodes.push({
            id,
            type: "branchGroupNode",
            position: { x: xPos, y: NODE_Y - 25 },
            style: { width: GROUP_WIDTH, height: groupHeight },
            data: {
              label: proc.label || "branch",
              type: "processor",
              componentId: "branch",
              component: "branch",
              nodeId: id,
              isGroup: true,
              childCount,
              configYaml: "",
              readOnly: true,
            },
          });

          let prevChildId: string | null = null;
          childConfigs.forEach((procObj, i) => {
            const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0];
            const config = procObj[componentName];
            const childLabel = procObj.label as string | undefined;
            const schema = componentSchemas.processor.find(
              (p) => p.component === componentName || p.id === componentName
            );
            const childId = uuidv4();
            flowNodes.push({
              id: childId,
              type: "childProcessorNode",
              position: {
                x: CHILD_X,
                y: CATCH_CHILD_Y_START + i * (CHILD_NODE_HEIGHT + CHILD_GAP_Y),
              },
              parentId: id,
              extent: "parent" as const,
              data: {
                label: childLabel || schema?.id || componentName,
                type: "processor",
                componentId: schema?.id || componentName,
                component: componentName,
                configYaml:
                  typeof config === "string"
                    ? config
                    : config && Object.keys(config).length > 0
                      ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                      : "",
                nodeId: childId,
                readOnly: true,
              },
            });
            if (prevChildId) {
              flowEdges.push({
                id: `e-internal-${prevChildId}-${childId}`,
                source: prevChildId,
                target: childId,
                type: "pipeline",
                data: { internal: true },
              });
            }
            prevChildId = childId;
          });

          connectFromPrev(id);
          prevNodeIds = [id];
          xPos += NODE_SPACING_X;
        } else if (proc.component === "catch") {
          let childConfigs: any[] = [];
          if (proc.config?.trim()) {
            try {
              const parsed = yaml.load(proc.config);
              if (Array.isArray(parsed)) childConfigs = parsed;
            } catch {}
          }

          const childCount = childConfigs.length;
          const groupHeight = calcCatchGroupHeight(childCount);

          flowNodes.push({
            id,
            type: "catchGroupNode",
            position: { x: xPos, y: NODE_Y - 25 },
            style: { width: GROUP_WIDTH, height: groupHeight },
            data: {
              label: proc.label || "catch",
              type: "processor",
              componentId: "catch",
              component: "catch",
              nodeId: id,
              isGroup: true,
              childCount,
              configYaml: "",
              readOnly: true,
            },
          });

          let prevChildId: string | null = null;
          childConfigs.forEach((procObj, i) => {
            const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0];
            const config = procObj[componentName];
            const childLabel = procObj.label as string | undefined;
            const schema = componentSchemas.processor.find(
              (p) => p.component === componentName || p.id === componentName
            );
            const childId = uuidv4();
            flowNodes.push({
              id: childId,
              type: "childProcessorNode",
              position: {
                x: CHILD_X,
                y: CATCH_CHILD_Y_START + i * (CHILD_NODE_HEIGHT + CHILD_GAP_Y),
              },
              parentId: id,
              extent: "parent" as const,
              data: {
                label: childLabel || schema?.id || componentName,
                type: "processor",
                componentId: schema?.id || componentName,
                component: componentName,
                configYaml:
                  typeof config === "string"
                    ? config
                    : config && Object.keys(config).length > 0
                      ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                      : "",
                nodeId: childId,
                readOnly: true,
              },
            });
            if (prevChildId) {
              flowEdges.push({
                id: `e-internal-${prevChildId}-${childId}`,
                source: prevChildId,
                target: childId,
                type: "pipeline",
                data: { internal: true },
              });
            }
            prevChildId = childId;
          });

          connectFromPrev(id);
          prevNodeIds = [id];
          xPos += NODE_SPACING_X;
        } else if (proc.component === "switch") {
          let casesConfigs: any[] = [];
          if (proc.config?.trim()) {
            try {
              const parsed = yaml.load(proc.config);
              if (Array.isArray(parsed)) casesConfigs = parsed;
            } catch {}
          }

          // Decision tree: switch node → branching edges → processor chains per case
          flowNodes.push({
            id,
            type: "processorNode",
            position: { x: xPos, y: NODE_Y },
            data: {
              label: proc.label || "switch",
              type: "processor",
              componentId: "switch",
              component: "Switch",
              configYaml: "",
              nodeId: id,
              readOnly: true,
            },
          });
          connectFromPrev(id);

          const caseCount = casesConfigs.length;
          const CASE_SPACING_Y = 120;
          const startY = NODE_Y - ((caseCount - 1) * CASE_SPACING_Y) / 2;
          const lastProcIds: string[] = [];
          let maxChainX = xPos;

          casesConfigs.forEach((caseObj, caseIdx) => {
            const check = caseObj.check || "";
            const fallthrough = caseObj.fallthrough === true;
            const processors = Array.isArray(caseObj.processors) ? caseObj.processors : [];
            const chainY = startY + caseIdx * CASE_SPACING_Y;
            let chainX = xPos + NODE_SPACING_X;
            let prevChainId = id;

            const edgeLabel = check
              ? (fallthrough ? `${check} ↓` : check)
              : (fallthrough ? "default ↓" : "default");

            if (processors.length === 0) {
              lastProcIds.push(id);
              return;
            }

            processors.forEach((procObj: any, procIdx: number) => {
              const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0] || "";
              const config = procObj[componentName];
              const procLabel = procObj.label as string | undefined;
              const schema = componentSchemas.processor.find(
                (p) => p.component === componentName || p.id === componentName
              );
              const procId = uuidv4();
              flowNodes.push({
                id: procId,
                type: "processorNode",
                position: { x: chainX, y: chainY },
                data: {
                  label: procLabel || schema?.id || componentName,
                  type: "processor",
                  componentId: schema?.id || componentName,
                  component: componentName,
                  configYaml:
                    typeof config === "string"
                      ? config
                      : config && Object.keys(config).length > 0
                        ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                        : "",
                  nodeId: procId,
                  readOnly: true,
                },
              });
              flowEdges.push({
                id: `e-${prevChainId}-${procId}`,
                source: prevChainId,
                target: procId,
                type: "pipeline",
                data: procIdx === 0 ? { edgeLabel } : undefined,
              });
              prevChainId = procId;
              chainX += NODE_SPACING_X;
            });

            lastProcIds.push(prevChainId);
            maxChainX = Math.max(maxChainX, chainX);
          });

          prevNodeIds = lastProcIds.length > 0 ? lastProcIds : [id];
          xPos = maxChainX;
        } else {
          flowNodes.push({
            id,
            type: "processorNode",
            position: { x: xPos, y: NODE_Y },
            data: {
              label: proc.label || "Processor",
              type: "processor",
              componentId: proc.component,
              component: getComponentDisplayName(proc.component, "processor"),
              configYaml: proc.config,
              nodeId: id,
              readOnly: true,
            },
          });
          connectFromPrev(id);
          prevNodeIds = [id];
          xPos += NODE_SPACING_X;
        }
      }
    }

    const outputId = uuidv4();

    if (flow.output_component === "broker") {
      let brokerConfig: any = {};
      let childConfigs: any[] = [];
      let brokerPattern = "fan_out";
      if (flow.output_config?.trim()) {
        try {
          brokerConfig = yaml.load(flow.output_config) || {};
          if (Array.isArray(brokerConfig.outputs)) childConfigs = brokerConfig.outputs;
          if (brokerConfig.pattern) brokerPattern = brokerConfig.pattern;
        } catch {}
      }

      const childCount = childConfigs.length;
      const groupHeight = calcBrokerGroupHeight(childCount);

      flowNodes.push({
        id: outputId,
        type: "brokerGroupNode",
        position: { x: xPos, y: NODE_Y - 25 },
        style: { width: GROUP_WIDTH, height: groupHeight },
        data: {
          label: flow.output_label || "broker",
          type: "output",
          componentId: "broker",
          component: "broker",
          nodeId: outputId,
          isGroup: true,
          childCount,
          configYaml: "",
          brokerPattern,
          readOnly: true,
        },
      });

      let prevChildId: string | null = null;
      childConfigs.forEach((outputObj, i) => {
        const componentName = Object.keys(outputObj).find((k) => k !== "label") || Object.keys(outputObj)[0];
        const config = outputObj[componentName];
        const childLabel = outputObj.label as string | undefined;
        const schema = componentSchemas.output.find(
          (o) => o.component === componentName || o.id === componentName
        );
        const childId = uuidv4();
        flowNodes.push({
          id: childId,
          type: "childOutputNode",
          position: {
            x: CHILD_X,
            y: BROKER_CHILD_Y_START + i * (CHILD_NODE_HEIGHT + CHILD_GAP_Y),
          },
          parentId: outputId,
          extent: "parent" as const,
          data: {
            label: childLabel || schema?.id || componentName,
            type: "output",
            componentId: schema?.id || componentName,
            component: componentName,
            configYaml:
              typeof config === "string"
                ? config
                : config && Object.keys(config).length > 0
                  ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                  : "",
            nodeId: childId,
            readOnly: true,
            childIndex: i,
            parentBrokerPattern: brokerPattern,
          },
        });
        const showEdges = brokerPattern === "fan_out_sequential" || brokerPattern === "fan_out_sequential_fail_fast";
        if (prevChildId && showEdges) {
          flowEdges.push({
            id: `e-internal-${prevChildId}-${childId}`,
            source: prevChildId,
            target: childId,
            type: "pipeline",
            data: { internal: true },
          });
        }
        prevChildId = childId;
      });

      connectFromPrev(outputId);
    } else if (flow.output_component === "switch") {
      let switchConfig: any = {};
      let caseConfigs: any[] = [];
      if (flow.output_config?.trim()) {
        try {
          switchConfig = yaml.load(flow.output_config) || {};
          if (Array.isArray(switchConfig.cases)) caseConfigs = switchConfig.cases;
        } catch {}
      }

      const childCount = caseConfigs.length;
      const CASE_SPACING_Y = 120;
      const CASE_X_OFFSET = 250;

      flowNodes.push({
        id: outputId,
        type: "switchProcessorGroupNode",
        position: { x: xPos, y: NODE_Y },
        data: {
          label: flow.output_label || "switch",
          type: "output",
          componentId: "switch",
          component: "switch",
          nodeId: outputId,
          isGroup: false,
          childCount,
          configYaml: "",
          readOnly: true,
        },
      });

      const startY = NODE_Y - ((childCount - 1) * CASE_SPACING_Y) / 2;
      caseConfigs.forEach((caseObj, i) => {
        const check = caseObj.check || "";
        const cont = caseObj.continue === true;
        const outputDef = caseObj.output || {};
        const componentName = Object.keys(outputDef).find((k) => k !== "label") || Object.keys(outputDef)[0] || "";
        const config = outputDef[componentName];
        const childLabel = outputDef.label as string | undefined;
        const schema = componentSchemas.output.find(
          (o) => o.component === componentName || o.id === componentName
        );
        const childId = uuidv4();
        flowNodes.push({
          id: childId,
          type: "switchCaseStartNode",
          position: {
            x: xPos + CASE_X_OFFSET,
            y: startY + i * CASE_SPACING_Y,
          },
          data: {
            label: childLabel || schema?.id || componentName || "Output",
            type: "output",
            componentId: schema?.id || componentName,
            component: componentName,
            configYaml:
              typeof config === "string"
                ? config
                : config && Object.keys(config).length > 0
                  ? yaml.dump(config, { lineWidth: -1, noRefs: true })
                  : "",
            nodeId: childId,
            readOnly: true,
            isCaseStart: true,
            switchId: outputId,
            caseCheck: check,
            caseContinue: cont,
          },
        });
        const caseLabel = check
          ? (cont ? `${check} →` : check)
          : (cont ? "default →" : "default");
        flowEdges.push({
          id: `e-${outputId}-${childId}`,
          source: outputId,
          target: childId,
          type: "pipeline",
          data: { switchCase: true, edgeLabel: caseLabel, edgeLabelColor: "amber" },
        });
      });

      connectFromPrev(outputId);
    } else {
      flowNodes.push({
        id: outputId,
        type: "outputNode",
        position: { x: xPos, y: NODE_Y },
        data: {
          label: flow.output_label || "Output",
          type: "output",
          componentId: flow.output_component,
          component: getComponentDisplayName(flow.output_component, "output"),
          configYaml: flow.output_config,
          nodeId: outputId,
          readOnly: true,
        },
      });
      connectFromPrev(outputId);
    }

    return { nodes: flowNodes, edges: flowEdges };
  }, [flow, componentSchemas]);

  return (
    <div style={{ width: "100%", height: "500px" }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        colorMode={colorMode}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        panOnDrag={true}
        zoomOnScroll={true}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        proOptions={{ hideAttribution: true }}
      >
        <Background variant={"dots" as any} />
      </ReactFlow>
    </div>
  );
}

export default function FlowPreview({ flow }: FlowPreviewProps) {
  return (
    <ReactFlowProvider>
      <FlowPreviewContent flow={flow} />
    </ReactFlowProvider>
  );
}
