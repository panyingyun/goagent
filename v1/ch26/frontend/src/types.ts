// 与设计稿 §6 一致的数据类型(后端返回同构 JSON)。

export type Intent =
  | 'plot' | 'add_equation' | 'update_equation' | 'remove_equation'
  | 'update_viewport' | 'analyze' | 'explain' | 'unknown'

export interface Viewport { xMin: number; xMax: number; yMin: number; yMax: number }

export interface GraphSettings {
  showGrid: boolean
  showAxis: boolean
  showLegend: boolean
  sampleCount: number
}

export interface EquationItem {
  id: string
  type: string
  expression: string
  normalizedExpression: string
  label: string
  color: string
  visible: boolean
  lineWidth: number
}

export interface EquationUpdate {
  color?: string
  lineWidth?: number
  visible?: boolean
  label?: string
  expression?: string
  normalizedExpression?: string
}

export interface KeyPoint { label: string; x: number; y: number }

export interface GraphAnalysis {
  functionType?: string
  keyPoints?: KeyPoint[] | unknown
  monotonicity?: string[] | string
  zeros?: number[] | string
  symmetry?: string
  asymptotes?: string[] | string
  description?: string
}

export interface LLMStructuredResult {
  intent: Intent
  equations?: EquationItem[]
  viewport?: Partial<Viewport>
  targetEquationId?: string
  updates?: EquationUpdate
  explanation?: string
  analysis?: GraphAnalysis
  error?: string
}

export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  structuredResult?: LLMStructuredResult
  createdAt: string
  status?: string
}

export interface GraphState {
  equations: EquationItem[]
  viewport: Viewport
  settings: GraphSettings
  analysis?: GraphAnalysis
}

export interface Session {
  id: string
  title: string
  messages: Message[]
  graphState: GraphState
  isFavorite: boolean
  createdAt: string
  updatedAt: string
}

export interface ChatResponse {
  message: Message
  graphState: GraphState
}
