// Package pdf provides PDF manipulation utilities for merging and cleaning PDF files.
//
// Functions:
//   - MergePDFs: Merges multiple PDF files into a single output file.
//     Inputs: slice of PDF file paths, output file path.
//     Output: error if merge fails.
//   - RemoveBookmarks: Removes bookmarks from a PDF file in-place.
//     Input: PDF file path.
//     Output: error if operation fails.
//
// These functions are used by the API handlers to process user-uploaded files.
package pdf

import (
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func MergePDFs(files []string, outputPath string) error {
	config := model.NewDefaultConfiguration()
	return pdfapi.MergeCreateFile(files, outputPath, false, config)
}

func RemoveBookmarks(pdfPath string) error {
	config := model.NewDefaultConfiguration()
	return pdfapi.RemoveBookmarksFile(pdfPath, pdfPath, config)
}
