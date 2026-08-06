/// <reference types="vite/client" />

declare module 'plotly.js-dist-min' {
  const Plotly: {
    react: (root: HTMLElement, data: unknown, layout?: unknown, config?: unknown) => Promise<unknown>
    toImage: (root: HTMLElement, opts: { format: string; width: number; height: number }) => Promise<string>
  }
  export default Plotly
}

declare module 'katex' {
  const katex: {
    renderToString: (tex: string, options?: { throwOnError?: boolean; displayMode?: boolean }) => string
  }
  export default katex
}
