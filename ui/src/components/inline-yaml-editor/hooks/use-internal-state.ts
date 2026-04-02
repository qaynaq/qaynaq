import { useState, useRef } from 'react';

export function useInternalState() {
  const [internalProcessors, setInternalProcessors] = useState<any[]>([]);
  const [internalInputs, setInternalInputs] = useState<any[]>([]);
  const [internalOutputs, setInternalOutputs] = useState<any[]>([]);
  const [internalOutputCases, setInternalOutputCases] = useState<any[]>([]);
  const [internalProcessorCases, setInternalProcessorCases] = useState<any[]>([]);
  const isInternalUpdateRef = useRef(false);

  return {
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
    isInternalUpdateRef
  };
} 