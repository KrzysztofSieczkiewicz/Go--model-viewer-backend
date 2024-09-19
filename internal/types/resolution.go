package types

import "fmt"

type Resolution struct {
	res string
}

func (r Resolution) String() string {
	return r.res
}

func FromString(s string) (Resolution, error) {
	switch s {
	case Resolution512.res:
		return Resolution512, nil
	case Resolution1024.res:
		return Resolution1024, nil
	case Resolution2048.res:
		return Resolution2048, nil
	}

	return ResolutionError, fmt.Errorf("Resolution does not exist")
}

var (
	ResolutionError	= Resolution{"0x0"}
	Resolution512  = Resolution{"512x512"}
	Resolution1024 = Resolution{"1024x1024"}
	Resolution2048 = Resolution{"2048x2048"}
)
