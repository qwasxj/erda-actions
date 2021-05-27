package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/utils"
	"github.com/erda-project/erda/apistructs"
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

	// write metafile
	WriteMetaFile()

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
		//PUBLIC: path.Join(GetRepoErdaReleasePath(), fmt.Sprintf(
		//	"dice-tools.%s.tar.gz", config.ErdaVersion())),
		PUBLIC: "",
		PRIVATE: path.Join(GetRepoToolsPath(), fmt.Sprintf(
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

func WriteMetaFile() {
	oss := NewOss()

	// write metafile
	metaInfos := make([]apistructs.MetadataField, 0, 1)
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaErdaVersion,
		Value: config.ErdaVersion(),
	})
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaPublicUrl,
		Value: oss.GenReleaseUrl(OssPkgReleasePublicPath),
	})
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  config.MetaPrivateUrl,
		Value: oss.GenReleaseUrl(OssPkgReleasePrivatePath),
	})

	metaByte, _ := json.Marshal(apistructs.ActionCallback{Metadata: metaInfos})
	if err := filehelper.CreateFile(config.MetaFile(), string(metaByte), 0644); err != nil {
		logrus.Warnf("failed to write metafile, %v", err)
	}
}
