package types

import "fmt"

// TODO: Add color space property to enable filtering out maps that do not match requred colorspace

type Image struct {
	imgType string
}

var (
	InvalidMap = Image{"Invalid"}
	// Standard material
	ColorMap            = Image{"color"}
	RoughnessMap        = Image{"roughness"}
	MetalnessMap        = Image{"metalness"}
	EmissiveMap         = Image{"emissive"}
	AmbientOcclusionMap = Image{"ambient occlusion"}
	NormalMap           = Image{"normal"}
	BumpMap             = Image{"bump"}
	DisplacementMap     = Image{"displacement"}

	// Physical material (extends standard)
	SpecularIntensityMap = Image{"specular intensity"}
	SpecularColorMap     = Image{"specular color"}
	ClearcoatMap         = Image{"clearcoat"}
	ClearcoatRoughness   = Image{"clearcoat roughness"}
	ClearcoatNormalMap   = Image{"clearcoat normal"}
	AnisotropyMap        = Image{"anisotropy"}
	IridescenceMap       = Image{"iridescence"}
	SheenColorMap        = Image{"sheen Color"}
	SheenRoughnessMap    = Image{"sheen roughness"}
	ThicknessMap         = Image{"thickness"}
	TransmissionMap      = Image{"transmission"}
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