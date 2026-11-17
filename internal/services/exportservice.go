// Package services contains the application business-logic layer.
//
// CSV formula-injection protection: any cell value whose first character is
// '=', '+', '-', or '@' is prefixed with a tab character ('\t') before being
// written to the CSV output. This prevents spreadsheet applications from
// evaluating the value as a formula when the file is opened.
package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExportResult is returned by export methods and describes what was written.
type ExportResult struct {
	Path     string `json:"path"`
	RowCount int    `json:"rowCount"`
}

// ExportService writes document data to JSON or CSV files. It is stateless and
// safe for concurrent use.
type ExportService struct{}

// NewExportService creates an ExportService.
func NewExportService() *ExportService {
	return &ExportService{}
}

// ExportDocuments writes a JSON array of documents to destPath.
//
// Supported formats:
//   - "json"  — indented JSON array
//   - "csv"   — comma-separated values with a header row
//
// documentsJSON must be a JSON array of objects.
// destPath must be an absolute path to the destination file.
func (s *ExportService) ExportDocuments(ctx context.Context, documentsJSON, format, destPath string) (ExportResult, error) {
	if destPath == "" {
		return ExportResult{}, fmt.Errorf("export: destPath must not be empty")
	}
	if err := ctx.Err(); err != nil {
		return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
	}
	if !json.Valid([]byte(documentsJSON)) {
		return ExportResult{}, fmt.Errorf("export: documentsJSON is not valid JSON")
	}

	var docs []json.RawMessage
	if err := json.Unmarshal([]byte(documentsJSON), &docs); err != nil {
		return ExportResult{}, fmt.Errorf("export: documentsJSON must be a JSON array: %w", err)
	}

	switch format {
	case "json":
		return s.writeJSONArray(ctx, docs, destPath)
	case "csv":
		return s.writeCSV(ctx, docs, destPath)
	default:
		return ExportResult{}, fmt.Errorf("export: unknown format %q (want \"json\" or \"csv\")", format)
	}
}

// ExportSearchResult writes Meilisearch search results to destPath.
//
// Supported formats:
//   - "json_hits"  — indented JSON array of the hits only
//   - "json_full"  — indented JSON of the complete search response
//
// searchResponseJSON must be the full Meilisearch search response JSON object.
// destPath must be an absolute path to the destination file.
func (s *ExportService) ExportSearchResult(ctx context.Context, searchResponseJSON, format, destPath string) (ExportResult, error) {
	if destPath == "" {
		return ExportResult{}, fmt.Errorf("export: destPath must not be empty")
	}
	if err := ctx.Err(); err != nil {
		return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
	}
	if !json.Valid([]byte(searchResponseJSON)) {
		return ExportResult{}, fmt.Errorf("export: searchResponseJSON is not valid JSON")
	}

	switch format {
	case "json_hits":
		var resp struct {
			Hits []json.RawMessage `json:"hits"`
		}
		if err := json.Unmarshal([]byte(searchResponseJSON), &resp); err != nil {
			return ExportResult{}, fmt.Errorf("export: parse search response: %w", err)
		}
		return s.writeJSONArray(ctx, resp.Hits, destPath)

	case "json_full":
		if err := ctx.Err(); err != nil {
			return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
		}
		// Re-indent the full response for consistent output.
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(searchResponseJSON), &raw); err != nil {
			return ExportResult{}, fmt.Errorf("export: parse search response: %w", err)
		}
		indented, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return ExportResult{}, fmt.Errorf("export: marshal: %w", err)
		}
		if err := atomicWriteFile(destPath, indented); err != nil {
			return ExportResult{}, err
		}
		// Row count is the number of hits when available; fall back to 1 (the
		// document itself) so the caller always has a meaningful number.
		var hitsOnly struct {
			Hits []json.RawMessage `json:"hits"`
		}
		rowCount := 1
		if json.Unmarshal([]byte(searchResponseJSON), &hitsOnly) == nil {
			rowCount = len(hitsOnly.Hits)
		}
		return ExportResult{Path: destPath, RowCount: rowCount}, nil

	default:
		return ExportResult{}, fmt.Errorf("export: unknown format %q (want \"json_hits\" or \"json_full\")", format)
	}
}

