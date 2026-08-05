export type SessionSummary = {
  id: string;
  title: string;
  version: number;
  updatedAt: string;
};

export type Message = {
  id: string;
  role: "user" | "assistant";
  content: string;
  traceId?: string;
  createdAt: string;
};

export type Equation = {
  id: string;
  expression: string;
  label: string;
  color: string;
  visible: boolean;
  lineWidth: number;
};

export type Viewport = { xMin: number; xMax: number; yMin: number; yMax: number };
export type GraphSettings = { showGrid: boolean; showAxes: boolean; showLegend: boolean; sampleCount: number };
export type GraphAnalysis = { summary: string; highlights: string[] };

export type GraphState = {
  equations: Equation[];
  viewport: Viewport;
  settings: GraphSettings;
  analysis?: GraphAnalysis;
};

export type Session = SessionSummary & {
  messages: Message[];
  graphState: GraphState;
  createdAt: string;
};

export type TraceRecord = {
  id: string;
  sessionId: string;
  createdAt: string;
  spans: Array<{ id: string; name: string; started: string; ended: string; error?: string }>;
  events: Array<{ kind: string; tool?: string; status: string; duration: number; error?: string }>;
};
