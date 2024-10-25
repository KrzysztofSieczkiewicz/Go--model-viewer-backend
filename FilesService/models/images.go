package models

type Images []*Image

// TODO -> should be deleted

// Deconstruct slice of filenames into slice of Images
func (i *Images) DeconstructImageNames(filenames []string) error {
	// Iterate over each filename and deconstruct it
	for _, filename := range filenames {
		image := &Image{}
		err := image.DeconstructImageName(filename)
		if err != nil {
			return err
		}
		*i = append(*i, image)
	}

	return nil
}