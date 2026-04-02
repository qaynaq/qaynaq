import type { Monaco } from "@monaco-editor/react";

let registered = false;

export function registerBloblangLanguage(monaco: Monaco) {
  if (registered) return;
  registered = true;

  monaco.languages.register({ id: "bloblang" });

  monaco.languages.setMonarchTokensProvider("bloblang", {
    keywords: ["if", "else", "match", "map", "import", "let", "meta"],
    constants: ["true", "false", "null", "deleted"],
    references: ["root", "this"],

    operators: [
      "==", "!=", ">=", "<=", "&&", "||",
      "=>", "->", "+", "-", "*", "/", "%",
      ">", "<", "!", "=", "|",
    ],
    symbols: /[=><!~?:&|+\-*/^%]+/,
    escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{2}|u[0-9A-Fa-f]{4})/,

    tokenizer: {
      root: [
        // comments
        [/#.*$/, "comment"],

        // import statements
        [/\b(import)\s+(".*?")/, ["keyword", "string"]],

        // variable references $var
        [/\$\w+/, "variable"],

        // metadata references @field
        [/@(?:\w+)?/, "variable"],

        // map declaration: map name {
        [/\b(map)\s+(\w+)/, ["keyword", "type"]],

        // let/meta declaration: let name =
        [/\b(let|meta)\s+(\w+)/, ["keyword", "variable"]],

        // function/method calls: name(
        [/\b(\w+)(\()/, [
          { cases: { "@keywords": "keyword", "@default": "entity.name.function" } },
          "@brackets",
        ]],

        // method calls after dot: .method(
        [/\.(\w+)(?=\()/, "entity.name.function"],

        // named parameters: name:
        [/\b(\w+)(:)/, ["variable.parameter", "delimiter"]],

        // arrow functions: x ->
        [/(\w+)\s+(->)/, ["variable.parameter", "operator"]],

        // identifiers
        [/[a-zA-Z_]\w*/, {
          cases: {
            "@keywords": "keyword",
            "@constants": "constant.language",
            "@references": "variable.language",
            "@default": "identifier",
          },
        }],

        // string accessor: ."field name"
        [/\.("(?:[^"\\]|\\.)*")/, "string"],

        // strings
        [/"/, "string", "@string_double"],
        [/'/, "string", "@string_single"],

        // numbers
        [/\b\d+\.\d+\b/, "number.float"],
        [/\b\d+\b/, "number"],

        // brackets
        [/[{}()\[\]]/, "@brackets"],

        // match arrow
        [/=>/, "keyword.operator"],

        // operators
        [/->/, "operator"],
        [/@symbols/, {
          cases: {
            "@operators": "operator",
            "@default": "",
          },
        }],

        // delimiters
        [/[,.]/, "delimiter"],

        // whitespace
        [/\s+/, "white"],
      ],

      string_double: [
        [/[^\\"]+/, "string"],
        [/@escapes/, "string.escape"],
        [/\\./, "string.escape.invalid"],
        [/"/, "string", "@pop"],
      ],

      string_single: [
        [/[^\\']+/, "string"],
        [/@escapes/, "string.escape"],
        [/\\./, "string.escape.invalid"],
        [/'/, "string", "@pop"],
      ],
    },
  });

  monaco.languages.setLanguageConfiguration("bloblang", {
    comments: {
      lineComment: "#",
    },
    brackets: [
      ["{", "}"],
      ["[", "]"],
      ["(", ")"],
    ],
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
      { open: "'", close: "'" },
    ],
    surroundingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
      { open: "'", close: "'" },
    ],
  });
}
