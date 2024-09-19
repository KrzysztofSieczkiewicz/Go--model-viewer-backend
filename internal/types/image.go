package types

import "fmt"

// TODO: Add color space property to enable filtering out maps that do not match requred colorspace

type Image struct {
	imgType string
}

var (
	InvalidMap = Image{"Invalid"}
	// Standard material
	ColorMap            = Image{"Color"}
	RoughnessMap        = Image{"Roughness"}
	MetalnessMap        = Image{"Metalness"}
	EmissiveMap         = Image{"Emissive"}
	AmbientOcclusionMap = Image{"Ambient Occlusion"}
	NormalMap           = Image{"Normal"}
	BumpMap             = Image{"Bump"}
	DisplacementMap     = Image{"Displacement"}

	// Physical material (extends standard)
	SpecularIntensityMap = Image{"Specular Intensity"}
	SpecularColorMap     = Image{"Specular Color"}
	ClearcoatMap         = Image{"Clearcoat"}
	ClearcoatRoughness   = Image{"Clearcoat Roughness"}
	ClearcoatNormalMap   = Image{"Clearcoat Normal"}
	AnisotropyMap        = Image{"Anisotropy"}
	IridescenceMap       = Image{"Iridescence"}
	SheenColorMap        = Image{"Sheen Color"}
	SheenRoughnessMap    = Image{"Sheen Roughness"}
	ThicknessMap         = Image{"Thickness"}
	TransmissionMap      = Image{"Transmission"}
)

func (i Image) String() string {
	return i.imgType
}

func (i Image) FromString(s string) (Image, error) {
	switch s {

	case ColorMap.imgType:
		return ColorMap, nil
	case RoughnessMap.imgType:
		return RoughnessMap, nil
	case MetalnessMap.imgType:
		return MetalnessMap, nil
	case EmissiveMap.imgType:
		return EmissiveMap, nil
	case AmbientOcclusionMap.imgType:
		return AmbientOcclusionMap, nil
	case NormalMap.imgType:
		return NormalMap, nil
	case BumpMap.imgType:
		return BumpMap, nil
	case DisplacementMap.imgType:
		return DisplacementMap, nil

	case SpecularIntensityMap.imgType:
		return SpecularIntensityMap, nil
	case SpecularColorMap.imgType:
		return SpecularColorMap, nil
	case ClearcoatMap.imgType:
		return ClearcoatMap, nil
	case ClearcoatRoughness.imgType:
		return ClearcoatRoughness, nil
	case ClearcoatNormalMap.imgType:
		return ClearcoatNormalMap, nil
	case AnisotropyMap.imgType:
		return AnisotropyMap, nil
	case IridescenceMap.imgType:
		return IridescenceMap, nil
	case SheenColorMap.imgType:
		return SheenColorMap, nil
	case SheenRoughnessMap.imgType:
		return SheenRoughnessMap, nil
	case ThicknessMap.imgType:
		return ThicknessMap, nil
	case TransmissionMap.imgType:
		return TransmissionMap, nil
	}
	

	return InvalidMap, fmt.Errorf("map type does not exist")
}