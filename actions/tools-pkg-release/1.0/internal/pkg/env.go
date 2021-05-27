package pkg

import (
	"os"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/pkg/errors"
)

const (
	RepoToolsPath   = "REPO_TOOLS_PATH"
	RepoReleasePath = "REPO_RELEASE_PATH"
	RepoVersionPath = "REPO_VERSION_PATH"

	ErdaVersion = "DICE_VERSION"
)

type Env struct{}

func NewEnv() *Env {
	return &Env{}
}

// Set action env
func (e *Env) InitEnv() error {

	if err := os.Setenv(RepoToolsPath, GetRepoToolsPath()); err != nil {
		return errors.WithMessage(err, "set env of tools tmp path failed")
	}

	if err := os.Setenv(RepoReleasePath, GetRepoErdaReleasePath()); err != nil {
		return errors.WithMessage(err, "set env of erda release tmp path failed")
	}

	if err := os.Setenv(RepoVersionPath, GetRepoVersionPath()); err != nil {
		return errors.WithMessage(err, "set env of version tmp path failed")
	}

	// init erda version env
	if err := os.Setenv(ErdaVersion, config.ErdaVersion()); err != nil {
		return errors.WithMessage(err, "set env of erda-version")
	}

	return nil
}
