import { lazy } from 'react';

export const LazyOutputCasesEditor = lazy(() => 
  import('./output-cases-editor').then(module => ({ default: module.OutputCasesEditor }))
);

export const LazyProcessorCasesEditor = lazy(() => 
  import('./processor-cases-editor').then(module => ({ default: module.ProcessorCasesEditor }))
);

export const LazyProcessorListEditor = lazy(() => 
  import('./processor-list-editor').then(module => ({ default: module.ProcessorListEditor }))
);

export const LazyCodeEditorField = lazy(() => 
  import('./code-editor-field').then(module => ({ default: module.CodeEditorField }))
);

export const LazyObjectEditor = lazy(() => 
  import('./object-editor').then(module => ({ default: module.ObjectEditor }))
);

export const LazyArrayEditor = lazy(() =>
  import('./array-editor').then(module => ({ default: module.ArrayEditor }))
);

export const LazyPropertyListEditor = lazy(() =>
  import('./property-list-editor').then(module => ({ default: module.PropertyListEditor }))
);