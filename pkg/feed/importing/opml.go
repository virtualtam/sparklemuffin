// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import "github.com/virtualtam/opml-go"

const (
	defaultCategoryName string = "Default"
)

func opmlToCategoriesFeeds(outlines []opml.Outline) map[string][]string {
	categories := map[string][]string{}

	var currentCategoryName string

	for _, outline := range outlines {
		currentCategoryName = defaultCategoryName

		if outline.IsDirectory() {
			currentCategoryName = outline.Text

			for _, childOutline := range outline.Outlines {
				if childOutline.OutlineType() == opml.OutlineTypeSubscription {
					categories[currentCategoryName] = append(categories[currentCategoryName], childOutline.XmlUrl)
				}
			}

		} else if outline.OutlineType() == opml.OutlineTypeSubscription {
			categories[defaultCategoryName] = append(categories[defaultCategoryName], outline.XmlUrl)
		}
	}

	return categories
}
