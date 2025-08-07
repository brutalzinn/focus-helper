package actions

import "github.com/brutalzinn/focus-helper/config"

type Action interface {
	Execute(level config.AlertLevel) error
}
