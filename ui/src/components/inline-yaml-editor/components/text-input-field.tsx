import React, { useState } from "react";
import { Input } from "@/components/ui/input";

interface TextInputFieldProps {
  value: string;
  onChange: (value: string) => void;
  previewMode?: boolean;
  placeholder?: string;
  small?: boolean;
  pattern?: string;
  patternMessage?: string;
}

export function TextInputField({
  value,
  onChange,
  previewMode = false,
  placeholder = "",
  small = false,
  pattern,
  patternMessage,
}: TextInputFieldProps) {
  const [showTooltip, setShowTooltip] = useState(false);

  const displayValue = value
    ? value.length > 50
      ? `${value.substring(0, 50)}...`
      : value
    : previewMode
      ? "No value configured"
      : placeholder;

  if (previewMode) {
    return (
      <div className="relative">
        <div
          className={`text-sm px-3 py-2 text-left border rounded-md cursor-default bg-muted/30
            ${value ? "text-foreground" : "text-muted-foreground"}`}
          onMouseEnter={() =>
            value && value.length > 50 && setShowTooltip(true)
          }
          onMouseLeave={() => setShowTooltip(false)}
        >
          {displayValue}
        </div>
        {showTooltip && value && value.length > 50 && (
          <div
            className="absolute z-50 bottom-full left-0 mb-2 p-3 bg-popover text-popover-foreground text-xs rounded-md shadow-lg max-w-md max-h-48 overflow-auto border"
            style={{
              whiteSpace: "pre-wrap",
              wordBreak: "break-word",
            }}
          >
            {value}
          </div>
        )}
      </div>
    );
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    if (pattern && val !== "") {
      const regex = new RegExp(`^${pattern}$`);
      if (!regex.test(val)) return;
    }
    onChange(val);
  };

  return (
    <Input
      value={value}
      onChange={handleChange}
      className={small ? "h-8 text-sm" : "text-sm"}
      placeholder={placeholder}
      title={pattern ? (patternMessage || `Must match: ${pattern}`) : undefined}
    />
  );
}
