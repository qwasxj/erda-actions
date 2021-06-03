package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/utils"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	// prepare patch release of some version specified by erda_version
	if err := oss.PreparePatchRelease(); err != nil {
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

	// tool-pack execute
	releaseInfo, err := ToolsRelease()
	if err != nil {
		return err
	}

	// upload release install pkg of erda
	if err := oss.ReleaseToolsPackage(releaseInfo[PUBLIC],
		releaseInfo[PRIVATE]); err != nil {
		return err
	}

	for i, _ := range make([]int, 10000) {
		time.Sleep(time.Second * 3)
		logrus.Println(i)
	}

	// write metafile
	WriteMetaFile()

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
		//PUBLIC: path.Join(GetRepoErdaReleasePath(), fmt.Sprintf(
		//	"erda-release.%s.tar.gz", config.ErdaVersion())),
		PUBLIC: "",
		PRIVATE: path.Join(GetRepoToolsPath(), fmt.Sprintf(
			"dice-tools.%s.tar.gz", config.ErdaVersion())),
	}

	return pkgInfo, nil
}

func EnterprisePkgRelease() error {

	// replace build script
	_, err := utils.ExecCmd(os.Stdout, os.Stderr, GetRepoToolsPath(), "bash", "-x", "build", "pack")
	if err != nil {
		return errors.WithMessage(err, "build enterprise erda install package")
	}

	// compress enterprise install package of erda
	_, err = utils.ExecCmd(os.Stdout, os.Stderr, GetRepoToolsPath(), "make", "tar")
	if err != nil {
		return errors.WithMessage(err, "build enterprise erda install package")
	}

	return nil
}

func PublicPkgRelease() error {
	return nil
}

func WriteMetaFile() {
	oss := NewOss()

	logrus.Infof("start to write metafile")

	// write metafile
	metaInfos := make([]apistructs.MetadataField, 0, 1)
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaErdaVersion,
		Value: config.ErdaVersion(),
	})
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaPrivateUrl,
		Value: oss.GenPrivateReleaseUrl(OssPkgReleasePrivatePath),
	})
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaPublicUrl,
		Value: oss.GenPublicReleaseUrl(OssPkgReleasePublicPath),
	})

	metaByte, _ := json.Marshal(apistructs.ActionCallback{Metadata: metaInfos})
	if err := filehelper.CreateFile(config.MetaFile(), string(metaByte), 0644); err != nil {
		logrus.Warnf("failed to write metafile, %v", err)
	}

	logrus.Infof("write metafile success...")
}
