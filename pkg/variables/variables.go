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
	// p.resolvers["username"] = getUsername
	// p.resolvers["level"] = getHyperfocusLevel
	// p.resolvers["date"] = formatDateToWords
	// p.resolvers["datetime"] = formatDateTime
	// p.resolvers["activity_duration"] = getActivityDuration
	// p.resolvers["person"] = getPersonName
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

// func getUsername(ctx ...string) string           { return "Alex" }
// func getHyperfocusLevel(ctx ...string) string    { return "87%" }
// func getActivityDuration(ctx ...string) string   { return "42 minutes" }
// func getPersonName(ctx ...string) string {
// 	if len(ctx) > 0 {
// 		return strings.Title(strings.Replace(ctx[0], "_", " ", -1))
// 	}
// 	return "System"
// }
// func formatDateTime(ctx ...string) string {
// 	loc, _ := time.LoadLocation("America/Sao_Paulo")
// 	return time.Now().In(loc).Format("3:04 PM")
// }
// func formatDateToWords(ctx ...string) string {
// 	now := time.Now()
// 	monthsPtBr := []string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
// 	return fmt.Sprintf("%d de %s de %d", now.Day(), monthsPtBr[now.Month()], now.Year())
// }
