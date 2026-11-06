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

	reOKPO        = regexp.MustCompile(`ОКПО\s+Участника\s+(\d+)`)
	reOKTMO       = regexp.MustCompile(`ОКТМО\s+Участника\s+(\d+)`)
	reOKOPF       = regexp.MustCompile(`ОКОПФ\s+Участника\s+(\d+)`)
	reOKOPFName   = regexp.MustCompile(`Организационно-правовая форма и фирменное наименование Участника\s+(.+)`)
	reLegalAddr   = regexp.MustCompile(`Юридический адрес\s+(.+)`)
	rePostalAddr  = regexp.MustCompile(`Почтовый адрес\s+(.+)`)
	reEGRULCert   = regexp.MustCompile(`серия\s+(\d+)\s+№\s*(\d+)`)
	reEGRULAuth   = regexp.MustCompile(`Свидетельство о внесении[^,]*,[^,]*,?\s*([А-ЯЁ][^,\n]+)`)
	reEGRULDate   = regexp.MustCompile(`Свидетельство о внесении в Единый государственный реестр[^,]*,\s*(\d{2}\.\d{2}\.\d{4})`)
	reTaxAuth     = regexp.MustCompile(`КПП:\s*\d+[^А-ЯЁ]*([А-ЯЁ][^\n.]+)`)
	reTaxRegDate  = regexp.MustCompile(`Дата постановки на учет[^\d]*(\d{2}\.\d{2}\.\d{4})`)
	reBranches    = regexp.MustCompile(`Филиалы:[^\n]*(нет|да)`)
	rePhoneMarker = regexp.MustCompile(`Телефоны\s+Участника[^\d]*(\+?[\d\s()\-]+)`)
	reFaxMarker   = regexp.MustCompile(`Факс\s+Участника[^\d]*([\d\s()\-]+|нет)`)
	reDirectorFIO = regexp.MustCompile(`Фамилия, Имя и Отчество руководителя[^\n]*?(Генеральный директор|Директор)\s+([А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+)`)
	reChiefFIO    = regexp.MustCompile(`Фамилия, Имя и Отчество главного бухгалтера\s+(.+)`)
	reRespFIO     = regexp.MustCompile(`Фамилия, Имя и Отчество ответственного лица[^\n]*?(Генеральный директор|Директор)\s+([А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+)[^\d]*(\+?[\d\s()\-]+)`)
	reEmailSite   = regexp.MustCompile(`Адрес электронной почты Участника, сайт Участника\s+(.+)`)
	reOKVED       = regexp.MustCompile(`(\d{2}\.\d{2})\s*-\s*(.+)`)
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
					fillFromMergedCell(out, p)
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
						if m := reEGRULCert.FindString(val); m != "" {
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

func fillFromMergedCell(out *model.Participant, p string) {
	if out.OKPO == nil {
		if m := reOKPO.FindStringSubmatch(p); m != nil {
			v := m[1]
			out.OKPO = &v
		}
	}
	if out.OKTMO == nil {
		if m := reOKTMO.FindStringSubmatch(p); m != nil {
			v := m[1]
			out.OKTMO = &v
		}
	}
	if out.OKOPF == nil {
		if m := reOKOPF.FindStringSubmatch(p); m != nil {
			v := m[1]
			out.OKOPF = &v
		}
	}
	if out.OKOPFName == nil {
		if m := reOKOPFName.FindStringSubmatch(p); m != nil {
			v := strings.TrimSpace(m[1])
			if v != "" {
				out.OKOPFName = &v
			}
		}
	}
	if out.LegalAddress == nil {
		if m := reLegalAddr.FindStringSubmatch(p); m != nil {
			v := strings.TrimSpace(m[1])
			if v != "" {
				out.LegalAddress = &v
			}
		}
	}
	if out.PostalAddress == nil {
		if m := rePostalAddr.FindStringSubmatch(p); m != nil {
			v := strings.TrimSpace(m[1])
			if v != "" {
				out.PostalAddress = &v
			}
		}
	}
	if out.EGRULCertificate == nil {
		if m := reEGRULCert.FindStringSubmatch(p); m != nil {
			v := "серия " + m[1] + " № " + m[2]
			out.EGRULCertificate = &v
		}
	}
	if out.EGRULRegistrationDate == nil {
		if m := reEGRULDate.FindStringSubmatch(p); m != nil {
			if d := normalizeDotDate(m[1]); d != nil {
				out.EGRULRegistrationDate = d
			}
		}
	}
	if out.EGRULRegistrationAuthority == nil {
		if m := reEGRULAuth.FindStringSubmatch(p); m != nil {
			v := strings.TrimSpace(m[1])
			if v != "" {
				out.EGRULRegistrationAuthority = &v
			}
		}
	}
	if out.TaxRegistrationDateCurrent == nil {
		if m := reTaxRegDate.FindStringSubmatch(p); m != nil {
			if d := normalizeDotDate(m[1]); d != nil {
				out.TaxRegistrationDateCurrent = d
			}
		}
	}
	if out.TaxAuthorityCurrent == nil {
		if m := reTaxAuth.FindStringSubmatch(p); m != nil {
			v := strings.TrimSpace(m[1])
			if v != "" {
				out.TaxAuthorityCurrent = &v
			}
		}
	}
	if out.HasBranches == nil {
		if m := reBranches.FindStringSubmatch(p); m != nil {
			b := !strings.Contains(strings.ToLower(m[1]), "нет")
			out.HasBranches = &b
		}
	}
	if out.Phone == nil {
		if m := rePhoneMarker.FindStringSubmatch(p); m != nil {
			if v := normalizePhone(m[1]); v != nil {
				out.Phone = v
			}
		}
	}
	if out.Fax == nil {
		if m := reFaxMarker.FindStringSubmatch(p); m != nil {
			if !strings.Contains(strings.ToLower(m[1]), "нет") {
				if v := normalizePhone(m[1]); v != nil {
					out.Fax = v
				}
			}
		}
	}
	if out.Email == nil || out.Website == nil {
		if m := reEmailSite.FindStringSubmatch(p); m != nil {
			if out.Email == nil {
				if e := reEmail.FindString(m[1]); e != "" {
					out.Email = &e
				}
			}
			if out.Website == nil {
				if w := reWebsite.FindString(m[1]); w != "" {
					out.Website = &w
				}
			}
		}
	}
	if out.DirectorName == nil || out.DirectorPosition == nil {
		if m := reDirectorFIO.FindStringSubmatch(p); m != nil {
			if out.DirectorPosition == nil {
				pos := m[1]
				out.DirectorPosition = &pos
			}
			if out.DirectorName == nil {
				fio := m[2]
				out.DirectorName = &fio
			}
		}
	}
	if out.ChiefAccountantName == nil {
		if m := reChiefFIO.FindStringSubmatch(p); m != nil {
			if fio := reFIO.FindString(m[1]); fio != "" {
				out.ChiefAccountantName = &fio
			}
		}
	}
	if out.ResponsiblePersonName == nil || out.ResponsiblePersonPhone == nil {
		if m := reRespFIO.FindStringSubmatch(p); m != nil {
			if out.ResponsiblePersonName == nil {
				fio := m[2]
				out.ResponsiblePersonName = &fio
			}
			if out.ResponsiblePersonPhone == nil {
				if v := normalizePhone(m[3]); v != nil {
					out.ResponsiblePersonPhone = v
				}
			}
		}
	}
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
	if out.PrimaryOKVED == nil || out.PrimaryOKVEDName == nil {
		if m := reOKVED.FindStringSubmatch(p); m != nil {
			if out.PrimaryOKVED == nil {
				v := m[1]
				out.PrimaryOKVED = &v
			}
			if out.PrimaryOKVEDName == nil {
				v := strings.TrimSpace(m[2])
				if v != "" {
					out.PrimaryOKVEDName = &v
				}
			}
		}
	}
}
