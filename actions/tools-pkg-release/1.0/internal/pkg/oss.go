package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/config"
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	OssArchiveBucket           = "xj"
	OssArchivePath             = "versions"
	OssPkgReleasePublicBucket  = "xj"
	OssPkgReleasePrivateBucket = "installer"
	OssPkgReleasePath          = "tools_pkg"
)

type OSS struct {
	OssEndPoint        string `json:"endpoint"`
	OssAccessKeyId     string `json:"accessKeyID"`
	OssAccessKeySecret string `json:"accessKeySecret"`
}

func NewOss() *OSS {
	return config.OssInfo()
}

func (o *OSS) OssRemotePath(bucket, path string) string {
	return fmt.Sprintf("oss://%s/%s/", bucket, path)
}

func (o *OSS) ReleaseToolsPackage(private, public string) error {

	// upload public release installing pkg of erda
	if err := o.UploadFile(private, OssPkgReleasePrivateBucket, OssPkgReleasePath); err != nil {
		return err
	}

	// upload enterprise release install pkg of erda
	if err := o.UploadFile(public, OssPkgReleasePublicBucket, OssPkgReleasePath); err != nil {
		return err
	}

	return nil
}

func (o *OSS) UploadFile(local, bucket, path string) error {
	remote := o.OssRemotePath(bucket, path)

	_, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", local, remote)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to %s failed", local, remote))
	}

	logrus.Infof("upload file %s to %s success", local, remote)
	return nil
}

func (o *OSS) InitOssConfig() error {

	logrus.Info("start to init oss config...")
	// current user in action
	u, err := user.Current()
	if err != nil {
		return errors.WithMessage(err, "get current user when init oss config")
	}

	// oss config path
	home := u.HomeDir
	ossConfigPath := path.Join(home, ".ossutilconfig")

	// oss config
	ossConfig := fmt.Sprintf("oss\\n[Credentials]\\nlanguage=CH\\nendpoint=%s\\naccessKeyID="+
		"%s\\naccessKeySecret=%s", o.OssEndPoint, o.OssAccessKeyId, o.OssAccessKeySecret)
	if err := ioutil.WriteFile(ossConfigPath, []byte(ossConfig), 0666); err != nil {
		fmt.Println(err)
		return err
	}
	logrus.Info("init oss config success!!")

	return nil
}
