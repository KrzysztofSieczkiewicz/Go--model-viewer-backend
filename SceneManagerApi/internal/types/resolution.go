package types

import (
	"encoding/json"
	"fmt"
)

type Resolution struct {
	res string
}

var (
	ResolutionError	= Resolution{"0x0"}
	Resolution512  = Resolution{"512x512"}
	Resolution1024 = Resolution{"1024x1024"}
	Resolution2048 = Resolution{"2048x2048"}
	Resolution4096 = Resolution{"4096x4096"}
)

func (r Resolution) String() string {
	return r.res
}

func (r Resolution) FromString(s string) (Resolution, error) {
	switch s {
	case Resolution512.res:
		return Resolution512, nil
	case Resolution1024.res:
		return Resolution1024, nil
	case Resolution2048.res:
		return Resolution2048, nil
	case Resolution4096.res:
		return Resolution4096, nil
	}

	return ResolutionError, fmt.Errorf("Resolution does not exist")
}

func (r Resolution) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.res)
}

func (r *Resolution) UnmarshalJSON(data []byte) error {
	var resString string
	if err := json.Unmarshal(data, &resString); err != nil {
		return err
	}

	resolution, err := r.FromString(resString)
	if err != nil {
		return err
	}

	*r = resolution
	return nil
}