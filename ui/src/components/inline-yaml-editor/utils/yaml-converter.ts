import * as yaml from "js-yaml";
import { ProcessorComponentSchema } from "../types";

export function convertOutputCasesToYaml(
  outputCases: any[],
  availableOutputs: ProcessorComponentSchema[]
): any[] {
  return outputCases
    .filter((caseItem: any) => caseItem.check || (caseItem.output?.componentId && caseItem.output?.component))
    .map((caseItem: any) => {
      const output = caseItem.output;
      
      const result: any = {
        check: caseItem.check || ""
      };
      
      if (output?.componentId && output?.component) {
        const outputObj: any = { [output.component]: {} };
        if (output.configYaml && output.configYaml.trim()) {
          try {
            outputObj[output.component] = yaml.load(output.configYaml) || {};
          } catch (error) {
            console.warn("Failed to parse output config YAML:", error);
          }
        }
        result.output = outputObj;
      }
      
      if (caseItem.continue) {
        result.continue = true;
      }
      
      return result;
    });
}

export function convertOutputListToYaml(
  outputList: any[],
  availableOutputs: ProcessorComponentSchema[]
): any[] {
  return outputList
    .filter((output: any) => output?.componentId && output?.component)
    .map((output: any) => {
      const outputObj: any = { [output.component]: {} };
      if (output.configYaml && output.configYaml.trim()) {
        try {
          outputObj[output.component] = yaml.load(output.configYaml) || {};
        } catch (error) {
          console.warn("Failed to parse output config YAML:", error);
        }
      }
      return outputObj;
    });
}

export function convertInputListToYaml(
  inputList: any[],
  availableInputs: ProcessorComponentSchema[]
): any[] {
  return inputList
    .filter((input: any) => input?.componentId && input?.component)
    .map((input: any) => {
      const inputObj: any = { [input.component]: {} };
      if (input.configYaml && input.configYaml.trim()) {
        try {
          inputObj[input.component] = yaml.load(input.configYaml) || {};
        } catch (error) {
          console.warn("Failed to parse input config YAML:", error);
        }
      }
      return inputObj;
    });
}

export function convertProcessorCasesToYaml(
  processorCases: any[],
  availableProcessors: ProcessorComponentSchema[]
): any[] {
  return processorCases
    .filter((caseItem: any) => (caseItem.processors && caseItem.processors.length > 0))
    .map((caseItem: any) => {
      const result: any = {
        check: caseItem.check || ""
      };
      
      if (caseItem.processors && Array.isArray(caseItem.processors)) {
        const validProcessors = caseItem.processors
          .filter((proc: any) => proc?.componentId && proc?.component)
          .map((proc: any) => {
            const selectedProcessor = availableProcessors.find(p => p.id === proc.componentId);
            
            if (selectedProcessor?.schema?.flat) {
              return { [proc.component]: proc.configYaml?.trim() || "" };
            } else {
              const processorObj: any = { [proc.component]: {} };
              if (proc.configYaml && proc.configYaml.trim()) {
                try {
                  processorObj[proc.component] = yaml.load(proc.configYaml) || {};
                } catch (error) {
                  console.warn("Failed to parse processor config YAML:", error);
                }
              }
              return processorObj;
            }
          });
        result.processors = validProcessors;
      } else {
        result.processors = [];
      }
      
      if (caseItem.fallthrough) {
        result.fallthrough = true;
      }
      
      return result;
    });
}

export function convertYamlToOutputCases(
  yamlValue: any[],
  availableOutputs: ProcessorComponentSchema[]
): any[] {
  return yamlValue.map((caseItem: any) => {
    if (typeof caseItem === 'object' && caseItem.output) {
      const outputComponent = Object.keys(caseItem.output)[0];
      const outputConfig = caseItem.output[outputComponent];
      const outputSchema = availableOutputs.find(o => o.component === outputComponent);
      
      return {
        check: caseItem.check || "",
        continue: caseItem.continue || false,
        output: {
          componentId: outputSchema?.id || outputComponent,
          component: outputComponent,
          configYaml: outputConfig && typeof outputConfig === 'object' && Object.keys(outputConfig).length > 0 
            ? yaml.dump(outputConfig) 
            : ""
        }
      };
    }
    return caseItem;
  });
}

export function convertYamlToOutputList(
  yamlValue: any[],
  availableOutputs: ProcessorComponentSchema[]
): any[] {
  return yamlValue.map((output: any) => {
    if (typeof output === 'object' && output !== null) {
      const componentName = Object.keys(output)[0];
      const config = output[componentName];
      const outputSchema = availableOutputs.find(o => o.component === componentName);
      
      return {
        componentId: outputSchema?.id || componentName,
        component: componentName,
        configYaml: config ? yaml.dump(config) : ""
      };
    }
    return { componentId: "", component: "", configYaml: "" };
  });
}

export function convertYamlToInputList(
  yamlValue: any[],
  availableInputs: ProcessorComponentSchema[]
): any[] {
  return yamlValue.map((input: any) => {
    if (typeof input === 'object' && input !== null) {
      const componentName = Object.keys(input)[0];
      const config = input[componentName];
      const inputSchema = availableInputs.find(i => i.component === componentName);
      
      return {
        componentId: inputSchema?.id || componentName,
        component: componentName,
        configYaml: config ? yaml.dump(config) : ""
      };
    }
    return { componentId: "", component: "", configYaml: "" };
  });
}

export function convertYamlToProcessorCases(
  yamlValue: any[],
  availableProcessors: ProcessorComponentSchema[]
): any[] {
  return yamlValue.map((caseItem: any) => {
    const result: any = {
      check: caseItem.check || "",
      fallthrough: caseItem.fallthrough || false,
      processors: []
    };
    
    if (caseItem.processors && Array.isArray(caseItem.processors)) {
      result.processors = caseItem.processors.map((proc: any) => {
        if (typeof proc === 'object' && proc !== null) {
          const componentName = Object.keys(proc)[0];
          const config = proc[componentName];
          const processorSchema = availableProcessors.find(p => p.component === componentName);
          
          if (processorSchema?.schema?.flat) {
            return {
              componentId: processorSchema?.id || componentName,
              component: componentName,
              configYaml: typeof config === 'string' ? config : (config ? yaml.dump(config) : "")
            };
          } else {
            return {
              componentId: processorSchema?.id || componentName,
              component: componentName,
              configYaml: config ? yaml.dump(config) : ""
            };
          }
        }
        return { componentId: "", component: "", configYaml: "" };
      });
    }
    
    return result;
  });
} 