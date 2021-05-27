package pkg

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/utils"
	"github.com/pkg/errors"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
)

func Execute() error {

	// oss init
	oss := NewOss()
	if err := oss.InitOssConfig(); err != nil {
		return err
	}

	// prepare repo to use
	repo := NewRepo()
	if err := repo.PrepareRepo(); err != nil {
		return err
	}

	// set init env
	var env = NewEnv()
	if err := env.InitEnv(); err != nil {
		return err
	}

	// env print
	env.ShowInitEnvs()

	// tool-pack execute
	releaseInfo, _ := ToolsRelease()
	//if err != nil {
	//	return err
	//}

	fmt.Println(releaseInfo)

	// wait
	for _, i := range make([]int, 10000) {
		fmt.Println(i)
		time.Sleep(time.Second * 3)
	}
	return nil
}

// ToolsRelease to build some erda installing package with
// some version specified by ERDA_VERSION
func ToolsRelease() (map[string]string, error) {

	if err := EnterprisePkgRelease(); err != nil {
		return nil, err
	}

	if err := PublicPkgRelease(); err != nil {
		return nil, err
	}

	pkgInfo := map[string]string{
		PUBLIC: path.Join(GetRepoToolsPath(), fmt.Sprintf(
			"dice-tools.%s.tar.gz", config.ErdaVersion())),
		PRIVATE: path.Join(GetRepoErdaReleasePath(), fmt.Sprintf(
			"erda-release.%s.tar.gz", config.ErdaVersion())),
	}

	return pkgInfo, nil
}

func EnterprisePkgRelease() error {

	// replace build script
	_, err := utils.ExecCmd(os.Stdout, os.Stderr, GetRepoToolsPath(), "bash", "-x", "build", "pack")
	if err != nil {
		return errors.WithMessage(err, "build enterprise erda install package")
	}

	return nil
}

func PublicPkgRelease() error {
	return nil
}
