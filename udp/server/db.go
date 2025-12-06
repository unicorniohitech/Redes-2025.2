package server

type Dictionary struct {
	terms map[string]string
	keys  []string
}

func NewDictionary() *Dictionary {
	return &Dictionary{
		terms: make(map[string]string),
		keys:  []string{},
	}
}

func (d *Dictionary) List() []string {
	return d.keys
}

func (d *Dictionary) LookUp(term string) (string, bool) {
	definition, exists := d.terms[term]
	return definition, exists
}

func (d *Dictionary) Insert(term, definition string) bool {
	if _, exists := d.terms[term]; exists {
		return false
	}
	d.terms[term] = definition
	d.keys = append(d.keys, term)
	return true
}

func (d *Dictionary) Update(term, newDefinition string) bool {
	if _, exists := d.terms[term]; !exists {
		return false
	}
	d.terms[term] = newDefinition
	return true
}
