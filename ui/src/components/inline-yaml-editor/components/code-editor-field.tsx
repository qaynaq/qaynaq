import { useState } from "react";
import Editor, { type Monaco } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { registerBloblangLanguage } from "@/lib/bloblang-language";

interface CodeEditorFieldProps {
  value: string;
  onChange: (value: string) => void;
  previewMode?: boolean;
  language?: string;
}

export function CodeEditorField({
  value,
  onChange,
  previewMode = false,
  language = "bloblang",
}: CodeEditorFieldProps) {
  const { resolvedTheme } = useTheme();
  const [showTooltip, setShowTooltip] = useState(false);

  const beforeMount = (monaco: Monaco) => {
    if (language === "bloblang") {
      registerBloblangLanguage(monaco);
    }
  };

  const displayValue = value
    ? value.length > 60
      ? `${value.substring(0, 60)}...`
      : value
    : "No code configured";

  if (previewMode) {
    return (
      <div className="relative">
        <div
          className={`text-sm px-3 py-2 text-left border rounded-md cursor-default bg-muted/30 font-mono
            ${value ? "text-foreground" : "text-muted-foreground"}`}
          onMouseEnter={() => value && setShowTooltip(true)}
          onMouseLeave={() => setShowTooltip(false)}
        >
          {displayValue}
        </div>
        {showTooltip && value && (
          <div
            className="absolute z-50 bottom-full left-0 mb-2 p-3 bg-popover text-popover-foreground text-xs rounded-md shadow-lg max-w-md max-h-48 overflow-auto border font-mono"
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

  return (
    <div className="rounded-md border overflow-hidden">
      <Editor
        value={value}
        language={language}
        theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
        onChange={(val) => onChange(val ?? "")}
        height="150px"
        beforeMount={beforeMount}
        options={{
          minimap: { enabled: false },
          fontSize: 13,
          lineNumbers: "on",
          scrollBeyondLastLine: false,
          wordWrap: "on",
          tabSize: 2,
          automaticLayout: true,
          padding: { top: 8 },
          overviewRulerLanes: 0,
          scrollbar: {
            vertical: "auto",
            horizontal: "auto",
          },
          renderLineHighlight: "none",
          folding: false,
        }}
      />
    </div>
  );
}
