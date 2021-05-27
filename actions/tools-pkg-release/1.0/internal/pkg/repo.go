package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/bin"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Repo struct{}

func NewRepo() *Repo {
	return &Repo{}
}

func (r *Repo) PrepareRepo() error {

	// source repo exists validate
	exists, err := utils.IsDirExists(config.RepoVersion())
	if err != nil {
		return errors.WithMessage(err, "stats repo version stat")
	}
	if !exists {
		return fmt.Errorf("repo version does not exists")
	}

	exists, err = utils.IsDirExists(config.RepoErdaTools())
	if err != nil {
		return errors.WithMessage(err, "stats-repo erda tools stat")
	}
	if !exists {
		return fmt.Errorf("repo erda-tools does not exists")
	}

	exists, err = utils.IsDirExists(config.RepoErdaRelease())
	if err != nil {
		return errors.WithMessage(err, "stats repo erda-release stat")
	}
	if !exists {
		return fmt.Errorf("repo erda-release does not exists")
	}

	// cp repo
	if _, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "cp",
		"-a", config.RepoVersion(), "/tmp/"); err != nil {
		return errors.WithMessage(err, "cp repo version to /tmp/")
	}

	if _, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "cp",
		"-a", config.RepoErdaTools(), "/tmp/"); err != nil {
		return errors.WithMessage(err, "cp repo tools to /tmp/")
	}

	if _, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "cp",
		"-a", config.RepoErdaRelease(), "/tmp/"); err != nil {
		return errors.WithMessage(err, "cp repo erda release to /tmp/")
	}

	// erda-tools execute script prepare
	if err := ReplaceBuildScript(); err != nil {
		return errors.WithMessage(err, "replace build script in tools")
	}

	return nil
}

func ReplaceBuildScript() error {

	logrus.Info("start to replace oss build script...")

	// build script in tools
	if err := ioutil.WriteFile(path.Join(GetRepoToolsPath(), "build"),
		[]byte(bin.PrivateExecuteScript), 0666); err != nil {
		fmt.Println(err)
		return err
	}

	logrus.Info("replace oss build script success!!")

	return nil
}

func RepoToolsName() string {
	_, name := path.Split(config.RepoErdaTools())

	return name
}

func RepoVersionName() string {
	_, name := path.Split(config.RepoVersion())

	return name
}

func RepoErdaReleaseName() string {
	_, name := path.Split(config.RepoErdaRelease())

	return name
}

func GetRepoToolsPath() string {
	return path.Join("/tmp", RepoToolsName())
}

func GetRepoVersionPath() string {
	return path.Join("/tmp", RepoVersionName())
}

func GetRepoErdaReleasePath() string {
	return path.Join("/tmp", RepoErdaReleaseName())
}
