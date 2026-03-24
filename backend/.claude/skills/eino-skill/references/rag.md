# Eino RAG 模式详解

RAG（Retrieval-Augmented Generation）是通过检索外部文档增强 LLM 回答质量的模式。

## RAG 基本流程

```
用户 Query → Embedding → 向量检索 → 相关文档 → Prompt 拼接 → LLM 生成 → 答案
```

## 完整示例

### 1. 基础 RAG（Redis 向量库）

```go
package main

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino/components/embedding/openai"
    "github.com/cloudwego/eino/components/model/openai"
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/components/retriever/redis"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    "github.com/redis/go-redis/v9"
)

// RAG 状态
type RAGState struct {
    Query     string
    Documents []*schema.Document
    Answer    string
}

func main() {
    ctx := context.Background()

    // 1. 创建 Embedding
    embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
        APIKey: "sk-key",
        Model:  "text-embedding-3-small",
    })
    if err != nil {
        panic(err)
    }

    // 2. 创建 Redis 客户端
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // 3. 创建 Retriever
    retriever, err := redis.NewRetriever(ctx, &redis.RetrieverConfig{
        Client:    redisClient,
        Index:     "knowledge_base",
        Embedding: embedder,
        TopK:      5,
    })
    if err != nil {
        panic(err)
    }

    // 4. 创建 ChatModel
    chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-key",
        Model:  "gpt-4o",
    })
    if err != nil {
        panic(err)
    }

    // 5. 创建 RAG Prompt 模板
    ragTemplate := prompt.FromMessages(
        schema.FString,
        &schema.Message{
            Role: "system",
            Content: `你是一个智能助手，根据提供的参考文档回答问题。
如果文档中没有相关信息，请如实告知。

参考文档：
{documents}

问题：{query}`,
        },
    )

    // 6. 构建 RAG Graph
    g := compose.NewGraph[string, string](
        compose.WithGenLocalState(func(ctx context.Context) *RAGState {
            return &RAGState{}
        }),
    )

    // 检索节点
    _ = g.AddRetrieverNode("retrieve", retriever,
        compose.WithStatePreHandler(func(ctx context.Context, query string, state *RAGState) (string, error) {
            state.Query = query
            return query, nil
        }),
        compose.WithNodeName("文档检索"),
    )

    // Prompt 节点
    _ = g.AddChatTemplateNode("prompt", ragTemplate,
        compose.WithStatePreHandler(func(ctx context.Context, docs []*schema.Document, state *RAGState) (map[string]any, error) {
            state.Documents = docs
            // 拼接文档内容
            var docText string
            for i, doc := range docs {
                docText += fmt.Sprintf("文档 %d:\n%s\n\n", i+1, doc.Content)
            }
            return map[string]any{
                "documents": docText,
                "query":     state.Query,
            }, nil
        }),
        compose.WithNodeName("RAG Prompt"),
    )

    // 生成节点
    _ = g.AddChatModelNode("generate", chatModel,
        compose.WithNodeName("LLM 生成"),
    )

    // 提取节点
    _ = g.AddLambdaNode("extract", compose.InvokableLambda(
        func(ctx context.Context, msg *schema.Message) (string, error) {
            return msg.Content, nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, answer string, state *RAGState) (string, error) {
            state.Answer = answer
            return answer, nil
        }),
        compose.WithNodeName("答案提取"),
    )

    // 边
    _ = g.AddEdge(compose.START, "retrieve")
    _ = g.AddEdge("retrieve", "prompt")
    _ = g.AddEdge("prompt", "generate")
    _ = g.AddEdge("generate", "extract")
    _ = g.AddEdge("extract", compose.END)

    // 编译
    runnable, err := g.Compile(ctx,
        compose.WithGraphName("RAGPipeline"),
        compose.WithMaxRunSteps(10),
    )
    if err != nil {
        panic(err)
    }

    // 执行
    query := "公司的年假政策是什么？"
    answer, err := runnable.Invoke(ctx, query)
    if err != nil {
        panic(err)
    }

    fmt.Println("答案:", answer)
}
```

### 2. 带对话历史的 RAG

