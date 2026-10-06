package extractor

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func syntheticMetadataPDF(t *testing.T, emptyForm ...bool) ([]byte, []byte) {
	t.Helper()
	content := "BT /F1 12 Tf 72 720 Td (Synthetic statement) Tj ET"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}
	if len(emptyForm) > 0 && emptyForm[0] {
		var compressed bytes.Buffer
		w := zlib.NewWriter(&compressed)
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		objects[2] = "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> /XObject << /Blank 6 0 R >> >> /Contents 5 0 R >>"
		content += " /Blank Do"
		objects[4] = fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content)
		objects = append(objects, fmt.Sprintf("<< /Type /XObject /Subtype /Form /FormType 1 /BBox [0 0 1 1] /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream", compressed.Len(), compressed.Bytes()))
	}
	var plain bytes.Buffer
	plain.WriteString("%PDF-1.7\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = plain.Len()
		fmt.Fprintf(&plain, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := plain.Len()
	fmt.Fprintf(&plain, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&plain, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&plain, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	conf := model.NewAESConfiguration("fixture-open", "fixture-owner", 256)
	conf.WriteObjectStream, conf.WriteXRefStream = false, false
	var encrypted bytes.Buffer
	if err := api.Encrypt(bytes.NewReader(plain.Bytes()), &encrypted, conf); err != nil {
		t.Fatal(err)
	}
	readConf := model.NewDefaultConfiguration()
	readConf.UserPW = "fixture-open"
	ctx, err := api.ReadContext(bytes.NewReader(encrypted.Bytes()), readConf)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.E.R != 5 {
		t.Fatalf("fixture must use revision 5, got %d", ctx.E.R)
	}
	enc, err := ctx.EncryptDict()
	if err != nil {
		t.Fatal(err)
	}
	if _, present := enc.Find("EncryptMetadata"); present {
		t.Fatal("fixture must omit EncryptMetadata")
	}
	return bytes.Clone(encrypted.Bytes()), bytes.Clone(ctx.EncKey)
}

func alterSyntheticPermissions(t *testing.T, data, key []byte, mutate func([]byte)) []byte {
	t.Helper()
	result := bytes.Clone(data)
	match := regexp.MustCompile(`/Perms\s*<([A-Fa-f0-9]+)>`).FindSubmatchIndex(result)
	if len(match) != 4 {
		t.Fatal("fixture missing hexadecimal permissions")
	}
	perms, err := hex.DecodeString(string(result[match[2]:match[3]]))
	if err != nil || len(perms) != aes.BlockSize {
		t.Fatal("fixture permissions malformed")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	decoded := make([]byte, aes.BlockSize)
	block.Decrypt(decoded, perms)
	mutate(decoded)
	block.Encrypt(perms, decoded)
	hex.Encode(result[match[2]:match[3]], perms)
	return result
}

// The fixture's encryption object is the final object before its classic xref.
// Insert an explicit flag and shift only the final startxref; object offsets
// still refer to the original positions because no preceding object moves.
func insertSyntheticMetadataFlag(t *testing.T, data []byte, flag string) []byte {
	t.Helper()
	match := finalPDFXRef.FindSubmatchIndex(data)
	if len(match) != 4 {
		t.Fatal("fixture missing final xref")
	}
	xref, err := strconv.Atoi(string(data[match[2]:match[3]]))
	if err != nil {
		t.Fatal(err)
	}
	end := bytes.LastIndex(data[:xref], []byte(">>\nendobj\n"))
	if end < 0 {
		t.Fatal("fixture missing encryption dictionary terminator")
	}
	flagBytes := []byte("/EncryptMetadata " + flag)
	result := append(bytes.Clone(data[:end]), flagBytes...)
	result = append(result, data[end:match[2]]...)
	result = append(result, []byte(strconv.Itoa(xref+len(flagBytes)))...)
	return append(result, data[match[3]:]...)
}

func TestPDFMissingMetadataFlagCompatibility(t *testing.T) {
	valid, key := syntheticMetadataPDF(t)
	quirk := alterSyntheticPermissions(t, valid, key, func(p []byte) { p[8] = 'F' })
	original := bytes.Clone(quirk)
	if _, err := normalizeMissingPDFMetadataFlag(quirk, "fixture-open"); err != nil {
		t.Fatalf("fixture normalization failed: %v", err)
	}
	conf := model.NewDefaultConfiguration()
	conf.UserPW = "fixture-open"
	var out bytes.Buffer
	if err := api.Decrypt(bytes.NewReader(quirk), &out, conf); err == nil || !strings.Contains(err.Error(), "invalid permissions") {
		t.Fatalf("fixture must reproduce the permission validation failure: %v", err)
	}
	for _, password := range []string{"fixture-open", "fixture-owner"} {
		text, err := ExtractPDFText(bytes.NewReader(quirk), password)
		if err != nil || !strings.Contains(text, "Synthetic statement") {
			t.Fatalf("authenticated fixture failed to extract: %q, %v", text, err)
		}
	}
	if !bytes.Equal(quirk, original) {
		t.Fatal("normalization changed the original source bytes")
	}
	for _, password := range []string{"", "wrong-password"} {
		if _, err := DecryptPDFIfNeeded(quirk, password); !errors.Is(err, ErrPDFPasswordRequired) {
			t.Fatalf("incorrect password must remain rejected: %v", err)
		}
	}
	normalized, err := normalizeMissingPDFMetadataFlag(quirk, "fixture-open")
	if err != nil || !bytes.HasPrefix(normalized, quirk) {
		t.Fatalf("normalization must append a revision: %v", err)
	}
	// An explicit conflicting flag in the original single revision is not
	// eligible for normalization, even with a correctly authenticated password.
	explicitTrue := insertSyntheticMetadataFlag(t, quirk, "true")
	if _, err := DecryptPDFIfNeeded(explicitTrue, "fixture-open"); err == nil || errors.Is(err, ErrPDFPasswordRequired) {
		t.Fatalf("explicit metadata mismatch must remain a decoding error: %v", err)
	}
}

func TestPDFMetadataNormalizationScope(t *testing.T) {
	valid, key := syntheticMetadataPDF(t)
	quirk := alterSyntheticPermissions(t, valid, key, func(p []byte) { p[8] = 'F' })
	for _, flag := range []string{"true", "false", "null"} {
		if _, err := normalizeMissingPDFMetadataFlag(insertSyntheticMetadataFlag(t, quirk, flag), "fixture-open"); err == nil {
			t.Fatalf("explicit metadata flag %s must not be normalized", flag)
		}
	}
	// Explicit false is already valid and takes the normal decryption path.
	if _, err := DecryptPDFIfNeeded(insertSyntheticMetadataFlag(t, quirk, "false"), "fixture-open"); err != nil {
		t.Fatalf("valid explicit metadata flag rejected: %v", err)
	}
	for _, key := range []string{"Prev", "XRefStm"} {
		unsupported := bytes.Replace(quirk, []byte("/Size 8"), []byte("/"+key+" 0/Size 8"), 1)
		if _, err := normalizeMissingPDFMetadataFlag(unsupported, "fixture-open"); err == nil {
			t.Fatalf("%s structure must not be normalized", key)
		}
	}
	match := finalPDFXRef.FindSubmatch(quirk)
	xref, _ := strconv.Atoi(string(match[1]))
	streamShape := bytes.Clone(quirk)
	copy(streamShape[xref:xref+4], []byte("9 0 "))
	if _, err := normalizeMissingPDFMetadataFlag(streamShape, "fixture-open"); err == nil {
		t.Fatal("xref stream structure must not be normalized")
	}
}

func TestPDFMetadataCompatibilityRejectsOtherPermissionDamage(t *testing.T) {
	valid, key := syntheticMetadataPDF(t)
	for _, tc := range []struct {
		name   string
		mutate func([]byte)
	}{
		{"magic", func(p []byte) { p[9] ^= 1 }},
		{"permission bits", func(p []byte) { p[0] ^= 4 }},
		{"reserved permission bytes", func(p []byte) { p[4] = 0 }},
		{"metadata marker", func(p []byte) { p[8] = 'X' }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			damaged := alterSyntheticPermissions(t, valid, key, func(p []byte) {
				p[8] = 'F'
				tc.mutate(p)
			})
			if _, err := DecryptPDFIfNeeded(damaged, "fixture-open"); err == nil || errors.Is(err, ErrPDFPasswordRequired) {
				t.Fatalf("permission damage must remain a decoding error: %v", err)
			}
		})
	}
}

func TestPDFMetadataCompatibilityWithEmptyEncryptedForm(t *testing.T) {
	valid, key := syntheticMetadataPDF(t, true)
	quirk := alterSyntheticPermissions(t, valid, key, func(p []byte) { p[8] = 'F' })
	// Rebuild only this synthetic fixture's classic xref after replacing its
	// encrypted blank Form with the producer's legacy IV-only representation.
	match := finalPDFXRef.FindSubmatch(quirk)
	xref, _ := strconv.Atoi(string(match[1]))
	rows := regexp.MustCompile(`(?m)^([0-9]{10}) ([0-9]{5}) ([nf]) `).FindAllSubmatch(quirk[xref:], -1)
	if len(rows) != 9 {
		t.Fatalf("expected nine synthetic xref entries, got %d", len(rows))
	}
	offsets := make([]int, len(rows))
	for n, row := range rows {
		offsets[n], _ = strconv.Atoi(string(row[1]))
	}
	start, end := offsets[6], xref
	for _, offset := range offsets {
		if offset > start && offset < end {
			end = offset
		}
	}
	object := string(quirk[start:end])
	object = strings.TrimPrefix(object, "6 0 obj\n")
	obj, err := model.ParseObject(&object)
	d, ok := obj.(types.Dict)
	if err != nil || !ok || d.NameEntry("Subtype") == nil || *d.NameEntry("Subtype") != "Form" {
		t.Fatal("synthetic fixture did not retain the blank Form")
	}
	d["Length"] = types.Integer(aes.BlockSize)
	replacement := []byte("6 0 obj\n" + d.PDFString() + "\nstream\n" + strings.Repeat("B", aes.BlockSize) + "\nendstream\nendobj\n")
	delta := len(replacement) - (end - start)
	var out bytes.Buffer
	out.Write(quirk[:start])
	out.Write(replacement)
	out.Write(quirk[end:xref])
	newXRef := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(rows))
	for n, row := range rows {
		offset := offsets[n]
		if offset >= end {
			offset += delta
		}
		fmt.Fprintf(&out, "%010d %s %s \n", offset, row[2], row[3])
	}
	trailerPos := bytes.Index(quirk[xref:], []byte("trailer"))
	trailer := string(quirk[xref+trailerPos+len("trailer"):])
	trailerObj, err := model.ParseObject(&trailer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(&out, "trailer\n%s\nstartxref\n%d\n%%%%EOF\n", trailerObj.PDFString(), newXRef)
	broken := bytes.Clone(out.Bytes())
	conf := model.NewDefaultConfiguration()
	conf.UserPW = "fixture-open"
	// Correct only the fixture's metadata flag to show the remaining failure
	// is the IV-only encrypted stream, independently of metadata validation.
	metadataOnly := insertSyntheticMetadataFlag(t, broken, "false")
	var decoded bytes.Buffer
	if err := api.Decrypt(bytes.NewReader(metadataOnly), &decoded, conf); err == nil || !strings.Contains(err.Error(), "ciphertext too short") {
		t.Fatalf("fixture must reproduce empty encrypted stream failure: %v", err)
	}
	for _, password := range []string{"fixture-open", "fixture-owner"} {
		text, err := ExtractPDFText(bytes.NewReader(broken), password)
		if err != nil || !strings.Contains(text, "Synthetic statement") {
			t.Fatalf("empty encrypted Form compatibility failed: %q, %v", text, err)
		}
	}
	if !bytes.Equal(broken, out.Bytes()) {
		t.Fatal("empty Form repair mutated its source")
	}
}

func TestIVOnlyFormCompatibilityScope(t *testing.T) {
	base := types.Dict{
		"Type": types.Name("XObject"), "Subtype": types.Name("Form"),
		"FormType": types.Integer(1), "Filter": types.Name("FlateDecode"),
		"Length": types.Integer(aes.BlockSize),
	}
	validBody := "\nstream\n" + strings.Repeat("B", aes.BlockSize) + "\nendstream\nendobj\n"
	if !ivOnlyEncryptedForm(base, validBody) {
		t.Fatal("exact synthetic IV-only Form rejected")
	}
	for _, length := range []int{0, 15, 17, 32} {
		d := base.Clone().(types.Dict)
		d["Length"] = types.Integer(length)
		if ivOnlyEncryptedForm(d, validBody) {
			t.Fatalf("declared length %d must not be repaired", length)
		}
	}
	for _, body := range []string{
		"stream\n" + strings.Repeat("B", 15) + "endstream\nendobj",
		"stream\n" + strings.Repeat("B", 17) + "\nendstream\nendobj",
		"stream\n" + strings.Repeat("B", 16) + "\nendstream\nendobjgarbage",
	} {
		if ivOnlyEncryptedForm(base, body) {
			t.Fatal("malformed IV-only boundary must not be repaired")
		}
	}
	for key, value := range map[string]types.Object{
		"Type": types.Name("Other"), "Subtype": types.Name("Image"),
		"FormType": types.Integer(2), "Filter": types.Name("ASCIIHexDecode"),
		"DecodeParms": types.Dict{"Predictor": types.Integer(12)},
	} {
		d := base.Clone().(types.Dict)
		d[key] = value
		if ivOnlyEncryptedForm(d, validBody) {
			t.Fatalf("unsupported %s must not be repaired", key)
		}
	}
}
