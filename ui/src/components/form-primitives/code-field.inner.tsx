import Editor, { type Monaco } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { registerBloblangLanguage } from "@/lib/bloblang-language";

interface Props {
  value: string;
  onChange: (next: string) => void;
  language: string;
  height: string;
}

export default function CodeFieldInner({
  value,
  onChange,
  language,
  height,
}: Props) {
  const { resolvedTheme } = useTheme();

  const beforeMount = (monaco: Monaco) => {
    if (language === "bloblang") {
      registerBloblangLanguage(monaco);
    }
  };

  return (
    <div className="rounded-md border overflow-hidden">
      <Editor
        value={value}
        language={language}
        theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
        onChange={(v) => onChange(v ?? "")}
        height={height}
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
          scrollbar: { vertical: "auto", horizontal: "auto" },
          renderLineHighlight: "none",
          folding: false,
        }}
      />
    </div>
  );
}
