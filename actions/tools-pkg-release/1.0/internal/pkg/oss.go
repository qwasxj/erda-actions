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
	OssArchiveBucket         = "erda-release"
	OssArchivePath           = "archived-versions"
	OssPkgReleaseBucket      = "erda-release"
	OssPkgReleasePublicPath  = "erda"
	OssPkgReleasePrivatePath = "enterprise"

	OssAclPublicReadWrite = "public-read-write"
	OssAclPublicRead      = "public-read"
	OssAclPrivate         = "private"
	OssAclDefault         = "default"
)

type OSS struct {
	oss *config.OSS
}

func NewOss() *OSS {
	return &OSS{
		config.OssInfo(),
	}
}

func (o *OSS) OssRemotePath(bucket, path string) string {

	return fmt.Sprintf("oss://%s/%s", bucket, path)
}

func (o *OSS) ReleaseToolsPackage(public, private string) error {

	// upload public release installing pkg of erda
	if public != "" {

		if !path.IsAbs(public) {
			return errors.Errorf("public release pkg path is "+
				"not a absolute path: %s", public)
		}

		_, publicPkgName := path.Split(public)
		publicPath := fmt.Sprintf("%s/%s", OssPkgReleasePublicPath, publicPkgName)
		if err := o.UploadFile(public, OssPkgReleaseBucket, publicPath, OssAclPublicRead); err != nil {
			return err
		}
	}

	// upload enterprise release install pkg of erda
	if private != "" {

		if !path.IsAbs(private) {
			return errors.Errorf("private release pkg path is "+
				"not a absolute path: %s", private)
		}

		_, privatePkgName := path.Split(private)
		privatePath := fmt.Sprintf("%s/%s", OssPkgReleasePrivatePath, privatePkgName)
		if err := o.UploadFile(private, OssPkgReleaseBucket, privatePath, OssAclPublicRead); err != nil {
			return err
		}
	}

	return nil
}

func (o *OSS) GenPrivateReleaseUrl(path string) string {
	return fmt.Sprintf("http://%s.%s/%s/dice-tools.%s.tar.gz",
		OssPkgReleaseBucket, o.oss.OssEndPoint, path, config.ErdaVersion())
}

func (o *OSS) GenPublicReleaseUrl(path string) string {
	return fmt.Sprintf("http://%s.%s/%s/erda-release.%s.tar.gz",
		OssPkgReleaseBucket, o.oss.OssEndPoint, path, config.ErdaVersion())
}

func (o *OSS) PreparePatchRelease() error {

	// download release
	releasePath := fmt.Sprintf("%s/%s", OssArchivePath, config.ErdaVersion())
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
		if _, err := utils.ExecCmd(os.Stdout, os.Stderr, fmt.Sprintf("/tmp/%s/extensions", config.ErdaVersion()),
			"tar", "-zxvf", tar); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("decompress %s failed", tar))
		}
	}

	return nil
}

func (o *OSS) UploadFile(local, bucket, path, acl string) error {

	exists, err := utils.FileExist(local)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to oss", local))
	}
	if !exists {
		return fmt.Errorf("the file %s waited to upload is not exists", local)
	}

	remote := o.OssRemotePath(bucket, path)

	if acl == "" {
		_, err = utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-f", local, remote)
	} else {
		_, err = utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64",
			"cp", "-f", fmt.Sprintf("--acl=%s", acl), local, remote)
	}
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to %s failed", local, remote))
	}

	logrus.Infof("upload file %s to %s success", local, remote)
	return nil
}

func (o *OSS) UploadDir(dir, bucket, path, acl string) error {

	exists, err := utils.IsDirExists(dir)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload dir %s to oss", dir))
	}
	if !exists {
		return fmt.Errorf("the dir %s waited to upload is not exists", dir)
	}

	remote := o.OssRemotePath(bucket, path)

	if acl == "" {
		_, err = utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf", dir, path)
	} else {
		_, err = utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf",
			fmt.Sprintf("--acl=%s", acl), dir, path)
	}
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("upload file %s to %s failed", dir, remote))
	}

	logrus.Infof("upload file %s to %s success", dir, remote)
	return nil
}

func (o *OSS) DownloadFile(local, bucket, path string) error {
	remote := o.OssRemotePath(bucket, path)

	_, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-f", remote, local)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("download file %s to %s failed", remote, local))
	}

	logrus.Infof("download file %s to %s success", remote, local)
	return nil
}

func (o *OSS) DownloadDir(parent, bucket, path string) error {
	remote := o.OssRemotePath(bucket, path)

	_, err := utils.ExecCmd(os.Stdout, os.Stderr, "", "ossutil64", "cp", "-rf", remote, parent)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("download dir %s from to %s failed", remote, parent))
	}

	logrus.Infof("download dir %s to %s success", remote, parent)
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
	ossConfig := fmt.Sprintf("[Credentials]\nlanguage=CH\nendpoint=%s\naccessKeyID="+
		"%s\naccessKeySecret=%s", o.oss.OssEndPoint, o.oss.OssAccessKeyId, o.oss.OssAccessKeySecret)
	if err := ioutil.WriteFile(ossConfigPath, []byte(ossConfig), 0666); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Info("init oss config success!!")

	return nil
}
