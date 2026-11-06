package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	_ "strings"

	"github.com/acheonline/tenderguide/internal/model"
	"github.com/acheonline/tenderguide/internal/storage"
)

func runShowDoc(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("show-doc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.String("id", "", "doc_id")
	dbPath := fs.String("db", "./tenderguide.db", "path to SQLite database")
	format := fs.String("format", "text", "output format: text or json")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *id == "" {
		fmt.Fprintln(os.Stderr, "--id is required")
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "invalid format: %s\n", *format)
		return 2
	}

	dsn := "file:" + *dbPath + "?_pragma=foreign_keys(1)"
	store, err := storage.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer store.Close()

	doc, err := store.GetDocByDocID(ctx, *id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doc not found: %s (%v)\n", *id, err)
		return 1
	}

	if *format == "json" {
		b, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(string(b))
		return 0
	}

	printDocText(doc)
	return 0
}

func printDocText(d *model.DocAll) {
	if d == nil || d.Info == nil {
		fmt.Println("(нет данных)")
		return
	}
	info := d.Info

	fmt.Println("=== Документация о закупке ===")
	fmt.Printf("ID:                 %s\n", info.DocID)
	fmt.Printf("Заказчик:           %s\n", strOrDash(info.CustomerName))
	fmt.Printf("ИНН заказчика:      %s\n", strOrDash(info.CustomerINN))
	fmt.Printf("Организатор:        %s\n", strOrDash(info.OrganizerName))
	fmt.Printf("Номер процедуры:    %s\n", strOrDash(info.ProcedureNumber))
	fmt.Printf("Предмет:            %s\n", strOrDash(info.ProcedureName))
	fmt.Printf("НМЦ:                %s RUB\n", strOrDash(info.NMCK))
	fmt.Printf("НМЦ без НДС:        %s\n", strOrDash(info.NMCKWithoutVAT))
	fmt.Printf("НДС:                %s\n", strOrDash(info.VAT))
	fmt.Printf("Срок подачи:        %s\n", strOrDash(info.SubmissionDeadline))
	fmt.Printf("Рассмотрение:       %s\n", strOrDash(info.ReviewDeadline))
	fmt.Printf("Обеспечение заявки: %s (%s%%)\n", strOrDash(info.BidSecurityAmount), strOrDash(info.BidSecurityPercent))
	fmt.Printf("Тип обеспечения:    %s\n", strOrDash(info.BidSecurityType))
	if info.ContractSecurityRequired != nil && *info.ContractSecurityRequired {
		fmt.Printf("Обеспечение договора: %s\n", strOrDash(info.ContractSecurityAmount))
	} else {
		fmt.Println("Обеспечение договора: не установлено")
	}
	fmt.Printf("Критерии:           цена %s%%, квалификация %s%%\n",
		strOrDash(info.PriceCriterionWeight), strOrDash(info.QualificationCriterionWeight))
	fmt.Printf("Правовая база:      %s\n", strOrDash(info.LawBasis))
	fmt.Printf("Создано:            %s\n", info.CreatedAt)
	fmt.Println()

	var reqs, docs []*model.DocRequirement
	for _, r := range d.Requirements {
		if r == nil {
			continue
		}
		if r.Kind == "requirement" {
			reqs = append(reqs, r)
		} else {
			docs = append(docs, r)
		}
	}
	sort.Slice(reqs, func(i, j int) bool { return strVal(reqs[i].Code) < strVal(reqs[j].Code) })
	sort.Slice(docs, func(i, j int) bool { return strVal(docs[i].Code) < strVal(docs[j].Code) })

	fmt.Printf("=== Требования к участнику (%d) ===\n", len(reqs))
	if len(reqs) == 0 {
		fmt.Println("(нет)")
	} else {
		for _, r := range reqs {
			fmt.Printf("  %-8s %s\n", strOrDash(r.Code), strOrDash(r.Name))
		}
	}
	fmt.Println()

	fmt.Printf("=== Документы в составе заявки (%d) ===\n", len(docs))
	if len(docs) == 0 {
		fmt.Println("(нет)")
	} else {
		for _, r := range docs {
			mark := "[обяз.]"
			if r.Required != nil && !*r.Required {
				mark = "[опц.] "
			}
			name := strOrDash(r.Name)
			if runes := []rune(name); len(runes) > 80 {
				name = string(runes[:80]) + "…"
			}
			fmt.Printf("  %-8s %s  %s\n", strOrDash(r.Code), mark, name)
		}
	}
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strOrDash(s *string) string {
	if s == nil || *s == "" {
		return "—"
	}
	return strings.TrimSpace(*s)
}
