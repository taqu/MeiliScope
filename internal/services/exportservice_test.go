package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- helpers ---

func newExportSvc() *ExportService {
	return NewExportService()
}

// readCSVFile opens path, parses it with encoding/csv, and returns all records
// (including the header row as records[0]).
func readCSVFile(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open CSV %q: %v", path, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV %q: %v", path, err)
	}
	return records
}

// readJSONArray opens path and unmarshals its content as a JSON array into out.
func readJSONArray(t *testing.T, path string, out interface{}) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON file %q: %v", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("unmarshal JSON file %q: %v", path, err)
	}
}

// readJSONObject opens path and unmarshals its content as a JSON object into out.
func readJSONObject(t *testing.T, path string, out interface{}) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON file %q: %v", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("unmarshal JSON object file %q: %v", path, err)
	}
}

// colIndex returns the 0-based position of header in headers, or fatals.
func colIndex(t *testing.T, headers []string, header string) int {
	t.Helper()
	for i, h := range headers {
		if h == header {
			return i
		}
	}
	t.Fatalf("header %q not found in %v", header, headers)
	return -1
}

// --- ExportDocuments JSON ---

func TestExportService_ExportDocuments_JSON(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	docs := `[{"id":1,"title":"Inception"},{"id":2,"title":"Interstellar"}]`
	destPath := filepath.Join(t.TempDir(), "out.json")

	result, err := svc.ExportDocuments(ctx, docs, "json", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}
	if result.Path != destPath {
		t.Errorf("Path: got %q, want %q", result.Path, destPath)
	}
	if result.RowCount != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount)
	}

	var got []map[string]interface{}
	readJSONArray(t, destPath, &got)
	if len(got) != 2 {
		t.Fatalf("array length: got %d, want 2", len(got))
	}
	if got[0]["title"] != "Inception" {
		t.Errorf("first title: got %v, want Inception", got[0]["title"])
	}
	if got[1]["title"] != "Interstellar" {
		t.Errorf("second title: got %v, want Interstellar", got[1]["title"])
	}
}

// --- ExportDocuments CSV ---

func TestExportService_ExportDocuments_CSV(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	docs := `[{"id":1,"title":"Inception"},{"id":2,"title":"Interstellar"},{"id":3,"title":"Tenet"}]`
	destPath := filepath.Join(t.TempDir(), "out.csv")

	result, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}
	if result.RowCount != 3 {
		t.Errorf("RowCount: got %d, want 3", result.RowCount)
	}

	records := readCSVFile(t, destPath)
	// 1 header + 3 data rows.
	if len(records) != 4 {
		t.Fatalf("record count (incl. header): got %d, want 4", len(records))
	}

	headers := records[0]
	idIdx := colIndex(t, headers, "id")
	titleIdx := colIndex(t, headers, "title")

	// "id" must be promoted to position 0.
	if idIdx != 0 {
		t.Errorf("id column index: got %d, want 0", idIdx)
	}

	// Verify data row values.
	row1 := records[1]
	if row1[idIdx] != "1" {
		t.Errorf("row1 id: got %q, want %q", row1[idIdx], "1")
	}
	if row1[titleIdx] != "Inception" {
		t.Errorf("row1 title: got %q, want %q", row1[titleIdx], "Inception")
	}
}

func TestExportService_ExportDocuments_CSV_Unicode(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	docs := `[{"id":1,"title":"日本映画"},{"id":2,"title":"黒澤明"}]`
	destPath := filepath.Join(t.TempDir(), "unicode.csv")

	result, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}
	if result.RowCount != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 3 {
		t.Fatalf("record count: got %d, want 3", len(records))
	}

	headers := records[0]
	titleIdx := colIndex(t, headers, "title")

	if records[1][titleIdx] != "日本映画" {
		t.Errorf("row1 title: got %q, want %q", records[1][titleIdx], "日本映画")
	}
	if records[2][titleIdx] != "黒澤明" {
		t.Errorf("row2 title: got %q, want %q", records[2][titleIdx], "黒澤明")
	}
}

func TestExportService_ExportDocuments_CSV_LargeNumbers(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	// This number exceeds float64 precision.
	const largeID = "9007199254740993"
	docs := `[{"id":` + largeID + `,"title":"Precision"}]`
	destPath := filepath.Join(t.TempDir(), "large.csv")

	_, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 2 {
		t.Fatalf("record count: got %d, want 2", len(records))
	}

	headers := records[0]
	idIdx := colIndex(t, headers, "id")

	if records[1][idIdx] != largeID {
		t.Errorf("large integer: got %q, want %q", records[1][idIdx], largeID)
	}
}

