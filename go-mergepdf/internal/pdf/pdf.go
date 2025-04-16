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
	"fmt"
	"io"
	"os"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func MergePDFs(files []string, outputPath string) error {
	config := model.NewDefaultConfiguration()
	return pdfapi.MergeCreateFile(files, outputPath, false, config)
}

func RemoveBookmarks(pdfPath string) error {
	config := model.NewDefaultConfiguration()
	return pdfapi.RemoveBookmarksFile(pdfPath, pdfPath, config)
}

// SignPDF stamps a signature image onto a PDF at the specified page, coordinates, and scale.
// pdfPath: input PDF file
// sigImgPath: signature image file (PNG/JPEG)
// pageNum: 1-based page number
// x, y: coordinates in points (72 points = 1 inch)
// scale: scale factor for the image (1.0 = original size)
// outputPath: output PDF file
func SignPDF(pdfPath, sigImgPath string, pageNum int, x, y, scale float64, outputPath string) error {
	// Copy the original file to the output first
	if err := copyFile(pdfPath, outputPath); err != nil {
		return fmt.Errorf("failed to copy PDF: %w", err)
	}

	// Use pos:full (absolute positioning), rot:0 (no rotation), op:1 (fully opaque)
	desc := fmt.Sprintf("scale:%.2f, pos:full, rot:0, op:1", scale)

	wm, err := pdfcpu.ParseImageWatermarkDetails(sigImgPath, desc, true, types.POINTS)
	if err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to parse image watermark: %w", err)
	}

	// Manually override positioning
	wm.Dx = x
	wm.Dy = y

	// Apply watermark on a specific page
	config := model.NewDefaultConfiguration()
	pages := []string{fmt.Sprintf("%d", pageNum)}
	if err := pdfapi.AddWatermarksFile(outputPath, "", pages, wm, config); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to apply signature: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
