export const architectureRoles = [
  "user",
  "orchestrator",
  "compute",
  "transform",
  "data",
  "automation",
] as const;

export type ArchitectureRole = (typeof architectureRoles)[number];

export interface ArchitectureMode {
  id: string;
  label: string;
}

export interface ArchitectureNodeVariant {
  label?: string;
  tech?: string;
  detail?: string;
}

export interface ArchitectureNode {
  id: string;
  role: ArchitectureRole;
  label: string;
  tech?: string;
  detail?: string;
  /** Percentage coordinates within the diagram stage. */
  x: number;
  y: number;
  /** Omit to show the node in every mode. */
  visibleIn?: string[];
  variants?: Record<string, ArchitectureNodeVariant>;
}

export interface ArchitectureStepVariant {
  title?: string;
  route?: string;
  description?: string;
  payload?: string;
  chips?: string[];
}

export interface ArchitectureStep extends ArchitectureStepVariant {
  from: string;
  to: string;
  title: string;
  description: string;
  variants?: Record<string, ArchitectureStepVariant>;
}

export interface ArchitectureFlow {
  id: string;
  label: string;
  note?: string;
  /** Omit to make the flow available in every mode. */
  visibleIn?: string[];
  steps: ArchitectureStep[];
}

export interface ArchitectureDiagramDefinition {
  /** Stable, kebab-case identifier; unique on the page. */
  id: string;
  title: string;
  summary?: string;
  nodes: ArchitectureNode[];
  flows: ArchitectureFlow[];
  modes?: ArchitectureMode[];
  initialMode?: string;
  autoplayMs?: number;
  theme?: "auto" | "light" | "dark";
}

