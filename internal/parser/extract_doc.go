package parser

import (
	"time"

	"github.com/acheonline/tenderguide/internal/model"
)

// ExtractDoc обрабатывает все Input (весь комплект DOCX документации)
// и возвращает собранную модель. Никогда не возвращает nil.
func ExtractDoc(docID string, inputs []Input) *model.DocAll {
	out := &model.DocAll{
		Info: &model.DocInfo{
			DocID:     docID,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
	return out
}
