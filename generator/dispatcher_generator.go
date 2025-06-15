package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/om3kk/slacky/model"
)

// GenerateDispatcher creates the dispatcher file that routes view submissions.
func GenerateDispatcher(
	structs []model.StructInfo,
	simpleActionIDs []string,
	tmplPath string,
	outputDir string,
	pkgName string,
) error {
	if len(structs) == 0 && len(simpleActionIDs) == 0 {
		log.Println("No modals or simple actions found, skipping dispatcher creation.")
		return nil
	}

	data := model.DispatcherTemplateData{
		PackageName:     pkgName,
		Structs:         structs,
		SimpleActionIDs: simpleActionIDs,
	}

	// 1. Read and parse the template file
	tmpl, err := template.New(filepath.Base(tmplPath)).Funcs(helperFunc).ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse dispatcher template file '%s': %w", tmplPath, err)
	}

	// 2. Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute dispatcher template: %w", err)
	}

	// 3. Format the generated Go code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("ERROR: Failed to format dispatcher code: %v. The unformatted code will be saved.", err)
		formattedCode = buf.Bytes() // Use unformatted code on error
	}

	// 4. Write the generated code to the dispatcher.go file
	outputFilePath := filepath.Join(outputDir, "dispatcher.go")
	if err := os.WriteFile(outputFilePath, formattedCode, 0o644); err != nil {
		return fmt.Errorf("failed to write dispatcher file '%s': %w", outputFilePath, err)
	}

	fmt.Printf("Successfully generated dispatcher: %s\n", outputFilePath)
	return nil
}
