package xmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Markdown2png 将 Markdown 转换为图片字节流
func Markdown2png(content string) ([]byte, error) {
	// 1. 准备 CSS 样式
	// 修改点：width 调整为 fit-content 并配合 min-width
	const customStyle = `
    /* 容器基础样式 */
    .markdown-image-container {
      /* [修改] 移除固定宽度，改为自适应 */
      width: fit-content; 
      /* [修改] 保持原有视觉宽度作为最小值，防止短文本导致图片过窄 */
      min-width: 980px; 
      
      padding: 32px 40px;
      background-color: #ffffff;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
      font-size: 16px;
      line-height: 1.6;
      color: #1f2328;
      box-sizing: border-box;
      margin: 0 auto;
    }

    /* 标题样式 */
    .markdown-image-container h1 { font-size: 2em; font-weight: 600; padding-bottom: 0.3em; border-bottom: 1px solid #d1d9e0; margin: 24px 0 16px 0; }
    .markdown-image-container h2 { font-size: 1.5em; font-weight: 600; padding-bottom: 0.3em; border-bottom: 1px solid #d1d9e0; margin: 24px 0 16px 0; }
    .markdown-image-container h3 { font-size: 1.25em; font-weight: 600; margin: 24px 0 16px 0; }
    .markdown-image-container h4, .markdown-image-container h5, .markdown-image-container h6 { font-weight: 600; margin: 24px 0 16px 0; }
    .markdown-image-container h1:first-child, .markdown-image-container h2:first-child, .markdown-image-container h3:first-child { margin-top: 0; }

    /* 段落 */
    .markdown-image-container p { margin: 0 0 16px 0; }
    .markdown-image-container p:last-child { margin-bottom: 0; }

    /* 代码块 */
    .markdown-image-container pre {
      background-color: #f6f8fa;
      border: 1px solid #d1d9e0;
      border-radius: 6px;
      padding: 16px;
      overflow-x: auto;
      margin: 16px 0;
      font-size: 14px;
      line-height: 1.45;
    }
    .markdown-image-container pre code {
      background-color: transparent;
      padding: 0;
      border: none;
      color: #1f2328;
      font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
      font-size: 14px;
      white-space: pre-wrap;
      word-break: break-all;
    }

    /* 行内代码 */
    .markdown-image-container code {
      background-color: #eff1f3;
      padding: 0.2em 0.4em;
      border-radius: 6px;
      font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
      font-size: 85%;
      color: #1f2328;
    }

    /* 列表 */
    .markdown-image-container ul, .markdown-image-container ol { margin: 0 0 16px 0; padding-left: 2em; }
    .markdown-image-container li { margin: 4px 0; }
    .markdown-image-container li + li { margin-top: 4px; }

    /* 引用块 */
    .markdown-image-container blockquote { margin: 16px 0; padding: 0 1em; border-left: 4px solid #d1d9e0; color: #656d76; }
    .markdown-image-container blockquote p { margin: 0; }

    /* 表格 - 确保宽度占满容器 */
    .markdown-image-container table { border-collapse: collapse; margin: 16px 0; width: 100%; border: 1px solid #d1d9e0; }
    .markdown-image-container th, .markdown-image-container td { border: 1px solid #d1d9e0; padding: 8px 16px; text-align: left; }
    .markdown-image-container th { background-color: #f6f8fa; font-weight: 600; }
    .markdown-image-container tr:nth-child(even) { background-color: #f6f8fa; }

    /* 链接与图片 */
    .markdown-image-container a { color: #0969da; text-decoration: none; }
    .markdown-image-container img { max-width: 100%; height: auto; border-radius: 6px; }
    .markdown-image-container hr { height: 0.25em; padding: 0; margin: 24px 0; background-color: #d1d9e0; border: 0; }
    .markdown-image-container strong { font-weight: 600; }

    /* 语法高亮 */
    .markdown-image-container .hljs-keyword, .markdown-image-container .hljs-selector-tag { color: #cf222e; }
    .markdown-image-container .hljs-string, .markdown-image-container .hljs-attr { color: #0a3069; }
    .markdown-image-container .hljs-number, .markdown-image-container .hljs-literal { color: #0550ae; }
    .markdown-image-container .hljs-comment { color: #6e7781; }
    .markdown-image-container .hljs-function, .markdown-image-container .hljs-title { color: #8250df; }
    .markdown-image-container .hljs-built_in { color: #0550ae; }
    .markdown-image-container .hljs-type, .markdown-image-container .hljs-class { color: #953800; }
    `

	// 2. 构建 HTML 模板
	htmlTemplate := `
    <!DOCTYPE html>
    <html>
    <head>
       <meta charset="UTF-8">
       <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
       <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
       <style>
          body { 
            margin: 0; 
            padding: 0; 
            background: #fff; 
            /* 确保 body 允许子元素溢出撑开宽度 */
            width: fit-content; 
            min-width: 100%;
          }
          ` + customStyle + `
       </style>
    </head>
    <body>
       <div id="wrapper" class="markdown-image-container"></div>
       <script>
            marked.setOptions({
                highlight: function(code, lang) {
                    const language = highlight.getLanguage(lang) ? lang : 'plaintext';
                    return highlight.highlight(code, { language }).value;
                },
                langPrefix: 'hljs language-' 
            });

          function render(md) {
             document.getElementById('wrapper').innerHTML = marked.parse(md);
          }
       </script>
    </body>
    </html>
    `

	// 3. 启动 Chrome
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1200, 1), // 初始宽度，后续会动态调整
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	mdJSON, _ := json.Marshal(content)
	jsInject := fmt.Sprintf("render(%s);", string(mdJSON))

	var buf []byte
	var contentWidth, contentHeight float64

	// 4. 执行渲染任务
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),

		// 写入 HTML
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlTemplate).Do(ctx)
		}),

		// 等待加载
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`typeof marked !== 'undefined' && typeof highlight !== 'undefined'`, nil),

		// 渲染内容
		chromedp.Evaluate(jsInject, nil),
		chromedp.WaitVisible("#wrapper"),

		// [核心修改] 获取实际渲染后的内容尺寸
		// 使用 scrollWidth/scrollHeight 获取完整尺寸（包含溢出部分）
		chromedp.Evaluate(`document.getElementById('wrapper').scrollWidth`, &contentWidth),
		chromedp.Evaluate(`document.getElementById('wrapper').scrollHeight`, &contentHeight),

		// [核心修改] 动态调整视口大小以匹配内容
		// 这样可以确保截图时不会因为视口过小而截断宽表格
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 宽度取 max(1200, 实际宽度)，确保不小于初始设置
			// 必须转换为 int64
			w := int64(math.Ceil(contentWidth))
			h := int64(math.Ceil(contentHeight))

			// 强制设置设备指标，scale: 2.0 保证高清
			return emulation.SetDeviceMetricsOverride(w, h, 2.0, false).Do(ctx)
		}),

		// 截图
		chromedp.Screenshot("#wrapper", &buf, chromedp.NodeVisible),
	)

	if err != nil {
		return nil, fmt.Errorf("render error: %v", err)
	}

	return buf, nil
}