// --- internal helpers ---

// writeJSONArray marshals docs as an indented JSON array and writes it atomically.
func (s *ExportService) writeJSONArray(ctx context.Context, docs []json.RawMessage, destPath string) (ExportResult, error) {
	if err := ctx.Err(); err != nil {
		return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return ExportResult{}, fmt.Errorf("export: marshal: %w", err)
	}
	if err := atomicWriteFile(destPath, data); err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Path: destPath, RowCount: len(docs)}, nil
}

// writeCSV converts docs to CSV and writes it atomically.
func (s *ExportService) writeCSV(ctx context.Context, docs []json.RawMessage, destPath string) (ExportResult, error) {
	if err := ctx.Err(); err != nil {
		return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
	}

	// Parse every document into a map so we can collect the union of all keys.
	parsed := make([]map[string]json.RawMessage, 0, len(docs))
	keySet := make(map[string]struct{})
	for _, raw := range docs {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			// Skip non-object elements rather than aborting the export.
			continue
		}
		parsed = append(parsed, obj)
		for k := range obj {
			keySet[k] = struct{}{}
		}
	}

	// Build a sorted header; promote "id" to the front if present.
	headers := make([]string, 0, len(keySet))
	for k := range keySet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	// Move "id" to position 0 if it exists.
	for i, h := range headers {
		if h == "id" && i != 0 {
			headers = append(headers[:i], headers[i+1:]...)
			headers = append([]string{"id"}, headers...)
			break
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(headers); err != nil {
		return ExportResult{}, fmt.Errorf("export: csv header: %w", err)
	}

	for _, obj := range parsed {
		if err := ctx.Err(); err != nil {
			return ExportResult{}, fmt.Errorf("export: context canceled: %w", err)
		}
		row := make([]string, len(headers))
		for i, h := range headers {
			raw, ok := obj[h]
			if !ok {
				row[i] = ""
				continue
			}
			row[i] = csvCellValue(raw)
		}
		if err := w.Write(row); err != nil {
			return ExportResult{}, fmt.Errorf("export: csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return ExportResult{}, fmt.Errorf("export: csv flush: %w", err)
	}

	if err := atomicWriteFile(destPath, buf.Bytes()); err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Path: destPath, RowCount: len(parsed)}, nil
}

// csvCellValue converts a raw JSON value to its CSV cell representation.
//
// Rules:
//   - JSON string  → string content without surrounding quotes
//   - JSON null    → empty string
//   - number/bool  → string representation as-is
//   - object/array → compact JSON
//
// Formula injection protection: values beginning with '=', '+', '-', or '@'
// are prefixed with a tab character so that spreadsheet applications do not
// evaluate them as formulas.
func csvCellValue(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var cell string

	switch raw[0] {
	case '"':
		// JSON string — unwrap the quotes.
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			cell = s
		} else {
			cell = string(raw)
		}
	case 'n':
		// JSON null → empty.
		cell = ""
	case '{', '[':
		// Object or array → compact JSON.
		var compacted bytes.Buffer
		if err := json.Compact(&compacted, raw); err == nil {
			cell = compacted.String()
		} else {
			cell = string(raw)
		}
	default:
		// number or bool — use the raw text directly.
		cell = strings.TrimSpace(string(raw))
	}

	// Formula injection protection.
	if len(cell) > 0 {
		switch cell[0] {
		case '=', '+', '-', '@':
			cell = "\t" + cell
		}
	}

	return cell
}

// atomicWriteFile writes data to destPath using a tmp-file + rename pattern.
// If the rename fails the tmp file is removed.
func atomicWriteFile(destPath string, data []byte) error {
	dir := filepath.Dir(destPath)
	_ = dir // dir is implicitly used via destPath; kept for documentation clarity.
	tmp := destPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("export: write tmp: %w", err)
	}
	if err := os.Rename(tmp, destPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("export: rename: %w", err)
	}
	return nil
}