```go
type ConversationRAGState struct {
    History   []*schema.Message
    Query     string
    Documents []*schema.Document
    Answer    string
}

func BuildConversationRAG(ctx context.Context) (compose.Runnable[[]*schema.Message, string], error) {
    g := compose.NewGraph[[]*schema.Message, string](
        compose.WithGenLocalState(func(ctx context.Context) *ConversationRAGState {
            return &ConversationRAGState{}
        }),
    )

    // 重写查询（结合历史）
    _ = g.AddLambdaNode("rewrite", compose.InvokableLambda(
        func(ctx context.Context, history []*schema.Message) (string, error) {
            // 从历史中提取最新问题，必要时进行重写
            if len(history) == 0 {
                return "", nil
            }
            return history[len(history)-1].Content, nil
        }),
        compose.WithStatePreHandler(func(ctx context.Context, history []*schema.Message, state *ConversationRAGState) ([]*schema.Message, error) {
            state.History = history
            return history, nil
        }),
    )

    // 检索
    _ = g.AddRetrieverNode("retrieve", retriever,
        compose.WithStatePreHandler(func(ctx context.Context, query string, state *ConversationRAGState) (string, error) {
            state.Query = query
            return query, nil
        }),
    )

    // 带历史的 Prompt
    ragTemplate := prompt.FromMessages(
        schema.FString,
        &schema.Message{
            Role: "system",
            Content: `根据参考文档和对话历史回答问题。

参考文档：
{documents}

对话历史：
{history}

问题：{query}`,
        },
    )

    _ = g.AddChatTemplateNode("prompt", ragTemplate,
        compose.WithStatePreHandler(func(ctx context.Context, docs []*schema.Document, state *ConversationRAGState) (map[string]any, error) {
            state.Documents = docs
            var docText, historyText string
            for _, doc := range docs {
                docText += doc.Content + "\n\n"
            }
            for _, msg := range state.History {
                historyText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
            }
            return map[string]any{
                "documents": docText,
                "history":   historyText,
                "query":     state.Query,
            }, nil
        }),
    )

    _ = g.AddChatModelNode("generate", chatModel)

    _ = g.AddLambdaNode("extract", compose.InvokableLambda(
        func(ctx context.Context, msg *schema.Message) (string, error) {
            return msg.Content, nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, answer string, state *ConversationRAGState) (string, error) {
            state.Answer = answer
            // 将新对话加入历史
            state.History = append(state.History, &schema.Message{
                Role:    "assistant",
                Content: answer,
            })
            return answer, nil
        }),
    )

    _ = g.AddEdge(compose.START, "rewrite")
    _ = g.AddEdge("rewrite", "retrieve")
    _ = g.AddEdge("retrieve", "prompt")
    _ = g.AddEdge("prompt", "generate")
    _ = g.AddEdge("generate", "extract")
    _ = g.AddEdge("extract", compose.END)

    return g.Compile(ctx)
}
```

## 索引文档

在使用 RAG 之前，需要先将文档索引到向量库：

```go
// 索引文档
func IndexDocuments(ctx context.Context, docs []*schema.Document) error {
    // 1. 创建 Embedding
    embedder, _ := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
        APIKey: apiKey,
        Model:  "text-embedding-3-small",
    })

    // 2. 创建 Indexer
    indexer, _ := redis.NewIndexer(ctx, &redis.IndexerConfig{
        Client:    redisClient,
        KeyPrefix: "doc:",
        Index:     "knowledge_base",
        Embedding: embedder,
    })

    // 3. 存储文档
    ids, err := indexer.Store(ctx, docs)
    if err != nil {
        return err
    }
    fmt.Printf("索引了 %d 个文档\n", len(ids))
    return nil
}

// 示例文档
docs := []*schema.Document{
    {
        ID:      "policy-001",
        Content: "公司年假政策：员工每年享有 15 天带薪年假...",
        MetaData: map[string]any{
            "category": "hr",
            "updated":  "2024-01-01",
        },
    },
    // ... 更多文档
}
```

## 多路检索（混合检索）

```go
// 合并多个 Retriever 的结果
type HybridRetriever struct {
    vectorRetriever  compose.Retriever
    keywordRetriever compose.Retriever
}

func (h *HybridRetriever) Retrieve(ctx context.Context, query string, opts ...compose.Option) ([]*schema.Document, error) {
    // 向量检索
    vectorDocs, _ := h.vectorRetriever.Retrieve(ctx, query)
    // 关键词检索
    keywordDocs, _ := h.keywordRetriever.Retrieve(ctx, query)

    // 合并结果（去重、重排序）
    return mergeAndRerank(vectorDocs, keywordDocs), nil
}
```
