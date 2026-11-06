package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reInnKpp      = regexp.MustCompile(`ИНН\s*[/\\]\s*КПП:?\s*(\d{10})\s*[/\\]\s*(\d{9})`)
	reInnOnly     = regexp.MustCompile(`ИНН:?\s*(\d{10})\b`)
	reKppOnly     = regexp.MustCompile(`КПП:?\s*(\d{9})\b`)
	reOGRN        = regexp.MustCompile(`ОГРН[:\s]+(\d{13,15})`)
	reFullName    = regexp.MustCompile(`Общество с ограниченной ответственностью\s+«[^»]+»`)
	reShortName   = regexp.MustCompile(`ООО\s+«[^»]+»`)
	reEmail       = regexp.MustCompile(`[\w.+-]+@[\w.-]+\.\w+`)
	reWebsite     = regexp.MustCompile(`https?://[\w./-]+`)
	reDirectorPos = regexp.MustCompile(`(Генеральный директор|Директор)`)
	reFIO         = regexp.MustCompile(`[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+`)
)

func extractParticipant(inputs []Input) *model.Participant {
	out := &model.Participant{}
	for _, in := range inputs {
		for _, raw := range in.Paragraphs {
			p := stripGarbage(raw)
			fillFromText(out, p)
		}
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				for _, cell := range row.Cells {
					p := stripGarbage(cell)
					fillFromText(out, p)
					if out.Email == nil {
						if m := reEmail.FindString(p); m != "" {
							out.Email = &m
						}
					}
					if out.Website == nil {
						if m := reWebsite.FindString(p); m != "" {
							out.Website = &m
						}
					}
				}
				if len(row.Cells) < 2 {
					continue
				}
				key := strings.TrimSpace(row.Cells[0])
				val := strings.TrimSpace(row.Cells[len(row.Cells)-1])
				switch {
				case strings.Contains(key, "ОКПО"):
					if out.OKPO == nil && val != "" {
						out.OKPO = &val
					}
				case strings.Contains(key, "ОКТМО"):
					if out.OKTMO == nil && val != "" {
						out.OKTMO = &val
					}
				case strings.Contains(key, "ОКОПФ"):
					if out.OKOPF == nil && val != "" {
						out.OKOPF = &val
					}
				case strings.Contains(key, "Организационно-правовая форма"):
					if out.OKOPFName == nil && val != "" {
						out.OKOPFName = &val
					}
				case strings.Contains(key, "Юридический адрес"):
					if out.LegalAddress == nil && val != "" {
						out.LegalAddress = &val
					}
				case strings.Contains(key, "Почтовый адрес"):
					if out.PostalAddress == nil && val != "" {
						out.PostalAddress = &val
					}
				case strings.Contains(key, "Свидетельство о внесении"):
					if out.EGRULRegistrationDate == nil {
						if d := normalizeDotDate(val); d != nil {
							out.EGRULRegistrationDate = d
						}
					}
					if out.EGRULCertificate == nil {
						if m := regexp.MustCompile(`серия\s+\d+\s+№\s*\d+`).FindString(val); m != "" {
							out.EGRULCertificate = &m
						}
					}
				case strings.Contains(key, "Филиалы"):
					b := !strings.Contains(strings.ToLower(val), "нет")
					out.HasBranches = &b
				case strings.Contains(key, "Телефоны Участника"):
					if out.Phone == nil {
						if v := normalizePhone(val); v != nil {
							out.Phone = v
						}
					}
				case strings.Contains(key, "Факс"):
					if out.Fax == nil && !strings.Contains(strings.ToLower(val), "нет") {
						if v := normalizePhone(val); v != nil {
							out.Fax = v
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество руководителя"):
					if out.DirectorName == nil {
						if fio := reFIO.FindString(val); fio != "" {
							out.DirectorName = &fio
						}
					}
					if out.DirectorPosition == nil {
						if m := reDirectorPos.FindString(val); m != "" {
							out.DirectorPosition = &m
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество главного бухгалтера"):
					if out.ChiefAccountantName == nil {
						if fio := reFIO.FindString(val); fio != "" {
							out.ChiefAccountantName = &fio
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество ответственного лица"):
					if out.ResponsiblePersonName == nil {
						if fio := reFIO.FindString(val); fio != "" {
							out.ResponsiblePersonName = &fio
						}
					}
					if out.ResponsiblePersonPhone == nil {
						if v := normalizePhone(val); v != nil {
							out.ResponsiblePersonPhone = v
						}
					}
				}
			}
		}
	}
	if out.INN == nil && out.KPP == nil && out.OGRN == nil && out.FullName == nil {
		return nil
	}
	return out
}

func fillFromText(out *model.Participant, p string) {
	if out.INN == nil || out.KPP == nil {
		if m := reInnKpp.FindStringSubmatch(p); m != nil {
			if out.INN == nil {
				v := m[1]
				out.INN = &v
			}
			if out.KPP == nil {
				v := m[2]
				out.KPP = &v
			}
		} else {
			if out.INN == nil {
				if m := reInnOnly.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.INN = &v
				}
			}
			if out.KPP == nil {
				if m := reKppOnly.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.KPP = &v
				}
			}
		}
	}
	if out.OGRN == nil {
		if m := reOGRN.FindStringSubmatch(p); m != nil {
			if v := normalizeOGRN(m[1]); v != nil {
				out.OGRN = v
			}
		}
	}
	if out.FullName == nil {
		if m := reFullName.FindString(p); m != "" {
			out.FullName = &m
		}
	}
	if out.ShortName == nil {
		if m := reShortName.FindString(p); m != "" {
			out.ShortName = &m
		}
	}
	if out.Email == nil {
		if m := reEmail.FindString(p); m != "" {
			out.Email = &m
		}
	}
	if out.Website == nil {
		if m := reWebsite.FindString(p); m != "" {
			out.Website = &m
		}
	}
}
