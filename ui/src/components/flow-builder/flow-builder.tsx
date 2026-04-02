import { useCallback, useState, useEffect, useMemo } from "react";
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  Controls,
  Panel,
  useNodesState,
  useEdgesState,
  useReactFlow,
  addEdge,
  MarkerType,
  type Connection,
  type Node,
  type Edge,
  type ColorMode,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { v4 as uuidv4 } from "uuid";
import { Button } from "@/components/ui/button";
import {
  Save,
  ShieldCheck,
  FlaskConical,
  Loader2,
  Plus,
  X,
  XCircle,
  Download,
  Workflow,
  Upload,
} from "lucide-react";
import {
  NodeConfigPanel,
  type AllComponentSchemas,
} from "./node-config-panel";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/toast";
import { fetchBuffers } from "@/lib/api";
import type { Buffer } from "@/lib/entities";
import * as yaml from "js-yaml";

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
import { ChildCaseNode } from "./nodes/child-case-node";
import { SwitchCaseStartNode } from "./nodes/switch-case-start-node";
import { PipelineEdge } from "./edges/pipeline-edge";

// --- Exported types (unchanged for compatibility) ---
export interface FlowNodeData {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
}

export interface SectionNodeData {
  label: string;
  type: "input" | "processor" | "output";
}

export interface CustomNode {
  id: string;
  type: "input" | "processor" | "output";
  parentId: string;
  data: FlowNodeData;
}

export interface CustomSection {
  id: string;
  type: "input" | "processor" | "output";
  label: string;
  nodes: CustomNode[];
  data: SectionNodeData;
}

// --- Internal types ---
export interface StreamFlowNodeData extends Record<string, unknown> {
  label: string;
  type: "input" | "processor" | "output";
  componentId?: string;
  component?: string;
  configYaml?: string;
  disconnected?: boolean;
  readOnly?: boolean;
  nodeId: string;
  isGroup?: boolean;
  childCount?: number;
  brokerPattern?: string;
  onAddAndConnect?: (sourceNodeId: string, sourceType: string) => void;
  onAddBefore?: (targetNodeId: string) => void;
  onAddChildProcessor?: (groupId: string) => void;
  onAddChildOutput?: (groupId: string) => void;
  onAddChildInput?: (groupId: string) => void;
  onAddChildCase?: (groupId: string) => void;
  childIndex?: number;
  parentBrokerPattern?: string;
  caseFallthrough?: boolean;
  caseCheck?: string;
  isDragging?: boolean;
  isDropTarget?: boolean;
}

type ValidationResult = { valid: boolean; error?: string } | null;

type TryFlowResult = {
  outputs: Array<{ content: string }>;
  error?: string;
};

interface FlowBuilderProps {
  allComponentSchemas: AllComponentSchemas;
  initialData?: {
    name: string;
    status: string;
    bufferId?: number;
    nodes: FlowNodeData[];
    builderState?: string;
  };
  onSave: (data: {
    name: string;
    status: string;
    bufferId?: number;
    nodes: FlowNodeData[];
    builderState: string;
    isReady: boolean;
  }) => void;
  onValidate?: (data: {
    name: string;
    status: string;
    bufferId?: number;
    nodes: FlowNodeData[];
  }) => Promise<ValidationResult>;
  onTry?: (data: {
    processors: Array<{ label: string; component: string; config: string }>;
    messages: Array<{ content: string }>;
  }) => Promise<TryFlowResult>;
}

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
  childCaseNode: ChildCaseNode,
  switchCaseStartNode: SwitchCaseStartNode,
};

const edgeTypes = {
  pipeline: PipelineEdge,
};

const NODE_SPACING_X = 350;
const NODE_Y = 200;

// Shared child node dimensions
const GROUP_WIDTH = 220;
const CHILD_NODE_WIDTH = 160;
const CHILD_NODE_HEIGHT = 72;
const CHILD_GAP_Y = 35;
const CHILD_X = Math.round((GROUP_WIDTH - CHILD_NODE_WIDTH) / 2);
const GROUP_BOTTOM_PAD = 50;

// Catch group layout
const CATCH_CHILD_Y_START = 68;
const CATCH_GROUP_MIN_HEIGHT = 140;

