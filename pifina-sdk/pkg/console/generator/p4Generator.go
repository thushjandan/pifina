// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package generator

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
)

//go:embed template/*.tpl
var p4CodeTemplate embed.FS

const (
	P4_TEMPLATE_DIR       = "template"
	P4_HEADER_FILE_NAME   = "p4Header.tpl"
	P4_APP_FILE_NAME      = "p4App.tpl"
	P4_SKELETON_FILE_NAME = "p4SkeletonApp.tpl"
)

// Generate a basic skeleton P4 file in a new folder
func GenerateSkeleton(logger hclog.Logger, templateOptions *model.P4CodeTemplate, outputDir string) error {
	// Create new app folder
	pfSkeletonFolderPath := filepath.Join(outputDir, "myp4app_with_pifina")
	if err := os.Mkdir(pfSkeletonFolderPath, os.ModePerm); err != nil {
		return err
	}

	p4SkeletonTemplate, err := template.New(P4_SKELETON_FILE_NAME).ParseFS(
		p4CodeTemplate,
		filepath.Join(P4_TEMPLATE_DIR, P4_SKELETON_FILE_NAME),
	)

	if err != nil {
		return err
	}

	pfSkeletonFilePath := filepath.Join(pfSkeletonFolderPath, "myp4app_with_pifina.p4")
	pfSkeletonFileHandle, err := os.Create(pfSkeletonFilePath)
	if err != nil {
		return err
	}
	defer pfSkeletonFileHandle.Close()

	// Parse p4 header template file
	if err = p4SkeletonTemplate.Execute(pfSkeletonFileHandle, templateOptions); err != nil {
		return err
	}

	logger.Info("P4 skeleton file has been successfully generated!", "file", pfSkeletonFilePath)

	// Create new includes folder
	includesFolderPath := filepath.Join(pfSkeletonFolderPath, "include")
	if err := os.Mkdir(includesFolderPath, os.ModePerm); err != nil {
		return err
	}
	// Generate Pifina probe files
	return GenerateP4App(logger, templateOptions, includesFolderPath)
}

func GenerateP4App(logger hclog.Logger, templateOptions *model.P4CodeTemplate, outputDir string) error {
	p4HeaderTemplate, err := template.New(P4_HEADER_FILE_NAME).ParseFS(
		p4CodeTemplate,
		filepath.Join(P4_TEMPLATE_DIR, P4_HEADER_FILE_NAME),
	)

	if err != nil {
		return err
	}

	pfHeaderFilePath := filepath.Join(outputDir, "pifina_headers.p4")
	pfHeaderFileHandle, err := os.Create(pfHeaderFilePath)
	if err != nil {
		return err
	}
	defer pfHeaderFileHandle.Close()

	// Parse p4 header template file
	if err = p4HeaderTemplate.Execute(pfHeaderFileHandle, templateOptions); err != nil {
		return err
	}
	logger.Info("P4 header file has been successfully generated!", "file", pfHeaderFilePath)

	pfAppFilePath := filepath.Join(outputDir, "pifina_probes.p4")
	pfAppFileHandle, err := os.Create(pfAppFilePath)
	if err != nil {
		return err
	}
	defer pfAppFileHandle.Close()

	p4AppTemplate, err := template.New(P4_APP_FILE_NAME).ParseFS(
		p4CodeTemplate,
		filepath.Join(P4_TEMPLATE_DIR, P4_APP_FILE_NAME),
	)

	if err != nil {
		return err
	}

	// Parse P4 app template file
	if err = p4AppTemplate.Execute(pfAppFileHandle, templateOptions); err != nil {
		return err
	}
	logger.Info("P4 probe file has been successfully generated!", "file", pfAppFilePath)

	return nil
}
