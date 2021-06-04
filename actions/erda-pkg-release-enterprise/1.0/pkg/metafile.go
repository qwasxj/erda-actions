package pkg

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// metafile keys
const (
	MetaErdaVersion = "erdaVersion"
	erdaPkgMapUrl   = "pkgMapUrl"

	// json string
	MetaReleaseInfoType = "ToolsPkgReleaseInfo"
)

// WriteMetaFile write metafile
func WriteMetaFile(o *Oss, metafile string, releaseInfo map[string]string,
	erdaVersion, releasePath string, splitOsArch bool) error {

	logrus.Infof("start to write metafile")

	metaInfos := make([]apistructs.MetadataField, 0, 1)

	oss := NewOss(o)
	urlInfo := map[string]string{}

	// 关联发版策略管理
	for osArch, pkg := range releaseInfo {
		if splitOsArch {
			urlInfo[osArch] = oss.GenReleaseUrl(fmt.Sprintf("%s/%s/%s", releasePath, osArch, pkg))
		} else {
			urlInfo[osArch] = oss.GenReleaseUrl(fmt.Sprintf("%s/%s", releasePath, pkg))
		}
	}

	// serialize url info
	vJson, err := json.Marshal(urlInfo)
	if err != nil {
		return errors.WithMessage(err, "change release pkg url info to json string")
	}

	sJson := string(vJson)

	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  MetaErdaVersion,
		Value: erdaVersion,
	})
	metaInfos = append(metaInfos, apistructs.MetadataField{
		Name:  erdaPkgMapUrl,
		Type:  MetaReleaseInfoType,
		Value: sJson,
	})

	metaByte, _ := json.Marshal(apistructs.ActionCallback{Metadata: metaInfos})
	if err := filehelper.CreateFile(metafile, string(metaByte), 0644); err != nil {
		logrus.Warnf("failed to write metafile, %v", err)
	}

	logrus.Infof("write metafile success...")

	return nil
}
