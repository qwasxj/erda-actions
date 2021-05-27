package pkg

import (
	"os"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RepoTools       = "ERDA_TOOLS"
	RepoRelease     = "ERDA_RELEASE"
	RepoVersion     = "ERDA_VERSION"
	RepoToolsPath   = "REPO_TOOLS_PATH"
	RepoReleasePath = "REPO_RELEASE_PATH"
	RepoVersionPath = "REPO_VERSION_PATH"

	OssEndPoint        = "OSS_ENDPOINT"
	OssAccessKeyId     = "OSS_ACCESS_KEY_ID"
	OssAccessKeySecret = "OSS_ACCESS_KEY_SECRET"

	ErdaVersion  = "DICE_VERSION"
	ArchivePatch = "ARCHIVE_PATH"
)

type Env struct{}

func NewEnv() *Env {
	return &Env{}
}

// Set action env
func (e *Env) InitEnv() error {

	// init repo env
	if err := os.Setenv(RepoTools, config.RepoErdaTools()); err != nil {
		return errors.WithMessage(err, "set env of erda-tools repo path")
	}

	if err := os.Setenv(RepoRelease, config.RepoErdaRelease()); err != nil {
		return errors.WithMessage(err, "set env of erda-release repo path")
	}

	if err := os.Setenv(RepoVersion, config.RepoVersion()); err != nil {
		return errors.WithMessage(err, "set env of erda-version repo path")
	}

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

	// init oss env
	if err := os.Setenv(OssEndPoint, config.OssInfo().OssEndPoint); err != nil {
		return errors.WithMessage(err, "set env of oss endpoint")
	}

	if err := os.Setenv(OssAccessKeyId, config.OssInfo().OssAccessKeyId); err != nil {
		return errors.WithMessage(err, "set env of oss ak")
	}

	if err := os.Setenv(OssAccessKeySecret, config.OssInfo().OssAccessKeySecret); err != nil {
		return errors.WithMessage(err, "set env of oss sk")
	}

	// init oss env
	if err := os.Setenv(ArchivePatch, NewOss().OssRemotePath(OssArchiveBucket, OssArchivePath)); err != nil {
		return errors.WithMessage(err, "set env of oss sk")
	}

	return nil
}

func (e *Env) ShowInitEnvs() {
	logrus.Infof("init envs:\n")
	logrus.Infof("%s: %v", RepoTools, os.Getenv(RepoTools))
	logrus.Infof("%s: %v", RepoRelease, os.Getenv(RepoRelease))
	logrus.Infof("%s: %v", RepoVersion, os.Getenv(RepoVersion))
	logrus.Infof("%s: %v", ErdaVersion, os.Getenv(ErdaVersion))
	logrus.Infof("%s: %v", OssEndPoint, os.Getenv(OssEndPoint))
	logrus.Infof("%s: %v", OssAccessKeyId, os.Getenv(OssAccessKeyId))
	logrus.Infof("%s: %v", OssAccessKeySecret, os.Getenv(OssAccessKeySecret))
	logrus.Infof("%s: %v", RepoToolsPath, os.Getenv(RepoToolsPath))
	logrus.Infof("%s: %v", RepoReleasePath, os.Getenv(RepoReleasePath))
	logrus.Infof("%s: %v", RepoVersionPath, os.Getenv(RepoVersionPath))
	logrus.Infof("%s: %v", ArchivePatch, os.Getenv(ArchivePatch))
}
