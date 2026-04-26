import React, { useState, useEffect, useMemo } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import * as yaml from "js-yaml";
import {
  fetchCaches,
  fetchConnections,
  fetchSecrets,
  fetchRateLimits,
  fetchFiles,
} from "@/lib/api";

import { InlineYamlEditorProps, FieldState, FieldSchema } from "./types";
import { getDefaultValue } from "./utils/defaults";
import { useInternalState } from "./hooks/use-internal-state";
import {
  convertOutputCasesToYaml,
  convertOutputListToYaml,
  convertInputListToYaml,
  convertProcessorCasesToYaml,
  convertYamlToOutputCases,
  convertYamlToOutputList,
  convertYamlToInputList,
  convertYamlToProcessorCases,
} from "./utils/yaml-converter";
import {
  TextInputField,
  KeyValueEditor,
  OutputListEditor,
  InputListEditor,
} from "./components";
import {
  LazyOutputCasesEditor,
  LazyProcessorCasesEditor,
  LazyProcessorListEditor,
  LazyCodeEditorField,
  LazyObjectEditor,
  LazyArrayEditor,
  LazyPropertyListEditor,
} from "./components/lazy-components";
import { Suspense } from "react";

function DynamicSelectField({
  fieldSchema,
  value,
  onChange,
}: {
  fieldSchema: FieldSchema;
  value: string;
  onChange: (value: string) => void;
}) {
  const [options, setOptions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const isSecret = fieldSchema.dataSource === "secrets";
  const isConnection = fieldSchema.dataSource === "connections";

  useEffect(() => {
    const fetchOptions = async () => {
      setLoading(true);
      try {
        if (fieldSchema.dataSource === "caches") {
          const caches = await fetchCaches();
          setOptions(caches.map((cache) => cache.label));
        } else if (isSecret) {
          const secrets = await fetchSecrets();
          setOptions(secrets.map((secret) => secret.key));
        } else if (isConnection) {
          const connections = await fetchConnections();
          setOptions(connections.map((conn) => conn.name));
        } else if (fieldSchema.dataSource === "rate_limits") {
          const rateLimits = await fetchRateLimits();
          setOptions(rateLimits.map((rateLimit) => rateLimit.label));
        }
      } catch (error) {
        console.error("Failed to fetch options:", error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    };

    fetchOptions();
  }, [fieldSchema.dataSource, isSecret, isConnection]);

  const displayValue = (isSecret || isConnection)
    ? value.replace(/^\$\{(?:QAYNAQ_CONN_)?(.+)\}$/, "$1")
    : value;

  const handleChange = (selected: string) => {
    if (isSecret) {
      onChange(`\${${selected}}`);
    } else if (isConnection) {
      onChange(`\${QAYNAQ_CONN_${selected}}`);
    } else {
      onChange(selected);
    }
  };

  if (loading) {
    return (
      <div className="text-sm text-muted-foreground">Loading...</div>
    );
  }

  return (
    <Select value={displayValue} onValueChange={handleChange}>
      <SelectTrigger className="h-9 text-sm">
        <SelectValue
          placeholder={options.length === 0 ? "No options" : "Select..."}
        />
      </SelectTrigger>
      <SelectContent>
        {options.length === 0 ? (
          <div className="px-2 py-1.5 text-sm text-muted-foreground">
            No {fieldSchema.dataSource} available
          </div>
        ) : (
          options.map((option) => (
            <SelectItem key={option} value={option}>
              {option}
            </SelectItem>
          ))
        )}
      </SelectContent>
    </Select>
  );
}

function FileSelectField({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const [options, setOptions] = useState<{ key: string; label: string }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const files = await fetchFiles();
        setOptions(
          files.map((f) => ({
            key: f.key,
            label: f.key,
          })),
        );
      } catch (error) {
        console.error("Failed to fetch files:", error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const displayValue = value?.startsWith("qaynaq://")
    ? value.replace("qaynaq://", "")
    : value;

  if (loading) {
    return (
      <div className="text-sm text-muted-foreground">Loading...</div>
    );
  }

  return (
    <Select
      value={displayValue || ""}
      onValueChange={(val) => onChange(`qaynaq://${val}`)}
    >
      <SelectTrigger className="h-9 text-sm">
        <SelectValue
          placeholder={options.length === 0 ? "No files" : "Select file..."}
        />
      </SelectTrigger>
      <SelectContent>
        {options.length === 0 ? (
          <div className="px-2 py-1.5 text-sm text-muted-foreground">
            No files available. Create files in the Files page first.
          </div>
        ) : (
          options.map((option) => (
            <SelectItem key={option.key} value={option.key}>
              {option.label}
            </SelectItem>
          ))
        )}
      </SelectContent>
    </Select>
  );
}

export function InlineYamlEditor({
  schema,
  value,
  onChange,
  availableProcessors = [],
  availableInputs = [],
  availableOutputs = [],
  previewMode = false,
}: InlineYamlEditorProps) {
  const [fieldStates, setFieldStates] = useState<Record<string, FieldState>>(
    {},
  );
  const lastOutputRef = React.useRef<string>("");

  const {
    internalProcessors,
    setInternalProcessors,
    internalInputs,
    setInternalInputs,
    internalOutputs,
    setInternalOutputs,
    internalOutputCases,
    setInternalOutputCases,
    internalProcessorCases,
    setInternalProcessorCases,
    isInternalUpdateRef,
  } = useInternalState();

  const isFlat = schema.flat === true;
  const actualSchema = schema.properties || {};
  const flatFieldKey = isFlat ? Object.keys(actualSchema)[0] : null;
  const flatFieldSchema = flatFieldKey ? actualSchema[flatFieldKey] : null;

  useEffect(() => {
    if (isFlat && flatFieldSchema?.type === "processor_list") {
      const currentValue = Array.isArray(getCurrentValue())
        ? (getCurrentValue() as any[])
        : [];
      if (currentValue.length > 0 && internalProcessors.length === 0) {
        setInternalProcessors(currentValue);
      }
      return;
    }

    if (isFlat) return;

    const initialStates: Record<string, FieldState> = {};
    let existingData: any = {};

    if (value && value.trim()) {
      try {
        existingData = yaml.load(value) || {};
      } catch (error) {
        console.warn("Failed to parse existing YAML:", error);
      }
    }

    Object.entries(actualSchema).forEach(([fieldKey, fieldSchema]) => {
      const hasExistingValue = existingData.hasOwnProperty(fieldKey);
      const isRequired = fieldSchema.required === true;

      let initialValue = hasExistingValue
        ? existingData[fieldKey]
        : getDefaultValue(fieldSchema);

      if (hasExistingValue && Array.isArray(existingData[fieldKey])) {
        if (fieldSchema.type === "output_cases") {
          initialValue = convertYamlToOutputCases(
            existingData[fieldKey],
            availableOutputs,
          );
          setInternalOutputCases(initialValue);
        } else if (fieldSchema.type === "output_list") {
          initialValue = convertYamlToOutputList(
            existingData[fieldKey],
            availableOutputs,
          );
          setInternalOutputs(initialValue);
        } else if (fieldSchema.type === "input_list") {
          initialValue = convertYamlToInputList(
            existingData[fieldKey],
            availableInputs,
          );
          setInternalInputs(initialValue);
        } else if (fieldSchema.type === "processor_cases") {
          initialValue = convertYamlToProcessorCases(
            existingData[fieldKey],
            availableProcessors,
          );
          setInternalProcessorCases(initialValue);
        }
      }

      initialStates[fieldKey] = {
        enabled: isRequired || hasExistingValue,
        value: initialValue,
      };
    });

    setFieldStates((prev) => {
      const hasChanges = Object.keys(initialStates).some(
        (key) =>
          !prev[key] ||
          prev[key].enabled !== initialStates[key].enabled ||
          JSON.stringify(prev[key].value) !==
            JSON.stringify(initialStates[key].value),
      );

      if (
        hasChanges ||
        Object.keys(prev).length !== Object.keys(initialStates).length
      ) {
        return initialStates;
      }

      return prev;
    });
  }, [
    actualSchema,
    value,
    isFlat,
    availableProcessors,
    availableInputs,
    availableOutputs,
    internalProcessors.length,
  ]);

  const yamlOutput = useMemo(() => {
    if (isFlat) return value || "";

    const data: any = {};

    Object.entries(fieldStates).forEach(([fieldKey, state]) => {
      if (state.enabled && state.value !== undefined) {
        const fieldSchema = actualSchema[fieldKey];

        if (fieldSchema?.type === "output_cases") {
          const sourceValue =
            internalOutputCases.length > 0 ? internalOutputCases : state.value;
          const validCases = convertOutputCasesToYaml(
            sourceValue,
            availableOutputs,
          );
          data[fieldKey] = validCases;
        } else if (fieldSchema?.type === "output_list") {
          const sourceValue =
            internalOutputs.length > 0 ? internalOutputs : state.value;
          const validOutputs = convertOutputListToYaml(
            sourceValue,
            availableOutputs,
          );
          data[fieldKey] = validOutputs;
        } else if (fieldSchema?.type === "input_list") {
          const sourceValue =
            internalInputs.length > 0 ? internalInputs : state.value;
          const validInputs = convertInputListToYaml(
            sourceValue,
            availableInputs,
          );
          data[fieldKey] = validInputs;
        } else if (fieldSchema?.type === "processor_cases") {
          const sourceValue =
            internalProcessorCases.length > 0
              ? internalProcessorCases
              : state.value;
          const validCases = convertProcessorCasesToYaml(
            sourceValue,
            availableProcessors,
          );
          data[fieldKey] = validCases;
        } else {
          data[fieldKey] = state.value;
        }
      }
    });

    if (Object.keys(data).length === 0) {
      return "";
    }

    try {
      return yaml.dump(data, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        quotingType: '"',
        forceQuotes: false,
      });
    } catch (error) {
      console.error("Failed to generate YAML:", error);
      return "";
    }
  }, [
    fieldStates,
    isFlat,
    value,
    internalOutputCases,
    internalOutputs,
    internalInputs,
    internalProcessorCases,
    availableOutputs,
    availableInputs,
    availableProcessors,
    actualSchema,
  ]);

  useEffect(() => {
    if (isFlat) return;

    if (lastOutputRef.current === yamlOutput) return;

    let currentData: any = {};
    let newData: any = {};

    try {
      if (value && value.trim()) {
        currentData = yaml.load(value) || {};
      }
      if (yamlOutput && yamlOutput.trim()) {
        newData = yaml.load(yamlOutput) || {};
      }
    } catch (error) {
      if (yamlOutput !== value && lastOutputRef.current !== yamlOutput) {
        lastOutputRef.current = yamlOutput;
        onChange(yamlOutput);
      }
      return;
    }

    if (JSON.stringify(currentData) !== JSON.stringify(newData)) {
      lastOutputRef.current = yamlOutput;
      onChange(yamlOutput);
    }
  }, [yamlOutput, value, isFlat]);

  const updateFieldState = (fieldKey: string, updates: Partial<FieldState>) => {
    setFieldStates((prev) => ({
      ...prev,
      [fieldKey]: { ...prev[fieldKey], ...updates },
    }));
  };

  const renderValueInput = (
    fieldSchema: FieldSchema,
    state: FieldState,
    handleValueChange: (value: any) => void,
  ) => {
    switch (fieldSchema.type) {
      case "input":
        return (
          <TextInputField
            value={state.value || ""}
            onChange={handleValueChange}
            previewMode={previewMode}
            placeholder={fieldSchema.description || "Enter value..."}
          />
        );

      case "number":
        return (
          <Input
            type="number"
            value={state.value || 0}
            onChange={(e) => handleValueChange(Number(e.target.value))}
            className="h-9 text-sm"
          />
        );

      case "bool":
        return (
          <div className="flex items-center gap-2">
            <Switch
              checked={state.value || false}
              onCheckedChange={handleValueChange}
            />
            <span className="text-sm text-muted-foreground">
              {state.value ? "Enabled" : "Disabled"}
            </span>
          </div>
        );

      case "select":
        return (
          <Select value={state.value || ""} onValueChange={handleValueChange}>
            <SelectTrigger className="h-9 text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {fieldSchema.options?.map((option) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case "dynamic_select":
        return (
          <DynamicSelectField
            fieldSchema={fieldSchema}
            value={state.value || ""}
            onChange={handleValueChange}
          />
        );

      case "file":
        return (
          <FileSelectField
            value={state.value || ""}
            onChange={handleValueChange}
          />
        );

      case "code":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyCodeEditorField
              value={state.value || ""}
              onChange={handleValueChange}
              previewMode={previewMode}
              language={fieldSchema.language}
            />
          </Suspense>
        );

      case "key_value":
        return (
          <KeyValueEditor
            value={state.value || {}}
            updateValue={handleValueChange}
            previewMode={previewMode}
          />
        );

      case "array":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyArrayEditor
              value={state.value || []}
              updateValue={handleValueChange}
              previewMode={previewMode}
            />
          </Suspense>
        );

      case "property_list":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyPropertyListEditor
              value={state.value || []}
              updateValue={handleValueChange}
              previewMode={previewMode}
            />
          </Suspense>
        );

      case "object":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyObjectEditor
              value={state.value || {}}
              updateValue={handleValueChange}
              fieldSchema={fieldSchema}
              previewMode={previewMode}
            />
          </Suspense>
        );

      case "processor_list":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyProcessorListEditor
              value={state.value || []}
              updateValue={handleValueChange}
              availableProcessors={availableProcessors}
              availableInputs={availableInputs}
              availableOutputs={availableOutputs}
              previewMode={previewMode}
            />
          </Suspense>
        );

      case "output_cases":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyOutputCasesEditor
              value={state.value || []}
              updateValue={handleValueChange}
              availableOutputs={availableOutputs}
              availableProcessors={availableProcessors}
              availableInputs={availableInputs}
              internalValue={internalOutputCases}
              setInternalValue={setInternalOutputCases}
              isInternalUpdateRef={isInternalUpdateRef}
              previewMode={previewMode}
            />
          </Suspense>
        );

      case "output_list":
        return (
          <OutputListEditor
            value={state.value || []}
            updateValue={handleValueChange}
            availableOutputs={availableOutputs}
            availableProcessors={availableProcessors}
            availableInputs={availableInputs}
            previewMode={previewMode}
          />
        );

      case "input_list":
        return (
          <InputListEditor
            value={state.value || []}
            updateValue={handleValueChange}
            availableInputs={availableInputs}
            availableProcessors={availableProcessors}
            availableOutputs={availableOutputs}
            previewMode={previewMode}
          />
        );

      case "processor_cases":
        return (
          <Suspense
            fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
          >
            <LazyProcessorCasesEditor
              value={state.value || []}
              updateValue={handleValueChange}
              availableProcessors={availableProcessors}
              availableInputs={availableInputs}
              availableOutputs={availableOutputs}
              previewMode={previewMode}
            />
          </Suspense>
        );

      default:
        return (
          <TextInputField
            value={String(state.value || "")}
            onChange={handleValueChange}
            previewMode={previewMode}
            placeholder="Enter value..."
          />
        );
    }
  };

  const renderField = (fieldKey: string, fieldSchema: FieldSchema) => {
    const state = fieldStates[fieldKey];
    const isRequired = fieldSchema.required === true;

    if (!state) return null;

    if (previewMode && !state.enabled) return null;

    const handleValueChange = (newValue: any) => {
      updateFieldState(fieldKey, { value: newValue });
    };

    return (
      <div key={fieldKey} className="space-y-1.5">
        <div className="flex items-center gap-2">
          {!previewMode && (
            <Checkbox
              checked={state.enabled}
              disabled={isRequired}
              onCheckedChange={(checked) =>
                updateFieldState(fieldKey, { enabled: checked as boolean })
              }
              className="h-4 w-4"
            />
          )}
          <Label className={`text-sm ${state.enabled ? "text-foreground" : "text-muted-foreground"}`}>
            {fieldSchema.title || fieldKey}
            {!previewMode && isRequired && (
              <span className="text-destructive ml-1">*</span>
            )}
          </Label>
        </div>
        {fieldSchema.description && state.enabled && (
          <p className={`text-xs text-muted-foreground ${!previewMode ? "pl-6" : ""}`}>
            {fieldSchema.description}
          </p>
        )}
        {state.enabled && (
          <div className={!previewMode ? "pl-6" : ""}>
            {renderValueInput(fieldSchema, state, handleValueChange)}
          </div>
        )}
      </div>
    );
  };

  const getCurrentValue = (): any[] | string => {
    if (flatFieldSchema?.type === "processor_list") {
      if (!value || !value.trim()) return [];
      try {
        const parsedYaml = yaml.load(value) || [];

        if (Array.isArray(parsedYaml)) {
          return parsedYaml.map((proc) => {
            if (typeof proc === "object" && proc !== null) {
              const componentName = Object.keys(proc)[0];
              const config = proc[componentName];
              const processorSchema = availableProcessors.find(
                (p) => p.component === componentName,
              );

              if (processorSchema?.schema?.flat) {
                return {
                  componentId: processorSchema?.id || componentName,
                  component: componentName,
                  configYaml:
                    typeof config === "string"
                      ? config
                      : config
                        ? yaml.dump(config)
                        : "",
                };
              } else {
                return {
                  componentId: processorSchema?.id || componentName,
                  component: componentName,
                  configYaml: config ? yaml.dump(config) : "",
                };
              }
            }
            return { componentId: "", component: "", configYaml: "" };
          });
        }

        return [];
      } catch (error) {
        console.warn("Failed to parse processor list YAML:", error);
        return [];
      }
    } else if (flatFieldSchema?.type === "processor_cases") {
      if (!value || !value.trim()) return [];
      try {
        const parsedYaml = yaml.load(value) || [];
        if (Array.isArray(parsedYaml)) {
          return convertYamlToProcessorCases(parsedYaml, availableProcessors);
        }
        return [];
      } catch (error) {
        console.warn("Failed to parse processor cases YAML:", error);
        return [];
      }
    }
    return value || "";
  };

  if (isFlat && flatFieldKey && flatFieldSchema) {
    const handleFlatValueChange = (newValue: string | any[]) => {
      if (typeof newValue === "string") {
        onChange(newValue);
      } else if (flatFieldSchema.type === "processor_list") {
        const currentProcessors = Array.isArray(newValue) ? newValue : [];

        try {
          const processorsForYaml = currentProcessors
            .filter((proc) => {
              if (!proc.componentId || !proc.component) {
                return false;
              }

              const selectedProcessor = availableProcessors.find(
                (p) => p.id === proc.componentId,
              );

              const hasNoFields = selectedProcessor?.schema?.properties &&
                Object.keys(selectedProcessor.schema.properties).length === 0;
              if (hasNoFields) {
                return true;
              }

              if (!proc.configYaml || !proc.configYaml.trim()) {
                return false;
              }

              if (selectedProcessor?.schema?.flat) {
                return proc.configYaml.trim().length > 0;
              } else {
                try {
                  const config = yaml.load(proc.configYaml);
                  return (
                    config &&
                    typeof config === "object" &&
                    Object.keys(config).length > 0
                  );
                } catch {
                  return false;
                }
              }
            })
            .map((proc) => {
              const selectedProcessor = availableProcessors.find(
                (p) => p.id === proc.componentId,
              );

              if (selectedProcessor?.schema?.flat) {
                return { [proc.component]: proc.configYaml.trim() };
              } else {
                const processorObj: any = { [proc.component]: {} };
                try {
                  const config = yaml.load(proc.configYaml) || {};
                  processorObj[proc.component] = config;
                } catch (error) {
                  console.warn("Failed to parse processor config YAML:", error);
                  processorObj[proc.component] = {};
                }
                return processorObj;
              }
            });

          const yamlOutput =
            processorsForYaml.length > 0
              ? yaml.dump(processorsForYaml, {
                  lineWidth: -1,
                  noRefs: true,
                  quotingType: '"',
                  forceQuotes: false,
                })
              : "";
          onChange(yamlOutput);
        } catch (error) {
          console.warn("Failed to convert processor list to YAML:", error);
          onChange("");
        }
      } else if (flatFieldSchema.type === "processor_cases") {
        try {
          const validCases = convertProcessorCasesToYaml(
            newValue,
            availableProcessors,
          );
          const yamlOutput =
            validCases.length > 0
              ? yaml.dump(validCases, {
                  lineWidth: -1,
                  noRefs: true,
                  quotingType: '"',
                  forceQuotes: false,
                })
              : "";
          onChange(yamlOutput);
        } catch (error) {
          console.warn("Failed to convert processor cases to YAML:", error);
          onChange("");
        }
      }
    };

    return (
      <div className="p-4 bg-background rounded-md border">
        <div className="space-y-2">
          <Label className="text-sm text-muted-foreground">
            {flatFieldSchema.title || flatFieldKey}
          </Label>
          {flatFieldSchema.type === "code" ? (
            <Suspense
              fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
            >
              <LazyCodeEditorField
                value={value || ""}
                onChange={handleFlatValueChange as (value: string) => void}
                previewMode={previewMode}
                language={flatFieldSchema.language}
              />
            </Suspense>
          ) : flatFieldSchema.type === "processor_list" ? (
            <Suspense
              fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
            >
              <LazyProcessorListEditor
                value={
                  internalProcessors.length > 0
                    ? internalProcessors
                    : Array.isArray(getCurrentValue())
                      ? (getCurrentValue() as any[])
                      : []
                }
                updateValue={(newValue: any[]) => {
                  setInternalProcessors(newValue);
                  isInternalUpdateRef.current = true;
                  handleFlatValueChange(newValue);
                }}
                availableProcessors={availableProcessors}
                previewMode={previewMode}
              />
            </Suspense>
          ) : flatFieldSchema.type === "processor_cases" ? (
            <Suspense
              fallback={<div className="text-sm text-muted-foreground">Loading...</div>}
            >
              <LazyProcessorCasesEditor
                value={
                  Array.isArray(getCurrentValue())
                    ? (getCurrentValue() as any[])
                    : []
                }
                updateValue={handleFlatValueChange as (value: any[]) => void}
                availableProcessors={availableProcessors}
                previewMode={previewMode}
              />
            </Suspense>
          ) : (
            <TextInputField
              value={value || ""}
              onChange={handleFlatValueChange as (value: string) => void}
              previewMode={previewMode}
              placeholder={flatFieldSchema.description}
            />
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 bg-background rounded-md border">
      <div className="space-y-4">
        {Object.entries(actualSchema).map(([fieldKey, fieldSchema]) =>
          renderField(fieldKey, fieldSchema),
        )}
      </div>
    </div>
  );
}
