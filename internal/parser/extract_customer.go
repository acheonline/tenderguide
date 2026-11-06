package parser

import (
	"regexp"

	"github.com/acheonline/tenderguide/internal/model"
)

var reCustomerName = regexp.MustCompile(`(АО|ПАО|ООО|ЗАО)\s+«[^»]+»`)

func extractCustomer(req *model.Request) *model.Customer {
	if req == nil || req.ProcedureName == nil {
		return nil
	}
	m := reCustomerName.FindString(*req.ProcedureName)
	if m == "" {
		return nil
	}
	return &model.Customer{Name: &m}
}
