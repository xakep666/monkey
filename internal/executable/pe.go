package executable

import (
	"debug/gosym"
	"debug/pe"
	"fmt"
	"io"
)

type PE struct {
	io.WriterAt

	goarch      string
	imageBase   uint64
	textSection *pe.Section

	symTabStart, symTabEnd *pe.Symbol
	symTabSection          *pe.Section

	pcLnTabStart, pcLnTabEnd *pe.Symbol
	pcLnTabSection           *pe.Section
}

func NewPE(rw ReadWriterAt) (*PE, error) {
	peFile, err := pe.NewFile(rw)
	if err != nil {
		return nil, fmt.Errorf("pe open: %w", err)
	}

	var imageBase uint64
	switch oh := peFile.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		imageBase = uint64(oh.ImageBase)
	case *pe.OptionalHeader64:
		imageBase = oh.ImageBase
	default:
		return nil, ErrNotGo("pe format not recognized")
	}

	text := peFile.Section(".text")
	if text == nil {
		return nil, ErrNotGo(".text section not found")
	}

	// go stores symtab and pclntab inside other section (currently .text) and their boundaries can be found in symbol values

	symTabStart, symTabEnd, err := startEndSymbols(peFile, "runtime.symtab", "runtime.esymtab")
	if err != nil {
		return nil, err
	}

	pcLnTabStart, pcLnTabEnd, err := startEndSymbols(peFile, "runtime.pclntab", "runtime.epclntab")
	if err != nil {
		return nil, err
	}

	goarch := getGOARCH(rw)
	if goarch == "" {
		if goarch = peGOARCH(peFile); goarch == "" {
			return nil, ErrNotGo("can't detect goarch")
		}
	}

	return &PE{
		WriterAt: rw,

		goarch:         goarch,
		imageBase:      imageBase,
		textSection:    text,
		symTabStart:    symTabStart,
		symTabEnd:      symTabEnd,
		symTabSection:  peFile.Sections[symTabStart.SectionNumber-1],
		pcLnTabStart:   pcLnTabStart,
		pcLnTabEnd:     pcLnTabEnd,
		pcLnTabSection: peFile.Sections[pcLnTabStart.SectionNumber-1],
	}, nil
}

func (pe *PE) GOARCH() string { return pe.goarch }

func (pe *PE) TextAddr() uint64 { return pe.imageBase + uint64(pe.textSection.VirtualAddress) }

func (pe *PE) GoSymTabData() io.Reader {
	return io.NewSectionReader(pe.symTabSection.ReaderAt,
		int64(pe.symTabStart.Value),
		int64(pe.symTabEnd.Value-pe.symTabStart.Value),
	)
}

func (pe *PE) GoPCLnTabData() io.Reader {
	return io.NewSectionReader(pe.pcLnTabSection.ReaderAt,
		int64(pe.pcLnTabStart.Value),
		int64(pe.pcLnTabEnd.Value-pe.pcLnTabStart.Value),
	)
}

func (pe *PE) Offset(p *gosym.Func) int64 {
	return int64(p.Entry-pe.imageBase) - int64(pe.textSection.VirtualAddress-pe.textSection.Offset)
}

func startEndSymbols(f *pe.File, startSymbol, endSymbol string) (ssym, esym *pe.Symbol, err error) {
	for _, s := range f.Symbols {
		switch s.Name {
		case startSymbol:
			if ssym == nil {
				ssym = s
			}
		case endSymbol:
			if esym == nil {
				esym = s
			}
		default:
			continue
		}

		if s.SectionNumber <= 0 || len(f.Sections) < int(s.SectionNumber) {
			return nil, nil, ErrNotGo(fmt.Sprintf("bad secion number %d", s.SectionNumber))
		}
	}

	if esym == nil || ssym == nil || ssym.SectionNumber != esym.SectionNumber {
		return nil, nil, ErrNotGo("no start/end symbol or they're in different sections")
	}

	return ssym, esym, nil
}

func peGOARCH(f *pe.File) string {
	switch f.Machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		return "386"
	case pe.IMAGE_FILE_MACHINE_AMD64:
		return "amd64"
	case pe.IMAGE_FILE_MACHINE_ARMNT:
		return "arm"
	case pe.IMAGE_FILE_MACHINE_ARM64:
		return "arm64"
	default:
		return ""
	}
}
