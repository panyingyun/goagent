import katex from 'katex'

export function MathTex({ tex, display = false }: { tex: string; display?: boolean }) {
  let html = tex
  try {
    html = katex.renderToString(tex, { throwOnError: false, displayMode: display })
  } catch {
    /* keep raw */
  }
  return <span className="font-mono text-[15px]" dangerouslySetInnerHTML={{ __html: html }} />
}
