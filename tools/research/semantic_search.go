package research

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/tools/research/websearch"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- Pinecone Connection Helper ---

func createPineconeConn(apiKey, host string) (*pinecone.IndexConnection, error) {
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	idxConn, err := pc.Index(pinecone.NewIndexConnParams{Host: host})
	if err != nil {
		return nil, err
	}
	return idxConn, nil
}

// --- Pinecone Indexer Implementation ---

type PineconeIndexer struct {
	indexConn *pinecone.IndexConnection
	embedder  embedding.Embedder
	namespace string
}

func (p *PineconeIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	options := indexer.GetCommonOptions(nil, opts...)
	embedder := p.embedder
	if options.Embedding != nil {
		embedder = options.Embedding
	}

	vectors := make([]*pinecone.Vector, 0, len(docs))
	ids = make([]string, 0, len(docs))

	for _, doc := range docs {
		docID := doc.ID
		if docID == "" {
			docID = fmt.Sprintf("doc_%d", time.Now().UnixNano())
		}
		ids = append(ids, docID)

		var vector []float32
		if doc.DenseVector() != nil {
			denseVector := doc.DenseVector()
			vector = make([]float32, len(denseVector))
			for i, v := range denseVector {
				vector[i] = float32(v)
			}
		} else {
			embeddings, err := embedder.EmbedStrings(ctx, []string{doc.Content})
			if err != nil {
				return nil, err
			}
			if len(embeddings) > 0 {
				vector = make([]float32, len(embeddings[0]))
				for i, v := range embeddings[0] {
					vector[i] = float32(v)
				}
			}
		}

		if vector == nil {
			continue
		}

		metadata := map[string]interface{}{"content": doc.Content}
		for k, v := range doc.MetaData {
			metadata[k] = v
		}

		pineconeMetadata, err := structpb.NewStruct(metadata)
		if err != nil {
			return nil, err
		}

		vectors = append(vectors, &pinecone.Vector{
			Id:       docID,
			Values:   &vector,
			Metadata: pineconeMetadata,
		})
	}

	if len(vectors) > 0 {
		batchSize := 10
		for i := 0; i < len(vectors); i += batchSize {
			end := i + batchSize
			if end > len(vectors) {
				end = len(vectors)
			}
			_, err = p.indexConn.UpsertVectors(ctx, vectors[i:end])
			if err != nil {
				return nil, err
			}
		}
	}

	return ids, nil
}

// --- Semantic Retriever Implementation ---

type SemanticRetriever struct {
	indexConn *pinecone.IndexConnection
	embedder  embedding.Embedder
	namespace string
}

func (r *SemanticRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	queryVectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	vector32 := make([]float32, len(queryVectors[0]))
	for i, v := range queryVectors[0] {
		vector32[i] = float32(v)
	}

	defaultTopK := 5
	options := retriever.GetCommonOptions(&retriever.Options{TopK: &defaultTopK}, opts...)

	queryReq := &pinecone.QueryByVectorValuesRequest{
		Vector:          vector32,
		TopK:            uint32(*options.TopK),
		IncludeValues:   false,
		IncludeMetadata: true,
	}

	res, err := r.indexConn.QueryByVectorValues(ctx, queryReq)
	if err != nil {
		return nil, err
	}

	docs := make([]*schema.Document, len(res.Matches))
	for i, match := range res.Matches {
		metadata := match.Vector.Metadata.AsMap()
		content, _ := metadata["content"].(string)
		docs[i] = &schema.Document{
			ID:       match.Vector.Id,
			Content:  content,
			MetaData: metadata,
		}
	}
	return docs, nil
}

// --- Tools and Factories ---

type SemanticSearchRequest struct {
	Query      string `json:"query" jsonschema:"required,description=搜索意图，如'权限校验逻辑'"`
	MaxResults int    `json:"max_results" jsonschema:"description=返回结果数，默认为5"`
}

type SemanticSearchResponse struct {
	Results []websearch.SearchResult `json:"results"`
}

