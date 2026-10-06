package extractor

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"golang.org/x/text/secure/precis"
	"golang.org/x/text/unicode/norm"
)

var finalPDFXRef = regexp.MustCompile(`startxref\s+([0-9]+)\s+%%EOF\s*$`)

// normalizeMissingPDFMetadataFlag handles an AES-256 revision-5 producer quirk:
// /Perms declares unencrypted metadata but /EncryptMetadata is omitted. It only
// adds the missing flag and repairs legacy IV-only empty Form streams in a new
// in-memory revision, without changing any password, encrypted permission, or
// original byte. The caller MUST
// retry full pdfcpu decryption; successful authentication and /Perms validation
// are what establish that false is the correct value.
//
// Keep the structural scope conservative: one classic xref table, an indirect
// encryption dictionary, and no previous/hybrid revisions. Unsupported shapes
// retain the original decoding error rather than guessing at PDF structure.
func normalizeMissingPDFMetadataFlag(data []byte, password string) ([]byte, error) {
	unsupported := errors.New("PDF metadata normalization is not applicable")
	match := finalPDFXRef.FindSubmatch(data)
	if len(match) != 2 {
		return nil, unsupported
	}
	xrefOffset, err := strconv.ParseInt(string(match[1]), 10, 64)
	if err != nil || xrefOffset < 0 || xrefOffset >= int64(len(data)) {
		return nil, unsupported
	}
	r := strings.NewReader(string(data[xrefOffset:]))
	var token string
	if _, err := fmt.Fscan(r, &token); err != nil || token != "xref" {
		return nil, unsupported
	}
	type entry struct {
		offset int64
		gen    int
	}
	entries := map[int]entry{}
	for {
		if _, err := fmt.Fscan(r, &token); err != nil {
			return nil, unsupported
		}
		if token == "trailer" {
			break
		}
		first, err := strconv.Atoi(token)
		var count int
		if err != nil || first < 0 {
			return nil, unsupported
		}
		if _, err := fmt.Fscan(r, &count); err != nil || count < 0 || count > len(data)/10 || first > len(data)/10 || first+count > len(data)/10 {
			return nil, unsupported
		}
		for n := 0; n < count; n++ {
			var offsetToken, genToken, status string
			if _, err := fmt.Fscan(r, &offsetToken, &genToken, &status); err != nil {
				return nil, unsupported
			}
			// Xref fields have leading zeroes but are decimal, never octal.
			offset, offsetErr := strconv.ParseInt(offsetToken, 10, 64)
			gen, genErr := strconv.Atoi(genToken)
			if offsetErr != nil || genErr != nil || (status != "n" && status != "f") || gen < 0 || gen > 65535 {
				return nil, unsupported
			}
			if status == "n" {
				if offset < 0 || offset >= xrefOffset {
					return nil, unsupported
				}
				if _, exists := entries[first+n]; exists {
					return nil, unsupported
				}
				entries[first+n] = entry{offset, gen}
			}
		}
	}
	trailerBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, unsupported
	}
	line := strings.TrimSpace(string(trailerBytes))
	obj, err := model.ParseObject(&line)
	trailer, ok := obj.(types.Dict)
	if err != nil || !ok || trailer["Prev"] != nil || trailer["XRefStm"] != nil {
		return nil, fmt.Errorf("%w: trailer", unsupported)
	}
	ref, ok := trailer["Encrypt"].(types.IndirectRef)
	if !ok || trailer["Root"] == nil || trailer.IntEntry("Size") == nil {
		return nil, fmt.Errorf("%w: reference", unsupported)
	}
	objectNumber, generation := ref.ObjectNumber.Value(), ref.GenerationNumber.Value()
	encEntry, ok := entries[objectNumber]
	if !ok || encEntry.gen != generation {
		return nil, unsupported
	}
	objectReader := strings.NewReader(string(data[encEntry.offset:xrefOffset]))
	var num, gen int
	if _, err := fmt.Fscan(objectReader, &num, &gen, &token); err != nil || num != objectNumber || gen != generation || token != "obj" {
		return nil, fmt.Errorf("%w: object header", unsupported)
	}
	objectBytes, err := io.ReadAll(objectReader)
	if err != nil {
		return nil, unsupported
	}
	line = strings.TrimSpace(string(objectBytes))
	metadataFlagPresent, err := pdfDictionaryHasKey(line, "EncryptMetadata")
	if err != nil {
		return nil, unsupported
	}
	obj, err = model.ParseObject(&line)
	enc, ok := obj.(types.Dict)
	if err != nil || !ok || !strings.HasPrefix(strings.TrimSpace(line), "endobj") {
		return nil, fmt.Errorf("%w: object dictionary", unsupported)
	}
	v, rev := enc.IntEntry("V"), enc.IntEntry("R")
	if metadataFlagPresent || enc.NameEntry("Filter") == nil || *enc.NameEntry("Filter") != "Standard" || v == nil || *v != 5 || rev == nil || *rev != 5 {
		return nil, fmt.Errorf("%w: encryption configuration", unsupported)
	}
	fileKey, err := authenticatedR5MetadataKey(enc, password)
	if err != nil {
		return nil, unsupported
	}
	enc["EncryptMetadata"] = types.Boolean(false)
	type revision struct {
		number int
		gen    int
		dict   types.Dict
		stream []byte
	}
	revisions := []revision{{number: objectNumber, gen: generation, dict: enc}}
	streamFilter := enc.NameEntry("StmF")
	streamsUseAESV3 := false
	if streamFilter != nil {
		method := enc.DictEntry("CF").DictEntry(*streamFilter).NameEntry("CFM")
		streamsUseAESV3 = method != nil && *method == "AESV3"
	}
	// Bound each original object by the next physical object so scanning the
	// dictionaries costs O(file size), rather than repeatedly copying the tail.
	var ordered []int
	for number := range entries {
		ordered = append(ordered, number)
	}
	sort.Slice(ordered, func(i, j int) bool { return entries[ordered[i]].offset < entries[ordered[j]].offset })
	for i, number := range ordered {
		if number == objectNumber || !streamsUseAESV3 {
			continue
		}
		end := xrefOffset
		if i+1 < len(ordered) {
			end = entries[ordered[i+1]].offset
		}
		e := entries[number]
		objectReader := strings.NewReader(string(data[e.offset:end]))
		var n, g int
		var keyword string
		if _, err := fmt.Fscan(objectReader, &n, &g, &keyword); err != nil || n != number || g != e.gen || keyword != "obj" {
			continue
		}
		body, _ := io.ReadAll(objectReader)
		line := strings.TrimSpace(string(body))
		obj, err := model.ParseObject(&line)
		d, ok := obj.(types.Dict)
		if err != nil || !ok || !ivOnlyEncryptedForm(d, line) {
			continue
		}
		stream, err := encryptEmptyFlateStream(fileKey)
		if err != nil {
			return nil, unsupported
		}
		d["Length"] = types.Integer(len(stream))
		revisions = append(revisions, revision{number, e.gen, d, stream})
	}
	var out bytes.Buffer
	out.Write(data)
	out.WriteByte('\n')
	revisionOffsets := make([]int, len(revisions))
	for i, rev := range revisions {
		revisionOffsets[i] = out.Len()
		fmt.Fprintf(&out, "%d %d obj\n%s\n", rev.number, rev.gen, rev.dict.PDFString())
		if rev.stream != nil {
			out.WriteString("stream\n")
			out.Write(rev.stream)
			out.WriteString("\nendstream\n")
		}
		out.WriteString("endobj\n")
	}
	newXRefOffset := out.Len()
	trailer["Prev"] = types.Integer(xrefOffset)
	out.WriteString("xref\n")
	for i, rev := range revisions {
		fmt.Fprintf(&out, "%d 1\n%010d %05d n \n", rev.number, revisionOffsets[i], rev.gen)
	}
	fmt.Fprintf(&out, "trailer\n%s\nstartxref\n%d\n%%%%EOF\n", trailer.PDFString(), newXRefOffset)
	return out.Bytes(), nil
}

