package executable

import (
	"debug/gosym"
	"debug/macho"
	"fmt"
	"io"
)

const (
	machoText      = "__text"
	machoGoSymTab  = "__gosymtab"
	machoGoPCLnTab = "__gopclntab"
)

type MachO struct {
	io.WriterAt

	goarch                string
	lcSegment             *macho.Segment
	text, symTab, pcLnTab *macho.Section
}

func NewMachO(rw ReadWriterAt) (*MachO, error) {
	machoFile, err := macho.NewFile(rw)
	if err != nil {
		return nil, fmt.Errorf("macho open: %w", err)
	}

	lcSegment := machoFile.Segment("LC_SEGMENT")
	if lcSegment == nil {
		return nil, ErrNotGo
	}

	var text, symTab, pcLnTab *macho.Section
	for _, section := range machoFile.Sections {
		switch {
		case text == nil && section.Name == machoText:
			text = section
		case symTab == nil && section.Name == machoGoSymTab:
			symTab = section
		case pcLnTab == nil && section.Name == machoGoPCLnTab:
			pcLnTab = section
		}
	}

	if text == nil || symTab == nil || pcLnTab == nil {
		return nil, ErrNotGo
	}

	goarch := getGOARCH(rw)
	if goarch == "" {
		if goarch = machoGOARCH(machoFile); goarch == "" {
			return nil, ErrNotGo
		}
	}

	return &MachO{
		WriterAt: rw,

		goarch:    goarch,
		lcSegment: lcSegment,
		text:      text,
		symTab:    symTab,
		pcLnTab:   pcLnTab,
	}, nil
}

func (m *MachO) GOARCH() string { return m.goarch }

func (m *MachO) TextAddr() uint64 { return m.text.Addr }

func (m *MachO) GoSymTabData() io.Reader { return m.symTab.Open() }

func (m *MachO) GoPCLnTabData() io.Reader { return m.pcLnTab.Open() }

func (m *MachO) Offset(p *gosym.Func) int64 {
	return int64(p.Entry - m.lcSegment.Addr + m.lcSegment.Offset)
}

func machoGOARCH(m *macho.File) string {
	switch m.Cpu {
	case macho.Cpu386:
		return "386"
	case macho.CpuAmd64:
		return "amd64"
	case macho.CpuArm:
		return "arm"
	case macho.CpuArm64:
		return "arm64"
	case macho.CpuPpc64:
		return "ppc64"
	}
	return ""
}
