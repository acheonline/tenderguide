package loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fumiama/go-docx"

	"github.com/acheonline/tenderguide/internal/parser"
)

// Load читает все *.docx в dir (рекурсивно), извлекает абзацы и таблицы,
// возвращает []parser.Input — по одному Input на файл.
func Load(dir string) ([]parser.Input, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("loader: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("loader: %s is not a directory", dir)
	}

	var paths []string
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".docx") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loader: walk %s: %w", dir, err)
	}

	if len(paths) == 0 {
		return nil, errors.New("loader: no *.docx files in directory")
	}

	inputs := make([]parser.Input, 0, len(paths))
	for _, p := range paths {
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return nil, fmt.Errorf("loader: rel %s: %w", p, err)
		}
		in, err := parseFile(p, rel)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].FileName < inputs[j].FileName
	})

	return inputs, nil
}

func parseFile(path, rel string) (parser.Input, error) {
	f, err := os.Open(path)
	if err != nil {
		return parser.Input{}, fmt.Errorf("loader: %s: %w", rel, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return parser.Input{}, fmt.Errorf("loader: %s: %w", rel, err)
	}

	doc, err := docx.Parse(f, stat.Size())
	if err != nil {
		return parser.Input{}, fmt.Errorf("loader: %s: %w", rel, err)
	}

	in := parser.Input{FileName: rel}

	for _, item := range doc.Document.Body.Items {
		switch v := item.(type) {
		case *docx.Paragraph:
			if text := strings.TrimSpace(v.String()); text != "" {
				in.Paragraphs = append(in.Paragraphs, text)
			}
		case *docx.Table:
			if rows := tableRows(v); len(rows) > 0 {
				in.Tables = append(in.Tables, rows)
			}
		}
	}

	return in, nil
}

func tableRows(t *docx.Table) []parser.Row {
	if t == nil || len(t.TableRows) == 0 {
		return nil
	}
	out := make([]parser.Row, 0, len(t.TableRows))
	for _, row := range t.TableRows {
		if row == nil {
			continue
		}
		var cells []string
		for _, cell := range row.TableCells {
			if cell == nil {
				cells = append(cells, "")
				continue
			}
			// У WTableCell нет String(). Текст собираем из его параграфов.
			var b strings.Builder
			for _, p := range cell.Paragraphs {
				if p == nil {
					continue
				}
				text := strings.TrimSpace(p.String())
				if text != "" {
					if b.Len() > 0 {
						b.WriteString("\n")
					}
					b.WriteString(text)
				}
			}
			cells = append(cells, b.String())
		}
		out = append(out, parser.Row{Cells: cells})
	}
	return out
}