// Recover the R5 file key only after matching the salted user or owner hash,
// then validate all twelve protected permission bytes, including the reserved
// FF bytes pdfcpu does not check. This accepts only the known absent/F quirk.
func authenticatedR5MetadataKey(enc types.Dict, password string) ([]byte, error) {
	invalid := errors.New("invalid PDF compatibility permissions")
	processed, err := precis.NewIdentifier(precis.BidiRule, precis.Norm(norm.NFKC)).String(password)
	if err != nil {
		return nil, invalid
	}
	pw := []byte(processed)
	if len(pw) > 127 {
		pw = pw[:127]
	}
	u, err := enc.StringEntryBytes("U")
	if err != nil || len(u) != 48 {
		return nil, invalid
	}
	o, err := enc.StringEntryBytes("O")
	if err != nil || len(o) != 48 {
		return nil, invalid
	}
	hash := func(salt, ownerUser []byte) []byte {
		h := sha256.New()
		h.Write(pw)
		h.Write(salt)
		h.Write(ownerUser)
		return h.Sum(nil)
	}
	var wrappingKey, wrapped []byte
	if subtle.ConstantTimeCompare(hash(o[32:40], u), o[:32]) == 1 {
		wrappingKey = hash(o[40:48], u)
		wrapped, err = enc.StringEntryBytes("OE")
	} else if subtle.ConstantTimeCompare(hash(u[32:40], nil), u[:32]) == 1 {
		wrappingKey = hash(u[40:48], nil)
		wrapped, err = enc.StringEntryBytes("UE")
	} else {
		return nil, invalid
	}
	if err != nil || len(wrapped) != 32 {
		return nil, invalid
	}
	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return nil, invalid
	}
	fileKey := make([]byte, 32)
	cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize)).CryptBlocks(fileKey, wrapped)
	perms, err := enc.StringEntryBytes("Perms")
	p := enc.IntEntry("P")
	if err != nil || len(perms) != aes.BlockSize || p == nil || int64(*p) < -1<<31 || int64(*p) > 1<<32-1 {
		return nil, invalid
	}
	block, err = aes.NewCipher(fileKey)
	if err != nil {
		return nil, invalid
	}
	protected := make([]byte, aes.BlockSize)
	block.Decrypt(protected, perms)
	expected := []byte{0, 0, 0, 0, 255, 255, 255, 255, 'F', 'a', 'd', 'b'}
	binary.LittleEndian.PutUint32(expected[:4], uint32(*p))
	if subtle.ConstantTimeCompare(protected[:12], expected) != 1 {
		return nil, invalid
	}
	return fileKey, nil
}