function calcCatchGroupHeight(childCount: number): number {
  if (childCount === 0) return CATCH_GROUP_MIN_HEIGHT;
  return CATCH_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

// Broker output group layout (taller header due to pattern line)
const BROKER_CHILD_Y_START = 96;
const BROKER_GROUP_MIN_HEIGHT = 170;

function calcBrokerGroupHeight(childCount: number): number {
  if (childCount === 0) return BROKER_GROUP_MIN_HEIGHT;
  return BROKER_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

// Broker input group layout (simpler header, no pattern)
const BROKER_INPUT_CHILD_Y_START = 68;
const BROKER_INPUT_GROUP_MIN_HEIGHT = 140;

function calcBrokerInputGroupHeight(childCount: number): number {
  if (childCount === 0) return BROKER_INPUT_GROUP_MIN_HEIGHT;
  return BROKER_INPUT_CHILD_Y_START + childCount * CHILD_NODE_HEIGHT + (childCount - 1) * CHILD_GAP_Y + GROUP_BOTTOM_PAD;
}

// Switch decision tree layout (processor and output switch)
const SWITCH_CASE_SPACING_Y = 120;
const SWITCH_CASE_X_OFFSET = 250;

function layoutSwitchCases(switchX: number, switchY: number, caseCount: number): { x: number; y: number }[] {
  const startY = switchY - ((caseCount - 1) * SWITCH_CASE_SPACING_Y) / 2;
  return Array.from({ length: caseCount }, (_, i) => ({
    x: switchX + SWITCH_CASE_X_OFFSET,
    y: startY + i * SWITCH_CASE_SPACING_Y,
  }));
}

function isBrokerSequential(pattern: string | undefined): boolean {
  return pattern === "fan_out_sequential" || pattern === "fan_out_sequential_fail_fast";
}

function makeInternalEdgeData(): Record<string, unknown> {
  return { internal: true };
}

// Migrate old nested processor switch groups to decision tree layout
function migrateProcessorSwitchGroups(flowNodes: Node[], flowEdges: Edge[]): void {
  const switchGroups = flowNodes.filter(
    (n) => n.type === "switchProcessorGroupNode" && (n.data as any).isGroup === true
  );
  for (const switchNode of switchGroups) {
    // Find old case groups (processorCaseGroupNode children)
    const caseGroups = flowNodes
      .filter((n) => n.parentId === switchNode.id && (n.type === "processorCaseGroupNode" || (n.data as any).componentId === "case"))
      .sort((a, b) => a.position.y - b.position.y);

    if (caseGroups.length === 0) continue;

    // Convert switch node: remove group properties
    switchNode.data = { ...switchNode.data, isGroup: false, childCount: caseGroups.length };
    switchNode.style = undefined;
    // Restore Y position (was offset by -25 for group)
    switchNode.position = { x: switchNode.position.x, y: switchNode.position.y + 25 };

    const casePositions = layoutSwitchCases(switchNode.position.x, switchNode.position.y, caseGroups.length);
    const caseChainEndIds: string[] = [];

    for (let caseIdx = 0; caseIdx < caseGroups.length; caseIdx++) {
      const caseGroup = caseGroups[caseIdx];
      const cd = caseGroup.data as any;
      const casePos = casePositions[caseIdx];

      // Convert case group → switchCaseStartNode
      caseGroup.type = "switchCaseStartNode";
      caseGroup.position = casePos;
      caseGroup.parentId = undefined;
      caseGroup.extent = undefined;
      caseGroup.style = undefined;
      caseGroup.data = {
        ...caseGroup.data,
        isCaseStart: true,
        switchId: switchNode.id,
        isGroup: undefined,
        childCount: undefined,
        caseCheck: cd.caseCheck || "",
        caseFallthrough: cd.caseFallthrough || false,
      };

      // Add edge from switch to case start with label
      const migCheck = cd.caseCheck || "";
      const migFall = cd.caseFallthrough === true;
      const migLabel = migCheck
        ? (migFall ? `${migCheck} ↓` : migCheck)
        : (migFall ? "default ↓" : "default");
      flowEdges.push({
        id: `e-${switchNode.id}-${caseGroup.id}`,
        source: switchNode.id,
        target: caseGroup.id,
        type: "pipeline",
        data: { switchCase: true, edgeLabel: migLabel },
      });

      // Find child processors of this case group
      const caseProcs = flowNodes
        .filter((n) => n.parentId === caseGroup.id)
        .sort((a, b) => a.position.y - b.position.y);

      // Convert each processor: remove parent, set absolute position, add switchCaseId
      let prevChainId = caseGroup.id;
      let chainX = casePos.x + SWITCH_CASE_X_OFFSET;
      for (const proc of caseProcs) {
        proc.type = "processorNode";
        proc.parentId = undefined;
        proc.extent = undefined;
        proc.position = { x: chainX, y: casePos.y };
        proc.data = { ...proc.data, switchCaseId: caseGroup.id };

        // Remove old internal edges involving this processor
        for (let i = flowEdges.length - 1; i >= 0; i--) {
          if ((flowEdges[i].data as any)?.internal && (flowEdges[i].source === proc.id || flowEdges[i].target === proc.id)) {
            flowEdges.splice(i, 1);
          }
        }

        // Add chain edge
        flowEdges.push({
          id: `e-${prevChainId}-${proc.id}`,
          source: prevChainId,
          target: proc.id,
          type: "pipeline",
        });

        prevChainId = proc.id;
        chainX += NODE_SPACING_X;
      }
      caseChainEndIds.push(prevChainId);
    }

    // Replace old switch→next edge with convergence edges (case chain ends → next)
    const oldOutEdgeIdx = flowEdges.findIndex(
      (e) => e.source === switchNode.id && !(e.data as any)?.switchCase
    );
    if (oldOutEdgeIdx >= 0 && caseChainEndIds.length > 0) {
      const nextTargetId = flowEdges[oldOutEdgeIdx].target;
      flowEdges.splice(oldOutEdgeIdx, 1);
      for (const endId of caseChainEndIds) {
        flowEdges.push({
          id: `e-${endId}-${nextTargetId}`,
          source: endId,
          target: nextTargetId,
          type: "pipeline",
        });
      }
    }
  }
}

// Migrate old output switch groups (switchGroupNode + childCaseNode) to decision tree layout
function migrateOutputSwitchGroups(flowNodes: Node[], flowEdges: Edge[]): void {
  const switchGroups = flowNodes.filter((n) => n.type === "switchGroupNode");
  for (const switchNode of switchGroups) {
    const children = flowNodes
      .filter((n) => n.parentId === switchNode.id)
      .sort((a, b) => a.position.y - b.position.y);

    if (children.length === 0) {
      // Just convert the node type
      switchNode.type = "switchProcessorGroupNode";
      switchNode.data = { ...switchNode.data, isGroup: false, childCount: 0 };
      switchNode.style = undefined;
      switchNode.position = { x: switchNode.position.x, y: switchNode.position.y + 25 };
      continue;
    }

    // Convert switch node
    switchNode.type = "switchProcessorGroupNode";
    switchNode.data = { ...switchNode.data, isGroup: false, childCount: children.length };
    switchNode.style = undefined;
    switchNode.position = { x: switchNode.position.x, y: switchNode.position.y + 25 };

    const casePositions = layoutSwitchCases(switchNode.position.x, switchNode.position.y, children.length);

    for (let i = 0; i < children.length; i++) {
      const child = children[i];
      const casePos = casePositions[i];

      // Convert childCaseNode → switchCaseStartNode
      child.type = "switchCaseStartNode";
      child.position = casePos;
      child.parentId = undefined;
      child.extent = undefined;
      child.data = {
        ...child.data,
        isCaseStart: true,
        switchId: switchNode.id,
      };

      // Add edge from switch to case with label
      const cd = child.data as any;
      const check = cd.caseCheck || "";
      const cont = cd.caseContinue === true;
      const migrateLabel = check
        ? (cont ? `${check} →` : check)
        : (cont ? "default →" : "default");
      flowEdges.push({
        id: `e-${switchNode.id}-${child.id}`,
        source: switchNode.id,
        target: child.id,
        type: "pipeline",
        data: { switchCase: true, edgeLabel: migrateLabel, edgeLabelColor: "amber" },
      });
    }
  }
}

// Re-snap child positions and fix internal edges for groups
function reSnapChildren(flowNodes: Node[], flowEdges: Edge[]): void {
  const groupNodes = flowNodes.filter((n) => n.type === "branchGroupNode" || n.type === "catchGroupNode" || n.type === "brokerGroupNode" || n.type === "brokerInputGroupNode");
  for (const group of groupNodes) {
    const isBrokerOutput = group.type === "brokerGroupNode";
    const isBrokerInput = group.type === "brokerInputGroupNode";

    const yStart = isBrokerOutput ? BROKER_CHILD_Y_START : isBrokerInput ? BROKER_INPUT_CHILD_Y_START : CATCH_CHILD_Y_START;
    const children = flowNodes
      .filter((n) => n.parentId === group.id)
      .sort((a, b) => a.position.y - b.position.y);
    const pattern = isBrokerOutput ? (group.data as StreamFlowNodeData)?.brokerPattern : undefined;
    children.forEach((child, i) => {
      child.position = { x: CHILD_X, y: yStart + i * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) };
      child.data = { ...child.data, childIndex: i, ...(isBrokerOutput ? { parentBrokerPattern: pattern } : {}) };
    });
    const calcHeight = isBrokerOutput ? calcBrokerGroupHeight : isBrokerInput ? calcBrokerInputGroupHeight : calcCatchGroupHeight;
    group.style = { ...group.style, width: GROUP_WIDTH, height: calcHeight(children.length) };

    // Fix internal edges: remove all, then re-add if sequential
    const childIds = new Set(children.map((c) => c.id));
    for (let i = flowEdges.length - 1; i >= 0; i--) {
      if ((flowEdges[i].data as any)?.internal && (childIds.has(flowEdges[i].source) || childIds.has(flowEdges[i].target))) {
        flowEdges.splice(i, 1);
      }
    }
    const shouldChain = isBrokerInput ? false : isBrokerOutput ? isBrokerSequential(pattern) : true;
    if (shouldChain && children.length > 1) {
      for (let i = 0; i < children.length - 1; i++) {
        flowEdges.push({
          id: `e-internal-${children[i].id}-${children[i + 1].id}`,
          source: children[i].id,
          target: children[i + 1].id,
          type: "pipeline",
          data: { internal: true },
        });
      }
    }
  }
}

function toFlowNodeType(type: "input" | "processor" | "output"): string {
  if (type === "input") return "inputNode";
  if (type === "processor") return "processorNode";
  return "outputNode";
}

function sortCasesDefaultLast(a: Node, b: Node): number {
  const aDefault = !(a.data as any).caseCheck;
  const bDefault = !(b.data as any).caseCheck;
  if (aDefault && !bDefault) return 1;
  if (!aDefault && bDefault) return -1;
  return a.position.y - b.position.y;
}

function insertCaseBeforeDefault(existing: Node[], newCase: Node): Node[] {
  const sorted = [...existing].sort(sortCasesDefaultLast);
  const defaultIdx = sorted.findIndex((n) => !(n.data as any).caseCheck);
  if (defaultIdx >= 0) {
    sorted.splice(defaultIdx, 0, newCase);
  } else {
    sorted.push(newCase);
  }
  return sorted;
}

function resolveFlowNodeType(
  baseType: "input" | "processor" | "output",
  componentId: string
): { nodeType: string; style?: Record<string, unknown>; extraData: Record<string, unknown>; yOffset: number } {
  if (componentId === "branch" && baseType === "processor") {
    return { nodeType: "branchGroupNode", style: { width: GROUP_WIDTH, height: CATCH_GROUP_MIN_HEIGHT }, extraData: { isGroup: true, childCount: 0 }, yOffset: -25 };
  }
  if (componentId === "catch" && baseType === "processor") {
    return { nodeType: "catchGroupNode", style: { width: GROUP_WIDTH, height: CATCH_GROUP_MIN_HEIGHT }, extraData: { isGroup: true, childCount: 0 }, yOffset: -25 };
  }
  if (componentId === "switch" && baseType === "processor") {
    return { nodeType: "switchProcessorGroupNode", extraData: { isGroup: false, childCount: 0 }, yOffset: 0 };
  }
  if (componentId === "switch" && baseType === "output") {
    return { nodeType: "switchProcessorGroupNode", extraData: { isGroup: false, childCount: 0 }, yOffset: 0 };
  }
  if (componentId === "broker" && baseType === "output") {
    return { nodeType: "brokerGroupNode", style: { width: GROUP_WIDTH, height: BROKER_GROUP_MIN_HEIGHT }, extraData: { isGroup: true, childCount: 0, brokerPattern: "fan_out" }, yOffset: -25 };
  }
  if (componentId === "broker" && baseType === "input") {
    return { nodeType: "brokerInputGroupNode", style: { width: GROUP_WIDTH, height: BROKER_INPUT_GROUP_MIN_HEIGHT }, extraData: { isGroup: true, childCount: 0 }, yOffset: -25 };
  }
  return { nodeType: toFlowNodeType(baseType), extraData: {}, yOffset: 0 };
}

function FlowBuilderContent({
  allComponentSchemas,
  initialData,
  onSave,
  onValidate,
  onTry,
}: FlowBuilderProps) {
  const { addToast } = useToast();
  const { fitView, screenToFlowPosition } = useReactFlow();

  const [name, setName] = useState(initialData?.name || "");
  const [status, setStatus] = useState(initialData?.status || "active");
  const [bufferId, setBufferId] = useState<number | undefined>(initialData?.bufferId);
  const [availableBuffers, setAvailableBuffers] = useState<Buffer[]>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [editingNodeId, setEditingNodeId] = useState<string | null>(null);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; nodeId: string } | null>(null);
  const [deleteConfirmNodeId, setDeleteConfirmNodeId] = useState<string | null>(null);
  const [pendingNode, setPendingNode] = useState<{
    kind:
      | "addAndConnect"    // + button on right side of a node
      | "addBefore"        // + button on left side of a node
      | "childProcessor"   // + in catch group
      | "childOutput"      // + in broker group
      | "childInput"       // + in broker input group
      | "childCase"        // + in switch group
      | "topLevel";        // toolbar buttons
    // Context for each kind
    sourceNodeId?: string;       // addAndConnect
    targetNodeId?: string;       // addBefore
    groupId?: string;            // child* and childCase
    topLevelType?: "input" | "processor" | "output";  // topLevel
    isProcessorSwitch?: boolean; // childCase
    // Form fields
    label: string;
    componentId: string;
    component: string;
    // Case-specific fields
    caseCheck?: string;
    caseFallthrough?: boolean;
    caseContinue?: boolean;
    defaultExists?: boolean;
    // Available components for the picker
    availableComponents: Array<{ id: string; name: string; component: string }>;
  } | null>(null);
  const [isValidating, setIsValidating] = useState(false);
  const [tryDialogOpen, setTryDialogOpen] = useState(false);
  const [tryMessages, setTryMessages] = useState<string[]>([""]);
  const [tryResult, setTryResult] = useState<TryFlowResult | null>(null);
  const [isTrying, setIsTrying] = useState(false);
  const [colorMode, setColorMode] = useState<ColorMode>("light");

  useEffect(() => {
    fetchBuffers().then(setAvailableBuffers).catch(() => setAvailableBuffers([]));
  }, []);

  useEffect(() => {
    const isDark = document.documentElement.classList.contains("dark");
    setColorMode(isDark ? "dark" : "light");
    const observer = new MutationObserver(() => {
      setColorMode(document.documentElement.classList.contains("dark") ? "dark" : "light");
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
    return () => observer.disconnect();
  }, []);

  // Build initial nodes and edges from initialData (prefer saved builderState)
  const [initialFlowNodes, initialFlowEdges] = useMemo(() => {
    if (initialData?.builderState) {
      try {
        const parsed = JSON.parse(initialData.builderState);
        if (parsed.nodes && parsed.edges) {
          const restoredNodes = parsed.nodes as Node[];
          const restoredEdges = parsed.edges as Edge[];
          migrateProcessorSwitchGroups(restoredNodes, restoredEdges);
          migrateOutputSwitchGroups(restoredNodes, restoredEdges);
          reSnapChildren(restoredNodes, restoredEdges);
          return [restoredNodes, restoredEdges];
        }
      } catch {}
    }

    if (!initialData?.nodes || initialData.nodes.length === 0) return [[], []];

    const inputNodes = initialData.nodes.filter((n) => n.type === "input");
    const processorNodes = initialData.nodes.filter((n) => n.type === "processor");
    const outputNodes = initialData.nodes.filter((n) => n.type === "output");

    const flowNodes: Node[] = [];
    const flowEdges: Edge[] = [];
    let xPos = 0;
    let prevNodeIds: string[] = [];

    const connectFromPrev = (targetId: string) => {
      for (const pid of prevNodeIds) {
        flowEdges.push({ id: `e-${pid}-${targetId}`, source: pid, target: targetId, type: "pipeline" });
      }
    };

    for (const nd of inputNodes) {
      const id = uuidv4();

      if (nd.componentId === "broker") {
        let childConfigs: any[] = [];
        let groupConfigYaml = "";
        if (nd.configYaml?.trim()) {
          try {
            const parsed = yaml.load(nd.configYaml) as any;
            if (parsed && Array.isArray(parsed.inputs)) childConfigs = parsed.inputs;
            const { inputs: _inputs, ...rest } = parsed || {};
            if (Object.keys(rest).length > 0) {
              groupConfigYaml = yaml.dump(rest, { lineWidth: -1, noRefs: true });
            }
          } catch {}
        }

        const childCount = childConfigs.length;
        const groupHeight = calcBrokerInputGroupHeight(childCount);

        flowNodes.push({
          id,
          type: "brokerInputGroupNode",
          position: { x: xPos, y: NODE_Y - 25 },
          style: { width: GROUP_WIDTH, height: groupHeight },
          data: { ...nd, nodeId: id, isGroup: true, childCount, configYaml: groupConfigYaml },
        });

        childConfigs.forEach((inputObj, i) => {
          const componentName = Object.keys(inputObj).find((k) => k !== "label") || Object.keys(inputObj)[0];
          const config = inputObj[componentName];
          const childLabel = inputObj.label as string | undefined;
          const schema = allComponentSchemas.input.find(
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
            parentId: id,
            extent: "parent" as const,
            data: {
              label: childLabel || schema?.id || componentName,
              type: "input",
              componentId: schema?.id || componentName,
              component: componentName,
              configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
              nodeId: childId,
            },
          });
        });
      } else {
        flowNodes.push({
          id,
          type: "inputNode",
          position: { x: xPos, y: NODE_Y },
          data: { ...nd, nodeId: id },
        });
      }

      prevNodeIds = [id];
      xPos += NODE_SPACING_X;
    }

    for (const nd of processorNodes) {
      const id = uuidv4();

      if (nd.componentId === "branch") {
        let branchConfig: any = {};
        if (nd.configYaml?.trim()) {
          try { branchConfig = yaml.load(nd.configYaml) || {}; } catch {}
        }
        const childConfigs: any[] = Array.isArray(branchConfig.processors) ? branchConfig.processors : [];
        const requestMap = branchConfig.request_map || "";
        const resultMap = branchConfig.result_map || "";

        const childCount = childConfigs.length;
        const groupHeight = calcCatchGroupHeight(childCount);

        flowNodes.push({
          id,
          type: "branchGroupNode",
          position: { x: xPos, y: NODE_Y - 25 },
          style: { width: GROUP_WIDTH, height: groupHeight },
          data: { ...nd, nodeId: id, isGroup: true, childCount, configYaml: yaml.dump({ request_map: requestMap, result_map: resultMap }, { lineWidth: -1, noRefs: true }) },
        });

        let prevChildId: string | null = null;
        childConfigs.forEach((procObj, i) => {
          const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0];
          const config = procObj[componentName];
          const childLabel = procObj.label as string | undefined;
          const schema = allComponentSchemas.processor.find(
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
              configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
              nodeId: childId,
              childIndex: i,
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
      } else if (nd.componentId === "catch") {
        // Parse catch children from configYaml
        let childConfigs: any[] = [];
        if (nd.configYaml?.trim()) {
          try {
            const parsed = yaml.load(nd.configYaml);
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
          data: { ...nd, nodeId: id, isGroup: true, childCount, configYaml: "" },
        });

        let prevChildId: string | null = null;
        childConfigs.forEach((procObj, i) => {
          const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0];
          const config = procObj[componentName];
          const childLabel = procObj.label as string | undefined;
          const schema = allComponentSchemas.processor.find(
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
              configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
              nodeId: childId,
              childIndex: i,
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
      } else if (nd.componentId === "switch") {
        // Parse processor switch config
        let casesConfigs: any[] = [];
        if (nd.configYaml?.trim()) {
          try {
            const parsed = yaml.load(nd.configYaml);
            if (Array.isArray(parsed)) casesConfigs = parsed;
          } catch {}
        }

        // Decision tree: switch node → case start nodes → processor chains
        flowNodes.push({
          id,
          type: "switchProcessorGroupNode",
          position: { x: xPos, y: NODE_Y },
          data: { ...nd, nodeId: id, isGroup: false, childCount: casesConfigs.length, configYaml: "" },
        });

        connectFromPrev(id);

        const casePositions = layoutSwitchCases(xPos, NODE_Y, casesConfigs.length);
        let maxChainX = xPos;
        const caseChainEndIds: string[] = [];

        casesConfigs.forEach((caseObj, caseIdx) => {
          const check = caseObj.check || "";
          const fallthrough = caseObj.fallthrough === true;
          const processors = Array.isArray(caseObj.processors) ? caseObj.processors : [];
          const caseId = uuidv4();
          const casePos = casePositions[caseIdx] || { x: xPos + SWITCH_CASE_X_OFFSET, y: NODE_Y };

          // Case start node
          flowNodes.push({
            id: caseId,
            type: "switchCaseStartNode",
            position: casePos,
            data: {
              label: `case_${caseIdx + 1}`,
              type: "processor",
              componentId: "case",
              component: "case",
              configYaml: "",
              nodeId: caseId,
              isCaseStart: true,
              caseCheck: check,
              caseFallthrough: fallthrough,
              switchId: id,
            },
          });
          const procEdgeLabel = check
            ? (fallthrough ? `${check} ↓` : check)
            : (fallthrough ? "default ↓" : "default");
          flowEdges.push({
            id: `e-${id}-${caseId}`,
            source: id,
            target: caseId,
            type: "pipeline",
            data: { switchCase: true, edgeLabel: procEdgeLabel },
          });

          // Processor chain after case start
          let prevChainId = caseId;
          let chainX = casePos.x + SWITCH_CASE_X_OFFSET;
          processors.forEach((procObj: any) => {
            const componentName = Object.keys(procObj).find((k) => k !== "label") || Object.keys(procObj)[0] || "";
            const config = procObj[componentName];
            const procLabel = procObj.label as string | undefined;
            const schema = allComponentSchemas.processor.find(
              (p) => p.component === componentName || p.id === componentName
            );
            const procId = uuidv4();
            flowNodes.push({
              id: procId,
              type: "processorNode",
              position: { x: chainX, y: casePos.y },
              data: {
                label: procLabel || schema?.id || componentName,
                type: "processor",
                componentId: schema?.id || componentName,
                component: componentName,
                configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
                nodeId: procId,
                switchCaseId: caseId,
              },
            });
            flowEdges.push({
              id: `e-${prevChainId}-${procId}`,
              source: prevChainId,
              target: procId,
              type: "pipeline",
            });
            prevChainId = procId;
            chainX += NODE_SPACING_X;
          });
          caseChainEndIds.push(prevChainId);
          maxChainX = Math.max(maxChainX, chainX);
        });

        // Convergence: all case chain ends connect to the next pipeline node
        prevNodeIds = caseChainEndIds.length > 0 ? caseChainEndIds : [id];
        xPos = Math.max(xPos + NODE_SPACING_X, maxChainX);
      } else {
        flowNodes.push({
          id,
          type: "processorNode",
          position: { x: xPos, y: NODE_Y },
          data: { ...nd, nodeId: id },
        });
        connectFromPrev(id);
        prevNodeIds = [id];
        xPos += NODE_SPACING_X;
      }
    }

    for (const nd of outputNodes) {
      const id = uuidv4();

      if (nd.componentId === "broker") {
        // Parse broker config to extract children and pattern
        let brokerConfig: any = {};
        let childConfigs: any[] = [];
        let brokerPattern = "fan_out";
        if (nd.configYaml?.trim()) {
          try {
            brokerConfig = yaml.load(nd.configYaml) || {};
            if (Array.isArray(brokerConfig.outputs)) childConfigs = brokerConfig.outputs;
            if (brokerConfig.pattern) brokerPattern = brokerConfig.pattern;
          } catch {}
        }

        const childCount = childConfigs.length;
        const groupHeight = calcBrokerGroupHeight(childCount);

        // Build group configYaml without outputs (those are in children)
        const groupConfig: any = {};
        if (brokerConfig.copies) groupConfig.copies = brokerConfig.copies;
        if (brokerConfig.batching) groupConfig.batching = brokerConfig.batching;
        const groupConfigYaml = Object.keys(groupConfig).length > 0
          ? yaml.dump(groupConfig, { lineWidth: -1, noRefs: true })
          : "";

        flowNodes.push({
          id,
          type: "brokerGroupNode",
          position: { x: xPos, y: NODE_Y - 25 },
          style: { width: GROUP_WIDTH, height: groupHeight },
          data: { ...nd, nodeId: id, isGroup: true, childCount, configYaml: groupConfigYaml, brokerPattern },
        });

        let prevChildId: string | null = null;
        childConfigs.forEach((outputObj, i) => {
          const componentName = Object.keys(outputObj).find((k) => k !== "label") || Object.keys(outputObj)[0];
          const config = outputObj[componentName];
          const childLabel = outputObj.label as string | undefined;
          const schema = allComponentSchemas.output.find(
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
            parentId: id,
            extent: "parent" as const,
            data: {
              label: childLabel || schema?.id || componentName,
              type: "output",
              componentId: schema?.id || componentName,
              component: componentName,
              configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
              nodeId: childId,
              childIndex: i,
              parentBrokerPattern: brokerPattern,
            },
          });
          if (prevChildId && isBrokerSequential(brokerPattern)) {
            flowEdges.push({
              id: `e-internal-${prevChildId}-${childId}`,
              source: prevChildId,
              target: childId,
              type: "pipeline",
              data: makeInternalEdgeData(),
            });
          }
          prevChildId = childId;
        });

        connectFromPrev(id);
      } else if (nd.componentId === "switch") {
        let switchConfig: any = {};
        let casesConfigs: any[] = [];
        if (nd.configYaml?.trim()) {
          try {
            switchConfig = yaml.load(nd.configYaml) || {};
            if (Array.isArray(switchConfig.cases)) casesConfigs = switchConfig.cases;
          } catch {}
        }

        const childCount = casesConfigs.length;

        const groupConfig: any = {};
        if (switchConfig.retry_until_success) groupConfig.retry_until_success = switchConfig.retry_until_success;
        if (switchConfig.strict_mode) groupConfig.strict_mode = switchConfig.strict_mode;
        const groupConfigYaml = Object.keys(groupConfig).length > 0
          ? yaml.dump(groupConfig, { lineWidth: -1, noRefs: true })
          : "";

        flowNodes.push({
          id,
          type: "switchProcessorGroupNode",
          position: { x: xPos, y: NODE_Y },
          data: { ...nd, nodeId: id, isGroup: false, childCount, configYaml: groupConfigYaml },
        });

        const casePositions = layoutSwitchCases(xPos, NODE_Y, childCount);
        casesConfigs.forEach((caseObj, i) => {
          const check = caseObj.check || "";
          const cont = caseObj.continue === true;
          const outputConfig = caseObj.output || {};
          const componentName = Object.keys(outputConfig)[0] || "";
          const config = outputConfig[componentName];
          const schema = allComponentSchemas.output.find(
            (o) => o.component === componentName || o.id === componentName
          );
          const childId = uuidv4();
          const casePos = casePositions[i];
          flowNodes.push({
            id: childId,
            type: "switchCaseStartNode",
            position: casePos,
            data: {
              label: schema?.id || componentName || `case_${i + 1}`,
              type: "output",
              componentId: schema?.id || componentName,
              component: componentName,
              configYaml: typeof config === "string" ? config : (config && Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : ""),
              nodeId: childId,
              isCaseStart: true,
              switchId: id,
              caseCheck: check,
              caseContinue: cont,
            },
          });
          const edgeLabel = check
            ? (cont ? `${check} →` : check)
            : (cont ? "default →" : "default");
          flowEdges.push({
            id: `e-${id}-${childId}`,
            source: id,
            target: childId,
            type: "pipeline",
            data: { switchCase: true, edgeLabel, edgeLabelColor: "amber" },
          });
        });

        connectFromPrev(id);
      } else {
        flowNodes.push({
          id,
          type: "outputNode",
          position: { x: xPos, y: NODE_Y },
          data: { ...nd, nodeId: id },
        });
        connectFromPrev(id);
      }
    }

    return [flowNodes, flowEdges];
  }, []); // Only compute once on mount

  const [nodes, setNodes, onNodesChange] = useNodesState(initialFlowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialFlowEdges);

  const handleDeleteEdge = useCallback(
    (edgeId: string) => {
      const edgeToDelete = edges.find((e) => e.id === edgeId);

      if (edgeToDelete) {
        const sourceNode = nodes.find((n) => n.id === edgeToDelete.source);
        const targetNode = nodes.find((n) => n.id === edgeToDelete.target);
        const sourceInCase = sourceNode?.type === "switchCaseStartNode" || !!(sourceNode?.data as any)?.switchCaseId;
        const targetHasCaseId = !!(targetNode?.data as any)?.switchCaseId;

        if (sourceInCase && targetHasCaseId) {
          // Detach: clear switchCaseId from target and all downstream case processors
          const detachIds = new Set<string>();
          const queue = [edgeToDelete.target];
          while (queue.length > 0) {
            const id = queue.shift()!;
            const node = nodes.find((n) => n.id === id);
            if (node && (node.data as any).switchCaseId) {
              detachIds.add(id);
              const outEdge = edges.find((e) => e.source === id);
              if (outEdge) {
                const nextNode = nodes.find((n) => n.id === outEdge.target);
                if (nextNode && (nextNode.data as any).switchCaseId) {
                  queue.push(outEdge.target);
                }
              }
            }
          }

          setNodes((nds) => nds.map((n) =>
            detachIds.has(n.id)
              ? { ...n, data: { ...n.data, switchCaseId: undefined, disconnected: false } }
              : (n.data as StreamFlowNodeData).disconnected
                ? { ...n, data: { ...n.data, disconnected: false } }
                : n
          ));

          // Just remove the deleted edge; detached processors become independent
          setEdges((eds) => eds.filter((e) => e.id !== edgeId));
          return;
        }
      }

      setEdges((eds) => eds.filter((e) => e.id !== edgeId));
      setNodes((nds) =>
        nds.map((n) =>
          (n.data as StreamFlowNodeData).disconnected
            ? { ...n, data: { ...n.data, disconnected: false } }
            : n
        )
      );
    },
    [edges, nodes, setEdges, setNodes]
  );

  // Clear disconnected state when edges change
  const handleEdgesChange: typeof onEdgesChange = useCallback(
    (changes) => {
      const filtered = changes.filter((change) => {
        if (change.type !== "remove") return true;
        const edge = edges.find((e) => e.id === change.id);
        if (!edge) return true;
        // Only protect switch→caseStart edges
        if ((edge.data as any)?.switchCase) return false;
        return true;
      });
      onEdgesChange(filtered);
      setNodes((nds) =>
        nds.map((n) =>
          (n.data as StreamFlowNodeData).disconnected
            ? { ...n, data: { ...n.data, disconnected: false } }
            : n
        )
      );
    },
    [onEdgesChange, setNodes, edges]
  );

  const scheduleAutoFit = useCallback(() => {
    setTimeout(() => fitView({ padding: 0.2, duration: 300 }), 50);
  }, [fitView]);

  // --- All add handlers now open the pending modal instead of adding immediately ---

  const handleAddAndConnect = useCallback(
    (sourceNodeId: string, _sourceType: string) => {
      const sourceNode = nodes.find((n) => n.id === sourceNodeId);
      if (!sourceNode) return;
      // Determine available components: switch case chain processors get filtered list
      const isCaseChain = sourceNode.type === "switchCaseStartNode" || !!(sourceNode.data as any).switchCaseId;
      const available = isCaseChain
        ? allComponentSchemas.processor.filter((p) => p.id !== "switch" && p.id !== "catch")
        : allComponentSchemas.processor;
      const count = nodes.filter((n) => (n.data as StreamFlowNodeData).type === "processor").length + 1;
      setPendingNode({
        kind: "addAndConnect",
        sourceNodeId,
        label: `new_processor_${count}`,
        componentId: "",
        component: "",
        availableComponents: available,
      });
    },
    [nodes, allComponentSchemas]
  );

  const handleAddBefore = useCallback(
    (targetNodeId: string) => {
      const count = nodes.filter((n) => (n.data as StreamFlowNodeData).type === "processor").length + 1;
      setPendingNode({
        kind: "addBefore",
        targetNodeId,
        label: `new_processor_${count}`,
        componentId: "",
        component: "",
        availableComponents: allComponentSchemas.processor,
      });
    },
    [nodes, allComponentSchemas]
  );

  const handleAddChildProcessor = useCallback(
    (groupId: string) => {
      const groupNode = nodes.find((n) => n.id === groupId);
      const isBranch = groupNode?.type === "branchGroupNode";
      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const available = allComponentSchemas.processor.filter((p) => p.id !== "catch" && p.id !== "branch");
      const prefix = isBranch ? "branch_proc" : "catch_proc";
      setPendingNode({
        kind: "childProcessor",
        groupId,
        label: `${prefix}_${existingChildren.length + 1}`,
        componentId: "",
        component: "",
        availableComponents: available,
      });
    },
    [nodes, allComponentSchemas]
  );

  const handleAddChildOutput = useCallback(
    (groupId: string) => {
      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const available = allComponentSchemas.output.filter((o) => o.id !== "broker" && o.id !== "switch");
      setPendingNode({
        kind: "childOutput",
        groupId,
        label: `broker_output_${existingChildren.length + 1}`,
        componentId: "",
        component: "",
        availableComponents: available,
      });
    },
    [nodes, allComponentSchemas]
  );

  const handleAddChildInput = useCallback(
    (groupId: string) => {
      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const available = allComponentSchemas.input.filter((i) => i.id !== "broker");
      setPendingNode({
        kind: "childInput",
        groupId,
        label: `broker_input_${existingChildren.length + 1}`,
        componentId: "",
        component: "",
        availableComponents: available,
      });
    },
    [nodes, allComponentSchemas]
  );

  const handleAddChildCase = useCallback(
    (groupId: string) => {
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) return;

      const isProcessorSwitch = groupNode.type === "switchProcessorGroupNode" && (groupNode.data as StreamFlowNodeData).type === "processor";
      const existingCaseStarts = nodes.filter(
        (n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === groupId
      );
      const defaultExists = existingCaseStarts.some((n) => !(n.data as any).caseCheck);

      // Output switch cases need a component picker for output components
      const availableComponents = !isProcessorSwitch
        ? allComponentSchemas.output
            .filter((o) => o.id !== "switch" && o.id !== "broker")
            .map((o) => ({ id: o.id, name: o.id, component: o.component }))
        : [];

      setPendingNode({
        kind: "childCase",
        groupId,
        isProcessorSwitch,
        label: "",
        componentId: "",
        component: "",
        caseCheck: "",
        caseFallthrough: false,
        caseContinue: false,
        defaultExists,
        availableComponents,
      });
    },
    [nodes, allComponentSchemas]
  );

  // --- Unified confirm handler that dispatches to actual creation logic ---
  const handleConfirmAddNode = useCallback(() => {
    if (!pendingNode) return;
    const { kind, label, componentId, component } = pendingNode;

    if (kind === "addAndConnect") {
      const sourceNodeId = pendingNode.sourceNodeId!;
      const sourceNode = nodes.find((n) => n.id === sourceNodeId);
      if (!sourceNode) { setPendingNode(null); return; }

      const newId = uuidv4();
      const isCaseStart = sourceNode.type === "switchCaseStartNode";
      const sourceCaseId = isCaseStart
        ? sourceNode.id
        : (sourceNode.data as any).switchCaseId as string | undefined;

      if (sourceNode.type === "switchProcessorGroupNode") {
        const caseStarts = nodes.filter(
          (n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === sourceNodeId
        );
        const chainEndIds: string[] = [];
        for (const cs of caseStarts) {
          let endId = cs.id;
          let outEdge = edges.find((e) => e.source === endId && !(e.data as any)?.switchCase);
          while (outEdge) {
            const nextNode = nodes.find((n) => n.id === outEdge!.target);
            if (!nextNode || !(nextNode.data as any).switchCaseId) break;
            endId = nextNode.id;
            outEdge = edges.find((e) => e.source === endId);
          }
          chainEndIds.push(endId);
        }
        if (chainEndIds.length === 0) chainEndIds.push(sourceNodeId);

        const maxEndX = Math.max(...chainEndIds.map((id) => nodes.find((n) => n.id === id)?.position.x || 0), sourceNode.position.x);
        const insertX = maxEndX + NODE_SPACING_X;

        const resolved = resolveFlowNodeType("processor", componentId);
        const newNode: Node = {
          id: newId,
          type: resolved.nodeType,
          position: { x: insertX, y: sourceNode.position.y + resolved.yOffset },
          ...(resolved.style ? { style: resolved.style } : {}),
          data: { label, type: "processor", componentId, component, configYaml: "", nodeId: newId, ...resolved.extraData },
        };

        setNodes((nds) => [
          ...nds.map((n) =>
            !n.parentId && n.id !== sourceNodeId && n.position.x >= insertX
              && n.type !== "switchCaseStartNode" && !(n.data as any).switchCaseId
              ? { ...n, position: { ...n.position, x: n.position.x + NODE_SPACING_X } }
              : n
          ),
          newNode,
        ]);

        setEdges((eds) => {
          const existingTargetId = chainEndIds.reduce<string | undefined>((found, endId) => {
            if (found) return found;
            const outEdge = eds.find((e) => e.source === endId && !(e.data as any)?.switchCase);
            return outEdge?.target;
          }, undefined);

          let filtered = eds.filter((e) => {
            if (chainEndIds.includes(e.source) && !(e.data as any)?.switchCase && !(e.data as any)?.internal) {
              const target = nodes.find((n) => n.id === e.target);
              if (!target || !(target.data as any).switchCaseId) return false;
            }
            return true;
          });

          const newEdges: Edge[] = chainEndIds.map((endId) => ({
            id: `e-${endId}-${newId}`,
            source: endId,
            target: newId,
            type: "pipeline",
          }));

          if (existingTargetId) {
            filtered = filtered.filter((e) => !(chainEndIds.includes(e.source) && e.target === existingTargetId));
            newEdges.push({ id: `e-${newId}-${existingTargetId}`, source: newId, target: existingTargetId, type: "pipeline" });
          }

          return [...filtered, ...newEdges];
        });

        setSelectedNodeId(newId);
        setEditingNodeId(newId);
        setPendingNode(null);
        scheduleAutoFit();
        return;
      }

      const insertX = sourceNode.position.x + NODE_SPACING_X;
      const resolved = resolveFlowNodeType("processor", componentId);
      const newNode: Node = {
        id: newId,
        type: resolved.nodeType,
        position: { x: insertX, y: sourceNode.position.y + resolved.yOffset },
        ...(resolved.style ? { style: resolved.style } : {}),
        data: {
          label, type: "processor", componentId, component, configYaml: "", nodeId: newId,
          ...resolved.extraData,
          ...(sourceCaseId ? { switchCaseId: sourceCaseId } : {}),
        } satisfies Partial<StreamFlowNodeData>,
      };

      setNodes((nds) => [
        ...nds.map((n) =>
          !n.parentId && n.id !== sourceNodeId && n.position.x >= insertX
            ? { ...n, position: { ...n.position, x: n.position.x + NODE_SPACING_X } }
            : n
        ),
        newNode,
      ]);
      setEdges((eds) => {
        const existingOutgoing = eds.find((e) => e.source === sourceNodeId && !(e.data as any)?.switchCase);
        const filtered = existingOutgoing ? eds.filter((e) => e.id !== existingOutgoing.id) : eds;
        const newEdges = [
          ...filtered,
          { id: `e-${sourceNodeId}-${newId}`, source: sourceNodeId, target: newId, type: "pipeline" } as Edge,
        ];
        if (existingOutgoing) {
          newEdges.push({ id: `e-${newId}-${existingOutgoing.target}`, source: newId, target: existingOutgoing.target, type: "pipeline" } as Edge);
        }
        return newEdges;
      });

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "addBefore") {
      const targetNodeId = pendingNode.targetNodeId!;
      const targetNode = nodes.find((n) => n.id === targetNodeId);
      if (!targetNode) { setPendingNode(null); return; }

      const newId = uuidv4();
      const insertX = targetNode.position.x;
      const resolved = resolveFlowNodeType("processor", componentId);
      const newNode: Node = {
        id: newId, type: resolved.nodeType,
        position: { x: insertX, y: targetNode.position.y + resolved.yOffset },
        ...(resolved.style ? { style: resolved.style } : {}),
        data: { label, type: "processor", componentId, component, configYaml: "", nodeId: newId, ...resolved.extraData } satisfies Partial<StreamFlowNodeData>,
      };

      setNodes((nds) => [
        ...nds.map((n) =>
          !n.parentId && n.position.x >= insertX
            ? { ...n, position: { ...n.position, x: n.position.x + NODE_SPACING_X } }
            : n
        ),
        newNode,
      ]);
      setEdges((eds) => {
        const incomingEdges = eds.filter((e) => e.target === targetNodeId);
        const filtered = eds.filter((e) => e.target !== targetNodeId);
        const newEdges = [
          ...filtered,
          { id: `e-${newId}-${targetNodeId}`, source: newId, target: targetNodeId, type: "pipeline" } as Edge,
        ];
        for (const incoming of incomingEdges) {
          newEdges.push({ id: `e-${incoming.source}-${newId}`, source: incoming.source, target: newId, type: "pipeline" } as Edge);
        }
        return newEdges;
      });

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "childProcessor") {
      const groupId = pendingNode.groupId!;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) { setPendingNode(null); return; }

      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const childCount = existingChildren.length;
      const newId = uuidv4();

      const newChild: Node = {
        id: newId, type: "childProcessorNode",
        position: { x: CHILD_X, y: CATCH_CHILD_Y_START + childCount * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) },
        parentId: groupId, extent: "parent" as const,
        data: { label, type: "processor", componentId, component, configYaml: "", nodeId: newId, childIndex: childCount } satisfies Partial<StreamFlowNodeData>,
      };

      const newHeight = calcCatchGroupHeight(childCount + 1);
      setNodes((nds) => [
        ...nds.map((n) => n.id === groupId
          ? { ...n, style: { ...n.style, width: GROUP_WIDTH, height: newHeight }, data: { ...n.data, childCount: childCount + 1 } }
          : n
        ),
        newChild,
      ]);

      if (childCount > 0) {
        const lastChild = existingChildren.sort((a, b) => a.position.y - b.position.y)[childCount - 1];
        setEdges((eds) => [...eds, {
          id: `e-internal-${lastChild.id}-${newId}`, source: lastChild.id, target: newId, type: "pipeline", data: { internal: true },
        }]);
      }

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "childOutput") {
      const groupId = pendingNode.groupId!;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) { setPendingNode(null); return; }

      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const childCount = existingChildren.length;
      const newId = uuidv4();

      const newChild: Node = {
        id: newId, type: "childOutputNode",
        position: { x: CHILD_X, y: BROKER_CHILD_Y_START + childCount * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) },
        parentId: groupId, extent: "parent" as const,
        data: {
          label, type: "output", componentId, component, configYaml: "", nodeId: newId,
          childIndex: childCount,
          parentBrokerPattern: (groupNode.data as StreamFlowNodeData).brokerPattern || "fan_out",
        } satisfies Partial<StreamFlowNodeData>,
      };

      const newHeight = calcBrokerGroupHeight(childCount + 1);
      setNodes((nds) => [
        ...nds.map((n) => n.id === groupId
          ? { ...n, style: { ...n.style, width: GROUP_WIDTH, height: newHeight }, data: { ...n.data, childCount: childCount + 1 } }
          : n
        ),
        newChild,
      ]);

      const pattern = (groupNode.data as StreamFlowNodeData).brokerPattern;
      if (childCount > 0 && isBrokerSequential(pattern)) {
        const lastChild = existingChildren.sort((a, b) => a.position.y - b.position.y)[childCount - 1];
        setEdges((eds) => [...eds, {
          id: `e-internal-${lastChild.id}-${newId}`, source: lastChild.id, target: newId, type: "pipeline", data: makeInternalEdgeData(),
        }]);
      }

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "childInput") {
      const groupId = pendingNode.groupId!;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) { setPendingNode(null); return; }

      const existingChildren = nodes.filter((n) => n.parentId === groupId);
      const childCount = existingChildren.length;
      const newId = uuidv4();

      const newChild: Node = {
        id: newId, type: "childInputNode",
        position: { x: CHILD_X, y: BROKER_INPUT_CHILD_Y_START + childCount * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) },
        parentId: groupId, extent: "parent" as const,
        data: { label, type: "input", componentId, component, configYaml: "", nodeId: newId } satisfies Partial<StreamFlowNodeData>,
      };

      const newHeight = calcBrokerInputGroupHeight(childCount + 1);
      setNodes((nds) => [
        ...nds.map((n) => n.id === groupId
          ? { ...n, style: { ...n.style, width: GROUP_WIDTH, height: newHeight }, data: { ...n.data, childCount: childCount + 1 } }
          : n
        ),
        newChild,
      ]);

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "childCase") {
      const groupId = pendingNode.groupId!;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) { setPendingNode(null); return; }

      const { caseCheck, caseFallthrough, caseContinue, isProcessorSwitch } = pendingNode;
      const newId = uuidv4();

      if (isProcessorSwitch) {
        const existingCaseStarts = nodes.filter(
          (n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === groupId
        );
        const caseCount = existingCaseStarts.length + 1;
        const positions = layoutSwitchCases(groupNode.position.x, groupNode.position.y, caseCount);

        const newCase: Node = {
          id: newId, type: "switchCaseStartNode", position: { x: 0, y: 0 },
          data: {
            label: `case_${caseCount}`, type: "processor", componentId: "case", component: "case",
            configYaml: "", nodeId: newId, isCaseStart: true,
            caseCheck, caseFallthrough, switchId: groupId,
          },
        };

        const allCases = insertCaseBeforeDefault(existingCaseStarts, newCase);

        setNodes((nds) => {
          let updated = nds.map((n) => {
            if (n.id === groupId) return { ...n, data: { ...n.data, childCount: caseCount } };
            const caseIdx = allCases.findIndex((c) => c.id === n.id);
            if (caseIdx >= 0) return { ...n, position: positions[caseIdx] };
            return n;
          });
          return [...updated, { ...newCase, position: positions[allCases.findIndex((c) => c.id === newId)] }];
        });

        const procCaseLabel = caseCheck
          ? (caseFallthrough ? `${caseCheck} ↓` : caseCheck)
          : (caseFallthrough ? "default ↓" : "default");
        setEdges((eds) => [
          ...eds,
          { id: `e-${groupId}-${newId}`, source: groupId, target: newId, type: "pipeline", data: { switchCase: true, edgeLabel: procCaseLabel } },
        ]);
      } else {
        // Output switch: same decision-tree layout as processor switch
        const existingCaseStarts = nodes.filter(
          (n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === groupId
        );
        const caseCount = existingCaseStarts.length + 1;
        const positions = layoutSwitchCases(groupNode.position.x, groupNode.position.y, caseCount);

        const newCase: Node = {
          id: newId, type: "switchCaseStartNode", position: { x: 0, y: 0 },
          data: {
            label: label || componentId || `case_${caseCount}`, type: "output", componentId, component,
            configYaml: "", nodeId: newId, isCaseStart: true,
            caseCheck, caseContinue, switchId: groupId,
          },
        };

        const allCases = insertCaseBeforeDefault(existingCaseStarts, newCase);

        setNodes((nds) => {
          let updated = nds.map((n) => {
            if (n.id === groupId) return { ...n, data: { ...n.data, childCount: caseCount } };
            const caseIdx = allCases.findIndex((c) => c.id === n.id);
            if (caseIdx >= 0) return { ...n, position: positions[caseIdx] };
            return n;
          });
          return [...updated, { ...newCase, position: positions[allCases.findIndex((c) => c.id === newId)] }];
        });

        const outCaseLabel = caseCheck
          ? (caseContinue ? `${caseCheck} →` : caseCheck)
          : (caseContinue ? "default →" : "default");
        setEdges((eds) => [
          ...eds,
          { id: `e-${groupId}-${newId}`, source: groupId, target: newId, type: "pipeline", data: { switchCase: true, edgeLabel: outCaseLabel, edgeLabelColor: "amber" } },
        ]);
      }

      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    } else if (kind === "topLevel") {
      const type = pendingNode.topLevelType!;
      const newId = uuidv4();
      let xPos = 50;

      if (type === "input") {
        xPos = 0;
      } else if (type === "output") {
        xPos = Math.max(...nodes.map((n) => n.position.x), 0) + NODE_SPACING_X;
      } else {
        const center = screenToFlowPosition({ x: window.innerWidth / 2, y: window.innerHeight / 2 });
        xPos = center.x - 90;
      }

      const yOffset = type === "processor"
        ? (() => {
            const center = screenToFlowPosition({ x: window.innerWidth / 2, y: window.innerHeight / 2 });
            return center.y - NODE_Y;
          })()
        : nodes.filter((n) => Math.abs(n.position.x - xPos) < 100).length * 100;

      const resolved = resolveFlowNodeType(type, componentId);
      const newNode: Node = {
        id: newId, type: resolved.nodeType,
        position: { x: xPos, y: NODE_Y + yOffset + resolved.yOffset },
        ...(resolved.style ? { style: resolved.style } : {}),
        data: { label, type, componentId, component, configYaml: "", nodeId: newId, ...resolved.extraData } satisfies Partial<StreamFlowNodeData>,
      };

      setNodes((nds) => [...nds, newNode]);
      setSelectedNodeId(newId);
      setEditingNodeId(newId);
    }

    setPendingNode(null);
    scheduleAutoFit();
  }, [pendingNode, nodes, edges, setNodes, setEdges, scheduleAutoFit, screenToFlowPosition]);

  // handleAddCaseProcessor is no longer needed — case processors are added via handleAddAndConnect from case start nodes

  const handleChildDragStart = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!node.parentId) return;
      const groupId = node.parentId;
      setNodes((nds) => nds.map((n) =>
        n.id === node.id ? { ...n, data: { ...n.data, isDragging: true } } : n
      ));
      setEdges((eds) => {
        const childIds = new Set(nodes.filter((n) => n.parentId === groupId).map((n) => n.id));
        return eds.map((e) =>
          (e.data as any)?.internal && (childIds.has(e.source) || childIds.has(e.target))
            ? { ...e, hidden: true }
            : e
        );
      });
    },
    [nodes, setNodes, setEdges]
  );

  const handleChildDrag = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!node.parentId) return;
      const groupId = node.parentId;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) return;

      const isBrokerOutput = groupNode.type === "brokerGroupNode";
      const isBrokerInput = groupNode.type === "brokerInputGroupNode";
      const childYStart = isBrokerOutput ? BROKER_CHILD_Y_START : isBrokerInput ? BROKER_INPUT_CHILD_Y_START : CATCH_CHILD_Y_START;
      const slotHeight = CHILD_NODE_HEIGHT + CHILD_GAP_Y;

      const siblings = nodes.filter((n) => n.parentId === groupId);
      const draggedCenter = node.position.y + CHILD_NODE_HEIGHT / 2;
      const newIndex = Math.max(0, Math.min(siblings.length - 1, Math.round((draggedCenter - childYStart - CHILD_NODE_HEIGHT / 2) / slotHeight)));

      setNodes((nds) => {
        const sibs = nds.filter((n) => n.parentId === groupId && n.id !== node.id).sort((a, b) => a.position.y - b.position.y);
        const ordered: string[] = [];
        let inserted = false;
        for (let i = 0; i < sibs.length; i++) {
          if (i === newIndex && !inserted) { ordered.push(node.id); inserted = true; }
          ordered.push(sibs[i].id);
        }
        if (!inserted) ordered.push(node.id);

        const draggedIdx = ordered.indexOf(node.id);
        return nds.map((n) => {
          if (n.id === node.id) {
            return { ...n, data: { ...n.data, childIndex: draggedIdx } };
          }
          const idx = ordered.indexOf(n.id);
          if (idx >= 0 && n.parentId === groupId) {
            return {
              ...n,
              position: { x: CHILD_X, y: childYStart + idx * slotHeight },
              data: { ...n.data, childIndex: idx },
            };
          }
          return n;
        });
      });
    },
    [nodes, setNodes]
  );

  const handleChildDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (!node.parentId) return;
      const groupId = node.parentId;
      const groupNode = nodes.find((n) => n.id === groupId);
      if (!groupNode) return;

      // Normal flat groups (catch, broker, branch)
      const isBrokerOutput = groupNode.type === "brokerGroupNode";
      const isBrokerInput = groupNode.type === "brokerInputGroupNode";
      const childYStart = isBrokerOutput ? BROKER_CHILD_Y_START : isBrokerInput ? BROKER_INPUT_CHILD_Y_START : CATCH_CHILD_Y_START;

      const sortFn = (a: Node, b: Node) => a.position.y - b.position.y;
      const siblings = nodes.filter((n) => n.parentId === groupId).sort(sortFn);

      setNodes((nds) => {
        const sorted = nds.filter((n) => n.parentId === groupId).sort(sortFn);
        return nds.map((n) => {
          const idx = sorted.findIndex((s) => s.id === n.id);
          if (idx >= 0) {
            return {
              ...n,
              position: { x: CHILD_X, y: childYStart + idx * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) },
              data: { ...n.data, childIndex: idx, isDragging: false, isDropTarget: false },
            };
          }
          return n;
        });
      });

      const brokerPattern = isBrokerOutput ? (groupNode.data as StreamFlowNodeData).brokerPattern : undefined;
      const shouldChain = isBrokerInput ? false : isBrokerOutput ? isBrokerSequential(brokerPattern) : true;

      const childIds = new Set(siblings.map((s) => s.id));
      setEdges((eds) => {
        const kept = eds.filter((e) => !((e.data as any)?.internal && (childIds.has(e.source) || childIds.has(e.target))));
        if (siblings.length <= 1 || !shouldChain) return kept;
        const sorted = [...siblings].sort((a, b) => (nodes.find((n) => n.id === a.id)?.position.y ?? 0) - (nodes.find((n) => n.id === b.id)?.position.y ?? 0));
        const newEdges: Edge[] = [];
        for (let i = 0; i < sorted.length - 1; i++) {
          newEdges.push({ id: `e-internal-${sorted[i].id}-${sorted[i + 1].id}`, source: sorted[i].id, target: sorted[i + 1].id, type: "pipeline", data: { internal: true } });
        }
        return [...kept, ...newEdges];
      });
    },
    [nodes, setNodes, setEdges]
  );

  // Inject callbacks into node data and compute child extent dynamically
  const nodesWithCallbacks = useMemo(() => {
    return nodes.map((n) => ({
      ...n,
      data: {
        ...n.data,
        nodeId: n.id,
        onAddAndConnect: handleAddAndConnect,
        onAddBefore: handleAddBefore,
        onAddChildProcessor: handleAddChildProcessor,
        onAddChildOutput: handleAddChildOutput,
        onAddChildInput: handleAddChildInput,
        onAddChildCase: handleAddChildCase,
      },
    }));
  }, [nodes, handleAddAndConnect, handleAddBefore, handleAddChildProcessor, handleAddChildOutput, handleAddChildInput, handleAddChildCase]);

  const handleEditCaseNode = useCallback((nodeId: string) => {
    setEditingNodeId(nodeId);
  }, []);

  const edgesWithCallbacks = useMemo(() => {
    return edges.map((e) => {
      // Only protect switch→caseStart edges (data.switchCase: true) and internal edges
      const isProtected = (e.data as any)?.internal || (e.data as any)?.switchCase;
      const isSwitchCase = (e.data as any)?.switchCase;
      return {
        ...e,
        data: {
          ...e.data,
          ...(!isProtected ? { onDeleteEdge: handleDeleteEdge } : {}),
          ...(isSwitchCase ? { onEditCaseNode: handleEditCaseNode, caseTargetId: e.target } : {}),
        },
        markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16 },
      };
    });
  }, [edges, handleDeleteEdge, handleEditCaseNode]);

  const isValidConnection = useCallback(
    (connection: Edge | Connection) => {
      // No self-connections
      if (connection.source === connection.target) return false;

      const source = nodes.find((n) => n.id === connection.source);
      const target = nodes.find((n) => n.id === connection.target);
      if (!source || !target) return false;

      // Prevent connections to/from child nodes
      if (source.parentId || target.parentId) return false;

      const sourceInCase = source.type === "switchCaseStartNode" || !!(source.data as any).switchCaseId;
      const targetInCase = target.type === "switchCaseStartNode" || !!(target.data as any).switchCaseId;

      if (sourceInCase && targetInCase) {
        // Allow connections within the same case chain (reordering)
        const sourceCaseId = source.type === "switchCaseStartNode"
          ? source.id : (source.data as any).switchCaseId;
        const targetCaseId = target.type === "switchCaseStartNode"
          ? target.id : (target.data as any).switchCaseId;
        if (sourceCaseId !== targetCaseId) return false;
      }

      // If one side is in a case chain, the other must be a processor or output (not input, not child, not same switch node)
      if (sourceInCase || targetInCase) {
        const independent = sourceInCase ? target : source;
        if (independent.parentId) return false;
        // Block connecting case children to their own parent switch node
        if (independent.type === "switchProcessorGroupNode") {
          const caseNode = sourceInCase ? source : target;
          const switchId = caseNode.type === "switchCaseStartNode"
            ? (caseNode.data as any).switchId
            : (nodes.find((n) => n.id === (caseNode.data as any).switchCaseId)?.data as any)?.switchId;
          // Block connecting to the same processor switch, but allow connecting to output switch (which is a valid target)
          if (switchId === independent.id) return false;
          if ((independent.data as StreamFlowNodeData).type === "processor") return false;
        }
        const indType = (independent.data as StreamFlowNodeData).type;
        // Case source can connect to processor or output; independent processor can connect to case target
        if (sourceInCase && indType !== "processor" && indType !== "output") return false;
        if (targetInCase && indType !== "processor") return false;
      }

      const isCaseRelated = sourceInCase || targetInCase;

      // Each source can only have one outgoing edge
      // Exception: case chain end nodes may have a convergence edge that gets rewired on connect
      const hasOutgoing = edges.some((e) => e.source === connection.source);
      if (hasOutgoing && !isCaseRelated) return false;

      // Each target can only have one incoming edge
      // Exception: when connecting to a node that receives convergence edges, onConnect rewires them
      const hasIncoming = edges.some((e) => e.target === connection.target);
      if (hasIncoming && !isCaseRelated) return false;

      const st = (source.data as StreamFlowNodeData).type;
      const tt = (target.data as StreamFlowNodeData).type;
      if (st === "input" && (tt === "processor" || tt === "output")) return true;
      if (st === "processor" && (tt === "processor" || tt === "output")) return true;
      return false;
    },
    [nodes, edges]
  );

  const onConnect = useCallback(
    (params: Connection) => {
      const sourceNode = nodes.find((n) => n.id === params.source);
      const targetNode = nodes.find((n) => n.id === params.target);

      const sourceInCase = sourceNode?.type === "switchCaseStartNode" || !!(sourceNode?.data as any)?.switchCaseId;
      const targetInCase = targetNode?.type === "switchCaseStartNode" || !!(targetNode?.data as any)?.switchCaseId;

      if (sourceInCase && !targetInCase && targetNode) {
        const caseId = sourceNode?.type === "switchCaseStartNode"
          ? sourceNode.id
          : (sourceNode?.data as any)?.switchCaseId;
        const targetType = (targetNode.data as StreamFlowNodeData).type;

        if (targetType === "processor") {
          // Attaching independent processor to end of case chain
          setNodes((nds) => nds.map((n) =>
            n.id === targetNode.id
              ? { ...n, data: { ...n.data, switchCaseId: caseId } }
              : n
          ));

          setEdges((eds) => {
            // Only remove outgoing from source, not all incoming to target
            const filtered = eds.filter((e) => e.source !== params.source);
            const newEdges = addEdge({ ...params, type: "pipeline" }, filtered);

            // Find old convergence edge from source to a main pipeline node
            const oldConvergence = eds.find((e) => {
              if (e.source !== params.source) return false;
              const t = nodes.find((n) => n.id === e.target);
              return t && !t.parentId && t.type !== "switchCaseStartNode" && !(t.data as any).switchCaseId;
            });
            if (oldConvergence) {
              // Move convergence from old chain end to new chain end
              return [
                ...newEdges.filter((e) => e.id !== oldConvergence.id),
                { id: `e-${params.target}-${oldConvergence.target}`, source: params.target!, target: oldConvergence.target, type: "pipeline" } as Edge,
              ];
            }
            return newEdges;
          });
        } else {
          // Connecting case chain end to output/main pipeline node (convergence edge)
          setEdges((eds) => {
            // Only remove outgoing from source, keep other incoming edges to target
            const filtered = eds.filter((e) => e.source !== params.source);
            return addEdge({ ...params, type: "pipeline" }, filtered);
          });
        }
      } else if (targetInCase && !sourceInCase && sourceNode) {
        // Attaching independent processor before a case chain node
        const caseId = targetNode?.type === "switchCaseStartNode"
          ? targetNode.id
          : (targetNode?.data as any)?.switchCaseId;

        setNodes((nds) => nds.map((n) =>
          n.id === sourceNode.id
            ? { ...n, data: { ...n.data, switchCaseId: caseId } }
            : n
        ));

        // Remove existing incoming edge to target and replace with chain through new node
        setEdges((eds) => {
          const filtered = eds.filter(
            (e) => e.source !== params.source && e.target !== params.target
          );
          const newEdges = addEdge({ ...params, type: "pipeline" }, filtered);

          // Find the edge that was previously going into the target
          const oldIncoming = eds.find((e) => e.target === params.target && e.source !== params.source);
          if (oldIncoming) {
            return [
              ...newEdges.filter((e) => e.id !== oldIncoming.id),
              { id: `e-${oldIncoming.source}-${params.source}`, source: oldIncoming.source, target: params.source!, type: "pipeline" } as Edge,
            ];
          }
          return newEdges;
        });
      } else {
        // Normal connection
        setEdges((eds) => {
          const filtered = eds.filter(
            (e) => e.source !== params.source && e.target !== params.target
          );
          return addEdge({ ...params, type: "pipeline" }, filtered);
        });
      }
    },
    [nodes, setNodes, setEdges]
  );

  const handleAddNode = useCallback(
    (type: "input" | "processor" | "output") => {
      const existingOfType = nodes.filter((n) => (n.data as StreamFlowNodeData).type === type && !n.parentId);
      if (type === "input" && existingOfType.length > 0) {
        addToast({ id: "input-exists", title: "Input Exists", description: "Only one input node allowed.", variant: "info" });
        return;
      }
      if (type === "output" && existingOfType.length > 0) {
        addToast({ id: "output-exists", title: "Output Exists", description: "Only one output node allowed.", variant: "info" });
        return;
      }

      const label = type === "processor"
        ? `new_processor_${existingOfType.length + 1}`
        : `new_${type}`;
      const available = allComponentSchemas[type === "processor" ? "processor" : type === "input" ? "input" : "output"];

      setPendingNode({
        kind: "topLevel",
        topLevelType: type,
        label,
        componentId: "",
        component: "",
        availableComponents: available,
      });
    },
    [nodes, addToast, allComponentSchemas]
  );

  const handleNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNodeId(node.id);
    setContextMenu(null);
  }, []);

  const handleNodeDoubleClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNodeId(node.id);
    setEditingNodeId(node.id);
  }, []);

  const handleNodeContextMenu = useCallback((event: React.MouseEvent, node: Node) => {
    event.preventDefault();
    setSelectedNodeId(node.id);
    setContextMenu({ x: event.clientX, y: event.clientY, nodeId: node.id });
  }, []);

  const handlePaneClick = useCallback(() => {
    setSelectedNodeId(null);
    setContextMenu(null);
  }, []);

  const editingNode = useMemo(() => {
    const n = nodes.find((n) => n.id === editingNodeId);
    if (!n) return null;
    const d = n.data as StreamFlowNodeData;
    const isCaseStart = (d as any).isCaseStart === true;
    const isOutputCaseStart = isCaseStart && d.type === "output";
    return {
      id: n.id,
      data: {
        label: d.label,
        type: d.type,
        componentId: d.componentId,
        component: d.component,
        configYaml: d.configYaml,
      } as FlowNodeData,
      isGroup: (d.isGroup === true || isCaseStart || n.type === "switchProcessorGroupNode") && !isOutputCaseStart,
      isGroupChild: !!n.parentId,
      isSwitchCaseProc: !!(d as any).switchCaseId,
      isOutputCaseStart,
    };
  }, [nodes, editingNodeId]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (editingNodeId || tryDialogOpen || deleteConfirmNodeId || pendingNode) return;
      if (!selectedNodeId) return;
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement || e.target instanceof HTMLSelectElement) return;

      if (e.key === "e" || e.key === "E") {
        e.preventDefault();
        setEditingNodeId(selectedNodeId);
      } else if (e.key === "Delete" || e.key === "Backspace") {
        e.preventDefault();
        setDeleteConfirmNodeId(selectedNodeId);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [selectedNodeId, editingNodeId, tryDialogOpen, deleteConfirmNodeId, pendingNode]);

  useEffect(() => {
    if (!contextMenu) return;
    const close = (e: Event) => {
      // Don't close if clicking inside the context menu itself
      const target = e.target as HTMLElement;
      if (target.closest("[data-context-menu]")) return;
      setContextMenu(null);
    };
    // Use setTimeout to avoid the current event from immediately closing the menu
    const timeoutId = setTimeout(() => {
      window.addEventListener("click", close);
      window.addEventListener("contextmenu", close);
    }, 0);
    return () => {
      clearTimeout(timeoutId);
      window.removeEventListener("click", close);
      window.removeEventListener("contextmenu", close);
    };
  }, [contextMenu]);

  const childFilteredSchemas = useMemo(() => {
    // Switch case chain processors: restrict to non-switch, non-catch
    if (editingNode?.isSwitchCaseProc) {
      return {
        ...allComponentSchemas,
        processor: allComponentSchemas.processor.filter((p) => p.id !== "switch" && p.id !== "catch"),
      };
    }
    // Output switch case starts: restrict output components (no switch, no broker)
    if (editingNode?.isOutputCaseStart) {
      return {
        ...allComponentSchemas,
        output: allComponentSchemas.output.filter((o) => o.id !== "switch" && o.id !== "broker"),
      };
    }
    if (!editingNode?.isGroupChild) return allComponentSchemas;
    const parentNode = nodes.find((n) => n.id === nodes.find((cn) => cn.id === editingNode.id)?.parentId);
    const parentComponentId = parentNode && (parentNode.data as StreamFlowNodeData).componentId;
    if (parentComponentId === "broker") {
      const parentType = (parentNode!.data as StreamFlowNodeData).type;
      if (parentType === "input") {
        return {
          ...allComponentSchemas,
          input: allComponentSchemas.input.filter((o) => o.id !== "broker"),
        };
      }
      return {
        ...allComponentSchemas,
        output: allComponentSchemas.output.filter((o) => o.id !== "broker" && o.id !== "switch"),
      };
    }
    if (parentComponentId === "switch") {
      return {
        ...allComponentSchemas,
        output: allComponentSchemas.output.filter((o) => o.id !== "switch" && o.id !== "broker"),
      };
    }
    return {
      ...allComponentSchemas,
      processor: allComponentSchemas.processor.filter((p) => p.id !== "catch"),
    };
  }, [allComponentSchemas, editingNode?.isGroupChild, editingNode?.isSwitchCaseProc, editingNode?.isOutputCaseStart, editingNode?.id, nodes]);

  const isMcpServer = useMemo(
    () => nodes.some((n) => (n.data as StreamFlowNodeData).type === "input" && (n.data as StreamFlowNodeData).componentId === "mcp_tool"),
    [nodes]
  );

  const handleUpdateNode = useCallback(
    (nodeId: string, data: FlowNodeData) => {
      setNodes((nds) => {
        const currentNode = nds.find((n) => n.id === nodeId);
        const prevComponentId = currentNode ? (currentNode.data as StreamFlowNodeData).componentId : undefined;

        // Clear disconnected state on all nodes when any config changes
        let updated = nds.map((n) =>
          n.id === nodeId
            ? { ...n, data: { ...n.data, ...data, nodeId: n.id, disconnected: false } }
            : (n.data as StreamFlowNodeData).disconnected
              ? { ...n, data: { ...n.data, disconnected: false } }
              : n
        );

        // Transform to branch group when component changes to "branch"
        if (data.componentId === "branch" && prevComponentId !== "branch" && data.type === "processor") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "branchGroupNode",
                  style: { width: GROUP_WIDTH, height: CATCH_GROUP_MIN_HEIGHT },
                  position: { ...n.position, y: n.position.y - 25 },
                  data: { ...n.data, isGroup: true, childCount: 0, configYaml: "" },
                }
              : n
          );
        }

        // Transform back from branch group when component changes away from "branch"
        if (prevComponentId === "branch" && data.componentId !== "branch" && data.type === "processor") {
          const childIds = updated.filter((n) => n.parentId === nodeId).map((n) => n.id);
          updated = updated.filter((n) => n.parentId !== nodeId);
          setEdges((eds) => eds.filter((e) => !childIds.includes(e.source) && !childIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "processorNode",
                  style: undefined,
                  position: { ...n.position, y: n.position.y + 25 },
                  data: { ...n.data, isGroup: false, childCount: undefined },
                }
              : n
          );
        }

        // Transform to catch group when component changes to "catch"
        if (data.componentId === "catch" && prevComponentId !== "catch" && data.type === "processor") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "catchGroupNode",
                  style: { width: GROUP_WIDTH, height: CATCH_GROUP_MIN_HEIGHT },
                  position: { ...n.position, y: n.position.y - 25 },
                  data: { ...n.data, isGroup: true, childCount: 0, configYaml: "" },
                }
              : n
          );
        }

        // Transform back from catch group when component changes away from "catch"
        if (prevComponentId === "catch" && data.componentId !== "catch" && data.type === "processor") {
          const childIds = updated.filter((n) => n.parentId === nodeId).map((n) => n.id);
          updated = updated.filter((n) => n.parentId !== nodeId);
          setEdges((eds) => eds.filter((e) => !childIds.includes(e.source) && !childIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "processorNode",
                  style: undefined,
                  position: { ...n.position, y: n.position.y + 25 },
                  data: { ...n.data, isGroup: false, childCount: undefined },
                }
              : n
          );
        }

        // Transform to broker group when output component changes to "broker"
        if (data.componentId === "broker" && prevComponentId !== "broker" && data.type === "output") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "brokerGroupNode",
                  style: { width: GROUP_WIDTH, height: BROKER_GROUP_MIN_HEIGHT },
                  position: { ...n.position, y: n.position.y - 25 },
                  data: { ...n.data, isGroup: true, childCount: 0, configYaml: "", brokerPattern: "fan_out" },
                }
              : n
          );
        }

        // Transform back from broker group when component changes away from "broker"
        if (prevComponentId === "broker" && data.componentId !== "broker" && data.type === "output") {
          const childIds = updated.filter((n) => n.parentId === nodeId).map((n) => n.id);
          updated = updated.filter((n) => n.parentId !== nodeId);
          setEdges((eds) => eds.filter((e) => !childIds.includes(e.source) && !childIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "outputNode",
                  style: undefined,
                  position: { ...n.position, y: n.position.y + 25 },
                  data: { ...n.data, isGroup: false, childCount: undefined, brokerPattern: undefined },
                }
              : n
          );
        }

        // Transform to broker input group when input component changes to "broker"
        if (data.componentId === "broker" && prevComponentId !== "broker" && data.type === "input") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "brokerInputGroupNode",
                  style: { width: GROUP_WIDTH, height: BROKER_INPUT_GROUP_MIN_HEIGHT },
                  position: { ...n.position, y: n.position.y - 25 },
                  data: { ...n.data, isGroup: true, childCount: 0, configYaml: "" },
                }
              : n
          );
        }

        // Transform back from broker input group when component changes away from "broker"
        if (prevComponentId === "broker" && data.componentId !== "broker" && data.type === "input") {
          const childIds = updated.filter((n) => n.parentId === nodeId).map((n) => n.id);
          updated = updated.filter((n) => n.parentId !== nodeId);
          setEdges((eds) => eds.filter((e) => !childIds.includes(e.source) && !childIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "inputNode",
                  style: undefined,
                  position: { ...n.position, y: n.position.y + 25 },
                  data: { ...n.data, isGroup: false, childCount: undefined },
                }
              : n
          );
        }

        // Transform to switch output node when output component changes to "switch"
        if (data.componentId === "switch" && prevComponentId !== "switch" && data.type === "output") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "switchProcessorGroupNode",
                  style: undefined,
                  data: { ...n.data, isGroup: false, childCount: 0, configYaml: "" },
                }
              : n
          );
        }

        // Transform back from switch output node when component changes away from "switch"
        if (prevComponentId === "switch" && data.componentId !== "switch" && data.type === "output") {
          const caseStartIds = updated
            .filter((n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === nodeId)
            .map((n) => n.id);
          updated = updated.filter((n) => !caseStartIds.includes(n.id));
          setEdges((eds) => eds.filter((e) => !caseStartIds.includes(e.source) && !caseStartIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "outputNode",
                  style: undefined,
                  data: { ...n.data, isGroup: false, childCount: undefined },
                }
              : n
          );
        }

        // Transform to processor switch node when processor component changes to "switch"
        if (data.componentId === "switch" && prevComponentId !== "switch" && data.type === "processor") {
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "switchProcessorGroupNode",
                  style: undefined,
                  data: { ...n.data, isGroup: false, childCount: 0, configYaml: "" },
                }
              : n
          );
        }

        // Transform back from processor switch when component changes away from "switch"
        if (prevComponentId === "switch" && data.componentId !== "switch" && data.type === "processor") {
          // Collect case start IDs and their chain processor IDs (edge-based)
          const caseStartIds = updated
            .filter((n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === nodeId)
            .map((n) => n.id);
          const caseProcIds = updated
            .filter((n) => caseStartIds.includes((n.data as any).switchCaseId))
            .map((n) => n.id);
          const allRemoveIds = [...caseStartIds, ...caseProcIds];
          updated = updated.filter((n) => !allRemoveIds.includes(n.id));
          setEdges((eds) => eds.filter((e) => !allRemoveIds.includes(e.source) && !allRemoveIds.includes(e.target)));
          updated = updated.map((n) =>
            n.id === nodeId
              ? {
                  ...n,
                  type: "processorNode",
                  style: undefined,
                  data: { ...n.data, isGroup: false, childCount: undefined },
                }
              : n
          );
        }

        if (data.type === "input" && data.componentId === "mcp_tool") {
          const outputNode = updated.find((n) => (n.data as StreamFlowNodeData).type === "output");
          if (outputNode) {
            if (
              (outputNode.data as StreamFlowNodeData).componentId !== "sync_response" ||
              (outputNode.data as StreamFlowNodeData).label !== "mcp_tool_response"
            ) {
              updated = updated.map((n) =>
                n.id === outputNode.id
                  ? { ...n, data: { ...n.data, label: "mcp_tool_response", componentId: "sync_response", component: "sync_response", configYaml: "" } }
                  : n
              );
            }
          } else {
            const newId = uuidv4();
            const maxX = Math.max(...updated.map((n) => n.position.x), 0);
            updated.push({
              id: newId,
              type: "outputNode",
              position: { x: maxX + NODE_SPACING_X, y: NODE_Y },
              data: { label: "mcp_tool_response", type: "output", componentId: "sync_response", component: "sync_response", configYaml: "", nodeId: newId },
            });
            const sourceNode = [...updated]
              .filter((n) => (n.data as StreamFlowNodeData).type !== "output")
              .sort((a, b) => b.position.x - a.position.x)[0];
            if (sourceNode) {
              setEdges((eds) => [
                ...eds,
                { id: `e-${sourceNode.id}-${newId}`, source: sourceNode.id, target: newId, type: "pipeline" },
              ]);
            }
          }
        }

        return updated;
      });
    },
    [setNodes, setEdges]
  );

  const handleDeleteNode = useCallback(
    (nodeId: string) => {
      const nodeToDelete = nodes.find((n) => n.id === nodeId);
      if (!nodeToDelete) return;

      // Deleting a switch case start node (decision tree)
      if (nodeToDelete.type === "switchCaseStartNode") {
        const switchId = (nodeToDelete.data as any).switchId as string;
        // Collect all processors in this case chain
        const caseProcIds = nodes
          .filter((n) => (n.data as any).switchCaseId === nodeId)
          .map((n) => n.id);
        const allRemoveIds = [nodeId, ...caseProcIds];
        setEdges((eds) => eds.filter((e) => !allRemoveIds.includes(e.source) && !allRemoveIds.includes(e.target)));
        setNodes((nds) => {
          const updated = nds.filter((n) => !allRemoveIds.includes(n.id));
          // Re-layout remaining case starts
          const switchNode = updated.find((n) => n.id === switchId);
          if (switchNode) {
            const remainingCases = updated.filter(
              (n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === switchId
            );
            const positions = layoutSwitchCases(switchNode.position.x, switchNode.position.y, remainingCases.length);
            const sorted = remainingCases.sort(sortCasesDefaultLast);
            return updated.map((n) => {
              if (n.id === switchId) return { ...n, data: { ...n.data, childCount: remainingCases.length } };
              const idx = sorted.findIndex((s) => s.id === n.id);
              if (idx >= 0) return { ...n, position: positions[idx] };
              return n;
            });
          }
          return updated;
        });
      }
      // Deleting a processor that belongs to a switch case chain
      else if ((nodeToDelete.data as any).switchCaseId) {
        // Re-link: find incoming and outgoing edges, bridge them
        const inEdge = edges.find((e) => e.target === nodeId);
        const outEdge = edges.find((e) => e.source === nodeId);
        setEdges((eds) => {
          let updated = eds.filter((e) => e.source !== nodeId && e.target !== nodeId);
          if (inEdge && outEdge) {
            updated.push({ id: `e-${inEdge.source}-${outEdge.target}`, source: inEdge.source, target: outEdge.target, type: "pipeline" });
          }
          return updated;
        });
        setNodes((nds) => nds.filter((n) => n.id !== nodeId));
      }
      else if (nodeToDelete.parentId) {
        const groupId = nodeToDelete.parentId;
        const groupNode = nodes.find((n) => n.id === groupId);

        // Flat group child (catch, broker)
        {
          const isBrokerOutput = groupNode?.type === "brokerGroupNode";
          const isBrokerInput = groupNode?.type === "brokerInputGroupNode";
          const childYStart = isBrokerOutput ? BROKER_CHILD_Y_START : isBrokerInput ? BROKER_INPUT_CHILD_Y_START : CATCH_CHILD_Y_START;
          const calcHeight = isBrokerOutput ? calcBrokerGroupHeight : isBrokerInput ? calcBrokerInputGroupHeight : calcCatchGroupHeight;
          const brokerPattern = isBrokerOutput ? (groupNode?.data as StreamFlowNodeData)?.brokerPattern : undefined;
          const shouldChain = isBrokerInput ? false : isBrokerOutput ? isBrokerSequential(brokerPattern) : true;

          setEdges((eds) => {
            const cleaned = eds.filter((e) => e.source !== nodeId && e.target !== nodeId);
            const remainingSiblings = nodes.filter((n) => n.parentId === groupId && n.id !== nodeId).sort((a, b) => a.position.y - b.position.y);
            const sibIds = new Set(remainingSiblings.map((s) => s.id));
            const kept = cleaned.filter((e) => !((e.data as any)?.internal && (sibIds.has(e.source) || sibIds.has(e.target))));
            if (!shouldChain) return kept;
            const newEdges: Edge[] = [];
            for (let i = 0; i < remainingSiblings.length - 1; i++) {
              newEdges.push({ id: `e-internal-${remainingSiblings[i].id}-${remainingSiblings[i + 1].id}`, source: remainingSiblings[i].id, target: remainingSiblings[i + 1].id, type: "pipeline", data: { internal: true } });
            }
            return [...kept, ...newEdges];
          });

          setNodes((nds) => {
            const remaining = nds.filter((n) => n.id !== nodeId);
            const siblings = remaining.filter((n) => n.parentId === groupId).sort((a, b) => a.position.y - b.position.y);
            return remaining.map((n) => {
              if (n.id === groupId) return { ...n, style: { ...n.style, width: GROUP_WIDTH, height: calcHeight(siblings.length) }, data: { ...n.data, childCount: siblings.length } };
              const sibIdx = siblings.findIndex((s) => s.id === n.id);
              if (sibIdx >= 0) return { ...n, position: { x: CHILD_X, y: childYStart + sibIdx * (CHILD_NODE_HEIGHT + CHILD_GAP_Y) }, data: { ...n.data, childIndex: sibIdx } };
              return n;
            });
          });
        }
      } else {
        // Deleting a top-level node — also delete children and switch tree nodes
        const childIds = nodes.filter((n) => n.parentId === nodeId).map((n) => n.id);
        // For switch processor: also collect case starts and their chain processors
        const caseStartIds = nodes
          .filter((n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === nodeId)
          .map((n) => n.id);
        const caseProcIds = nodes
          .filter((n) => caseStartIds.includes((n.data as any).switchCaseId))
          .map((n) => n.id);
        const allIds = [nodeId, ...childIds, ...caseStartIds, ...caseProcIds];
        setNodes((nds) => nds.filter((n) => !allIds.includes(n.id)));
        setEdges((eds) => eds.filter((e) => !allIds.includes(e.source) && !allIds.includes(e.target)));
      }

      if (selectedNodeId === nodeId) setSelectedNodeId(null);
      if (editingNodeId === nodeId) setEditingNodeId(null);
    },
    [nodes, edges, setNodes, setEdges, selectedNodeId, editingNodeId]
  );

  const validateRequiredFields = useCallback(
    (nodeData: FlowNodeData): { isValid: boolean; missingFields: string[]; error?: string } => {
      if (!nodeData.componentId) {
        return { isValid: false, missingFields: ["component selection"] };
      }

      const component = allComponentSchemas[nodeData.type]?.find((c) => c.id === nodeData.componentId);
      const schema = component?.schema || {};
      const properties = schema.properties || {};

      if (Object.keys(properties).length === 0) {
        return { isValid: true, missingFields: [] };
      }

      let configData: any = {};
      if (nodeData.configYaml && nodeData.configYaml.trim()) {
        try {
          configData = yaml.load(nodeData.configYaml) || {};
        } catch {
          return { isValid: false, missingFields: [], error: "Invalid YAML format" };
        }
      }

      const missingFields: string[] = [];

      if (schema.flat === true) {
        const hasContent =
          Object.keys(configData).length > 0 ||
          (Array.isArray(configData) && configData.length > 0) ||
          (typeof configData === "string" && configData.trim().length > 0);
        if (!hasContent) {
          Object.entries(properties).forEach(([fieldKey, fieldSchema]) => {
            if ((fieldSchema as any).required === true && missingFields.length === 0) {
              missingFields.push(fieldKey);
            }
          });
        }
      } else {
        Object.entries(properties).forEach(([fieldKey, fieldSchema]) => {
          if ((fieldSchema as any).required === true) {
            const value = configData[fieldKey];
            if (value === undefined || value === null || value === "") {
              missingFields.push(fieldKey);
            }
          }
        });
      }

      return { isValid: missingFields.length === 0, missingFields };
    },
    [allComponentSchemas]
  );

  const findDisconnectedNodes = useCallback((): Set<string> => {
    // Only check main pipeline nodes (not children inside groups, not switch tree nodes)
    const topLevelNodes = nodes.filter(
      (n) => !n.parentId && n.type !== "switchCaseStartNode" && !(n.data as any).switchCaseId
    );
    if (topLevelNodes.length === 0) return new Set();

    const adj = new Map<string, string[]>();
    const radj = new Map<string, string[]>();
    for (const n of topLevelNodes) {
      adj.set(n.id, []);
      radj.set(n.id, []);
    }
    const topLevelIds = new Set(topLevelNodes.map((n) => n.id));
    const externalEdges = edges.filter((e) => !(e.data as any)?.internal);
    for (const e of externalEdges) {
      if (topLevelIds.has(e.source) && topLevelIds.has(e.target)) {
        adj.get(e.source)?.push(e.target);
        radj.get(e.target)?.push(e.source);
      }
    }

    // Convergence: map switch tree edges to their parent switch node
    // Case chain end nodes (switchCaseId) may connect to main pipeline nodes
    const switchTreeNodes = nodes.filter(
      (n) => n.type === "switchCaseStartNode" || (n.data as any).switchCaseId
    );
    const switchTreeIds = new Set(switchTreeNodes.map((n) => n.id));
    for (const e of externalEdges) {
      if (switchTreeIds.has(e.source) && topLevelIds.has(e.target)) {
        // Find the switch node this tree node belongs to
        const srcNode = nodes.find((n) => n.id === e.source);
        if (!srcNode) continue;
        let switchId = (srcNode.data as any).switchId;
        if (!switchId && (srcNode.data as any).switchCaseId) {
          const caseStart = nodes.find((n) => n.id === (srcNode.data as any).switchCaseId);
          switchId = caseStart ? (caseStart.data as any).switchId : undefined;
        }
        if (switchId && topLevelIds.has(switchId)) {
          if (!adj.get(switchId)?.includes(e.target)) {
            adj.get(switchId)?.push(e.target);
          }
          if (!radj.get(e.target)?.includes(switchId)) {
            radj.get(e.target)?.push(switchId);
          }
        }
      }
    }

    const inputNode = topLevelNodes.find((n) => (n.data as StreamFlowNodeData).type === "input");
    const outputNode = topLevelNodes.find((n) => (n.data as StreamFlowNodeData).type === "output");

    if (!inputNode || !outputNode) return new Set(topLevelNodes.map((n) => n.id));

    const forwardReachable = new Set<string>();
    const queue = [inputNode.id];
    forwardReachable.add(inputNode.id);
    while (queue.length > 0) {
      const curr = queue.shift()!;
      for (const next of adj.get(curr) || []) {
        if (!forwardReachable.has(next)) {
          forwardReachable.add(next);
          queue.push(next);
        }
      }
    }

    const backwardReachable = new Set<string>();
    const bQueue = [outputNode.id];
    backwardReachable.add(outputNode.id);
    while (bQueue.length > 0) {
      const curr = bQueue.shift()!;
      for (const prev of radj.get(curr) || []) {
        if (!backwardReachable.has(prev)) {
          backwardReachable.add(prev);
          bQueue.push(prev);
        }
      }
    }

    const disconnected = new Set<string>();
    for (const n of topLevelNodes) {
      if (!forwardReachable.has(n.id) || !backwardReachable.has(n.id)) {
        disconnected.add(n.id);
      }
    }

    return disconnected;
  }, [nodes, edges]);

  const serializeCatchGroup = useCallback(
    (groupNode: Node): FlowNodeData => {
      const d = groupNode.data as StreamFlowNodeData;
      const children = nodes.filter((cn) => cn.parentId === groupNode.id).sort((a, b) => a.position.y - b.position.y);
      const childConfigs = children.map((child) => {
        const cd = child.data as StreamFlowNodeData;
        const comp = allComponentSchemas.processor.find((c) => c.id === cd.componentId);
        const componentName = comp?.component || cd.componentId || "";
        const entry: any = {};
        if (comp?.schema?.flat) {
          entry[componentName] = cd.configYaml?.trim() || "";
        } else {
          let config: any = {};
          if (cd.configYaml?.trim()) { try { config = yaml.load(cd.configYaml) || {}; } catch {} }
          entry[componentName] = config;
        }
        if (cd.label) {
          entry.label = cd.label;
        }
        return entry;
      });
      const catchYaml = childConfigs.length > 0
        ? yaml.dump(childConfigs, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false })
        : "";
      return { label: d.label, type: "processor", componentId: "catch", component: "catch", configYaml: catchYaml };
    },
    [nodes, allComponentSchemas]
  );

  const serializeBranchGroup = useCallback(
    (groupNode: Node): FlowNodeData => {
      const d = groupNode.data as StreamFlowNodeData;
      const children = nodes.filter((cn) => cn.parentId === groupNode.id).sort((a, b) => a.position.y - b.position.y);
      const childConfigs = children.map((child) => {
        const cd = child.data as StreamFlowNodeData;
        const comp = allComponentSchemas.processor.find((c) => c.id === cd.componentId);
        const componentName = comp?.component || cd.componentId || "";
        const entry: any = {};
        if (comp?.schema?.flat) {
          entry[componentName] = cd.configYaml?.trim() || "";
        } else {
          let config: any = {};
          if (cd.configYaml?.trim()) { try { config = yaml.load(cd.configYaml) || {}; } catch {} }
          entry[componentName] = config;
        }
        if (cd.label) {
          entry.label = cd.label;
        }
        return entry;
      });

      let groupConfig: any = {};
      if (d.configYaml?.trim()) {
        try { groupConfig = yaml.load(d.configYaml) || {}; } catch {}
      }

      const branchObj: any = {};
      if (groupConfig.request_map) branchObj.request_map = groupConfig.request_map;
      branchObj.processors = childConfigs;
      if (groupConfig.result_map) branchObj.result_map = groupConfig.result_map;

      const branchYaml = yaml.dump(branchObj, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false });
      return { label: d.label, type: "processor", componentId: "branch", component: "branch", configYaml: branchYaml };
    },
    [nodes, allComponentSchemas]
  );

  const serializeBrokerGroup = useCallback(
    (groupNode: Node): FlowNodeData => {
      const d = groupNode.data as StreamFlowNodeData;
      const children = nodes.filter((cn) => cn.parentId === groupNode.id).sort((a, b) => a.position.y - b.position.y);

      // Build outputs array from children
      const outputsList = children.map((child) => {
        const cd = child.data as StreamFlowNodeData;
        const comp = allComponentSchemas.output.find((c) => c.id === cd.componentId);
        const componentName = comp?.component || cd.componentId || "";
        const entry: any = {};
        if (comp?.schema?.flat) {
          entry[componentName] = cd.configYaml?.trim() || "";
        } else {
          let config: any = {};
          if (cd.configYaml?.trim()) { try { config = yaml.load(cd.configYaml) || {}; } catch {} }
          entry[componentName] = config;
        }
        if (cd.label) {
          entry.label = cd.label;
        }
        return entry;
      });

      // Build broker config with pattern + outputs
      const brokerConfig: any = {};
      if (d.brokerPattern) brokerConfig.pattern = d.brokerPattern;

      // Parse any extra config from group's configYaml (copies, batching)
      if (d.configYaml?.trim()) {
        try {
          const extra = yaml.load(d.configYaml) as any;
          if (extra && typeof extra === "object") {
            Object.assign(brokerConfig, extra);
          }
        } catch {}
      }

      brokerConfig.outputs = outputsList;

      const brokerYaml = yaml.dump(brokerConfig, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false });
      return { label: d.label, type: "output", componentId: "broker", component: "broker", configYaml: brokerYaml };
    },
    [nodes, allComponentSchemas]
  );

  const serializeBrokerInputGroup = useCallback(
    (groupNode: Node): FlowNodeData => {
      const d = groupNode.data as StreamFlowNodeData;
      const children = nodes.filter((cn) => cn.parentId === groupNode.id).sort((a, b) => a.position.y - b.position.y);

      const inputsList = children.map((child) => {
        const cd = child.data as StreamFlowNodeData;
        const comp = allComponentSchemas.input.find((c) => c.id === cd.componentId);
        const componentName = comp?.component || cd.componentId || "";
        const entry: any = {};
        if (comp?.schema?.flat) {
          entry[componentName] = cd.configYaml?.trim() || "";
        } else {
          let config: any = {};
          if (cd.configYaml?.trim()) { try { config = yaml.load(cd.configYaml) || {}; } catch {} }
          entry[componentName] = config;
        }
        if (cd.label) {
          entry.label = cd.label;
        }
        return entry;
      });

      const brokerConfig: any = {};

      if (d.configYaml?.trim()) {
        try {
          const extra = yaml.load(d.configYaml) as any;
          if (extra && typeof extra === "object") {
            Object.assign(brokerConfig, extra);
          }
        } catch {}
      }

      brokerConfig.inputs = inputsList;

      const brokerYaml = yaml.dump(brokerConfig, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false });
      return { label: d.label, type: "input", componentId: "broker", component: "broker", configYaml: brokerYaml };
    },
    [nodes, allComponentSchemas]
  );

  const serializeSwitchGroup = useCallback(
    (switchNode: Node): FlowNodeData => {
      const d = switchNode.data as StreamFlowNodeData;
      const caseStarts = nodes
        .filter((n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === switchNode.id)
        .sort(sortCasesDefaultLast);

      const casesList = caseStarts.map((caseNode) => {
        const cd = caseNode.data as StreamFlowNodeData;
        const comp = allComponentSchemas.output.find((c) => c.id === cd.componentId);
        const componentName = comp?.component || cd.componentId || "";
        const caseEntry: any = {};
        const check = (cd as any).caseCheck;
        if (check) caseEntry.check = check;
        const output: any = {};
        if (comp?.schema?.flat) {
          output[componentName] = cd.configYaml?.trim() || "";
        } else {
          let config: any = {};
          if (cd.configYaml?.trim()) { try { config = yaml.load(cd.configYaml) || {}; } catch {} }
          output[componentName] = config;
        }
        caseEntry.output = output;
        if ((cd as any).caseContinue === true) caseEntry.continue = true;
        return caseEntry;
      });

      const switchConfig: any = {};

      if (d.configYaml?.trim()) {
        try {
          const extra = yaml.load(d.configYaml) as any;
          if (extra && typeof extra === "object") {
            Object.assign(switchConfig, extra);
          }
        } catch {}
      }

      switchConfig.cases = casesList;

      const switchYaml = yaml.dump(switchConfig, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false });
      return { label: d.label, type: "output", componentId: "switch", component: "switch", configYaml: switchYaml };
    },
    [nodes, allComponentSchemas]
  );

  const serializeProcessorSwitchGroup = useCallback(
    (switchNode: Node): FlowNodeData => {
      const d = switchNode.data as StreamFlowNodeData;
      // Find case start nodes connected via edges
      const caseStarts = nodes
        .filter((n) => n.type === "switchCaseStartNode" && (n.data as any).switchId === switchNode.id)
        .sort(sortCasesDefaultLast);

      const casesList = caseStarts.map((caseNode) => {
        const cd = caseNode.data as StreamFlowNodeData;
        const caseEntry: any = {};
        const check = (cd as any).caseCheck;
        if (check) caseEntry.check = check;

        // Walk edge chain from case start to collect processors
        const processorsList: any[] = [];
        let currentId: string | undefined = caseNode.id;
        while (currentId) {
          const outEdge = edges.find((e) => e.source === currentId && !(e.data as any)?.switchCase);
          if (!outEdge) break;
          const nextNode = nodes.find((n) => n.id === outEdge.target);
          if (!nextNode || nextNode.type !== "processorNode") break;
          const pd = nextNode.data as StreamFlowNodeData;
          const comp = allComponentSchemas.processor.find((c) => c.id === pd.componentId);
          const componentName = comp?.component || pd.componentId || "";
          const procObj: any = {};
          if (comp?.schema?.flat) {
            procObj[componentName] = pd.configYaml?.trim() || "";
          } else {
            let config: any = {};
            if (pd.configYaml?.trim()) { try { config = yaml.load(pd.configYaml) || {}; } catch {} }
            procObj[componentName] = config;
          }
          processorsList.push(procObj);
          currentId = nextNode.id;
        }

        caseEntry.processors = processorsList;
        if ((cd as any).caseFallthrough === true) caseEntry.fallthrough = true;
        return caseEntry;
      });

      const switchYaml = yaml.dump(casesList, { lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false });
      return { label: d.label, type: "processor", componentId: "switch", component: "switch", configYaml: switchYaml };
    },
    [nodes, edges, allComponentSchemas]
  );

  const handleValidate = useCallback(async () => {
    if (!onValidate) return;

    const inputNode = nodes.find((n) => (n.data as StreamFlowNodeData).type === "input");
    const outputNode = nodes.find((n) => (n.data as StreamFlowNodeData).type === "output");

    if (!inputNode || !outputNode) {
      addToast({ id: "validate-err", title: "Invalid", description: "Flow must have both an input and output node.", variant: "error", duration: 5000 });
      return;
    }

    const disconnected = findDisconnectedNodes();
    if (disconnected.size > 0) {
      setNodes((nds) =>
        nds.map((n) =>
          disconnected.has(n.id)
            ? { ...n, data: { ...n.data, disconnected: true } }
            : { ...n, data: { ...n.data, disconnected: false } }
        )
      );
      addToast({ id: "validate-err", title: "Invalid", description: `${disconnected.size} node(s) are not connected to the pipeline.`, variant: "error", duration: 5000 });
      return;
    }

    setIsValidating(true);
    try {
      const nodeDataList = nodes
        .filter((n) => !n.parentId && n.type !== "switchCaseStartNode" && !(n.data as any).switchCaseId)
        .map((n) => {
          const d = n.data as StreamFlowNodeData;
          if (d.isGroup && d.componentId === "branch") return serializeBranchGroup(n);
          if (d.isGroup && d.componentId === "catch") return serializeCatchGroup(n);
          if (d.isGroup && d.componentId === "broker" && d.type === "output") return serializeBrokerGroup(n);
          if (d.isGroup && d.componentId === "broker" && d.type === "input") return serializeBrokerInputGroup(n);
          if (n.type === "switchProcessorGroupNode" && d.type === "output") return serializeSwitchGroup(n);
          if (n.type === "switchProcessorGroupNode" && d.type === "processor") return serializeProcessorSwitchGroup(n);
          return { label: d.label, type: d.type, componentId: d.componentId, component: d.component, configYaml: d.configYaml } as FlowNodeData;
        });
      const result = await onValidate({ name, status, bufferId, nodes: nodeDataList });
      if (result?.valid) {
        addToast({ id: "validate-ok", title: "Valid", description: "Configuration is valid.", variant: "success", duration: 3000 });
      } else {
        addToast({ id: "validate-err", title: "Invalid", description: result?.error || "Configuration is invalid.", variant: "error", duration: 5000 });
      }
    } catch {
      addToast({ id: "validate-fail", title: "Error", description: "Failed to reach validation endpoint.", variant: "error" });
    } finally {
      setIsValidating(false);
    }
  }, [onValidate, name, status, bufferId, nodes, addToast, serializeBranchGroup, serializeCatchGroup, serializeBrokerGroup, serializeBrokerInputGroup, serializeSwitchGroup, serializeProcessorSwitchGroup, findDisconnectedNodes, setNodes]);

  const handleTrySubmit = useCallback(async () => {
    if (!onTry) return;
    const nonEmpty = tryMessages.filter((m) => m.trim());
    if (nonEmpty.length === 0) {
      addToast({ id: "try-no-msg", title: "No Messages", description: "Enter at least one test message.", variant: "warning" });
      return;
    }
    setIsTrying(true);
    setTryResult(null);
    try {
      const topProcessors = nodes.filter(
        (n) => (n.data as StreamFlowNodeData).type === "processor" && !n.parentId
          && n.type !== "switchCaseStartNode" && !(n.data as any).switchCaseId
      );
      const processors = topProcessors.map((n) => {
        const d = n.data as StreamFlowNodeData;
        if (d.isGroup && d.componentId === "branch") {
          const serialized = serializeBranchGroup(n);
          return { label: serialized.label, component: "branch", config: serialized.configYaml || "" };
        }
        if (d.isGroup && d.componentId === "catch") {
          const serialized = serializeCatchGroup(n);
          return { label: serialized.label, component: "catch", config: serialized.configYaml || "" };
        }
        if (n.type === "switchProcessorGroupNode") {
          const serialized = serializeProcessorSwitchGroup(n);
          return { label: serialized.label, component: "switch", config: serialized.configYaml || "" };
        }
        const comp = allComponentSchemas.processor.find((c) => c.id === d.componentId);
        return { label: d.label, component: comp?.component || d.componentId || "", config: d.configYaml || "" };
      });
      const result = await onTry({ processors, messages: nonEmpty.map((content) => ({ content })) });
      setTryResult(result);
    } catch {
      setTryResult({ outputs: [], error: "Failed to reach try endpoint." });
    } finally {
      setIsTrying(false);
    }
  }, [onTry, tryMessages, nodes, allComponentSchemas, addToast, serializeBranchGroup, serializeCatchGroup, serializeProcessorSwitchGroup]);

  const handleSave = useCallback(() => {
    if (!name.trim()) {
      addToast({ id: "name-req", title: "Validation Error", description: "Flow name is required.", variant: "warning" });
      return;
    }

    const inputNode = nodes.find((n) => (n.data as StreamFlowNodeData).type === "input");
    const outputNode = nodes.find((n) => (n.data as StreamFlowNodeData).type === "output");

    // Serialize flow state for persistence
    const builderState = JSON.stringify({ nodes, edges });

    // Determine if stream is ready (fully connected and configured)
    let isReady = true;

    if (!inputNode || !outputNode) {
      isReady = false;
    }

    const disconnected = findDisconnectedNodes();
    if (disconnected.size > 0) {
      isReady = false;
      setNodes((nds) =>
        nds.map((n) =>
          disconnected.has(n.id)
            ? { ...n, data: { ...n.data, disconnected: true } }
            : { ...n, data: { ...n.data, disconnected: false } }
        )
      );
      addToast({
        id: "disconnected",
        title: "Disconnected Nodes",
        description: `${disconnected.size} node(s) are not connected. Flow will be saved as not ready.`,
        variant: "warning",
        duration: 5000,
      });
    }

    // Validate nodes (skip switch case starts and case chain processors — they're validated separately)
    for (const n of nodes) {
      if (n.parentId) continue;
      // Skip case start nodes and case chain processors — handled below
      if (n.type === "switchCaseStartNode") continue;
      if ((n.data as any).switchCaseId) continue;
      const d = n.data as StreamFlowNodeData;
      if (d.isGroup) {
        const groupType = d.componentId === "switch" ? "Switch Output" : d.componentId === "broker" ? (d.type === "input" ? "Broker Input" : "Broker Output") : "Catch Processor";
        const leafNodes = nodes.filter((cn) => cn.parentId === n.id);
        for (const child of leafNodes) {
          const cd = child.data as StreamFlowNodeData;
          if (cd.isGroup) continue;
          const childData: FlowNodeData = { label: cd.label, type: cd.type, componentId: cd.componentId, component: cd.component, configYaml: cd.configYaml };
          const validation = validateRequiredFields(childData);
          if (!validation.isValid) {
            isReady = false;
            const errorMessage = validation.error || `Missing required fields: ${validation.missingFields.join(", ")}`;
            addToast({
              id: `validation-${child.id}`,
              title: `${groupType} Validation Error`,
              description: `"${cd.label}" in ${d.componentId} "${d.label}" - ${errorMessage}`,
              variant: "warning",
            });
          }
        }
        continue;
      }
      // For switch nodes (processor or output), validate case contents
      if (n.type === "switchProcessorGroupNode") {
        const isOutputSwitch = d.type === "output";
        if (isOutputSwitch) {
          // Output switch: validate case start nodes (each holds output component config)
          const caseStarts = nodes.filter(
            (cn) => cn.type === "switchCaseStartNode" && (cn.data as any).switchId === n.id
          );
          for (const caseNode of caseStarts) {
            const cd = caseNode.data as StreamFlowNodeData;
            const caseData: FlowNodeData = { label: cd.label, type: cd.type, componentId: cd.componentId, component: cd.component, configYaml: cd.configYaml };
            const validation = validateRequiredFields(caseData);
            if (!validation.isValid) {
              isReady = false;
              const errorMessage = validation.error || `Missing required fields: ${validation.missingFields.join(", ")}`;
              addToast({
                id: `validation-${caseNode.id}`,
                title: "Switch Output Validation Error",
                description: `"${cd.label}" in switch "${d.label}" - ${errorMessage}`,
                variant: "warning",
              });
            }
          }
        } else {
          // Processor switch: validate case chain processors
          const caseProcs = nodes.filter((cn) => {
            const caseStart = nodes.find((cs) => cs.id === (cn.data as any).switchCaseId);
            return caseStart && (caseStart.data as any).switchId === n.id;
          });
          for (const proc of caseProcs) {
            const pd = proc.data as StreamFlowNodeData;
            const procData: FlowNodeData = { label: pd.label, type: pd.type, componentId: pd.componentId, component: pd.component, configYaml: pd.configYaml };
            const validation = validateRequiredFields(procData);
            if (!validation.isValid) {
              isReady = false;
              const errorMessage = validation.error || `Missing required fields: ${validation.missingFields.join(", ")}`;
              addToast({
                id: `validation-${proc.id}`,
                title: "Switch Processor Validation Error",
                description: `"${pd.label}" in switch "${d.label}" - ${errorMessage}`,
                variant: "warning",
              });
            }
          }
        }
        continue;
      }
      const nodeData: FlowNodeData = { label: d.label, type: d.type, componentId: d.componentId, component: d.component, configYaml: d.configYaml };
      const validation = validateRequiredFields(nodeData);
      if (!validation.isValid) {
        isReady = false;
        const errorMessage = validation.error || `Missing required fields: ${validation.missingFields.join(", ")}`;
        addToast({
          id: `validation-${n.id}`,
          title: `${d.type.charAt(0).toUpperCase() + d.type.slice(1)} Validation Error`,
          description: `"${d.label}" - ${errorMessage}`,
          variant: "warning",
        });
      }
    }

    // Build ordered node data: input, processors in edge-order, output
    const orderedNodes: FlowNodeData[] = [];

    if (inputNode) {
      const inputD = inputNode.data as StreamFlowNodeData;
      if (inputD.isGroup && inputD.componentId === "broker") {
        orderedNodes.push(serializeBrokerInputGroup(inputNode));
      } else {
        orderedNodes.push({ label: inputD.label, type: inputD.type, componentId: inputD.componentId, component: inputD.component, configYaml: inputD.configYaml });
      }
    }

    if (inputNode && outputNode) {
      // Build adjacency for main pipeline nodes only
      // For switch nodes: find the convergence target (the node case chain ends connect to)
      const mainPipelineNodes = new Set(
        nodes
          .filter((n) => !n.parentId && n.type !== "switchCaseStartNode" && !(n.data as any).switchCaseId)
          .map((n) => n.id)
      );
      const adj = new Map<string, string>();
      for (const e of edges) {
        if ((e.data as any)?.internal) continue;
        if ((e.data as any)?.switchCase) continue;
        if (mainPipelineNodes.has(e.source) && mainPipelineNodes.has(e.target)) {
          adj.set(e.source, e.target);
        }
        // Convergence: case chain end → main pipeline node → map switch to that target
        if (!mainPipelineNodes.has(e.source) && mainPipelineNodes.has(e.target)) {
          // Find which switch this source belongs to
          const sourceNode = nodes.find((n) => n.id === e.source);
          const switchId = sourceNode?.type === "switchCaseStartNode"
            ? (sourceNode?.data as any)?.switchId
            : (sourceNode?.data as any)?.switchCaseId
              ? (nodes.find((n) => n.id === (sourceNode?.data as any)?.switchCaseId)?.data as any)?.switchId
              : undefined;
          if (switchId && !adj.has(switchId)) {
            adj.set(switchId, e.target);
          }
        }
      }
      let current = adj.get(inputNode.id);
      while (current && current !== outputNode.id) {
        const node = nodes.find((n) => n.id === current);
        if (node) {
          const d = node.data as StreamFlowNodeData;
          if (d.isGroup && d.componentId === "branch") {
            orderedNodes.push(serializeBranchGroup(node));
          } else if (d.isGroup && d.componentId === "catch") {
            orderedNodes.push(serializeCatchGroup(node));
          } else if (node.type === "switchProcessorGroupNode") {
            orderedNodes.push(serializeProcessorSwitchGroup(node));
          } else {
            orderedNodes.push({ label: d.label, type: d.type, componentId: d.componentId, component: d.component, configYaml: d.configYaml });
          }
        }
        current = adj.get(current!);
      }
    }

    if (outputNode) {
      const outputD = outputNode.data as StreamFlowNodeData;
      if (outputD.isGroup && outputD.componentId === "broker") {
        orderedNodes.push(serializeBrokerGroup(outputNode));
      } else if (outputNode.type === "switchProcessorGroupNode" && outputD.componentId === "switch") {
        orderedNodes.push(serializeSwitchGroup(outputNode));
      } else {
        orderedNodes.push({ label: outputD.label, type: outputD.type, componentId: outputD.componentId, component: outputD.component, configYaml: outputD.configYaml });
      }
    }

    onSave({ name, status, bufferId, nodes: orderedNodes, builderState, isReady });
  }, [name, status, bufferId, nodes, edges, onSave, addToast, findDisconnectedNodes, validateRequiredFields, setNodes, serializeBranchGroup, serializeCatchGroup, serializeBrokerGroup, serializeBrokerInputGroup, serializeSwitchGroup, serializeProcessorSwitchGroup]);

  const hasInput = nodes.some((n) => (n.data as StreamFlowNodeData).type === "input");
  const hasOutput = nodes.some((n) => (n.data as StreamFlowNodeData).type === "output");
  const hasProcessors = nodes.some((n) => (n.data as StreamFlowNodeData).type === "processor" && !n.parentId);

  return (
    <div className="flex flex-col h-[calc(100vh-10rem)] w-full">
      {/* Top bar */}
      <div className="flex items-end gap-4 mb-4 flex-wrap">
        <div className="flex-1 min-w-[200px]">
          <Label htmlFor="stream-name">Flow Name</Label>
          <Input id="stream-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Enter stream name" />
        </div>
        <div className="w-40">
          <Label htmlFor="stream-status">Status</Label>
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger><SelectValue placeholder="Select status" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="paused">Paused</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="w-48">
          <Label htmlFor="stream-buffer">Buffer</Label>
          <Select
            value={bufferId ? String(bufferId) : "none"}
            onValueChange={(val) => setBufferId(val === "none" ? undefined : Number(val))}
          >
            <SelectTrigger><SelectValue placeholder="None" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="none">None</SelectItem>
              {availableBuffers.map((buffer) => (
                <SelectItem key={buffer.id} value={buffer.id}>{buffer.label} ({buffer.component})</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {onValidate && (
          <Button variant="outline" onClick={handleValidate} disabled={isValidating || !hasInput || !hasOutput} className="flex items-center gap-1">
            <ShieldCheck className="h-4 w-4" />
            {isValidating ? "Validating..." : "Validate"}
          </Button>
        )}
        {onTry && (
          <Button variant="outline" onClick={() => { setTryResult(null); setTryDialogOpen(true); }} disabled={!hasProcessors} className="flex items-center gap-1">
            <FlaskConical className="h-4 w-4" />
            Try
          </Button>
        )}
        <Button onClick={handleSave} disabled={!name.trim() || !hasInput || !hasOutput} className="flex items-center gap-1">
          <Save className="h-4 w-4" /> Save
        </Button>
      </div>

      {/* Canvas (full width) */}
      <div className="flex-1 w-full h-full min-h-0 rounded-md border bg-background overflow-hidden">
        <ReactFlow
          nodes={nodesWithCallbacks}
          edges={edgesWithCallbacks}
          onNodesChange={onNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={onConnect}
          onNodeDragStart={handleChildDragStart}
          onNodeDrag={handleChildDrag}
          onNodeDragStop={handleChildDragStop}
          onNodeClick={handleNodeClick}
          onNodeDoubleClick={handleNodeDoubleClick}
          onNodeContextMenu={handleNodeContextMenu}
          onPaneClick={handlePaneClick}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          isValidConnection={isValidConnection}
          fitView
          colorMode={colorMode}
          deleteKeyCode={null}
          defaultEdgeOptions={{ type: "pipeline" }}
          proOptions={{ hideAttribution: true }}
        >
          <Background gap={16} size={1} />
          <Controls />
          <Panel position="top-left" className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAddNode("input")}
              disabled={hasInput}
              className="flex items-center gap-1.5 bg-background/80 backdrop-blur-sm"
            >
              <Download className="h-3.5 w-3.5 text-green-500" />
              Input
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAddNode("processor")}
              className="flex items-center gap-1.5 bg-background/80 backdrop-blur-sm"
            >
              <Workflow className="h-3.5 w-3.5 text-blue-500" />
              Processor
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAddNode("output")}
              disabled={hasOutput}
              className="flex items-center gap-1.5 bg-background/80 backdrop-blur-sm"
            >
              <Upload className="h-3.5 w-3.5 text-amber-500" />
              Output
            </Button>
          </Panel>
        </ReactFlow>
      </div>

      {/* Config modal */}
      <Dialog open={!!editingNodeId} onOpenChange={(open: boolean) => { if (!open) setEditingNodeId(null); }}>
        <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>
              {editingNode?.isGroup
                ? editingNode.data.componentId === "case"
                  ? "Switch Case Configuration"
                  : editingNode.data.componentId === "switch"
                    ? editingNode.data.type === "processor" ? "Switch Processor Configuration" : "Switch Output Configuration"
                    : editingNode.data.componentId === "broker"
                      ? editingNode.data.type === "input" ? "Broker Input Configuration" : "Broker Output Configuration"
                      : editingNode.data.componentId === "branch"
                        ? "Branch Configuration"
                        : "Catch Group Configuration"
                : `Configure ${editingNode?.data.type ? editingNode.data.type.charAt(0).toUpperCase() + editingNode.data.type.slice(1) : ""} Node`}
            </DialogTitle>
            <DialogDescription>
              Changes are applied automatically. Close this dialog and use the Save button to persist.
            </DialogDescription>
          </DialogHeader>
          <div className="flex-1 min-h-0 overflow-y-auto">
            {editingNode?.isGroup ? (
              <div className="space-y-4 p-1">
                <div className="space-y-2">
                  <Label htmlFor="group-label">Label</Label>
                  <Input
                    id="group-label"
                    value={editingNode.data.label}
                    onChange={(e) => {
                      const val = e.target.value;
                      if (/^[a-z0-9_-]*$/.test(val)) {
                        handleUpdateNode(editingNode.id, { ...editingNode.data, label: val });
                      }
                    }}
                    placeholder={editingNode.data.componentId === "switch" ? "switch label" : editingNode.data.componentId === "broker" ? "broker label" : editingNode.data.componentId === "case" ? "case label" : editingNode.data.componentId === "branch" ? "branch label" : "catch label"}
                  />
                </div>
                {editingNode.data.componentId === "branch" && (() => {
                  let branchConfig: any = {};
                  const cfgYaml = editingNode.data.configYaml?.trim();
                  if (cfgYaml) { try { branchConfig = yaml.load(cfgYaml) || {}; } catch {} }
                  const updateBranchField = (field: string, value: string) => {
                    const updated = { ...branchConfig, [field]: value };
                    if (!updated[field]) delete updated[field];
                    const newYaml = Object.keys(updated).length > 0
                      ? yaml.dump(updated, { lineWidth: -1, noRefs: true })
                      : "";
                    handleUpdateNode(editingNode.id, { ...editingNode.data, configYaml: newYaml });
                  };
                  return (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="branch-request-map">Request Map</Label>
                        <Textarea
                          id="branch-request-map"
                          value={branchConfig.request_map || ""}
                          onChange={(e) => updateBranchField("request_map", e.target.value)}
                          placeholder="root = this"
                          className="font-mono text-sm min-h-[80px]"
                        />
                        <p className="text-xs text-muted-foreground">
                          A Bloblang mapping that transforms the message before sending it to the branch processors. Leave empty to use the original message.
                        </p>
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="branch-result-map">Result Map</Label>
                        <Textarea
                          id="branch-result-map"
                          value={branchConfig.result_map || ""}
                          onChange={(e) => updateBranchField("result_map", e.target.value)}
                          placeholder="root.result = this.content"
                          className="font-mono text-sm min-h-[80px]"
                        />
                        <p className="text-xs text-muted-foreground">
                          A Bloblang mapping that merges the branch processor results back into the original message. 'this' refers to the branch result, 'root' refers to the original message.
                        </p>
                      </div>
                    </>
                  );
                })()}
                {editingNode.data.componentId === "case" && (() => {
                  const nd = nodes.find((n) => n.id === editingNode.id)?.data as StreamFlowNodeData | undefined;
                  const editNode = nodes.find((n) => n.id === editingNode.id);
                  const switchId = (editNode?.data as any)?.switchId;
                  const siblingDefaultExists = switchId
                    ? nodes.some(
                        (n) => n.id !== editingNode.id && n.type === "switchCaseStartNode" && (n.data as any).switchId === switchId && !(n.data as any).caseCheck
                      )
                    : editNode?.parentId && nodes.some(
                        (n) => n.id !== editingNode.id && n.parentId === editNode.parentId && !(n.data as any).caseCheck
                      );
                  const isCurrentDefault = !(nd as any)?.caseCheck;
                  return (
                    <>
                      <div className="space-y-1">
                        <Label htmlFor="case-check-group" className="text-xs font-medium">Check Condition</Label>
                        <Input
                          id="case-check-group"
                          placeholder='e.g. this.type == "foo"'
                          value={(nd as any)?.caseCheck || ""}
                          onChange={(e) => {
                            const newVal = e.target.value;
                            if (!newVal && siblingDefaultExists) return;
                            setNodes((nds) => nds.map((n) =>
                              n.id === editingNode.id
                                ? { ...n, data: { ...n.data, caseCheck: newVal } }
                                : n
                            ));
                            const fall = (nd as any)?.caseFallthrough === true;
                            const lbl = newVal ? (fall ? `${newVal} ↓` : newVal) : (fall ? "default ↓" : "default");
                            setEdges((eds) => eds.map((e) =>
                              (e.data as any)?.switchCase && e.target === editingNode.id
                                ? { ...e, data: { ...e.data, edgeLabel: lbl } }
                                : e
                            ));
                          }}
                        />
                        {siblingDefaultExists && !isCurrentDefault ? (
                          <p className="text-[10px] text-amber-500">A default case already exists.</p>
                        ) : (
                          <p className="text-[10px] text-muted-foreground">Bloblang condition. Leave empty for the default case.</p>
                        )}
                      </div>
                      <div className="flex items-center justify-between">
                        <div>
                          <Label htmlFor="case-fallthrough" className="text-xs font-medium">Fallthrough</Label>
                          <p className="text-[10px] text-muted-foreground">Also execute subsequent cases.</p>
                        </div>
                        <input
                          id="case-fallthrough"
                          type="checkbox"
                          checked={(nd as any)?.caseFallthrough === true}
                          onChange={(e) => {
                            const checked = e.target.checked;
                            setNodes((nds) => nds.map((n) =>
                              n.id === editingNode.id
                                ? { ...n, data: { ...n.data, caseFallthrough: checked } }
                                : n
                            ));
                            const chk = (nd as any)?.caseCheck || "";
                            const lbl = chk ? (checked ? `${chk} ↓` : chk) : (checked ? "default ↓" : "default");
                            setEdges((eds) => eds.map((e) =>
                              (e.data as any)?.switchCase && e.target === editingNode.id
                                ? { ...e, data: { ...e.data, edgeLabel: lbl } }
                                : e
                            ));
                          }}
                          className="h-4 w-4"
                        />
                      </div>
                    </>
                  );
                })()}
                {editingNode.data.componentId === "broker" && editingNode.data.type === "output" && (() => {
                  const currentPattern = (nodes.find((n) => n.id === editingNode.id)?.data as StreamFlowNodeData)?.brokerPattern || "fan_out";
                  const patternDescriptions: Record<string, string> = {
                    fan_out: "Sends each message to all outputs in parallel. If an output applies back-pressure, it blocks other outputs from receiving new messages until it catches up.",
                    fan_out_fail_fast: "Sends each message to all outputs in parallel. Unlike fan_out, a failing output does not block other outputs and the message is abandoned immediately.",
                    fan_out_sequential: "Sends each message to all outputs one by one in the order they are listed. Each output must acknowledge the message before it is sent to the next.",
                    fan_out_sequential_fail_fast: "Sends each message to all outputs one by one in order. Unlike fan_out_sequential, a failing output causes the message to be abandoned immediately.",
                    round_robin: "Distributes messages across outputs by alternating between them in a round-robin fashion. Each message goes to exactly one output.",
                    greedy: "Sends each message to the first available output that is ready to accept it. Useful when outputs have varying throughput.",
                  };
                  const groupNode = nodes.find((n) => n.id === editingNode.id);
                  const groupData = groupNode?.data as StreamFlowNodeData | undefined;
                  let batchingConfig: any = {};
                  if (groupData?.configYaml?.trim()) {
                    try {
                      const parsed = yaml.load(groupData.configYaml) as any;
                      if (parsed?.batching) batchingConfig = parsed.batching;
                    } catch {}
                  }
                  const updateBatching = (field: string, value: any) => {
                    setNodes((nds) => nds.map((n) => {
                      if (n.id !== editingNode.id) return n;
                      let existing: any = {};
                      if ((n.data as StreamFlowNodeData).configYaml?.trim()) {
                        try { existing = yaml.load((n.data as StreamFlowNodeData).configYaml!) as any || {}; } catch {}
                      }
                      const batching = { ...(existing.batching || {}), [field]: value };
                      // Remove empty/default values
                      if (batching[field] === 0 || batching[field] === "" || batching[field] === undefined) delete batching[field];
                      const newConfig = { ...existing, batching: Object.keys(batching).length > 0 ? batching : undefined };
                      if (!newConfig.batching) delete newConfig.batching;
                      const configYaml = Object.keys(newConfig).length > 0
                        ? yaml.dump(newConfig, { lineWidth: -1, noRefs: true })
                        : "";
                      return { ...n, data: { ...n.data, configYaml } };
                    }));
                  };
                  return (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="broker-pattern">Pattern</Label>
                        <Select
                          value={currentPattern}
                          onValueChange={(val) => {
                            setNodes((nds) => nds.map((n) =>
                              n.id === editingNode.id
                                ? { ...n, data: { ...n.data, brokerPattern: val } }
                                : n.parentId === editingNode.id
                                  ? { ...n, data: { ...n.data, parentBrokerPattern: val } }
                                  : n
                            ));
                            // Add or remove internal edges based on pattern
                            const children = nodes
                              .filter((n) => n.parentId === editingNode.id)
                              .sort((a, b) => a.position.y - b.position.y);
                            const childIds = new Set(children.map((c) => c.id));
                            setEdges((eds) => {
                              const kept = eds.filter(
                                (e) => !((e.data as any)?.internal && (childIds.has(e.source) || childIds.has(e.target)))
                              );
                              if (!isBrokerSequential(val) || children.length <= 1) return kept;
                              const newEdges: Edge[] = [];
                              for (let i = 0; i < children.length - 1; i++) {
                                newEdges.push({
                                  id: `e-internal-${children[i].id}-${children[i + 1].id}`,
                                  source: children[i].id,
                                  target: children[i + 1].id,
                                  type: "pipeline",
                                  data: makeInternalEdgeData(),
                                });
                              }
                              return [...kept, ...newEdges];
                            });
                          }}
                        >
                          <SelectTrigger id="broker-pattern">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="fan_out">fan_out</SelectItem>
                            <SelectItem value="fan_out_fail_fast">fan_out_fail_fast</SelectItem>
                            <SelectItem value="fan_out_sequential">fan_out_sequential</SelectItem>
                            <SelectItem value="fan_out_sequential_fail_fast">fan_out_sequential_fail_fast</SelectItem>
                            <SelectItem value="round_robin">round_robin</SelectItem>
                            <SelectItem value="greedy">greedy</SelectItem>
                          </SelectContent>
                        </Select>
                        <p className="text-xs text-muted-foreground">{patternDescriptions[currentPattern]}</p>
                      </div>
                      <div className="space-y-3 border rounded-md p-3">
                        <p className="text-sm font-medium">Batching</p>
                        <p className="text-xs text-muted-foreground">Configure how messages are batched before being sent to outputs.</p>
                        <div className="grid grid-cols-2 gap-3">
                          <div className="space-y-1">
                            <Label htmlFor="batch-count" className="text-xs">Count</Label>
                            <Input
                              id="batch-count"
                              type="number"
                              min={0}
                              placeholder="0"
                              value={batchingConfig.count || ""}
                              onChange={(e) => updateBatching("count", e.target.value ? parseInt(e.target.value) : 0)}
                            />
                            <p className="text-[10px] text-muted-foreground">Flush after this many messages. 0 to disable.</p>
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="batch-byte-size" className="text-xs">Byte Size</Label>
                            <Input
                              id="batch-byte-size"
                              type="number"
                              min={0}
                              placeholder="0"
                              value={batchingConfig.byte_size || ""}
                              onChange={(e) => updateBatching("byte_size", e.target.value ? parseInt(e.target.value) : 0)}
                            />
                            <p className="text-[10px] text-muted-foreground">Flush after this many bytes. 0 to disable.</p>
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="batch-period" className="text-xs">Period</Label>
                            <Input
                              id="batch-period"
                              placeholder="e.g. 500ms, 1s"
                              value={batchingConfig.period || ""}
                              onChange={(e) => updateBatching("period", e.target.value)}
                            />
                            <p className="text-[10px] text-muted-foreground">Flush after this time regardless of size.</p>
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="batch-jitter" className="text-xs">Jitter</Label>
                            <Input
                              id="batch-jitter"
                              type="number"
                              min={0}
                              step={0.1}
                              placeholder="0"
                              value={batchingConfig.jitter || ""}
                              onChange={(e) => updateBatching("jitter", e.target.value ? parseFloat(e.target.value) : 0)}
                            />
                            <p className="text-[10px] text-muted-foreground">Random delay factor for flush intervals.</p>
                          </div>
                        </div>
                        <div className="space-y-1">
                          <Label htmlFor="batch-check" className="text-xs">Check</Label>
                          <Input
                            id="batch-check"
                            placeholder="Bloblang query"
                            value={batchingConfig.check || ""}
                            onChange={(e) => updateBatching("check", e.target.value)}
                          />
                          <p className="text-[10px] text-muted-foreground">A Bloblang query returning a boolean to decide if a message should end the batch.</p>
                        </div>
                      </div>
                    </>
                  );
                })()}
                {editingNode.data.componentId === "broker" && editingNode.data.type === "input" && (() => {
                  let currentConfig: any = {};
                  const configYaml = (nodes.find((n) => n.id === editingNode.id)?.data as StreamFlowNodeData)?.configYaml;
                  if (configYaml?.trim()) { try { currentConfig = yaml.load(configYaml) || {}; } catch {} }

                  const updateInputBrokerConfig = (field: string, value: any) => {
                    setNodes((nds) => nds.map((n) => {
                      if (n.id !== editingNode.id) return n;
                      const d = n.data as StreamFlowNodeData;
                      let config: any = {};
                      if (d.configYaml?.trim()) { try { config = yaml.load(d.configYaml) || {}; } catch {} }
                      if (field.startsWith("batching.")) {
                        const batchField = field.split(".")[1];
                        if (!config.batching) config.batching = {};
                        if (value === "" || value === 0) { delete config.batching[batchField]; }
                        else { config.batching[batchField] = value; }
                        if (Object.keys(config.batching).length === 0) delete config.batching;
                      } else {
                        if (value === "" || value === 0 || value === 1) { delete config[field]; }
                        else { config[field] = value; }
                      }
                      const newYaml = Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : "";
                      return { ...n, data: { ...n.data, configYaml: newYaml } };
                    }));
                  };

                  const batchingConfig = currentConfig.batching || {};

                  return (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="input-broker-copies" className="text-xs">Copies</Label>
                        <Input
                          id="input-broker-copies"
                          type="number"
                          min={1}
                          placeholder="1"
                          value={currentConfig.copies || ""}
                          onChange={(e) => updateInputBrokerConfig("copies", e.target.value ? parseInt(e.target.value) : 1)}
                        />
                        <p className="text-[10px] text-muted-foreground">Number of copies of each configured input to spawn.</p>
                      </div>
                      <div className="space-y-2 border rounded-md p-3">
                        <span className="text-xs font-medium">Batching</span>
                        <div className="grid grid-cols-2 gap-3">
                          <div className="space-y-1">
                            <Label htmlFor="ib-batch-count" className="text-xs">Count</Label>
                            <Input id="ib-batch-count" type="number" min={0} placeholder="0" value={batchingConfig.count || ""} onChange={(e) => updateInputBrokerConfig("batching.count", e.target.value ? parseInt(e.target.value) : 0)} />
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="ib-batch-byte-size" className="text-xs">Byte Size</Label>
                            <Input id="ib-batch-byte-size" type="number" min={0} placeholder="0" value={batchingConfig.byte_size || ""} onChange={(e) => updateInputBrokerConfig("batching.byte_size", e.target.value ? parseInt(e.target.value) : 0)} />
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="ib-batch-period" className="text-xs">Period</Label>
                            <Input id="ib-batch-period" placeholder="e.g. 1s" value={batchingConfig.period || ""} onChange={(e) => updateInputBrokerConfig("batching.period", e.target.value)} />
                          </div>
                          <div className="space-y-1">
                            <Label htmlFor="ib-batch-jitter" className="text-xs">Jitter</Label>
                            <Input id="ib-batch-jitter" type="number" min={0} step={0.1} placeholder="0" value={batchingConfig.jitter || ""} onChange={(e) => updateInputBrokerConfig("batching.jitter", e.target.value ? parseFloat(e.target.value) : 0)} />
                          </div>
                        </div>
                        <div className="space-y-1">
                          <Label htmlFor="ib-batch-check" className="text-xs">Check</Label>
                          <Input id="ib-batch-check" placeholder="Bloblang query" value={batchingConfig.check || ""} onChange={(e) => updateInputBrokerConfig("batching.check", e.target.value)} />
                        </div>
                      </div>
                    </>
                  );
                })()}
                {editingNode.data.componentId === "switch" && editingNode.data.type === "output" && (() => {
                  let currentConfig: any = {};
                  const configYaml = (nodes.find((n) => n.id === editingNode.id)?.data as StreamFlowNodeData)?.configYaml;
                  if (configYaml?.trim()) { try { currentConfig = yaml.load(configYaml) || {}; } catch {} }

                  const updateSwitchConfig = (field: string, value: any) => {
                    setNodes((nds) => nds.map((n) => {
                      if (n.id !== editingNode.id) return n;
                      const d = n.data as StreamFlowNodeData;
                      let config: any = {};
                      if (d.configYaml?.trim()) { try { config = yaml.load(d.configYaml) || {}; } catch {} }
                      if (value === false || value === "" || value === undefined) { delete config[field]; }
                      else { config[field] = value; }
                      const newYaml = Object.keys(config).length > 0 ? yaml.dump(config, { lineWidth: -1, noRefs: true }) : "";
                      return { ...n, data: { ...n.data, configYaml: newYaml } };
                    }));
                  };

                  return (
                    <>
                      <div className="flex items-center justify-between">
                        <Label htmlFor="switch-retry" className="text-xs">Retry Until Success</Label>
                        <input
                          id="switch-retry"
                          type="checkbox"
                          checked={currentConfig.retry_until_success === true}
                          onChange={(e) => updateSwitchConfig("retry_until_success", e.target.checked)}
                          className="h-4 w-4"
                        />
                      </div>
                      <p className="text-[10px] text-muted-foreground -mt-1">If a selected output fails, reattempt indefinitely.</p>
                      <div className="flex items-center justify-between">
                        <Label htmlFor="switch-strict" className="text-xs">Strict Mode</Label>
                        <input
                          id="switch-strict"
                          type="checkbox"
                          checked={currentConfig.strict_mode === true}
                          onChange={(e) => updateSwitchConfig("strict_mode", e.target.checked)}
                          className="h-4 w-4"
                        />
                      </div>
                      <p className="text-[10px] text-muted-foreground -mt-1">Report an error if no condition is met.</p>
                    </>
                  );
                })()}
                <p className="text-sm text-muted-foreground">
                  {editingNode.data.componentId === "case"
                    ? <>Use the <strong>+</strong> button inside the case on the canvas to add processors. Click on a processor to configure it.</>
                    : editingNode.data.componentId === "switch"
                      ? <>Use the <strong>+ Add</strong> button inside the group on the canvas to add cases. Click on a case to configure its condition and {editingNode.data.type === "processor" ? "processors" : "output"}.</>
                      : editingNode.data.componentId === "broker" && editingNode.data.type === "input"
                        ? <>Use the <strong>+ Add</strong> button inside the group on the canvas to add child inputs. Click on a child input to configure it.</>
                        : editingNode.data.componentId === "broker"
                          ? <>Use the <strong>+ Add</strong> button inside the group on the canvas to add child outputs. Click on a child output to configure it.</>
                          : <>Use the <strong>+ Add</strong> button inside the group on the canvas to add child processors. Click on a child processor to configure it.</>
                  }
                </p>
                <Button
                  variant="destructive"
                  className="w-full"
                  onClick={() => { handleDeleteNode(editingNode.id); setEditingNodeId(null); }}
                >
                  {editingNode.data.componentId === "case"
                    ? "Delete Case"
                    : editingNode.data.componentId === "switch"
                      ? (editingNode.data.type === "processor" ? "Delete Switch Processor Group" : "Delete Switch Output Group")
                      : editingNode.data.componentId === "broker"
                        ? (editingNode.data.type === "input" ? "Delete Broker Input Group" : "Delete Broker Output Group")
                        : "Delete Catch Group"}
                </Button>
              </div>
            ) : (
              <>
                {/* Output switch case: check condition + continue toggle */}
                {(() => {
                  if (!editingNode?.isOutputCaseStart) return null;
                  const editNode = nodes.find((n) => n.id === editingNode.id);
                  if (!editNode) return null;
                  const nd = editNode.data as StreamFlowNodeData;
                  const switchId = (nd as any).switchId;
                  const siblingDefaultExists = nodes.some(
                    (n) => n.id !== editingNode.id && n.type === "switchCaseStartNode" && (n.data as any).switchId === switchId && !(n.data as any).caseCheck
                  );
                  const isCurrentDefault = !(nd as any).caseCheck;
                  return (
                    <div className="space-y-3 mb-4 p-3 rounded-md border bg-muted/30">
                      <div className="space-y-1">
                        <Label htmlFor="case-check" className="text-xs font-medium">Check Condition</Label>
                        <Input
                          id="case-check"
                          placeholder='e.g. this.type == "foo"'
                          value={(nd as any).caseCheck || ""}
                          onChange={(e) => {
                            const newVal = e.target.value;
                            if (!newVal && siblingDefaultExists) return;
                            setNodes((nds) => nds.map((n) =>
                              n.id === editingNode.id
                                ? { ...n, data: { ...n.data, caseCheck: newVal } }
                                : n
                            ));
                            const cont = (nd as any).caseContinue === true;
                            const lbl = newVal ? (cont ? `${newVal} →` : newVal) : (cont ? "default →" : "default");
                            setEdges((eds) => eds.map((e) =>
                              (e.data as any)?.switchCase && e.target === editingNode.id
                                ? { ...e, data: { ...e.data, edgeLabel: lbl, edgeLabelColor: "amber" } }
                                : e
                            ));
                          }}
                        />
                        {siblingDefaultExists && !isCurrentDefault ? (
                          <p className="text-[10px] text-amber-500">A default case already exists.</p>
                        ) : (
                          <p className="text-[10px] text-muted-foreground">Bloblang condition. Leave empty for the default case.</p>
                        )}
                      </div>
                      <div className="flex items-center justify-between">
                        <div>
                          <Label htmlFor="case-continue" className="text-xs font-medium">Continue</Label>
                          <p className="text-[10px] text-muted-foreground">Also evaluate subsequent cases.</p>
                        </div>
                        <input
                          id="case-continue"
                          type="checkbox"
                          checked={(nd as any).caseContinue === true}
                          onChange={(e) => {
                            const checked = e.target.checked;
                            setNodes((nds) => nds.map((n) =>
                              n.id === editingNode.id
                                ? { ...n, data: { ...n.data, caseContinue: checked } }
                                : n
                            ));
                            const chk = (nd as any).caseCheck || "";
                            const lbl = chk ? (checked ? `${chk} →` : chk) : (checked ? "default →" : "default");
                            setEdges((eds) => eds.map((e) =>
                              (e.data as any)?.switchCase && e.target === editingNode.id
                                ? { ...e, data: { ...e.data, edgeLabel: lbl, edgeLabelColor: "amber" } }
                                : e
                            ));
                          }}
                          className="h-4 w-4"
                        />
                      </div>
                    </div>
                  );
                })()}
                <NodeConfigPanel
                  key={editingNodeId || "none"}
                  allComponentSchemas={editingNode?.isGroupChild || editingNode?.isSwitchCaseProc || editingNode?.isOutputCaseStart ? childFilteredSchemas : allComponentSchemas}
                  selectedNode={editingNode}
                  onUpdateNode={handleUpdateNode}
                  onDeleteNode={(id) => { handleDeleteNode(id); setEditingNodeId(null); }}
                  lockedComponentId={
                    editingNode?.data.type === "output" && isMcpServer ? "sync_response" : undefined
                  }
                />
              </>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={!!deleteConfirmNodeId} onOpenChange={(open: boolean) => { if (!open) setDeleteConfirmNodeId(null); }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Node</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this node? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" onClick={() => setDeleteConfirmNodeId(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                if (deleteConfirmNodeId) {
                  handleDeleteNode(deleteConfirmNodeId);
                  setEditingNodeId(null);
                }
                setDeleteConfirmNodeId(null);
              }}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Add node confirmation dialog */}
      <Dialog open={!!pendingNode} onOpenChange={(open: boolean) => { if (!open) setPendingNode(null); }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>
              {pendingNode?.kind === "childCase" ? "Add Switch Case"
                : pendingNode?.kind === "childProcessor" ? "Add Catch Processor"
                : pendingNode?.kind === "childOutput" ? "Add Broker Output"
                : pendingNode?.kind === "childInput" ? "Add Broker Input"
                : pendingNode?.kind === "topLevel" ? `Add ${(pendingNode.topLevelType || "").charAt(0).toUpperCase() + (pendingNode.topLevelType || "").slice(1)}`
                : "Add Processor"}
            </DialogTitle>
            <DialogDescription>
              Configure the new node before adding it to the flow.
            </DialogDescription>
          </DialogHeader>
          {pendingNode && (() => {
            const isCase = pendingNode.kind === "childCase";
            const hasComponents = pendingNode.availableComponents.length > 0;
            const isDefault = isCase && !pendingNode.caseCheck;
            const defaultConflict = isDefault && !!pendingNode.defaultExists;
            const needsComponent = hasComponents && !pendingNode.componentId;
            return (
              <div className="space-y-4 mt-2">
                {/* Label */}
                {!isCase && (
                  <div className="space-y-1">
                    <Label htmlFor="pending-label" className="text-xs font-medium">Label</Label>
                    <Input
                      id="pending-label"
                      value={pendingNode.label}
                      onChange={(e) => {
                        const val = e.target.value;
                        if (/^[a-z0-9_-]*$/.test(val)) {
                          setPendingNode({ ...pendingNode, label: val });
                        }
                      }}
                      placeholder="Node label (e.g., my_kafka_input)"
                    />
                  </div>
                )}

                {/* Component selector */}
                {hasComponents && (
                  <div className="space-y-1">
                    <Label htmlFor="pending-component" className="text-xs font-medium">Component</Label>
                    <Select
                      value={pendingNode.componentId || ""}
                      onValueChange={(val) => {
                        const comp = pendingNode.availableComponents.find((c) => c.id === val);
                        setPendingNode({
                          ...pendingNode,
                          componentId: val,
                          component: comp ? (comp.name === comp.component ? comp.component : `${comp.name} (${comp.component})`) : "",
                        });
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select component" />
                      </SelectTrigger>
                      <SelectContent>
                        {pendingNode.availableComponents.map((comp) => (
                          <SelectItem key={comp.id} value={comp.id}>
                            {comp.name === comp.component ? comp.component : `${comp.name} (${comp.component})`}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )}

                {/* Case-specific: check condition */}
                {isCase && (
                  <>
                    <div className="space-y-1">
                      <Label htmlFor="pending-case-check" className="text-xs font-medium">Check Condition</Label>
                      <Input
                        id="pending-case-check"
                        placeholder='e.g. this.type == "foo"'
                        value={pendingNode.caseCheck || ""}
                        onChange={(e) => setPendingNode({ ...pendingNode, caseCheck: e.target.value })}
                      />
                      {defaultConflict ? (
                        <p className="text-[10px] text-destructive">A default case already exists. Provide a condition or remove the existing default first.</p>
                      ) : (
                        <p className="text-[10px] text-muted-foreground">Bloblang condition. Leave empty for the default case.</p>
                      )}
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <Label htmlFor="pending-case-flag" className="text-xs font-medium">
                          {pendingNode.isProcessorSwitch ? "Fallthrough" : "Continue"}
                        </Label>
                        <p className="text-[10px] text-muted-foreground">
                          {pendingNode.isProcessorSwitch ? "Also execute subsequent cases." : "Also evaluate subsequent cases."}
                        </p>
                      </div>
                      <input
                        id="pending-case-flag"
                        type="checkbox"
                        checked={pendingNode.isProcessorSwitch ? pendingNode.caseFallthrough : pendingNode.caseContinue}
                        onChange={(e) =>
                          setPendingNode({
                            ...pendingNode,
                            ...(pendingNode.isProcessorSwitch
                              ? { caseFallthrough: e.target.checked }
                              : { caseContinue: e.target.checked }),
                          })
                        }
                        className="h-4 w-4 rounded border-gray-300"
                      />
                    </div>
                  </>
                )}

                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={() => setPendingNode(null)}>
                    Cancel
                  </Button>
                  <Button
                    disabled={defaultConflict || needsComponent}
                    onClick={handleConfirmAddNode}
                  >
                    Add to Flow
                  </Button>
                </div>
              </div>
            );
          })()}
        </DialogContent>
      </Dialog>

      {/* Context menu */}
      {contextMenu && (
        <div
          data-context-menu
          className="fixed z-50 min-w-[140px] rounded-md border bg-popover p-1 shadow-md"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
            onClick={() => {
              setEditingNodeId(contextMenu.nodeId);
              setContextMenu(null);
            }}
          >
            Edit <span className="ml-auto text-xs text-muted-foreground">E</span>
          </button>
          <button
            type="button"
            className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-destructive hover:bg-destructive hover:text-destructive-foreground"
            onClick={() => {
              setDeleteConfirmNodeId(contextMenu.nodeId);
              setContextMenu(null);
            }}
          >
            Delete <span className="ml-auto text-xs text-muted-foreground">Del</span>
          </button>
        </div>
      )}

      {/* Try Flow Dialog */}
      <Dialog open={tryDialogOpen} onOpenChange={(open) => { if (!open) setTryDialogOpen(false); }}>
        <DialogContent className="max-w-3xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>Try Flow Processors</DialogTitle>
          </DialogHeader>
          <div className="flex-1 min-h-0 overflow-y-auto space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-sm font-medium">Input Messages</Label>
                <Button variant="outline" size="sm" onClick={() => setTryMessages([...tryMessages, ""])} className="flex items-center gap-1 h-7 text-xs">
                  <Plus className="h-3 w-3" /> Add Message
                </Button>
              </div>
              <div className="space-y-2">
                {tryMessages.map((msg, idx) => (
                  <div key={idx} className="flex gap-2">
                    <Textarea
                      value={msg}
                      onChange={(e) => {
                        const updated = [...tryMessages];
                        updated[idx] = e.target.value;
                        setTryMessages(updated);
                      }}
                      placeholder='{"key": "value"}'
                      className="font-mono text-sm min-h-[80px]"
                    />
                    {tryMessages.length > 1 && (
                      <Button variant="ghost" size="icon" className="shrink-0 h-8 w-8 mt-1" onClick={() => setTryMessages(tryMessages.filter((_, i) => i !== idx))}>
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <Button onClick={handleTrySubmit} disabled={isTrying || !tryMessages.some((m) => m.trim())} className="w-full flex items-center justify-center gap-2">
              {isTrying ? (
                <><Loader2 className="h-4 w-4 animate-spin" /> Running...</>
              ) : (
                <><FlaskConical className="h-4 w-4" /> Run Test</>
              )}
            </Button>

            {tryResult && (
              <div className="space-y-2">
                <Label className="text-sm font-medium">Result</Label>
                {tryResult.error ? (
                  <div className="flex items-start gap-2 p-3 rounded-md border bg-red-50 border-red-200 text-red-800 dark:bg-red-950 dark:border-red-800 dark:text-red-200 text-sm">
                    <XCircle className="h-4 w-4 mt-0.5 shrink-0 text-red-600 dark:text-red-400" />
                    <pre className="whitespace-pre-wrap break-all font-mono text-xs flex-1">{tryResult.error}</pre>
                  </div>
                ) : !tryResult.outputs || tryResult.outputs.length === 0 ? (
                  <div className="p-3 rounded-md border bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-950 dark:border-yellow-800 dark:text-yellow-200 text-sm">
                    No output produced. The message may have been filtered out by the processors.
                  </div>
                ) : (
                  <div className="space-y-2">
                    {tryResult.outputs.map((output, idx) => (
                      <div key={idx} className="rounded-md border bg-green-50 border-green-200 dark:bg-green-950 dark:border-green-800">
                        <div className="px-3 py-1.5 border-b border-green-200 dark:border-green-800">
                          <span className="text-xs font-medium text-green-700 dark:text-green-300">Output {idx + 1}</span>
                        </div>
                        <pre className="p-3 text-sm font-mono whitespace-pre-wrap break-all text-green-900 dark:text-green-100">{output.content}</pre>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export function FlowBuilder(props: FlowBuilderProps) {
  return (
    <ReactFlowProvider>
      <FlowBuilderContent {...props} />
    </ReactFlowProvider>
  );
}
