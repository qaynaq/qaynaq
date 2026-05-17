import type { ComponentCategory, FlowComponent } from "../types";
import { getComponent } from "../registry";

export interface ListItem {
  componentId: string;
  config: unknown;
}

export function rawToListItem(
  category: ComponentCategory,
  raw: unknown,
): ListItem | null {
  if (typeof raw !== "object" || raw === null) return null;
  const entries = Object.entries(raw as Record<string, unknown>);
  if (entries.length === 0) return null;
  const [componentId, inner] = entries[0];
  const component = getComponent(category, componentId) as
    | FlowComponent<unknown>
    | undefined;
  if (!component) {
    return { componentId, config: inner };
  }
  const config = component.fromListItem
    ? component.fromListItem(inner)
    : (inner ?? structuredClone(component.defaultConfig));
  return { componentId, config };
}

export function listItemToRaw(
  category: ComponentCategory,
  item: ListItem,
): Record<string, unknown> {
  const component = getComponent(category, item.componentId) as
    | FlowComponent<unknown>
    | undefined;
  if (!component) {
    return { [item.componentId]: item.config };
  }
  const inner = component.toListItem
    ? component.toListItem(item.config)
    : item.config;
  return { [item.componentId]: inner };
}

export function newListItem(
  category: ComponentCategory,
  componentId: string,
): ListItem {
  const component = getComponent(category, componentId);
  if (!component) {
    return { componentId, config: {} };
  }
  return {
    componentId,
    config: structuredClone(component.defaultConfig),
  };
}
