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
	P4_TEMPLATE_DIR     = "template"
	P4_HEADER_FILE_NAME = "p4Header.tpl"
	P4_APP_FILE_NAME    = "p4App.tpl"
)

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
