package internal

type CodeMap map[string]string

func (m CodeMap) GetName(key string) string {
	val, ok := m[key]

	if !ok {
		return ""
	}

	return val
}

type AdminCodeNames struct {
	Admin1    *CodeMap
	Admin2    *CodeMap
	Countries *CodeMap
}

func (n *AdminCodeNames) ExpandPlace(p *Place) {
	p.AdminName1 = n.Admin1.GetName(p.Admin1Key())
	p.AdminName2 = n.Admin2.GetName(p.Admin2Key())
	p.Country = n.Countries.GetName(p.CountryCode)
}
