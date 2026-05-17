import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";

interface Props {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  error?: string;
}

export function CheckboxField({
  label,
  description,
  checked,
  onChange,
  disabled,
  error,
}: Props) {
  return (
    <div className="space-y-1">
      <label
        className={cn(
          "flex items-start gap-2 text-sm",
          disabled ? "cursor-default" : "cursor-pointer",
        )}
      >
        <Checkbox
          checked={checked}
          onCheckedChange={(c) => onChange(!!c)}
          disabled={disabled}
          className="mt-0.5 h-3.5 w-3.5"
        />
        <span>
          <span
            className={cn(
              "font-medium",
              disabled && "text-muted-foreground",
            )}
          >
            {label}
          </span>
          {description && (
            <span className="block text-xs text-muted-foreground">
              {description}
            </span>
          )}
        </span>
      </label>
      {error && <p className="text-xs text-destructive ml-6">{error}</p>}
    </div>
  );
}