// Older exporters encode an empty encrypted Form as only its 16-byte IV, with
// no mandatory padding block. Accept only that exact declared legacy shape,
// never another malformed ciphertext length or stream type. Replace it with a
// valid encrypted, compressed empty stream.
func ivOnlyEncryptedForm(d types.Dict, remainder string) bool {
	nameIs := func(key, wanted string) bool {
		n := d.NameEntry(key)
		return n != nil && *n == wanted
	}
	length, formType := d.IntEntry("Length"), d.IntEntry("FormType")
	if length == nil || *length != aes.BlockSize || formType == nil || *formType != 1 || !nameIs("Type", "XObject") || !nameIs("Subtype", "Form") || !nameIs("Filter", "FlateDecode") || d["DecodeParms"] != nil {
		return false
	}
	remainder = strings.TrimLeft(remainder, " \t\r\n")
	if !strings.HasPrefix(remainder, "stream") {
		return false
	}
	remainder = remainder[len("stream"):]
	consumeEOL := func(s string) (string, bool) {
		if strings.HasPrefix(s, "\r\n") {
			return s[2:], true
		}
		if strings.HasPrefix(s, "\n") || strings.HasPrefix(s, "\r") {
			return s[1:], true
		}
		return s, false
	}
	var ok bool
	if remainder, ok = consumeEOL(remainder); !ok {
		return false
	}
	if len(remainder) < aes.BlockSize {
		return false
	}
	remainder = remainder[aes.BlockSize:]
	remainder, _ = consumeEOL(remainder)
	if !strings.HasPrefix(remainder, "endstream") {
		return false
	}
	return strings.TrimSpace(remainder[len("endstream"):]) == "endobj"
}

func encryptEmptyFlateStream(key []byte) ([]byte, error) {
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if err := w.Close(); err != nil {
		return nil, err
	}
	plain := compressed.Bytes()
	padding := aes.BlockSize - len(plain)%aes.BlockSize
	plain = append(plain, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	result := make([]byte, aes.BlockSize+len(plain))
	if _, err := io.ReadFull(rand.Reader, result[:aes.BlockSize]); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, result[:aes.BlockSize]).CryptBlocks(result[aes.BlockSize:], plain)
	return result, nil
}

// ParseObject discards dictionary entries whose value is null. Inspect the
// original top-level keys with the same object parser so even an explicit null
// metadata flag (including an escaped PDF name) is outside this fallback.
func pdfDictionaryHasKey(line, wanted string) (bool, error) {
	if !strings.HasPrefix(line, "<<") {
		return false, errors.New("not a PDF dictionary")
	}
	line = line[2:]
	for {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ">>") {
			return false, nil
		}
		obj, err := model.ParseObject(&line)
		name, ok := obj.(types.Name)
		if err != nil || !ok {
			return false, errors.New("invalid PDF dictionary key")
		}
		decoded, err := types.DecodeName(string(name))
		if err != nil {
			return false, err
		}
		if decoded == wanted {
			return true, nil
		}
		if _, err := model.ParseObject(&line); err != nil {
			return false, err
		}
	}
}
