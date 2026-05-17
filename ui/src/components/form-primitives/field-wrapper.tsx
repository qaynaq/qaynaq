import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

export interface FieldWrapperProps {
  label?: string;
  description?: string;
  error?: string;
  required?: boolean;
  size?: "sm" | "default";
  className?: string;
  children: ReactNode;
}

export function FieldWrapper({
  label,
  description,
  error,
  required,
  size = "default",
  className,
  children,
}: FieldWrapperProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      {label && (
        <Label
          className={cn(
            size === "sm" && "text-xs font-normal text-muted-foreground",
          )}
        >
          {label}
          {required && <span className="text-destructive ml-1">*</span>}
        </Label>
      )}
      {description && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
      {children}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
