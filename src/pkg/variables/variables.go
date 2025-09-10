package variables

import (
	"regexp"
	"strings"
)

type ResolverFunc func(context ...string) string

type Processor struct {
	resolvers map[string]ResolverFunc
	regex     *regexp.Regexp
}

func NewProcessor() *Processor {
	p := &Processor{
		resolvers: make(map[string]ResolverFunc),
		regex:     regexp.MustCompile(`%(\w+)%`),
	}
	return p
}

func (p *Processor) RegisterHandler(key string, callback ResolverFunc) {
	p.resolvers[key] = callback
}

func (p *Processor) Process(template string, personaName string) string {
	return p.regex.ReplaceAllStringFunc(template, func(match string) string {
		varName := strings.Trim(match, "%")
		if resolver, exists := p.resolvers[varName]; exists {
			if varName == "person" {
				return resolver(personaName)
			}
			return resolver()
		}
		return match
	})
}
