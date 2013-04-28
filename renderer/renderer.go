// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"bufio"
	"fmt"
	"github.com/andreaskoch/allmark/config"
	"github.com/andreaskoch/allmark/mapper"
	"github.com/andreaskoch/allmark/path"
	"github.com/andreaskoch/allmark/repository"
	"github.com/andreaskoch/allmark/templates"
	"github.com/andreaskoch/allmark/view"
	"github.com/andreaskoch/allmark/watcher"
	"os"
	"text/template"
)

type Renderer struct {
	repositoryPath string
	pathProvider   *path.Provider
	config         *config.Config
}

func New(repositoryPath string, config *config.Config, useTempDir bool) *Renderer {

	return &Renderer{
		repositoryPath: repositoryPath,
		pathProvider:   path.NewProvider(repositoryPath, useTempDir),
		config:         config,
	}

}

func (renderer *Renderer) Execute() *repository.ItemIndex {
	itemIndex, err := repository.NewItemIndex(renderer.repositoryPath)
	if err != nil {
		fmt.Printf("Cannot create an item index for folder %q. Error: %v", renderer.repositoryPath, err)
	}

	for _, item := range itemIndex.Items() {
		renderer.renderItem(item)
	}

	return itemIndex
}

func (renderer *Renderer) renderItem(item *repository.Item) {

	// render child items first
	for _, child := range item.Childs() {

		// attach change listener
		child.OnChange("Throw Item Events on Child Item change", func(event *watcher.WatchEvent) {
			item.Throw(event)
		})

		renderer.renderItem(child)
	}

	// attach change listener
	item.OnChange("Render item on change", func(event *watcher.WatchEvent) {
		renderer.renderItem(item)
	})

	// create the viewmodel
	viewModel := mapper.Map(item, renderer.pathProvider)

	// get a template
	templateText := templates.GetTemplate(viewModel.Type)

	// render the template
	targetPath := renderer.pathProvider.GetRenderTargetPath(item)
	renderer.writeOutput(viewModel, templateText, targetPath)
}

func (renderer *Renderer) writeOutput(viewModel view.Model, templateText string, targetPath string) {
	file, err := os.Create(targetPath)
	if err != nil {
		fmt.Errorf("%s", err)
	}

	writer := bufio.NewWriter(file)

	defer func() {
		writer.Flush()
		file.Close()
	}()

	template := template.New(viewModel.Type)
	template.Parse(templateText)
	template.Execute(writer, viewModel)
}
