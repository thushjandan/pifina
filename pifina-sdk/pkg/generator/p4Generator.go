package generator

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/thushjandan/pifina/pkg/model"
)

//go:embed template/*.tpl
var p4CodeTemplate embed.FS

const (
	P4_TEMPLATE_DIR     = "template"
	P4_HEADER_FILE_NAME = "p4Header.tpl"
	P4_APP_FILE_NAME    = "p4App.tpl"
)

func GenerateP4App(templateOptions *model.P4CodeTemplate) error {
	p4HeaderTemplate, err := template.New(P4_HEADER_FILE_NAME).ParseFS(
		p4CodeTemplate,
		filepath.Join(P4_TEMPLATE_DIR, P4_HEADER_FILE_NAME),
	)

	if err != nil {
		return err
	}

	// Parse p4 header template file
	if err = p4HeaderTemplate.Execute(os.Stdout, templateOptions); err != nil {
		return err
	}

	p4AppTemplate, err := template.New(P4_APP_FILE_NAME).ParseFS(
		p4CodeTemplate,
		filepath.Join(P4_TEMPLATE_DIR, P4_APP_FILE_NAME),
	)

	if err != nil {
		return err
	}

	// Parse P4 app template file
	if err = p4AppTemplate.Execute(os.Stdout, templateOptions); err != nil {
		return err
	}

	return nil
}
