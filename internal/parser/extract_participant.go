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
	reOKVED       = regexp.MustCompile(`^(\d{2}\.\d{2})\s*-\s*(.+)$`)
)

func extractParticipant(inputs []Input) *model.Participant {
	out := &model.Participant{}
	for _, in := range inputs {
		for _, raw := range in.Paragraphs {
			fillFromText(out, stripGarbage(raw))
		}
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				if len(row.Cells) < 2 {
					continue
				}
				key := strings.TrimSpace(row.Cells[1])
				val := strings.TrimSpace(row.Cells[len(row.Cells)-1])
				if key == "" {
					continue
				}
				valClean := stripGarbage(val)
				allEmpty := true
				for _, c := range row.Cells {
					if strings.TrimSpace(c) != "" {
						allEmpty = false
						break
					}
				}
				if allEmpty {
					continue
				}
				fillFromText(out, valClean)

				switch {
				case strings.Contains(key, "ОКПО"):
					if out.OKPO == nil {
						if v := onlyDigits(val); v != "" {
							out.OKPO = &v
						}
					}
				case strings.Contains(key, "ОКТМО"):
					if out.OKTMO == nil {
						if v := onlyDigits(val); v != "" {
							out.OKTMO = &v
						}
					}
				case strings.Contains(key, "ОКОПФ"):
					if out.OKOPF == nil {
						if v := onlyDigits(val); v != "" {
							out.OKOPF = &v
						}
					}
				case strings.Contains(key, "Организационно-правовая форма"):
					if out.OKOPFName == nil && val != "" {
						v := val
						out.OKOPFName = &v
					}
				case strings.Contains(key, "Юридический адрес"):
					if out.LegalAddress == nil && valClean != "" {
						v := valClean
						out.LegalAddress = &v
					}
				case strings.Contains(key, "Почтовый адрес"):
					if out.PostalAddress == nil && valClean != "" {
						v := valClean
						out.PostalAddress = &v
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
					if out.EGRULRegistrationAuthority == nil {
						parts := strings.Split(val, ",")
						if len(parts) >= 3 {
							auth := strings.TrimSpace(parts[len(parts)-1])
							if auth != "" {
								out.EGRULRegistrationAuthority = &auth
							}
						}
					}
				case strings.Contains(key, "Филиалы"):
					b := !strings.Contains(strings.ToLower(valClean), "нет")
					out.HasBranches = &b
				case strings.Contains(key, "Телефоны Участника"):
					if out.Phone == nil {
						if v := normalizePhone(valClean); v != nil {
							out.Phone = v
						}
					}
				case strings.Contains(key, "Факс"):
					if out.Fax == nil && !strings.Contains(strings.ToLower(valClean), "нет") {
						if v := normalizePhone(valClean); v != nil {
							out.Fax = v
						}
					}
				case strings.Contains(key, "Адрес электронной почты"):
					// Сначала по сырому val (там mailto: и http://).
					if out.Email == nil {
						if m := reEmail.FindString(val); m != "" {
							out.Email = &m
						}
					}
					if out.Website == nil {
						if m := reWebsite.FindString(val); m != "" {
							out.Website = &m
						}
					}
					// Fallback по valClean, если stripGarbage оставил что-то.
					if out.Email == nil {
						if m := reEmail.FindString(valClean); m != "" {
							out.Email = &m
						}
					}
					if out.Website == nil {
						if m := reWebsite.FindString(valClean); m != "" {
							out.Website = &m
						}
					}
				case strings.Contains(key, "Дата постановки на учет"):
					if out.TaxRegistrationDateCurrent == nil {
						dates := reDotDate.FindAllString(val, -1)
						if len(dates) > 0 {
							out.TaxRegistrationDateCurrent = normalizeDotDate(dates[len(dates)-1])
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество руководителя"):
					if out.DirectorName == nil {
						if fio := reFIO.FindString(valClean); fio != "" {
							out.DirectorName = &fio
						}
					}
					if out.DirectorPosition == nil {
						if m := reDirectorPos.FindString(valClean); m != "" {
							out.DirectorPosition = &m
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество главного бухгалтера"):
					if out.ChiefAccountantName == nil {
						if fio := reFIO.FindString(valClean); fio != "" {
							out.ChiefAccountantName = &fio
						}
					}
				case strings.Contains(key, "Сведения о производимых субъектами малого и среднего предпринимательства товарах"):
					if out.PrimaryOKVED == nil && len(row.Cells) >= 3 {
						v := strings.TrimSpace(row.Cells[2])
						if m := regexp.MustCompile(`^(\d{2}\.\d{2})\s*-\s*(.+)$`).FindStringSubmatch(v); m != nil {
							code := m[1]
							name := strings.TrimSpace(m[2])
							out.PrimaryOKVED = &code
							if name != "" {
								out.PrimaryOKVEDName = &name
							}
						}
					}
				case strings.Contains(key, "Фамилия, Имя и Отчество ответственного лица"):
					if out.ResponsiblePersonName == nil {
						if fio := reFIO.FindString(valClean); fio != "" {
							out.ResponsiblePersonName = &fio
						}
					}
					if out.ResponsiblePersonPhone == nil {
						if v := normalizePhone(valClean); v != nil {
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
