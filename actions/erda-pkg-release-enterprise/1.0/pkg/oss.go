package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	OssArchiveBucket         = "erda-release"
	OssArchivePath           = "archived-versions"
	OssPkgReleaseBucket      = "erda-release"
	OssPkgReleasePublicPath  = "erda"
	OssPkgReleasePrivatePath = "enterprise"

	//OssAclPublicReadWrite = "public-read-write"
	OssAclPublicRead = "public-read"
	//OssAclPrivate         = "private"
	//OssAclDefault         = "default"
)

type Oss struct {
	OssEndPoint        string `json:"endpoint"`
	OssAccessKeyId     string `json:"accessKeyID"`
	OssAccessKeySecret string `json:"accessKeySecret"`
}

type oss struct {
	oss *Oss
}

func NewOss(o *Oss) *oss {
	return &oss{
		o,
	}
}

func (o *oss) GetOss() *Oss {
	return o.oss
}

func (o *oss) GenReleaseUrl(path string) string {
	return fmt.Sprintf("http://%s.%s/%s",
		OssPkgReleaseBucket, o.oss.OssEndPoint, path)
}

func (o *oss) OssRemotePath(bucket, path string) string {

	return fmt.Sprintf("oss://%s/%s", bucket, path)
}

func (o *oss) ReleasePackage(releasePathInfo map[string]string, releaseBucket,
	releasePath string, splitOsArch bool) error {

	// upload release installing pkg of erda
	if len(releasePathInfo) != 0 {

		for osArch, pkgPath := range releasePathInfo {
			if !path.IsAbs(pkgPath) {
				return errors.Errorf("release pkg path is "+
					"not a absolute path: %s", pkgPath)
			}

			_, pkgName := path.Split(pkgPath)

			ossReleasePath := ""

			// 发布包管理策略
			if splitOsArch {
				ossReleasePath = fmt.Sprintf("%s/%s/%s", releasePath, osArch, pkgName)
			} else {
				ossReleasePath = fmt.Sprintf("%s/%s", releasePath, pkgName)
			}

			if err := o.UploadFile(pkgPath, releaseBucket, ossReleasePath, OssAclPublicRead); err != nil {
				return err
			}
		}

	}

	return nil
}

func (o *oss) PreparePatchRelease(version string) error {

	// download release
	releasePath := fmt.Sprintf("%s/%s", OssArchivePath, version)
	if err := o.DownloadDir("/tmp", OssArchiveBucket, releasePath); err != nil {
		return errors.WithMessage(err, "cp release patch to /tmp/")
	}

	tars := []string{
		"erda-actions-enterprise.tar.gz",
		"erda-actions.tar.gz",
		"erda-addons-enterprise.tar.gz",
		"erda-addons.tar.gz",
	}

	// tar release
	for _, tar := range tars {
		if _, err := ExecCmd(os.Stdout, os.Stderr, fmt.Sprintf("/tmp/%s/extensions", version),
			"tar", "-zxvf", tar); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("decompress %s failed", tar))
		}
	}

	return nil
}

func (o *oss) UploadFile(local, bucket, path, acl string) error {

	exists, err := FileExist(local)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to oss", local))
	}
	if !exists {
		return fmt.Errorf("the file %s waited to upload is not exists", local)
	}

	remote := o.OssRemotePath(bucket, path)

	if acl == "" {
		_, err = ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-f", local, remote)
	} else {
		_, err = ExecCmd(os.Stdout, os.Stderr, "", "ossutil64",
			"cp", "-f", fmt.Sprintf("--acl=%s", acl), local, remote)
	}
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to %s failed", local, remote))
	}

	logrus.Infof("upload file %s to %s success", local, remote)
	return nil
}

func (o *oss) UploadDir(dir, bucket, path, acl string) error {

	exists, err := IsDirExists(dir)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload dir %s to oss", dir))
	}
	if !exists {
		return fmt.Errorf("the dir %s waited to upload is not exists", dir)
	}

	remote := o.OssRemotePath(bucket, path)

	if acl == "" {
		_, err = ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf", dir, path)
	} else {
		_, err = ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf",
			fmt.Sprintf("--acl=%s", acl), dir, path)
	}
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to %s failed", dir, remote))
	}

	logrus.Infof("upload file %s to %s success", dir, remote)
	return nil
}

func (o *oss) DownloadFile(local, bucket, path string) error {
	remote := o.OssRemotePath(bucket, path)

	_, err := ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-f", remote, local)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("download file %s to %s failed", remote, local))
	}

	logrus.Infof("download file %s to %s success", remote, local)
	return nil
}

func (o *oss) DownloadDir(parent, bucket, path string) error {
	remote := o.OssRemotePath(bucket, path)

	_, err := ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf", remote, parent)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("download dir %s from to %s failed", remote, parent))
	}

	logrus.Infof("download dir %s to %s success", remote, parent)
	return nil
}

func (o *oss) InitOssConfig() error {

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
	ossConfig := fmt.Sprintf("[Credentials]\nlanguage=CH\nendpoint=%s\naccessKeyID="+
		"%s\naccessKeySecret=%s", o.oss.OssEndPoint, o.oss.OssAccessKeyId, o.oss.OssAccessKeySecret)
	if err := ioutil.WriteFile(ossConfigPath, []byte(ossConfig), 0666); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Info("init oss config success!!")

	return nil
}
