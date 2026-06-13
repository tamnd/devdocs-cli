package cli

import (
	"errors"

	"github.com/tamnd/devdocs-cli/devdocs"
)

func isNotFound(err error) bool {
	return errors.Is(err, devdocs.ErrNotFound)
}
