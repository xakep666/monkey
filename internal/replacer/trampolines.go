package replacer

import (
	"debug/gosym"
	"encoding/binary"
	"math"
)

const (
	immediate32bit = math.MaxUint32
	immediate26bit = immediate32bit >> 6
	immediate24bit = immediate26bit >> 2
	immediate20bit = immediate24bit >> 4
)

type x86 struct{}

func (x86) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target) > immediate32bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry - (source.Entry + 5))

	ret := make([]byte, 5)
	ret[0] = 0xe9 // jmp
	binary.LittleEndian.PutUint32(ret[1:], to)

	return ret, nil
}

type arm struct{}

func (arm) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target) > immediate24bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry-source.Entry-8) >> 2

	ret := make([]byte, 4)
	binary.LittleEndian.PutUint32(ret, to)
	ret[3] = 0xea // bal to

	return ret, nil
}

type armbe struct{}

func (armbe) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target) > immediate24bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry-source.Entry-8) >> 2

	ret := make([]byte, 4)
	binary.BigEndian.PutUint32(ret, to)
	ret[0] = 0xea // b to

	return ret, nil
}

type arm64 struct{}

func (arm64) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target)>>2 > immediate26bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry-source.Entry) >> 2

	ret := make([]byte, 4)
	to &= 0x3ffffff
	binary.LittleEndian.PutUint32(ret, to)
	ret[3] |= 0x14 // bal to

	return ret, nil
}

type arm64be struct{}

func (arm64be) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target)>>2 > immediate26bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry-source.Entry) >> 2

	ret := make([]byte, 4)
	to &= 0x3ffffff
	binary.LittleEndian.PutUint32(ret, to)
	ret[0] |= 0x14 // bal to

	return ret, nil
}

type mipsle struct{}

func (mipsle) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if target.Entry > immediate26bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry) >> 2

	ret := make([]byte, 4)
	to &= 0x3ffffff
	binary.LittleEndian.PutUint32(ret, to)
	ret[3] |= 0x8 // j to

	return ret, nil
}

type mips struct{}

func (mips) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if target.Entry > immediate26bit {
		return nil, ErrLongDistance
	}

	to := uint32(target.Entry) >> 2

	ret := make([]byte, 4)
	to &= 0x3ffffff
	binary.BigEndian.PutUint32(ret, to)
	ret[0] |= 0x8 // j to

	return ret, nil
}

type riscv struct{}

func (riscv) GenerateTrampoline(source, target *gosym.Func) ([]byte, error) {
	if distance(source, target)>>1 > immediate20bit {
		return nil, ErrLongDistance
	}

	diff := uint32(target.Entry - source.Entry)

	ret := make([]byte, 4)
	// immediate format is diff[20|10:1|11|19:12]
	instr := (diff>>20)<<31 | ((diff>>1)&0x3ff)<<21 | ((diff>>11)&0x1)<<20 | ((diff>>12)&0xff)<<12
	instr |= 0x6f // jal to (rd=0)
	binary.LittleEndian.PutUint32(ret, instr)

	return ret, nil
}

func distance(source, target *gosym.Func) uint64 {
	if target.Entry > source.Entry {
		return target.Entry - source.Entry
	}

	return source.Entry - target.Entry
}

func trampolineFromGOARCH(goarch string) (trampolineGenerator, error) {
	switch goarch {
	// x86
	case "amd64", "386":
		return x86{}, nil
	// arm
	case "arm":
		return arm{}, nil
	case "armbe":
		return armbe{}, nil
	case "arm64":
		return arm64{}, nil
	case "arm64be":
		return arm64be{}, nil
	// mips
	case "mipsle", "mips64le":
		return mipsle{}, nil
	case "mips", "mips64":
		return mips{}, nil
	// riscv
	case "riscv", "riscv64":
		return riscv{}, nil
	// TODO: other architectures
	default:
		return nil, ErrUnsupportedArchitecture
	}
}