func TestExportService_ExportDocuments_CSV_NestedJSON(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	docs := `[{"id":1,"meta":{"year":2010,"genres":["action","sci-fi"]},"tags":["classic"]}]`
	destPath := filepath.Join(t.TempDir(), "nested.csv")

	_, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 2 {
		t.Fatalf("record count: got %d, want 2", len(records))
	}

	headers := records[0]
	metaIdx := colIndex(t, headers, "meta")
	tagsIdx := colIndex(t, headers, "tags")

	// Nested object and array should be compact JSON strings.
	metaCell := records[1][metaIdx]
	if !strings.HasPrefix(metaCell, "{") {
		t.Errorf("meta cell should be compact JSON object, got %q", metaCell)
	}
	if !strings.Contains(metaCell, "genres") {
		t.Errorf("meta cell missing 'genres' key, got %q", metaCell)
	}

	tagsCell := records[1][tagsIdx]
	if !strings.HasPrefix(tagsCell, "[") {
		t.Errorf("tags cell should be compact JSON array, got %q", tagsCell)
	}
	if !strings.Contains(tagsCell, "classic") {
		t.Errorf("tags cell missing 'classic', got %q", tagsCell)
	}
}

func TestExportService_ExportDocuments_CSV_NullAndMissing(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	// doc1 has null rating; doc2 is missing rating entirely.
	docs := `[{"id":1,"title":"Alpha","rating":null},{"id":2,"title":"Beta"}]`
	destPath := filepath.Join(t.TempDir(), "nullmissing.csv")

	_, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 3 {
		t.Fatalf("record count: got %d, want 3", len(records))
	}

	headers := records[0]
	ratingIdx := colIndex(t, headers, "rating")

	// null → empty cell.
	if records[1][ratingIdx] != "" {
		t.Errorf("null rating: got %q, want empty string", records[1][ratingIdx])
	}
	// missing → empty cell.
	if records[2][ratingIdx] != "" {
		t.Errorf("missing rating: got %q, want empty string", records[2][ratingIdx])
	}
}

func TestExportService_ExportDocuments_CSV_FormulaInjection(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	docs := `[{"id":1,"a":"=SUM(A1)","b":"+1234","c":"-note","d":"@user","e":"safe"}]`
	destPath := filepath.Join(t.TempDir(), "formula.csv")

	_, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 2 {
		t.Fatalf("record count: got %d, want 2", len(records))
	}

	headers := records[0]
	row := records[1]

	dangerous := []struct {
		col    string
		prefix byte
	}{
		{"a", '='},
		{"b", '+'},
		{"c", '-'},
		{"d", '@'},
	}
	for _, d := range dangerous {
		idx := colIndex(t, headers, d.col)
		cell := row[idx]
		if len(cell) == 0 || cell[0] != '\t' {
			t.Errorf("column %q: expected tab prefix for formula injection protection, got %q", d.col, cell)
		}
		// The original character should still be present after the tab.
		if len(cell) < 2 || cell[1] != d.prefix {
			t.Errorf("column %q: original prefix %q missing after tab, got %q", d.col, string(d.prefix), cell)
		}
	}

	// "safe" should not be prefixed.
	eIdx := colIndex(t, headers, "e")
	if row[eIdx] != "safe" {
		t.Errorf("safe cell: got %q, want %q", row[eIdx], "safe")
	}
}

func TestExportService_ExportDocuments_CSV_Quotes(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	// Titles with embedded quotes and commas — encoding/csv must handle these.
	docs := `[{"id":1,"title":"He said \"hello\", world"},{"id":2,"title":"comma, separated, title"}]`
	destPath := filepath.Join(t.TempDir(), "quotes.csv")

	_, err := svc.ExportDocuments(ctx, docs, "csv", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	records := readCSVFile(t, destPath)
	if len(records) != 3 {
		t.Fatalf("record count: got %d, want 3", len(records))
	}

	headers := records[0]
	titleIdx := colIndex(t, headers, "title")

	wantTitle1 := `He said "hello", world`
	wantTitle2 := "comma, separated, title"

	if records[1][titleIdx] != wantTitle1 {
		t.Errorf("row1 title: got %q, want %q", records[1][titleIdx], wantTitle1)
	}
	if records[2][titleIdx] != wantTitle2 {
		t.Errorf("row2 title: got %q, want %q", records[2][titleIdx], wantTitle2)
	}
}

func TestExportService_ExportDocuments_EmptyArray(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	t.Run("json", func(t *testing.T) {
		destPath := filepath.Join(t.TempDir(), "empty.json")
		result, err := svc.ExportDocuments(ctx, `[]`, "json", destPath)
		if err != nil {
			t.Fatalf("ExportDocuments JSON empty: %v", err)
		}
		if result.RowCount != 0 {
			t.Errorf("RowCount: got %d, want 0", result.RowCount)
		}
		var got []interface{}
		readJSONArray(t, destPath, &got)
		if len(got) != 0 {
			t.Errorf("array length: got %d, want 0", len(got))
		}
	})

	t.Run("csv", func(t *testing.T) {
		destPath := filepath.Join(t.TempDir(), "empty.csv")
		result, err := svc.ExportDocuments(ctx, `[]`, "csv", destPath)
		if err != nil {
			t.Fatalf("ExportDocuments CSV empty: %v", err)
		}
		if result.RowCount != 0 {
			t.Errorf("RowCount: got %d, want 0", result.RowCount)
		}
		// File must exist (may contain just a newline or be truly empty after
		// writing an empty CSV).
		if _, statErr := os.Stat(destPath); statErr != nil {
			t.Errorf("output file must exist: %v", statErr)
		}
	})
}

func TestExportService_ExportDocuments_UnknownFormat(t *testing.T) {
	svc := newExportSvc()
	destPath := filepath.Join(t.TempDir(), "out.xml")

	_, err := svc.ExportDocuments(context.Background(), `[{"id":1}]`, "xml", destPath)
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

// --- ExportSearchResult ---

func TestExportService_ExportSearchResult_JSONHits(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	searchResp := `{"hits":[{"id":1,"title":"Inception"},{"id":2,"title":"Interstellar"}],"query":"inc","processingTimeMs":3,"limit":20,"offset":0,"estimatedTotalHits":2}`
	destPath := filepath.Join(t.TempDir(), "hits.json")

	result, err := svc.ExportSearchResult(ctx, searchResp, "json_hits", destPath)
	if err != nil {
		t.Fatalf("ExportSearchResult json_hits: %v", err)
	}
	if result.RowCount != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount)
	}

	var got []map[string]interface{}
	readJSONArray(t, destPath, &got)
	if len(got) != 2 {
		t.Fatalf("hits array length: got %d, want 2", len(got))
	}
	if got[0]["title"] != "Inception" {
		t.Errorf("first hit title: got %v, want Inception", got[0]["title"])
	}
}

