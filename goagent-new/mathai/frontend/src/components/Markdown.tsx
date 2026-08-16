import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeKatex from 'rehype-katex'

/**
 * 渲染 AI 回复的 Markdown：加粗、列表、代码等由 react-markdown 处理，
 * `$...$` / `$$...$$` 数学公式由 remark-math + rehype-katex 渲染。
 */
export function Markdown({ children }: { children: string }) {
  return (
    <div className="markdown-body">
      <ReactMarkdown remarkPlugins={[remarkGfm, remarkMath]} rehypePlugins={[rehypeKatex]}>
        {children}
      </ReactMarkdown>
    </div>
  )
}
