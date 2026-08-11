import katex from 'katex'
import 'katex/dist/katex.min.css'

// 用 KaTeX 渲染公式(设计稿要求)。
export function Math({ tex }: { tex: string }) {
  try {
    const html = katex.renderToString(tex, { throwOnError: false })
    return <span dangerouslySetInnerHTML={{ __html: html }} />
  } catch {
    return <span>{tex}</span>
  }
}
