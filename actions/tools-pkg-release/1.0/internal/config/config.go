// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// defines configurations
package config

import (
	"github.com/erda-project/erda-actions/actions/tools-pkg-release/1.0/internal/pkg"
	"github.com/erda-project/erda-actions/pkg/log"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/sirupsen/logrus"
)

// metafile keys
const (
	ERDA_VERSION = "ERDA_VERSION"
	PUBLIC_URL   = "PUBLIC_URL"
	PRIVATE_URL  = "PRIVATE_URL"
)

var c *config

func init() {
	initLog()
}

type config struct {
	// basic envs
	OrgID             uint64 `env:"DICE_ORG_ID" required:"true"`
	CiOpenapiToken    string `env:"DICE_OPENAPI_TOKEN" required:"true"`
	DiceOpenapiPrefix string `env:"DICE_OPENAPI_ADDR" required:"true"`
	ProjectName       string `env:"DICE_PROJECT_NAME" required:"true"`
	AppName           string `env:"DICE_APPLICATION_NAME" required:"true"`
	ProjectID         int64  `env:"DICE_PROJECT_ID" required:"true"`
	AppID             uint64 `env:"DICE_APPLICATION_ID" required:"true"`
	Workspace         string `env:"DICE_WORKSPACE" required:"true"`

	// pipeline parameters
	PipelineDebugMode bool   `env:"PIPELINE_DEBUG_MODE"`
	PipelineID        string `env:"PIPELINE_ID"`
	PipelineTaskLogID string `env:"PIPELINE_TASK_LOG_ID"`
	PipelineTaskID    string `env:"PIPELINE_TASK_ID"`

	// action parameters
	ErdaVersion     string   `env:"ACTION_ERDA_VERSION"`
	RepoErdaTools   string   `env:"ACTION_REPO_ERDA_TOOLS"`
	RepoErdaRelease string   `env:"ACTION_REPO_ERDA_RELEASE"`
	RepoVersion     string   `env:"ACTION_REPO_VERSION"`
	OSS             *pkg.OSS `env:"ACTION_OSS"`

	// other parameters
	MetaFilename string `env:"METAFILE"`
}

type RegistryReplacement struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func configuration() *config {
	if c == nil {
		c = new(config)
		if err := envconf.Load(c); err != nil {
			logrus.Errorf("failed to load configuration, err: %v", err)
		}
	}

	return c
}

func OrgID() uint64 {
	return configuration().OrgID
}

func OpenapiToken() string {
	return configuration().CiOpenapiToken
}

func OpenapiPrefix() string {
	return configuration().DiceOpenapiPrefix
}

func ErdaVersion() string {
	return configuration().ErdaVersion
}

func RepoErdaTools() string {
	return configuration().RepoErdaTools
}

func RepoErdaRelease() string {
	return configuration().RepoErdaRelease
}

func RepoVersion() string {
	return configuration().RepoVersion
}

func OssInfo() *pkg.OSS {

	oss := configuration().OSS

	if oss.OssEndPoint == "" {
		oss.OssEndPoint = "oss.aliyuncs.com"
	}

	return oss
}

func ProjectName() string {
	return configuration().ProjectName
}

func ApplicationID() uint64 {
	return configuration().AppID
}

func ApplicationName() string {
	return configuration().AppName
}

func initLog() {
	log.Init()
	if configuration().PipelineDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
