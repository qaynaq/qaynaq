import { Suspense, useMemo } from "react";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
  listScanners,
  getScanner,
} from "@/components/flow-components/scanners/registry";
import {
  describeScanner,
  groupScanners,
} from "@/lib/scanner-catalog";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

export type ScannerValue = Record<string, unknown>;

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: ScannerValue;
  onChange: (next: ScannerValue) => void;
}

function pickScannerId(value: ScannerValue): string {
  const keys = Object.keys(value ?? {});
  return keys[0] ?? "lines";
}

export function ScannerField({ value, onChange, size, ...wrapper }: Props) {
  const groups = useMemo(() => groupScanners(listScanners()), []);
  const activeId = pickScannerId(value);
  const active = getScanner(activeId);
  const description = active?.description ?? describeScanner(activeId);

  const handleSelect = (id: string) => {
    if (id === activeId) return;
    const next = getScanner(id);
    onChange({ [id]: (next?.defaultConfig as object) ?? {} });
  };

  const handleInnerChange = (next: unknown) => {
    onChange({ [activeId]: next });
  };

  const Editor = active?.Editor;
  const innerValue = (value?.[activeId] as object) ?? active?.defaultConfig ?? {};

  return (
    <FieldWrapper size={size} {...wrapper}>
      <div className="space-y-3">
        <Select value={activeId} onValueChange={handleSelect}>
          <SelectTrigger className={cn(size === "sm" && "h-8 text-sm")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {groups.map(({ group, items }, idx) => (
              <SelectGroup
                key={group}
                className={cn(idx > 0 && "mt-1 border-t border-border pt-1")}
              >
                <SelectLabel className="pl-2 pr-2 text-[10px] font-medium uppercase tracking-wider text-muted-foreground/80">
                  {group}
                </SelectLabel>
                {items.map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            ))}
          </SelectContent>
        </Select>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
        {Editor && (
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <Suspense fallback={null}>
              <Editor value={innerValue} onChange={handleInnerChange} />
            </Suspense>
          </div>
        )}
      </div>
    </FieldWrapper>
  );
}