func SemanticSearch(ctx context.Context, req *SemanticSearchRequest) (*SemanticSearchResponse, error) {
	cfg := config.Get().Tools

	idxConn, err := createPineconeConn(cfg.Pinecone.ApiKey, cfg.Pinecone.Host)
	if err != nil {
		return nil, err
	}

	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     cfg.Embedding.ApiKey,
		BaseURL:    cfg.Embedding.BaseUrl,
		Model:      cfg.Embedding.Model,
		Dimensions: &cfg.Embedding.Dimensions,
	})
	if err != nil {
		return nil, err
	}

	retr := &SemanticRetriever{
		indexConn: idxConn,
		embedder:  embedder,
		namespace: cfg.Pinecone.Namespace,
	}

	topK := req.MaxResults
	if topK <= 0 {
		topK = 5
	}

	docs, err := retr.Retrieve(ctx, req.Query, retriever.WithTopK(topK))
	if err != nil {
		return nil, err
	}

	var results []websearch.SearchResult
	for _, doc := range docs {
		results = append(results, websearch.SearchResult{
			Title:       doc.ID,
			URL:         doc.MetaData["path"].(string), // 假设分块时带了 path
			Description: doc.Content,
		})
	}

	return &SemanticSearchResponse{Results: results}, nil
}

type CodeIndexRequest struct {
	Path string `json:"path" jsonschema:"description=要索引的目录路径，默认为当前项目根目录"`
}

type CodeIndexResponse struct {
	Message string `json:"message"`
}

func CodeIndex(ctx context.Context, req *CodeIndexRequest) (*CodeIndexResponse, error) {
	cfg := config.Get().Tools
	searchPath := req.Path
	if searchPath == "" {
		searchPath = "."
	}

	// 1. Loader
	ldr, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{})
	if err != nil {
		return nil, err
	}

	// 2. Splitter
	spl, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{"#": "h1", "##": "h2"},
	})
	if err != nil {
		return nil, err
	}

	// 3. Indexer
	idxConn, err := createPineconeConn(cfg.Pinecone.ApiKey, cfg.Pinecone.Host)
	if err != nil {
		return nil, err
	}

	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     cfg.Embedding.ApiKey,
		BaseURL:    cfg.Embedding.BaseUrl,
		Model:      cfg.Embedding.Model,
		Dimensions: &cfg.Embedding.Dimensions,
	})
	if err != nil {
		return nil, err
	}

	idx := &PineconeIndexer{
		indexConn: idxConn,
		embedder:  embedder,
		namespace: cfg.Pinecone.Namespace,
	}

	// 4. Graph
	g := compose.NewGraph[document.Source, []string]()
	_ = g.AddLoaderNode("loader", ldr)
	_ = g.AddDocumentTransformerNode("splitter", spl)
	_ = g.AddIndexerNode("indexer", idx)

	_ = g.AddEdge(compose.START, "loader")
	_ = g.AddEdge("loader", "splitter")
	_ = g.AddEdge("splitter", "indexer")
	_ = g.AddEdge("indexer", compose.END)

	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}

	// 5. Execution Logic (Dir vs File)
	var totalDocs int

	info, err := os.Stat(searchPath)
	if err != nil {
		return nil, fmt.Errorf("无法访问路径: %w", err)
	}

	processFile := func(path string) error {
		_, err := runnable.Invoke(ctx, document.Source{URI: path})
		if err != nil {
			// 记录错误但继续处理
			fmt.Printf("索引文件失败 %s: %v\n", path, err)
			return nil
		}
		totalDocs++
		return nil
	}

	if info.IsDir() {
		err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				// 忽略隐藏目录 (如 .git, .idea)
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}
			// 忽略隐藏文件
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			// 简单的文件类型过滤
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".go" && ext != ".md" && ext != ".txt" && ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}

			return processFile(path)
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := processFile(searchPath); err != nil {
			return nil, err
		}
	}

	return &CodeIndexResponse{Message: fmt.Sprintf("索引任务已完成，共处理 %d 个文件", totalDocs)}, nil
}

func NewSemanticSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_semantic_search", "在语义库中进行模糊查询，返回最相关的代码片段", SemanticSearch)
}

func NewCodeIndexTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_code_index", "扫描指定目录并更新语义索引库", CodeIndex)
}
