package investment

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/extrame/ole2"
	"github.com/extrame/xls"
	"io"
	"local-finance/internal/models"
	"math"
	"strconv"
	"strings"
)

func readLegacyInvestmentSheets(data []byte) (sheets []models.InvestmentSheet, err error) {
	// The legacy reader can panic on corrupt BIFF records. Return a normal import error.
	defer func() {
		if recover() != nil {
			sheets = nil
			err = fmt.Errorf("cannot read legacy investment workbook")
		}
	}()
	if !bytes.HasPrefix(data, []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}) {
		return nil, fmt.Errorf("not a legacy Excel workbook")
	}
	f, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil {
		return nil, fmt.Errorf("cannot open legacy investment workbook: %w", err)
	}
	if f.NumSheets() == 0 || f.NumSheets() > 128 {
		return nil, fmt.Errorf("invalid worksheet count")
	}
	numericCells, err := legacyNumericCells(data)
	if err != nil {
		return nil, err
	}
	sheets = []models.InvestmentSheet{}
	cells := 0
	for i := 0; i < f.NumSheets(); i++ {
		sheet := f.GetSheet(i)
		if sheet == nil {
			return nil, fmt.Errorf("cannot read investment worksheet")
		}
		rows := make([][]string, int(sheet.MaxRow)+1)
		for r := range rows {
			rows[r] = []string{}
			row := legacyInvestmentRow(sheet, r)
			if row == nil {
				continue
			}
			cells += row.LastCol()
			if cells > 1_000_000 {
				return nil, fmt.Errorf("investment workbook contains too many cells")
			}
			for c := 0; c < row.LastCol(); c++ {
				value, numeric := numericCells[i][[2]int{r, c}]
				if !numeric {
					value = strings.TrimSpace(row.Col(c))
				}
				rows[r] = append(rows[r], value)
			}
		}
		sheets = append(sheets, models.InvestmentSheet{Name: sheet.Name, Rows: rows})
	}
	return sheets, nil
}

// Read raw numeric BIFF cells. The legacy library treats every custom format as
// a date and fails to apply RK's divide-by-100 flag to integers. Neither formatted
// strings nor Excel date guesses are appropriate for acquisition costs/quantities.
func legacyNumericCells(data []byte) ([]map[[2]int]string, error) {
	ole, err := ole2.Open(bytes.NewReader(data), "utf-8")
	if err != nil {
		return nil, err
	}
	dirs, err := ole.ListDir()
	if err != nil {
		return nil, err
	}
	var book, root *ole2.File
	for _, dir := range dirs {
		if dir.Name() == "Workbook" || dir.Name() == "Book" {
			book = dir
		}
		if dir.Name() == "Root Entry" {
			root = dir
		}
	}
	if book == nil || root == nil {
		return nil, fmt.Errorf("missing Excel workbook stream")
	}
	stream, err := io.ReadAll(io.LimitReader(ole.OpenFile(book, root), 10<<20+1))
	if err != nil || len(stream) > 10<<20 {
		return nil, fmt.Errorf("invalid Excel workbook stream")
	}
	var offsets []int
	for pos := 0; pos+4 <= len(stream); {
		id, size := binary.LittleEndian.Uint16(stream[pos:]), int(binary.LittleEndian.Uint16(stream[pos+2:]))
		end := pos + 4 + size
		if end > len(stream) {
			return nil, fmt.Errorf("truncated Excel record")
		}
		if id == 0x85 && size >= 4 {
			offsets = append(offsets, int(binary.LittleEndian.Uint32(stream[pos+4:])))
		}
		pos = end
		if id == 0xa {
			break
		}
	}
	result := make([]map[[2]int]string, len(offsets))
	for i, start := range offsets {
		result[i] = map[[2]int]string{}
		for pos := start; pos+4 <= len(stream); {
			id, size := binary.LittleEndian.Uint16(stream[pos:]), int(binary.LittleEndian.Uint16(stream[pos+2:]))
			end := pos + 4 + size
			if end > len(stream) {
				return nil, fmt.Errorf("truncated Excel cell")
			}
			record := stream[pos+4 : end]
			put := func(row, col uint16, n float64) {
				result[i][[2]int{int(row), int(col)}] = strconv.FormatFloat(n, 'f', -1, 64)
			}
			switch id {
			case 0x203: // NUMBER
				if size < 14 {
					return nil, fmt.Errorf("invalid Excel number")
				}
				put(binary.LittleEndian.Uint16(record), binary.LittleEndian.Uint16(record[2:]), math.Float64frombits(binary.LittleEndian.Uint64(record[6:])))
			case 0x27e: // RK
				if size < 10 {
					return nil, fmt.Errorf("invalid Excel RK number")
				}
				put(binary.LittleEndian.Uint16(record), binary.LittleEndian.Uint16(record[2:]), decodeInvestmentRK(binary.LittleEndian.Uint32(record[6:])))
			case 0xbd: // MULRK: row, first column, repeated XF/RK pairs, last column.
				if size < 12 || (size-6)%6 != 0 {
					return nil, fmt.Errorf("invalid Excel MULRK numbers")
				}
				row, first := binary.LittleEndian.Uint16(record), binary.LittleEndian.Uint16(record[2:])
				for j := 0; j < (size-6)/6; j++ {
					put(row, first+uint16(j), decodeInvestmentRK(binary.LittleEndian.Uint32(record[6+6*j:])))
				}
			}
			pos = end
			if id == 0xa {
				break
			}
		}
	}
	return result, nil
}

func decodeInvestmentRK(value uint32) float64 {
	var n float64
	if value&2 != 0 {
		n = float64(int32(value) >> 2)
	} else {
		n = math.Float64frombits(uint64(value&^3) << 32)
	}
	if value&1 != 0 {
		n /= 100
	}
	return n
}

func legacyInvestmentRow(sheet *xls.WorkSheet, index int) (row *xls.Row) {
	// xls.Row dereferences missing rows; empty BIFF rows are valid export padding.
	defer func() {
		if recover() != nil {
			row = nil
		}
	}()
	return sheet.Row(index)
}
