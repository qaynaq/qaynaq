export type { McpToolParameter, McpToolTemplate, SharedConfigField, TemplatePack } from "./types";
export { param } from "./types";

import { googleCalendarPack } from "./google-calendar";
import { googleDrivePack } from "./google-drive";
import { googleSheetsPack } from "./google-sheets";
import { shopifyPack } from "./shopify";
import type { TemplatePack } from "./types";

export const templatePacks: TemplatePack[] = [
  googleCalendarPack,
  googleDrivePack,
  googleSheetsPack,
  shopifyPack,
];
