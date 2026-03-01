package backendmodule

import (
	"embed"

	"github.com/go-go-golems/go-go-os-backend/pkg/docmw"
)

//go:embed docs/*.md
var embeddedDocsFS embed.FS

func loadDocStore() (*docmw.DocStore, error) {
	return docmw.ParseFS(AppID, embeddedDocsFS, docmw.ParseOptions{})
}
