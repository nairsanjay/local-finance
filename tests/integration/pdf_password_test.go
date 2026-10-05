package integration_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"local-finance/internal/service"
)

func TestPDFPreviewDistinguishesAuthenticationAndDecodingFailures(t *testing.T) {
	svc := service.NewTransactionService(testDatabase(t))
	plain := syntheticBankHistoryPDF(t, "ICICI Bank", "123456781234")
	var encrypted bytes.Buffer
	// Separate passwords ensure document-open authentication works without
	// requiring the owner/permissions password.
	conf := model.NewAESConfiguration("fixture-open", "fixture-owner", 256)
	if err := api.Encrypt(bytes.NewReader(plain), &encrypted, conf); err != nil {
		t.Fatal(err)
	}
	for _, password := range []string{"", "incorrect", "fixture-open"} {
		preview, err := svc.PreviewStatement("statement.pdf", bytes.NewReader(encrypted.Bytes()), "", "", password)
		if err != nil {
			t.Fatal(err)
		}
		if password == "fixture-open" {
			if preview.RequiresPassword || preview.Error != "" || preview.TotalTransactions != 2 {
				t.Fatalf("correct document-open password rejected: %+v", preview)
			}
		} else if !preview.RequiresPassword {
			t.Fatalf("failed authentication did not request a password: %+v", preview)
		}
	}

	// Alter only synthetic encryption permission metadata, preserving byte
	// offsets and password authentication. A decoding error must not claim the
	// already authenticated document-open password is absent.
	malformed := bytes.Clone(encrypted.Bytes())
	match := regexp.MustCompile(`/Perms\s*<([A-Fa-f0-9]+)>`).FindSubmatchIndex(malformed)
	if len(match) != 4 {
		t.Fatal("synthetic AES PDF does not expose its permission metadata")
	}
	if malformed[match[2]] == '0' {
		malformed[match[2]] = '1'
	} else {
		malformed[match[2]] = '0'
	}
	preview, err := svc.PreviewStatement("statement.pdf", bytes.NewReader(malformed), "", "", "fixture-open")
	if err != nil {
		t.Fatal(err)
	}
	if preview.RequiresPassword || !strings.Contains(preview.Error, "permissions") {
		t.Fatalf("permission metadata failure mislabeled as authentication: %+v", preview)
	}
	preview, err = svc.PreviewStatement("statement.pdf", bytes.NewReader([]byte("%PDF-1.7\ninvalid")), "", "", "fixture-open")
	if err != nil {
		t.Fatal(err)
	}
	if preview.RequiresPassword || preview.Error == "" {
		t.Fatalf("structural failure mislabeled as authentication: %+v", preview)
	}
}
