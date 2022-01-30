package executable

import (
	"debug/elf"
	"debug/gosym"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	elfText      = ".text"
	elfGoSymTab  = ".gosymtab"
	elfGoPCLnTab = ".gopclntab"
)

type ELF struct {
	io.WriterAt

	goarch                string
	load                  *elf.Prog
	text, symTab, pcLnTab *elf.Section
}

func NewELF(rw ReadWriterAt) (*ELF, error) {
	elfFile, err := elf.NewFile(rw)
	if err != nil {
		return nil, fmt.Errorf("elf open: %w", err)
	}

	var loadProg *elf.Prog
	for _, prog := range elfFile.Progs {
		if prog.Type == elf.PT_LOAD {
			loadProg = prog
			break
		}
	}

	if loadProg == nil {
		return nil, ErrNotGo
	}

	var text, symTab, pcLnTab *elf.Section
	for _, section := range elfFile.Sections {
		switch {
		case text == nil && section.Name == elfText:
			text = section
		case symTab == nil && section.Name == elfGoSymTab:
			symTab = section
		case pcLnTab == nil && section.Name == elfGoPCLnTab:
			pcLnTab = section
		}
	}

	if text == nil || symTab == nil || pcLnTab == nil {
		return nil, ErrNotGo
	}

	goarch := getGOARCH(rw)
	if goarch == "" {
		if goarch = elfGOARCH(elfFile); goarch == "" {
			return nil, ErrNotGo
		}
	}

	return &ELF{
		WriterAt: rw,

		goarch:  goarch,
		load:    loadProg,
		text:    text,
		symTab:  symTab,
		pcLnTab: pcLnTab,
	}, nil
}

func (elf *ELF) GOARCH() string { return elf.goarch }

func (elf *ELF) TextAddr() uint64 { return elf.text.Addr }

func (elf *ELF) GoSymTabData() io.Reader { return elf.symTab.Open() }

func (elf *ELF) GoPCLnTabData() io.Reader { return elf.pcLnTab.Open() }

func (elf *ELF) Offset(p *gosym.Func) int64 { return int64(p.Entry - elf.load.Vaddr) }

func elfGOARCH(f *elf.File) string {
	switch f.Machine {
	case elf.EM_386:
		return "386"
	case elf.EM_X86_64:
		return "amd64"
	case elf.EM_ARM:
		return "arm"
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_PPC64:
		if f.ByteOrder == binary.LittleEndian {
			return "ppc64le"
		}
		return "ppc64"
	case elf.EM_S390:
		return "s390x"
	}
	return ""
}