func TestExportService_ExportSearchResult_JSONFull(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	searchResp := `{"hits":[{"id":1},{"id":2},{"id":3}],"query":"q","processingTimeMs":5,"limit":20,"offset":0,"estimatedTotalHits":3}`
	destPath := filepath.Join(t.TempDir(), "full.json")

	result, err := svc.ExportSearchResult(ctx, searchResp, "json_full", destPath)
	if err != nil {
		t.Fatalf("ExportSearchResult json_full: %v", err)
	}
	// Row count should equal the number of hits.
	if result.RowCount != 3 {
		t.Errorf("RowCount: got %d, want 3", result.RowCount)
	}

	// File must contain the full response including processingTimeMs.
	var got map[string]interface{}
	readJSONObject(t, destPath, &got)
	if _, ok := got["processingTimeMs"]; !ok {
		t.Error("full response must contain processingTimeMs")
	}
	if _, ok := got["query"]; !ok {
		t.Error("full response must contain query")
	}
	hitsRaw, ok := got["hits"]
	if !ok {
		t.Fatal("full response must contain hits")
	}
	hits, ok := hitsRaw.([]interface{})
	if !ok {
		t.Fatalf("hits should be an array, got %T", hitsRaw)
	}
	if len(hits) != 3 {
		t.Errorf("hits length: got %d, want 3", len(hits))
	}
}

func TestExportService_ExportSearchResult_UnknownFormat(t *testing.T) {
	svc := newExportSvc()
	searchResp := `{"hits":[],"query":"","processingTimeMs":1,"limit":20,"offset":0,"estimatedTotalHits":0}`
	destPath := filepath.Join(t.TempDir(), "out.csv")

	_, err := svc.ExportSearchResult(context.Background(), searchResp, "csv", destPath)
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

// --- Atomic write ---

func TestExportService_AtomicWrite(t *testing.T) {
	svc := newExportSvc()
	ctx := context.Background()

	destPath := filepath.Join(t.TempDir(), "atomic.json")
	_, err := svc.ExportDocuments(ctx, `[{"id":1}]`, "json", destPath)
	if err != nil {
		t.Fatalf("ExportDocuments: %v", err)
	}

	// The .tmp file must not remain after a successful write.
	tmpPath := destPath + ".tmp"
	if _, statErr := os.Stat(tmpPath); statErr == nil {
		t.Errorf("temp file %q must not remain after successful export", tmpPath)
	}

	// The destination file must exist.
	if _, statErr := os.Stat(destPath); statErr != nil {
		t.Errorf("destination file must exist: %v", statErr)
	}
}

// --- Empty destPath ---

func TestExportService_EmptyDestPath(t *testing.T) {
	svc := newExportSvc()

	t.Run("ExportDocuments", func(t *testing.T) {
		_, err := svc.ExportDocuments(context.Background(), `[{"id":1}]`, "json", "")
		if err == nil {
			t.Error("expected error for empty destPath")
		}
	})

	t.Run("ExportSearchResult", func(t *testing.T) {
		resp := `{"hits":[],"query":"","processingTimeMs":1,"limit":20,"offset":0,"estimatedTotalHits":0}`
		_, err := svc.ExportSearchResult(context.Background(), resp, "json_hits", "")
		if err == nil {
			t.Error("expected error for empty destPath")
		}
	})
}

// --- Context cancellation ---

func TestExportService_ContextCanceled(t *testing.T) {
	svc := newExportSvc()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before any work.

	destPath := filepath.Join(t.TempDir(), "canceled.json")
	_, err := svc.ExportDocuments(ctx, `[{"id":1}]`, "json", destPath)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}
